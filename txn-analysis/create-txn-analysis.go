package main 

import (
    "fmt"
    "os"
    "log"
    "io/ioutil"
    "stockRedis"
)


func main() {

    if (len(os.Args) < 2) {
        fmt.Println("Usage : go run create-db.go <transaction-csv-file-dir>")
    }

    f, err := os.Create("txn-analysis.csv")
    if err != nil {
        log.Fatal(err)
        return
    }

    _, err = f.WriteString("Date,Stock,Demat,Action,Quantity,Price,Value,Brokerage,STT,Transaction Charges,GST,Stamp Duty,Buy Date,Buy Cost,COA,P/L,Days,P/L Type\n")

    files, err := ioutil.ReadDir(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }

    for _, x := range files {
        fmt.Println(os.Args[1]+"/"+x.Name())
        stockRedis.PushAllTxnData(os.Args[1]+"/"+x.Name(), f)
        stockRedis.RecordSellInfo(f)
        stockRedis.CleanStockList()
    }

    f.Close()
}
