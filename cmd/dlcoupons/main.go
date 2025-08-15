// dlcoupons downloads coupon data from S3 and save them to data/coupon/ directory.
package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	urls := map[string]string{
		"couponbase1": "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"couponbase2": "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"couponbase3": "https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}
	var wg sync.WaitGroup
	for name, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := downloadCoupon(u, name); err != nil {
				log.Println("Failed to download", name, ":", err)
			}
		}()
	}
	wg.Wait()
}

func downloadCoupon(url, filename string) error {
	if err := os.MkdirAll("data/coupon", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	log.Println("Downloading", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading file: %d", resp.StatusCode)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzReader.Close()
	}()

	outputPath := filepath.Join("data", "coupon", filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	if _, err := io.Copy(outFile, gzReader); err != nil {
		return fmt.Errorf("failed to write uncompressed data: %w", err)
	}

	log.Println("Downlad finish", url, "to", outputPath)
	return nil
}
