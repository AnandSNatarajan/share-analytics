package main

import (
    "fmt"
    "os"
    "log"
    "bufio"
    "github.com/shareanalytics/utils"
)

func main() {

    var dates []string
    var err error

    if (len(os.Args) < 3) {
      fmt.Println("Invalid arguments, usage: main.go <stock-list-file>")
      os.Exit(1)
    }

    // Retreive all valid dates for processing 
    readFile, err := os.Open(os.Args[2])
    if (err != nil) {
        log.Fatal(err)
    }

    scanner := bufio.NewScanner(readFile)
    for (scanner.Scan()) {
        date := scanner.Text()
        dates = append(dates, date)
    }

    // Retreive all valid dates for processing 
    readFile2, err2 := os.Open(os.Args[1])
    if (err2 != nil) {
        log.Fatal(err2)
    }

    // Open stocks file and process each stock separately
    scanner = bufio.NewScanner(readFile2)
    for (scanner.Scan()) {
        stock := scanner.Text()
        log.Print("Processing ", stock)
        utils.PullAllStockData(stock, dates)
    }
}

