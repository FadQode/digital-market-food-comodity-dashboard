package main

import (
    "fmt"
    "ingestion/scraper-go/internal/scraper"
    "ingestion/scraper-go/internal/storage"
)

func main() {

    data, err := scraper.ScrapeTokopedia("beras")
    if err != nil {
        panic(err)
    }

    err = storage.SaveJSON("../../data/raw/tokopedia.json", data)
    if err != nil {
        panic(err)
    }

    fmt.Println("Scraping completed")
}