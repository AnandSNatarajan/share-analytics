package main

import (
    "fmt"
    "os"
    "stockRedis"
)

func main() {

    if (len(os.Args) < 2) {
      fmt.Println("Invalid arguments, usage: main.go <stock-list-file>")
      os.Exit(1)
    }

    stockRedis.PushAllStockData(os.Args[1])
}
