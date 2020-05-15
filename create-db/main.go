package main

import (
    "fmt"
    "os"
    "github.com/shareanalytics/utils"
)

func main() {

    if (len(os.Args) < 2) {
      fmt.Println("Invalid arguments, usage: main.go <stock-list-file>")
      os.Exit(1)
    }

    utils.PushAllStockData(os.Args[1])
}
