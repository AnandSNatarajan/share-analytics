package main

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
)

// Contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func main() {

	if len(os.Args) < 4 {
		fmt.Println("Invalid arguments, usage: dump-analysis.go <stock-list-file> <last-date> <current-holdings>")
		os.Exit(1)
	}

	// Store current holdings
	var holdings []string
	holdingsFile, err := os.Open(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(holdingsFile)
	for scanner.Scan() {
		stock := scanner.Text()
		holdings = append(holdings, stock)
	}

	fmt.Println(holdings)

	// Retreive all stocks for processing
	readFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialDatabase(14))
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	f, err := os.Create("stock-analysis.csv")
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = f.WriteString("Stock,Price,HA-Today,HA-Pattern,HAC-T,HAC-1D,HAC-2D,HAC-3D,HA-Trend-Days,HA-Chg-10D,HA-PriceChg,EMA5-EMA13-S,EMA5-EMA13-D,EMA50-EMA200-S,EMA50-EMA200-D,RSI,10D-Trend,Holding,52-Week-High,52-Week-Low,1-D,1-W,2-W,1-M,3-M,6-M,1-Y\n")
	if err != nil {
		log.Fatal(err)
		f.Close()
		return
	}

	// Open stocks file and process each stock separately
	scanner = bufio.NewScanner(readFile)
	for scanner.Scan() {
		stock := scanner.Text()
		j, err := conn.Do("HGETALL", stock+":Analysis")
		if err != nil {
			log.Fatal(err)
		}
		price, err := redis.String(conn.Do("HGET", stock+":close", os.Args[2]))
		if err != nil {
			log.Fatal(err)
		}
		analysis := stock + "," + price + ","
		values, _ := redis.StringMap(j, err)
		analysis += values["HA-Today"] + ","
		analysis += values["HA-Pattern"] + ","
		analysis += values["HAC-T"] + ","
		analysis += values["HAC-1D"] + ","
		analysis += values["HAC-2D"] + ","
		analysis += values["HAC-3D"] + ","
		analysis += values["HA-Trend-Days"] + ","
		analysis += values["HA-Chg-10D"] + ","
		analysis += values["HA-PriceChg"] + ","
		analysis += values["EMA5-EMA13-Status"] + ","
		analysis += values["EMA5-EMA13-Days"] + ","
		analysis += values["EMA50-EMA200-Status"] + ","
		analysis += values["EMA50-EMA200-Days"] + ","
		analysis += values["RSI"] + ","
		analysis += values["10D-Trend"] + ","
		if contains(holdings, stock) {
			analysis += "Y,"
		} else {
			analysis += "N,"
		}
		analysis += values["52-Week-High"] + ","
		analysis += values["52-Week-Low"] + ","
		analysis += values["1-D"] + ","
		analysis += values["1-W"] + ","
		analysis += values["2-W"] + ","
		analysis += values["1-M"] + ","
		analysis += values["3-M"] + ","
		analysis += values["6-M"] + ","
		analysis += values["1-Y"] + "\n"
		fmt.Println(analysis)
		f.WriteString(analysis)
	}

	err = f.Close()
}
