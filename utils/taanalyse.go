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

// Ema - Exponential Moving Average
func ema(inReal []float64, inTimePeriod int, k1 float64) []float64 {

	outReal := make([]float64, len(inReal))

	lookbackTotal := inTimePeriod - 1
	startIdx := lookbackTotal
	today := startIdx - lookbackTotal
	i := inTimePeriod
	tempReal := 0.0
	for i > 0 {
		tempReal += inReal[today]
		today++
		i--
	}

	prevMA := tempReal / float64(inTimePeriod)

	for today <= startIdx {
		//prevMA = ((inReal[today] - prevMA) * k1) + prevMA
		prevMA = ((inReal[today]-prevMA)*k1 + prevMA*(1-k1))
		today++
	}
	outReal[startIdx] = prevMA
	outIdx := startIdx + 1
	for today < len(inReal) {
		prevMA = ((inReal[today]-prevMA)*k1 + prevMA*(1-k1))
		//prevMA = ((inReal[today] - prevMA) * k1) + prevMA
		outReal[outIdx] = prevMA
		today++
		outIdx++
	}

	return outReal
}

// HtTrendline - Hilbert Transform - Instantaneous Trendline (lookback=63)
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

	var color = make([]int, len(ha_close))
	var ctype = make([]string, len(ha_close))

	// Current candle parameters are printed separately (whether it is bullish/bearish and type of candle.
	// The color of previous candle determines the current Trend. The color change that happens before that
	// determines the period of that trend.

	heikenAshiPrepareCandles(ha_open, ha_high, ha_low, ha_close, color, ctype)

	today := len(ha_close) - 1
	redis.String(conn.Do("HSET", stock, "HAC-T", ctype[today]))
	redis.String(conn.Do("HSET", stock, "HAC-1D", ctype[today-1]))
	redis.String(conn.Do("HSET", stock, "HAC-2D", ctype[today-2]))
	redis.String(conn.Do("HSET", stock, "HAC-3D", ctype[today-3]))
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

func sarPattern(shigh []float64, slow []float64, sclose []float64) []string {
	pattern := make([]string, len(shigh))

	sar := talib.Sar(shigh, slow, 0.02, 0.2)
	for i := 0; i <= len(shigh)-1; i++ {
		if sar[i] > sclose[i] {
			pattern[i] = "Downtrend"
		} else {
			pattern[i] = "Uptrend"
		}
	}

	return pattern
}

func rsiAnalysis(conn redis.Conn, stock string, rsi []float64, slowrsi []float64, fastrsi []float64) {
	rsi_str := fmt.Sprintf("%0.2f", rsi[len(rsi)-1])
	conn.Do("HSET", stock+":"+"Analysis", "RSI", rsi_str)

	var i int
	if slowrsi[len(rsi)-1] > 50 {
		for i = 2; i < len(slowrsi)-1; i++ {
			if slowrsi[len(slowrsi)-i] < 50 {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "RSI-50-Cross", i-1)
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "RSI-50-Cross", 0)
	}
}

func emaTrend(ema []float64, ema_avg []float64) int {

	trend_days := 1

	if ema[len(ema)-1] > ema_avg[len(ema_avg)-1] {
		for i := 2; i < len(ema)-2; i++ {
			if ema[len(ema)-i] > ema_avg[len(ema_avg)-i] {
				trend_days++
			} else {
				break
			}
		}
	} else {
		for i := 2; i < len(ema)-2; i++ {
			if ema[len(ema)-i] < ema_avg[len(ema_avg)-i] {
				trend_days++
			} else {
				break
			}
		}
	}

	return trend_days
}

func mamaAnalysis(conn redis.Conn, stock string, sclose []float64) {
	mama, fama := talib.Mama(sclose, 0.5, 0.05)

	today := len(mama) - 1
	if fama[today] > mama[today] {
		conn.Do("HSET", stock+":"+"Analysis", "Mama-Trend", "Uptrend")
		var k int
		for k = k; k < 99; k++ {
			if mama[today-k] < fama[today-k] {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "Mama-T1", k)
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "Mama-Trend", "Downtrend")
		var k int
		for k = 2; k < 99; k++ {
			if mama[today-k] > fama[today-k] {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "Mama-T1", k)
	}
}

func rsAnalysis(conn redis.Conn, stock string, sclose []float64, idxclose []float64) {
	rsmap := make([]float64, 0)

	for i := 0; i < len(sclose); i++ {
		if idxclose[i] == 0 {
			continue
		}
		rsmap = append(rsmap, sclose[i]/idxclose[i])
	}

	rsmap_ema8 := talib.Ema(rsmap, 8)
	rsmap_ema8_ema3 := talib.Ema(rsmap_ema8, 3)

	var rsstr string
	var rslen int
	t1 := len(rsmap_ema8_ema3) - 1
	if rsmap_ema8_ema3[t1-1] < rsmap_ema8[t1-1] &&
		rsmap_ema8_ema3[t1] > rsmap_ema8[t1] {
		rsstr = "Crossover"
		rslen = t1
	} else if rsmap_ema8_ema3[t1] > rsmap_ema8[t1] {
		rsstr = "Uptrend"
		for rslen = t1; rslen > 0; rslen-- {
			if rsmap_ema8_ema3[rslen] < rsmap_ema8[rslen] {
				break
			}
		}
	} else {
		rsstr = "Downtrend"
		for rslen = t1; rslen > 0; rslen-- {
			if rsmap_ema8_ema3[rslen] > rsmap_ema8[rslen] {
				break
			}
		}
	}
	conn.Do("HSET", stock+":"+"Analysis", "RS-Trend", rsstr)
	conn.Do("HSET", stock+":"+"Analysis", "RS-T1", t1-rslen)
}

func macdAnalysis(conn redis.Conn, stock string, sclose []float64, dates []string) {

	m, s, _ := talib.MacdExt(sclose, 8, talib.EMA, 26, talib.EMA, 9, talib.EMA)
	today := len(m) - 1

	prev_change := 0

	if m[today] > s[today] {
		for k := 0; k <= today-1; k++ {
			if s[today] > m[today] {
				break
			}
			prev_change = k
		}
		conn.Do("HSET", stock+":"+"Analysis", "MACD-Trend", "Uptrend")
	} else {
		for k := 0; k <= today-1; k++ {
			if s[today] < m[today] {
				break
			}
			prev_change = k
		}
		conn.Do("HSET", stock+":"+"Analysis", "MACD-Trend", "Downtrend")
	}
	conn.Do("HSET", stock+":"+"Analysis", "MACD-T1", prev_change)
}

func StockCommentary(conn redis.Conn, stock string, sopen []float64, shigh []float64, slow []float64, sclose []float64, idxclose []float64, volume []float64, dates []string) {

	ha_high, ha_open, ha_close, ha_low := MyHeikinashiCandles(shigh, sopen, sclose, slow)

	//rsi := talib.Rsi(ha_close, 14)
	//slowrsi := talib.Wma(rsi, 21)
	//fastrsi := talib.Sma(rsi, 3)

	//emaAnalysis(conn, stock, sclose)
	//rsiAnalysis(conn, stock, rsi, slowrsi, fastrsi)
	//rsAnalysis(conn, stock, sclose, idxclose)
	//macdAnalysis(conn, stock, sclose, dates)
	mamaAnalysis(conn, stock, sclose)

	heikenAshiAnalysis(conn, stock+":"+"Analysis", sclose, ha_open, ha_high, ha_low, ha_close, dates)
}
