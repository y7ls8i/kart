// Package server contains the server logic.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/y7ls8i/kart/config"
	"github.com/y7ls8i/kart/server/order"
	"github.com/y7ls8i/kart/server/product"
)

// DB is the interface for the database layer.
type DB interface {
	product.DB
}

// Business is the interface for the business layer.
type Business interface {
	order.Business
}

// Server is the server struct.
type Server struct {
	config config.Server
	router *gin.Engine
	db     DB
	buss   Business
}

// NewServer creates a new server.
func NewServer(config config.Server, db DB, buss Business) *Server {
	server := &Server{
		config: config,
		db:     db,
		buss:   buss,
	}

	switch server.config.Mode {
	case gin.DebugMode, gin.TestMode:
		gin.SetMode(server.config.Mode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	server.router = gin.Default()

	server.router.Use(AuthMiddleware())

	// setup routes
	productHandler := product.NewProduct(server.db)
	orderHandler := order.NewOrder(server.buss)
	server.router.GET("/api/product", productHandler.List)
	server.router.GET("/api/product/:id", productHandler.Get)
	server.router.POST("/api/order", orderHandler.Create)

	return server
}

// Start starts listening for HTTP requests
func (s *Server) Start(ctx context.Context) {
	srv := &http.Server{
		Addr:    s.config.Listen,
		Handler: s.router.Handler(),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // TLS 1.2 or higher
		},
	}

	// start server
	go func() {
		if len(s.config.Certfile) > 0 && len(s.config.Keyfile) > 0 {
			if err := srv.ListenAndServeTLS(s.config.Certfile, s.config.Keyfile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Error starting HTTPS server", "error", err, "listen", s.config.Listen)
				os.Exit(1)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Error starting HTTP server", "error", err, "listen", s.config.Listen)
				os.Exit(1)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-quit:
	}
	slog.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server Shutdown:", "error", err)
	}
}

// AuthMiddleware is a middleware that checks if the request has the correct API key.
// The key is hardcoded for this exercise.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Api_key")

		if authHeader != "apitest" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
