package utils

import (
    "fmt"
    "log"
    "os"
    "bufio"
    "strings"
    "io/ioutil"
    "github.com/gomodule/redigo/redis"

)

var verbose bool = false

func WriteOHLC (conn redis.Conn, stock string, open string, high string, low string, sclose string, volume string, date string) {

    _, err := conn.Do("HMSET", stock+":open", date, open)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.Do("HMSET", stock+":high", date, high)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.Do("HMSET", stock+":low", date, low)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.Do("HMSET", stock+":close", date, sclose)
    if err != nil {
        log.Fatal(err)
    }
    _, err = conn.Do("HMSET", stock+":volume", date, volume)
    if err != nil {
        log.Fatal(err)
    }
}

func PushAllStockData (data_dir string) {

    files, err := ioutil.ReadDir(data_dir)
    if err != nil {
        log.Fatal(err)
    }

    conn, err := redis.Dial("tcp", "localhost:6379")
    if err != nil {
        log.Fatal(err)
    }

    defer conn.Close()

    for _, f := range files {
        fmt.Println(f.Name())
        readFile, err := os.Open(data_dir+"/"+f.Name())
        if (err != nil) {
          log.Fatal(err)
        }

        scanner := bufio.NewScanner(readFile)
        for (scanner.Scan()) {

            s := strings.Split(scanner.Text(), ",")
            if (len(s) <= 1) {
                fmt.Println("ERROR: Received blank line")
                continue
            }
            d := s[1]
            date := d[0:4]+"-"+d[4:6]+"-"+d[6:8]

            // Write OHLC data 
            WriteOHLC(conn, s[0], s[2], s[3], s[4], s[5], s[6], date)
        }
    }
}
