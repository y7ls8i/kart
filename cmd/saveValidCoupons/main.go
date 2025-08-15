// saveValidCoupons saves valid coupons in data/coupon/valid/valid to MongoDB.
package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"github.com/y7ls8i/kart/adapter/mongo"
)

func main() {
	client, err := mongo.NewClient("mongodb://127.0.0.1:27017", "kart")
	if err != nil {
		log.Fatalf("Error connecting to mongo: %v", err)
	}

	if err := client.EnsureIndexes(); err != nil {
		log.Fatalf("Error ensuring indexes: %v", err)
	}

	file, err := os.Open("data/coupon/valid/valid")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Read all lines from the file because we know the file is small.
	var coupons []mongo.Coupon
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		code := scanner.Text()
		coupons = append(coupons, mongo.Coupon{Code: code})
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	if err := client.InsertCoupons(context.Background(), coupons); err != nil {
		log.Fatalf("Error inserting coupons: %v", err)
	}

	log.Printf("%d coupons inserted\n", len(coupons))
}
