package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

func main() {
	log.Println("> start")

	if len(os.Args[1:]) < 1 {
		log.Fatalf("expecting 1 argument got %d\n", len(os.Args[1:]))
	}

	filename := os.Args[1]
	backtestResult, err := loadBacktestResultFromFilename(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("> loaded backtest result\n")

	backtestResult.Print()
}

func loadBacktestResultFromFilename(filename string) (*BacktestResult, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	log.Printf("> opened %s\n", filename)

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	log.Printf("> read %d bytes from %s\n", len(data), filename)

	err = f.Close()
	if err != nil {
		log.Printf("> WARNING: failed to close file %s: %v\n", filename, err)
	}

	var backtestResult *BacktestResult
	err = json.Unmarshal(data, &backtestResult)
	if err != nil {
		return nil, err
	}

	return backtestResult, nil
}
