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
	_, err = f.WriteString("Stock,Price,HA-Today,HA-Pattern,HAC-T,HAC-1D,HAC-2D,HAC-3D,HA-Trend-Days,ST-Pattern,Zone-Status,Zone-Length,Zone-Sequence,NewHigh,RSI-Pattern,Holding\n")
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
		analysis += values["ST-Pattern"] + ","
		analysis += values["Zone-Status"] + ","
		analysis += values["Zone-Length"] + ","
		analysis += values["Zone-Sequence"] + ","
		analysis += values["NewHigh"] + ","
		analysis += values["RSI-Pattern"] + ","
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
