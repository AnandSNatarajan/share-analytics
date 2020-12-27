package utils

import (
	"fmt"
	//"time"
	"github.com/gomodule/redigo/redis"
	"github.com/markcheno/go-talib"
	"math"
)

const (
	GREEN = 1
	RED   = 2
)

func heikenAshiPrepareCandles(ha_open []float64, ha_high []float64, ha_low []float64, ha_close []float64, color []int, ctype []string) {

	for key, _ := range ha_close {

		if ha_close[key] > ha_open[key] {
			color[key] = GREEN
		} else {
			color[key] = RED
		}

		// Get current candle type, whether it is a doji, almost full candle or nuetral one.
		var candle_range float64
		if ha_close[key] > ha_open[key] {
			candle_range = ha_close[key] - ha_open[key]
		} else {
			candle_range = ha_open[key] - ha_close[key]
		}

		total_range := ha_high[key] - ha_low[key]
		candle_size := (candle_range * 100) / total_range

		if (math.Round(ha_open[key]) == math.Round(ha_high[key])) || (math.Round(ha_open[key]) == math.Round(ha_low[key])) {
			if color[key] == GREEN {
				ctype[key] = "Green-Full"
			} else {
				ctype[key] = "Red-Full"
			}
		} else if candle_size < 20 {
			if color[key] == GREEN {
				ctype[key] = "Green-Doji"
			} else {
				ctype[key] = "Red-Doji"
			}
		} else {
			if color[key] == GREEN {
				ctype[key] = "Green-Nuetral"
			} else {
				ctype[key] = "Red-Nuetral"
			}
		}
	}

	return
}

func heikenAshiAnalysis(conn redis.Conn, stock string, sclose []float64, ha_open []float64, ha_high []float64, ha_low []float64, ha_close []float64, dates []string) {

	var trend_changes = 0
	var last_trend_change = 0
	var today_trend_change = false

	var color = make([]int, len(ha_close))
	var ctype = make([]string, len(ha_close))

	// Current candle parameters are printed separately (whether it is bullish/bearish and type of candle.
	// The color of previous candle determines the current Trend. The color change that happens before that
	// determines the period of that trend.

	heikenAshiPrepareCandles(ha_open, ha_high, ha_low, ha_close, color, ctype)
	today := len(ha_close) - 1
	yest := len(ha_close) - 2

	if color[today] == GREEN && color[yest] == GREEN {
		redis.String(conn.Do("HSET", stock, "HA-Today", "Bullish"))
	} else if color[today] == GREEN && color[yest] == RED {
		redis.String(conn.Do("HSET", stock, "HA-Today", "Bullish-Turn"))
		today_trend_change = true
	} else if color[today] == RED && color[yest] == RED {
		redis.String(conn.Do("HSET", stock, "HA-Today", "Bearish"))
	} else if color[today] == RED && color[yest] == GREEN {
		redis.String(conn.Do("HSET", stock, "HA-Today", "Bearish-Turn"))
		today_trend_change = true
	}

	for i := today; i > today-11; i-- {
		if color[i] != color[i-1] {
			trend_changes++
		}
	}

	for i := yest; i > 0; i-- {
		if color[i] != color[i-1] {
			last_trend_change = i
			break
		}
	}

	// Calculate change in price during current/previous trend. This is current trend if there is no change in
	// trend from yesterday. If there is a change today, this is actually for the trend that got closed yesterday.
	if today_trend_change {
		chg_str := fmt.Sprintf("%.02f", ((sclose[today-1]-sclose[last_trend_change-1])/sclose[last_trend_change-1])*100)
		redis.String(conn.Do("HSET", stock, "HA-PriceChg", chg_str))
		redis.String(conn.Do("HSET", stock, "HA-Trend-Days", (yest)-last_trend_change))
	} else {
		chg_str := fmt.Sprintf("%.02f", ((sclose[today]-sclose[last_trend_change-1])/sclose[last_trend_change-1])*100)
		redis.String(conn.Do("HSET", stock, "HA-Trend-Days", (today)-last_trend_change))
		redis.String(conn.Do("HSET", stock, "HA-PriceChg", chg_str))
	}

	redis.String(conn.Do("HSET", stock, "HA-Chg-10D", trend_changes))
	redis.String(conn.Do("HSET", stock, "HAC-T", ctype[today]))
	redis.String(conn.Do("HSET", stock, "HAC-1D", ctype[yest]))
	redis.String(conn.Do("HSET", stock, "HAC-2D", ctype[today-2]))
	redis.String(conn.Do("HSET", stock, "HAC-3D", ctype[today-3]))

	var pattern string
	if color[today] == GREEN {
		pattern = "G-"
	} else {
		pattern = "R-"
	}
	if color[yest] == GREEN {
		pattern += "G-"
	} else {
		pattern += "R-"
	}
	if color[today-2] == GREEN {
		pattern += "G-"
	} else {
		pattern += "R-"
	}
	if color[today-3] == GREEN {
		pattern += "G"
	} else {
		pattern += "R"
	}
	redis.String(conn.Do("HSET", stock, "HA-Pattern", pattern))
}

func cdrString(ema3 []float64, ema8 []float64, index int) string {
	if ema3[index-1] > ema8[index-1] && ema3[index] < ema8[index] {
		return "RD"
	}

	if ema3[index-1] < ema8[index-1] && ema3[index] > ema8[index] {
		return "RU"
	}

	if ema3[index] > ema8[index] {
		if (ema3[index-1] - ema8[index-1]) > (ema3[index] - ema8[index]) {
			return "C"
		} else {
			return "D"
		}
	}

	if ema8[index] > ema3[index] {
		if (ema8[index-1] - ema3[index-1]) > (ema8[index] - ema3[index]) {
			return "C"
		} else {
			return "D"
		}
	}
	return "U"
}

func emaAnalysis(ema1 []float64, ema2 []float64) (string, int) {

	days := len(ema1) - 1

	var pattern = ""
	var chg int

	for i := 7; i >= 1; i-- {
		cc := cdrString(ema1, ema2, days-i)
		pattern += (cc + "-")
	}
	pattern += cdrString(ema1, ema2, days)

	if ema1[days] > ema2[days] {
		chg = 1
		for i := 1; i < days-1; i++ {
			if ema1[days-i] < ema2[days-i] {
				break
			}
			chg++
		}
	} else {
		chg = 1
		for i := 1; i < days-1; i++ {
			if ema1[days-i] > ema2[days-i] {
				break
			}
			chg++
		}
	}

	return pattern, chg
}

func ema3Status(ema3 []float64, ema8 []float64, ema13 []float64, ema20 []float64, ema50 []float64, today int) string {
	pattern := ""

	if ema3[today] > ema8[today] {
		pattern += "A-"
	} else {
		pattern += "B-"
	}

	if ema3[today] > ema13[today] {
		pattern += "A-"
	} else {
		pattern += "B-"
	}

	if ema3[today] > ema20[today] {
		pattern += "A-"
	} else {
		pattern += "B-"
	}

	if ema3[today] > ema50[today] {
		pattern += "A"
	} else {
		pattern += "B"
	}

	return pattern
}

func MyHeikinashiCandles(highs []float64, opens []float64, closes []float64, lows []float64) ([]float64, []float64, []float64, []float64) {
	N := len(highs)

	heikinHighs := make([]float64, N)
	heikinOpens := make([]float64, N)
	heikinCloses := make([]float64, N)
	heikinLows := make([]float64, N)

	for currentCandle := 1; currentCandle < N; currentCandle++ {
		previousCandle := currentCandle - 1

		heikinCloses[currentCandle] = (highs[currentCandle] + opens[currentCandle] + closes[currentCandle] + lows[currentCandle]) / 4
		heikinOpens[currentCandle] = (heikinOpens[previousCandle] + heikinCloses[previousCandle]) / 2
		heikinHighs[currentCandle] = math.Max(highs[currentCandle], math.Max(heikinOpens[currentCandle], heikinCloses[currentCandle]))
		heikinLows[currentCandle] = math.Min(lows[currentCandle], math.Min(heikinOpens[currentCandle], heikinCloses[currentCandle]))

		/* Buggy implementation in original is below
				heikinHighs[currentCandle] = math.Max(highs[currentCandle], math.Max(opens[currentCandle], closes[currentCandle]))
				heikinOpens[currentCandle] = (opens[previousCandle] + closes[previousCandle]) / 2
				heikinCloses[currentCandle] = (highs[currentCandle] + opens[currentCandle] + closes[currentCandle] + lows[currentCandle]) / 4
		        heikinLows[currentCandle] = math.Min(highs[currentCandle], math.Min(opens[currentCandle], closes[currentCandle]))
		*/
	}

	return heikinHighs, heikinOpens, heikinCloses, heikinLows
}

func sarPattern(shigh []float64, slow []float64, sclose []float64) string {
	pattern := ""

	sar := talib.Sar(shigh, slow, 0.02, 0.2)
	if sar[len(shigh)-4] > sclose[len(shigh)-4] {
		pattern += "D-"
	} else {
		pattern += "U-"
	}

	if sar[len(shigh)-3] > sclose[len(shigh)-3] {
		pattern += "D-"
	} else {
		pattern += "U-"
	}
	if sar[len(shigh)-2] > sclose[len(shigh)-2] {
		pattern += "D-"
	} else {
		pattern += "U-"
	}
	if sar[len(shigh)-1] > sclose[len(shigh)-1] {
		pattern += "D"
	} else {
		pattern += "U"
	}

	return pattern
}

func StockCommentary(conn redis.Conn, stock string, sopen []float64, shigh []float64, slow []float64, sclose []float64, dates []string) {

	ha_high, ha_open, ha_close, ha_low := MyHeikinashiCandles(shigh, sopen, sclose, slow)

	var ema50, ema20, ema13, ema8, ema3 []float64
	if len(ha_close)-1 > 3 {
		ema3 = talib.Ema(ha_close, 3)
	}
	if len(ha_close)-1 > 20 {
		ema20 = talib.Ema(ha_close, 20)
	}
	if len(ha_close)-1 > 13 {
		ema13 = talib.Ema(ha_close, 13)
	}
	if len(ha_close)-1 > 8 {
		ema8 = talib.Ema(ha_close, 8)
	}
	if len(ha_close)-1 > 50 {
		ema50 = talib.Ema(ha_close, 50)
	}

	color := GREEN
	if len(ha_close)-1 > 8 {
		if ha_close[len(ha_close)-1] > ha_open[len(ha_close)-1] {
			color = GREEN
		} else {
			color = RED
		}

		rsi8 := talib.Rsi(ha_close, 8)
		today := len(rsi8) - 1
		if (rsi8[today] > rsi8[today-1]) && (color != GREEN) {
			conn.Do("HSET", stock+":"+"Analysis", "RSI-Pattern", "Abnormal")
		} else if (rsi8[today] < rsi8[today-1]) && (color != RED) {
			conn.Do("HSET", stock+":"+"Analysis", "RSI-Pattern", "Abnormal")
		} else {
			conn.Do("HSET", stock+":"+"Analysis", "RSI-Pattern", "Normal")
		}
	}

	if len(ha_close) > 8 {
		pattern, _ := emaAnalysis(ema3, ema8)
		conn.Do("HSET", stock+":"+"Analysis", "ST-Pattern", pattern)
	}

	if ema3[len(ema3)-1] > ema50[len(ema3)-1] {
		conn.Do("HSET", stock+":"+"Analysis", "LT-Trend", "Uptrend")
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "LT-Trend", "Downtrend")
	}

	positive := 0
	for i := 1; i <= 13; i++ {
		if ha_close[len(ha_close)-1] > ha_close[len(ha_close)-i-1] {
			positive++
		}
	}

	conn.Do("HSET", stock+":"+"Analysis", "Sar", sarPattern(shigh, slow, sclose))
	conn.Do("HSET", stock+":"+"Analysis", "Ema3Status", ema3Status(ema3, ema8, ema13, ema20, ema50, len(ema3)-1))
	conn.Do("HSET", stock+":"+"Analysis", "Ema3Status-1", ema3Status(ema3, ema8, ema13, ema20, ema50, len(ema3)-2))
	conn.Do("HSET", stock+":"+"Analysis", "Ema3Status-2", ema3Status(ema3, ema8, ema13, ema20, ema50, len(ema3)-3))
	conn.Do("HSET", stock+":"+"Analysis", "Ema3Status-3", ema3Status(ema3, ema8, ema13, ema20, ema50, len(ema3)-4))
	conn.Do("HSET", stock+":"+"Analysis", "NewHigh", positive)

	heikenAshiAnalysis(conn, stock+":"+"Analysis", sclose, ha_open, ha_high, ha_low, ha_close, dates)
	//priceActionAnalysis(conn, stock+":"+"Analysis", sopen, shigh, slow, sclose)
}
