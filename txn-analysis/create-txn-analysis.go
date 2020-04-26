package main 

import (
    "fmt"
    "os"
    "log"
    "stockRedis"
)


func main() {

    if (len(os.Args) < 2) {
        fmt.Println("Usage : go run create-db.go <transaction-csv-file>")
    }

    f, err := os.Create("txn-analysis.csv")
    if err != nil {
        log.Fatal(err)
        return
    }

    _, err = f.WriteString("Date,Stock,Demat,Action,Quantity,Price,Value,Brokerage,STT,Transaction Charges,GST,Stamp Duty,Buy Date,Buy Cost,COA,P/L,Days,P/L Type\n")
    stockRedis.PushAllTxnData(os.Args[1], f)

    if err != nil {
        log.Fatal(err)
        f.Close()
        return
    }


    stockRedis.RecordSellInfo(f)
}
