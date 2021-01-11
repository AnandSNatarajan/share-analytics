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
	_, err = f.WriteString("Stock,Price,Volume-DMA,HA-Today,HA-Pattern,HAC-T,HAC-1D,HAC-2D,HAC-3D,HA-Trend-Days,RSI-50-Cross,RSI,RSI-Trend,EMA-Pattern,LEMA>LEMA_AVG,LEMA>HEMA,LEMA-LEMA-AVG-Diff,LEMA-HEMA-Diff,SAR-Pattern,LEMA-T1,Pullback-T1,Pullback-N,ADX-Trend,ADX-T1,ADX,RSI-T1,Strength,Corona-Change,Holding\n")

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
		analysis += values["LEMA>LEMA-AVG"] + ","
		analysis += values["LEMA>HEMA"] + ","
		analysis += values["LEMA-LEMA-AVG-Diff"] + ","
		analysis += values["LEMA-HEMA-Diff"] + ","

		/* Other information */
		analysis += values["Sar-Pattern"] + ","
		analysis += values["LEMA-T1"] + ","
		analysis += values["Pullback-T1"] + ","
		analysis += values["Pullback-N"] + ","
		analysis += values["ADX-Trend"] + ","
		analysis += values["ADX-T1"] + ","
		analysis += values["ADX"] + ","
		analysis += values["RSI-T1"] + ","
		analysis += values["Strength"] + ","
		analysis += values["Corona-Change"] + ","
		if contains(holdings, stock) {
			analysis += "Y,"
		} else {
			analysis += "N,"
		}
		analysis += "\n"
		fmt.Println(analysis)
		skip_check := false
		if strings.Contains(s[l-1], "Holdings") || strings.Contains(s[l-1], "FnO") {
			skip_check = true
		}
		if !skip_check && strings.Contains(values["LEMA>HEMA"], "N") {
			continue
		}
		if !skip_check && strings.Contains(values["LEMA>LEMA-AVG"], "N") {
			continue
		}
		if !skip_check && strings.Contains(values["RSI-Trend"], "N") {
			continue
		}
		if !skip_check && strings.Contains(values["ADX-Trend"], "N") {
			continue
		}
		f.WriteString(analysis)
	}

	err = f.Close()
}
