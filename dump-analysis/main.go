package main

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
	"strings"
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

	s := strings.Split(os.Args[1], "/")
	l := len(s)
	f, err := os.Create(s[l-1] + ".csv")
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = f.WriteString("Stock,Price,Volume-DMA,HA-Today,HA-Pattern,HAC-T,HAC-1D,HAC-2D,HAC-3D,HA-Trend-Days,RSI-50-Cross,RSI,RSI-Trend,EMA-Pattern,EMA8>EMA88,EMA8>EMA50,EMA8-88-Diff,EMA-8-50-Diff,NewHigh,SAR-Pattern,EMA88-T1,EMA88-T2,EMA88-T3,RSI-T1,Corona-Change,Holding")

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
		/* Heiken Ashi Data */
		analysis += values["Volume-DMA"] + ","
		analysis += values["HA-Today"] + ","
		analysis += values["HA-Pattern"] + ","
		analysis += values["HAC-T"] + ","
		analysis += values["HAC-1D"] + ","
		analysis += values["HAC-2D"] + ","
		analysis += values["HAC-3D"] + ","
		analysis += values["HA-Trend-Days"] + ","

		/* RSI Data */
		analysis += values["RSI-50-Cross"] + ","
		analysis += values["RSI"] + ","
		analysis += values["RSI-Trend"] + ","

		/* EMA8 Data */
		analysis += values["EMA8-Pattern"] + ","
		analysis += values["EMA8>EMA88"] + ","
		analysis += values["EMA8>EMA50"] + ","
		analysis += values["EMA-8-88-Diff"] + ","
		analysis += values["EMA-8-50-Diff"] + ","

		/* Long term trend data */
		analysis += values["NewHigh"] + ","

		/* Other information */
		analysis += values["Sar-Pattern"] + ","
		analysis += values["EMA88-T1"] + ","
		analysis += values["EMA88-T2"] + ","
		analysis += values["EMA88-T3"] + ","
		analysis += values["RSI-T1"] + ","
		analysis += values["Corona-Change"] + ","
		if contains(holdings, stock) {
			analysis += "Y,"
		} else {
			analysis += "N,"
		}
		analysis += "\n"
		fmt.Println(analysis)
		if !contains(holdings, stock) && strings.Contains(values["EMA-8-50-Diff"], "-") {
			continue
		}
		if !contains(holdings, stock) && (values["RSI-Trend"] == "N") {
			continue
		}
		f.WriteString(analysis)
	}

	err = f.Close()
}
