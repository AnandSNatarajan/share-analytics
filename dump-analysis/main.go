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
	//_, err = f.WriteString("Stock,Price,HA-Today,HA-Pattern,HAC-T,HAC-1D,HAC-2D,HAC-3D,HA-Trend-Days,RSI-Trend,RSI,RSI-50-Cross,RSI-T1,ADX-Trend,ADX-T1,ADX,Slow-EMA-Trend,Slow-EMA-T1,Fast-EMA-Trend,Fast-EMA-T1,50-EMA-Diff,200-EMA-Diff,Holding\n")
	_, err = f.WriteString("Stock,Price,HA-3D,HA-2D,HA-1D,HA-T,Mama-Trend,Mama-T1,Holding\n")

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
		//analysis += values["HA-Pattern"] + ","
		analysis += values["HAC-3D"] + ","
		analysis += values["HAC-2D"] + ","
		analysis += values["HAC-1D"] + ","
		analysis += values["HAC-T"] + ","
		//analysis += values["HA-Trend-Days"] + ","

		analysis += values["Mama-Trend"] + ","
		analysis += values["Mama-T1"] + ","

		/* Wclose Data */
		//analysis += values["KST-Trend"] + ","
		//analysis += values["KST-T1"] + ","
		//analysis += values["KST-Signal-Trend"] + ","
		//analysis += values["KST-Signal-T1"] + ","

		/* RS Data */
		//analysis += values["RS-Trend"] + ","
		//analysis += values["RS-T1"] + ","

		/* RSI Data */
		//analysis += values["RSI-Trend"] + ","
		//analysis += values["RSI"] + ","
		//analysis += values["RSI-50-Cross"] + ","
		//analysis += values["RSI-T1"] + ","

		/* Adx Data */
		//analysis += values["ADX-Trend"] + ","
		//analysis += values["ADX-T1"] + ","
		//analysis += values["ADX"] + ","

		/* EMA8 Data */
		//analysis += values["Slow-EMA-Trend"] + ","
		//analysis += values["Slow-EMA-T1"] + ","
		//analysis += values["Fast-EMA-Trend"] + ","
		//analysis += values["Fast-EMA-T1"] + ","
		//analysis += values["50-EMA-Diff"] + ","
		//analysis += values["200-EMA-Diff"] + ","

		/* MACD Data */
		//analysis += values["MACD-Trend"] + ","
		//analysis += values["MACD-T1"] + ","

		/* Other information */
		if contains(holdings, stock) {
			analysis += "Y,"
		} else {
			analysis += "N,"
		}
		analysis += "\n"
		fmt.Println(analysis)
		f.WriteString(analysis)
	}

	err = f.Close()
}
