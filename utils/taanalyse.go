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

func cdrString(ema1 []float64, ema2 []float64, index int) string {
	if ema1[index-1] > ema2[index-1] && ema1[index] < ema2[index] {
		return "RD"
	}

	if ema1[index-1] < ema2[index-1] && ema1[index] > ema2[index] {
		return "RU"
	}

	if ema1[index] > ema2[index] {
		if (ema1[index-1] - ema2[index-1]) > (ema1[index] - ema2[index]) {
			return "C"
		} else {
			return "D"
		}
	}

	if ema2[index] > ema1[index] {
		if (ema2[index-1] - ema1[index-1]) > (ema2[index] - ema1[index]) {
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

	for i := 2; i >= 1; i-- {
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

func rsiAnalysis(conn redis.Conn, stock string, rsi9 []float64, wrsi9 []float64, mwrsi9 []float64) {
	rsi_str := fmt.Sprintf("%0.2f", rsi9[len(rsi9)-1])
	conn.Do("HSET", stock+":"+"Analysis", "RSI", rsi_str)
	wtoday := len(wrsi9) - 1
	mwtoday := len(mwrsi9) - 1

	if mwrsi9[wtoday] > wrsi9[mwtoday] {
		conn.Do("HSET", stock+":"+"Analysis", "RSI-Trend", "Y")
		var i int
		for i = 1; i <= len(wrsi9)-1; i++ {
			if mwrsi9[wtoday-i] < wrsi9[mwtoday-i] {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "RSI-T1", i)
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "RSI-Trend", "N")
		var i int
		for i = 1; i <= len(wrsi9)-1; i++ {
			if mwrsi9[wtoday-i] < mwrsi9[mwtoday-i] {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "RSI-T1", i)
	}

	var i int
	if rsi9[len(rsi9)-1] > 50 {
		for i = 2; i < len(rsi9)-1; i++ {
			if rsi9[len(rsi9)-i] < 50 {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "RSI-50-Cross", i-1)
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "RSI-50-Cross", 0)
	}

}

func StockCommentary(conn redis.Conn, stock string, sopen []float64, shigh []float64, slow []float64, sclose []float64, volume []float64, dates []string) {

	ha_high, ha_open, ha_close, ha_low := MyHeikinashiCandles(shigh, sopen, sclose, slow)

	var ema50, ema88, ema8 []float64
	if len(sclose)-1 > 8 {
		ema8 = talib.Ema(ha_close, 8)
	}
	if len(sclose)-1 > 50 {
		ema50 = talib.Ema(ha_close, 50)
	}
	if len(ema8)-1 > 8 {
		ema88 = talib.Ema(ha_close, 3)
	}

	if len(sclose)-1 > 50 {
		vol_dma := talib.Sma(volume, 50)
		vol_dma_str := fmt.Sprintf("%0.0f", vol_dma[len(vol_dma)-1])
		conn.Do("HSET", stock+":"+"Analysis", "Volume-DMA", vol_dma_str)
	}

	rsi9 := talib.Rsi(sclose, 9)
	wrsi9 := talib.Wma(rsi9, 21)
	mwrsi9 := talib.Sma(rsi9, 3)

	sar_pattern := sarPattern(shigh, slow, sclose)

	positive := 0
	for i := 1; i <= len(ha_close)-2; i++ {
		if ha_close[len(ha_close)-1] > ha_close[len(ha_close)-i-1] {
			positive++
		} else {
			break
		}
	}

	corona_high := 0.0
	for i := 120; i <= 240; i++ {
		if sclose[len(sclose)-i] > corona_high {
			corona_high = sclose[len(sclose)-i]
		}
	}
	corona_chg := ((sclose[len(sclose)-1] - corona_high) / corona_high) * 100
	corona_chg_str := fmt.Sprintf("%0.2f", corona_chg)
	conn.Do("HSET", stock+":"+"Analysis", "Corona-Change", corona_chg_str)

	if ema8[len(ema8)-1] > ema50[len(ema50)-1] {
		conn.Do("HSET", stock+":"+"Analysis", "EMA8>EMA50", "Y")
	} else {
		conn.Do("HSET", stock+":"+"Analysis", "EMA8>EMA50", "N")
	}

	ema_pattern, _ := emaAnalysis(ema88, ema8)
	conn.Do("HSET", stock+":"+"Analysis", "EMA8-Pattern", ema_pattern)

	if ema8[len(ema8)-1] > ema88[len(ema88)-1] {
		conn.Do("HSET", stock+":"+"Analysis", "EMA8>EMA88", "Y")

		trend_days := 1
		prev_trend_days := 1
		prev2_trend_days := 1
		var i, j int
		for i = 2; i < len(ema8)-2; i++ {
			if ema8[len(ema8)-i] > ema88[len(ema88)-i] {
				trend_days++
			} else {
				break
			}
		}
		for j = i; j < len(ema8)-i-2; j++ {
			if ema8[len(ema8)-j] < ema88[len(ema88)-j] {
				prev_trend_days++
			} else {
				break
			}
		}
		for k := j; k < len(ema8)-i-j-2; k++ {
			if ema8[len(ema8)-k] > ema88[len(ema88)-k] {
				prev2_trend_days++
			} else {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T1", trend_days)
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T2", prev_trend_days)
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T3", prev2_trend_days)

	} else {
		conn.Do("HSET", stock+":"+"Analysis", "EMA8>EMA88", "N")
		trend_days := 1
		prev_trend_days := 1
		prev2_trend_days := 1
		var i, j int
		for i = 2; i < len(ema8)-2; i++ {
			if ema8[len(ema8)-i] < ema88[len(ema88)-i] {
				trend_days++
			} else {
				break
			}
		}
		for j = i; j < len(ema8)-i-2; j++ {
			if ema8[len(ema8)-j] > ema88[len(ema88)-j] {
				prev_trend_days++
			} else {
				break
			}
		}
		for k := j; k < len(ema8)-i-j-2; k++ {
			if ema8[len(ema8)-k] < ema88[len(ema88)-k] {
				prev2_trend_days++
			} else {
				break
			}
		}
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T1", trend_days)
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T2", prev_trend_days)
		conn.Do("HSET", stock+":"+"Analysis", "EMA88-T3", prev2_trend_days)
	}

	var ema_diff float64
	var ema_diff_str string

	ema_diff = ((ema8[len(ema8)-1] - ema50[len(ema50)-1]) / ema50[len(ema50)-1]) * 100
	ema_diff_str = fmt.Sprintf("%0.2f", ema_diff)
	conn.Do("HSET", stock+":"+"Analysis", "EMA-8-50-Diff", ema_diff_str)
	ema_diff = ((ema88[len(ema88)-1] - ema8[len(ema8)-1]) / ema8[len(ema8)-1]) * 100
	ema_diff_str = fmt.Sprintf("%0.2f", ema_diff)
	conn.Do("HSET", stock+":"+"Analysis", "EMA-8-88-Diff", ema_diff_str)
	conn.Do("HSET", stock+":"+"Analysis", "Sar-Pattern", sar_pattern[len(sar_pattern)-1])
	conn.Do("HSET", stock+":"+"Analysis", "NewHigh", positive)

	rsiAnalysis(conn, stock, rsi9, wrsi9, mwrsi9)

	heikenAshiAnalysis(conn, stock+":"+"Analysis", sclose, ha_open, ha_high, ha_low, ha_close, dates)
	//priceActionAnalysis(conn, stock+":"+"Analysis", sopen, shigh, slow, sclose)
}
