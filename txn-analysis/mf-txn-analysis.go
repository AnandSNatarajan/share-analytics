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
        fmt.Println("Usage : go run create-mf-txn-analysis.go <transaction-csv-file-dir>")
    }

    f, err := os.Create("mf-txn-analysis.csv")
    if err != nil {
        log.Fatal(err)
        return
    }

    _, err = f.WriteString("Date,MF,Folio,Category,Demat,Action,Quantity,Price,Value,Buy Date,Buy Cost,P/L,Days,P/L Type\n")

    files, err := ioutil.ReadDir(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }

    for _, x := range files {
        fmt.Println(os.Args[1]+"/"+x.Name())
        stockRedis.PushAllMFTxnData(os.Args[1]+"/"+x.Name(), f)
        stockRedis.MFRecordSellInfo(f)
        stockRedis.CleanMFList()
    }

    f.Close()
}
