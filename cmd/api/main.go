// api is a simple API server for kart.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/y7ls8i/kart/adapter/mongo"
	"github.com/y7ls8i/kart/business"
	"github.com/y7ls8i/kart/config"
	"github.com/y7ls8i/kart/server"
)

var configPath = kingpin.Flag("config", "Path to config file.").Short('c').ExistingFile()

func main() {
	kingpin.Parse()

	conf := config.ReadConfig(*configPath)

	client, err := mongo.NewClient(conf.MongoDB.URI, conf.MongoDB.DB)
	if err != nil {
		slog.Error("Error connecting to mongo", "error", err)
		os.Exit(1)
	}

	buss := business.NewBusiness(client)

	s := server.NewServer(conf.Server, client, buss)
	s.Start(context.Background())
}
