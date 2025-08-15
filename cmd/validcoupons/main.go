// validcoupons find valid coupons from a list of coupons in data/coupon/ directory.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	File1Bit = 1 << 0 // 001
	File2Bit = 1 << 1 // 010
	File3Bit = 1 << 2 // 100
)

func processFile(path string, bit byte, lineMap map[string]byte) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineLength := len(line)
		if 8 <= lineLength && lineLength <= 10 {
			lineMap[line] |= bit
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning %s: %w", path, err)
	}
	return nil
}

func main() {
	lineMap := make(map[string]byte, 100_000_000) // Estimate initial size

	{
		log.Println("couponbase3 start")
		err := processFile("data/coupon/couponbase3", File3Bit, lineMap)
		if err != nil {
			log.Fatalf("couponbase2 failed: %v", err)
		}
		log.Println("couponbase3 done")
	}

	{
		log.Println("couponbase2 start")
		err := processFile("data/coupon/couponbase2", File2Bit, lineMap)
		if err != nil {
			log.Fatalf("couponbase2 failed: %v", err)
		}
		log.Println("couponbase2 done")
	}

	{
		log.Println("couponbase1 start")
		err := processFile("data/coupon/couponbase1", File1Bit, lineMap)
		if err != nil {
			log.Fatalf("couponbase1 failed: %v", err)
		}
		log.Println("couponbase1 done")
	}

	// Output lines that appear in at least 2 files
	for line, b := range lineMap {
		if b == 3 || b == 5 || b == 6 || b == 7 {
			fmt.Println(line)
		}
	}
}
