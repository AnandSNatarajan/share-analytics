package utils

import (
	//"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"strconv"
)

func PullAllStockData(stock string, dates []string) ([]float64, []float64, []float64, []float64, []float64) {

	var sopen, shigh, slow, sclose, volume []float64

	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialDatabase(14))
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for _, date := range dates {
		var price string
		var err error
		var f float64
		var i int64

		price, err = redis.String(conn.Do("HGET", stock+":close", date))
		if err != nil {
			log.Print("Unable to retrieve close for ", stock, date)
		}
		f, err = strconv.ParseFloat(price, 64)
		sclose = append(sclose, f)

		price, err = redis.String(conn.Do("HGET", stock+":open", date))
		if err != nil {
			log.Print("Unable to retrieve open for ", stock, date)
		}
		f, err = strconv.ParseFloat(price, 64)
		sopen = append(sopen, f)

		price, err = redis.String(conn.Do("HGET", stock+":low", date))
		if err != nil {
			log.Print("Unable to retrieve low for ", stock, date)
		}
		f, err = strconv.ParseFloat(price, 64)
		slow = append(slow, f)

		price, err = redis.String(conn.Do("HGET", stock+":high", date))
		if err != nil {
			log.Print("Unable to retrieve high for ", stock, date)
		}
		f, err = strconv.ParseFloat(price, 64)
		shigh = append(shigh, f)

		price, err = redis.String(conn.Do("HGET", stock+":volume", date))
		if err != nil {
			log.Print("Unable to retrieve volume for ", stock, date)
			i = 0
		} else {
			i, err = strconv.ParseInt(price, 10, 64)
		}
		volume = append(volume, float64(i))
	}

	StockCommentary(conn, stock, sopen, shigh, slow, sclose, volume, dates)

	return sopen, shigh, slow, sclose, volume
}
