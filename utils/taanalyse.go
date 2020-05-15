package utils

import (
    "fmt"
    //"time"
    "math"
    "github.com/markcheno/go-talib"
    "github.com/gomodule/redigo/redis"
)

const (
    GREEN = 1
    RED = 2
)

var anotes map[string]string
var adata map[string]int

func heikenAshiPrepareCandles (ha_open []float64, ha_high []float64, ha_low []float64, ha_close []float64, color []int, ctype []string) () {

     for key, _ := range ha_close {

        if (ha_close[key] > ha_open[key]) {
            color[key] = GREEN
        } else {
            color[key] = RED
        }

        // Get current candle type, whether it is a doji, almost full candle or nuetral one.
        var candle_range float64
        if (ha_close[key] > ha_open[key]) {
            candle_range = ha_close[key] - ha_open[key]
        } else {
            candle_range = ha_open[key] - ha_close[key]
        }

        total_range := ha_high[key] - ha_low[key]
        candle_size := (candle_range * 100) / total_range


        if ((math.Round(ha_open[key]) == math.Round(ha_high[key])) || (math.Round(ha_open[key]) == math.Round(ha_low[key]))) {
            if (color[key] == GREEN) {
                ctype[key] = "Green-Full"
            } else {
                ctype[key] = "Red-Full"
            }
        } else if (candle_size < 20) {
            if (color[key] == GREEN) {
                ctype[key] = "Green-Doji"
            } else {
                ctype[key] = "Red-Doji"
            }
        } else {
            if (color[key] == GREEN) {
                ctype[key] = "Green-Nuetral"
            } else {
                ctype[key] = "Red-Nuetral"
            }
        }
     }

     return 
}

func heikenAshiAnalysis (conn redis.Conn, stock string, sclose []float64, ha_open []float64, ha_high []float64, ha_low []float64, ha_close []float64, dates []string) {

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

    if (color[today] == GREEN && color[yest] == GREEN) {
         redis.String(conn.Do("HSET", stock, "HA-Today", "Bullish"))
    } else if (color[today] == GREEN && color[yest] == RED) {
         redis.String(conn.Do("HSET", stock, "HA-Today", "Bullish-Turn"))
         today_trend_change = true
    } else if (color[today] == RED && color[yest] == RED) {
         redis.String(conn.Do("HSET", stock, "HA-Today", "Bearish"))
    } else if (color[today] == RED && color[yest] == GREEN) {
         redis.String(conn.Do("HSET", stock, "HA-Today", "Bearish-Turn"))
         today_trend_change = true
    }

    for i := today; i > today-11; i-- {
        if (color[i] != color[i-1]) {
             trend_changes++
        }
    }

    for i := yest; i > 0; i-- {
        if (color[i] != color[i-1]) {
             last_trend_change = i
             break
        }
    }

    // Calculate change in price during current/previous trend. This is current trend if there is no change in
    // trend from yesterday. If there is a change today, this is actually for the trend that got closed yesterday.
    if (today_trend_change) {
        chg_str := fmt.Sprintf("%.02f", ((sclose[today-1] - sclose[last_trend_change-1]) / sclose[last_trend_change-1]) * 100)
        redis.String(conn.Do("HSET", stock, "HA-PriceChg", chg_str))
        redis.String(conn.Do("HSET", stock, "HA-Trend-Days", (yest) - last_trend_change))
    } else  {
        chg_str := fmt.Sprintf("%.02f", ((sclose[today] - sclose[last_trend_change-1]) / sclose[last_trend_change-1]) * 100)
        redis.String(conn.Do("HSET", stock, "HA-Trend-Days", (today) - last_trend_change))
        redis.String(conn.Do("HSET", stock, "HA-PriceChg", chg_str))
    }

    redis.String(conn.Do("HSET", stock, "HA-Chg-10D", trend_changes))
    redis.String(conn.Do("HSET", stock, "HAC-T", ctype[today]))
    redis.String(conn.Do("HSET", stock, "HAC-1D", ctype[yest]))
    redis.String(conn.Do("HSET", stock, "HAC-2D", ctype[today-2]))
    redis.String(conn.Do("HSET", stock, "HAC-3D", ctype[today-3]))

    var pattern string
    if (color[today] == GREEN ) { pattern = "G-" }  else { pattern = "R-" }
    if (color[yest] == GREEN ) { pattern += "G-" }  else { pattern += "R-" }
    if (color[today-2] == GREEN ) { pattern += "G-" }  else { pattern += "R-" }
    if (color[today-3] == GREEN ) { pattern += "G" }  else { pattern += "R" }
    redis.String(conn.Do("HSET", stock, "HA-Pattern", pattern))
}

func emaAnalysis (conn redis.Conn, stock string, ema1 []float64, ema2 []float64, period1 int, period2 int, dates []string) {
    var last_bullish = 0
    var last_bearish = 0

    if (len(ema1) == 0 || len(ema2) == 0) {
        return
    }

    for i := 0; i <= len(ema1) - 2; i++ {
        if (ema1[i] < ema2[i] && ema1[i+1] > ema2[i+1]) {
            last_bullish = i
        }
        if (ema1[i] > ema2[i] && ema1[i+1] < ema2[i+1]) {
            last_bearish = i
        }
    }

    if (last_bearish > last_bullish) {
        temp := fmt.Sprintf("EMA%d-EMA%d-Status", period1, period2)
        conn.Do("HSET", stock, temp, "Bearish")
        temp = fmt.Sprintf("EMA%d-EMA%d-Days", period1, period2)
        conn.Do("HSET", stock, temp, len(ema1) - last_bearish)
    } else {
        temp := fmt.Sprintf("EMA%d-EMA%d-Status", period1, period2)
        redis.String(conn.Do("HSET", stock, temp, "Bullish"))
        temp = fmt.Sprintf("EMA%d-EMA%d-Days", period1, period2)
        redis.String(conn.Do("HSET", stock, temp, len(ema1) - last_bullish))
    }
}

func priceActionAnalysis (conn redis.Conn, stock string, sopen []float64, shigh []float64, slow []float64, sclose []float64) {

    var min, max float64

    min = sclose[len(sclose)-2]
    max = sclose[len(sclose)-2]
    for i := len(sclose)-2; i > len(sclose)-10; i-- {
        if (min > slow[i]) {
           min = slow[i]
        }
        if (max < shigh[i]) {
           max = shigh[i]
        }
    }

    crange := max - min
    pct := math.Round(((sclose[len(sclose) - 1] -  min))/crange *100)
    conn.Do("HSET", stock, "10D-Trend", pct)

    i := len(sclose)-1
    chg := ((sclose[i] - sclose[i-1])/sclose[i-1])*100
    d_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-5])/sclose[i-5])*100
    w_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-10])/sclose[i-10])*100
    w2_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-20])/sclose[i-20])*100
    m_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-60])/sclose[i-60])*100
    m3_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-125])/sclose[i-125])*100
    m6_str := fmt.Sprintf("%.02f", chg)
    chg = ((sclose[i] - sclose[i-250])/sclose[i-250])*100
    y_str := fmt.Sprintf("%.02f", chg)
    conn.Do("HSET", stock, "1-D", d_str, "1-W", w_str, "2-W", w2_str, "1-M", m_str, "3-M", m3_str, "6-M", m6_str, "1-Y", y_str)

    w52 := sclose[:53]
    high := w52[0]
    low := w52[0]
    for _, value := range w52 {
        if (value > high) {high = value}
        if (value < low) {high = value}
    }
    conn.Do("HSET", stock, "52-Week-High", high, "52-Week-Low", low)
}

func StockCommentary (conn redis.Conn, stock string, sopen []float64, shigh []float64, slow []float64, sclose []float64, dates []string) {

    anotes = make(map[string]string, 40)
    adata = make(map[string]int, 40)

    ha_high, ha_open, ha_close, ha_low := talib.HeikinashiCandles(shigh, sopen, sclose, slow)

    var ema5, ema13, ema50, ema200, rsi14 []float64
    /* Compute EMA */
    if (len(sclose) > 5) {
        ema5 = talib.Ema(sclose, 5)
    }
    if (len(sclose) > 5) {
        ema13 = talib.Ema(sclose, 13)
    }
    if (len(sclose) > 5) {
        ema50 = talib.Ema(sclose, 50)
    }
    if (len(sclose) > 5) {
        ema200 = talib.Ema(sclose, 200)
    }
    if (len(sclose) > 15) {
        rsi14 = talib.Rsi(sclose, 14)
        rsi_str := fmt.Sprintf("%.02f",rsi14[len(rsi14)-1])
        redis.String(conn.Do("HSET", stock+":"+"Analysis", "RSI", rsi_str))
    }

    emaAnalysis(conn, stock+":"+"Analysis", ema5, ema13, 5, 13, dates)
    emaAnalysis(conn, stock+":"+"Analysis", ema50, ema200, 50, 200, dates)
    heikenAshiAnalysis(conn, stock+":"+"Analysis", sclose, ha_open, ha_high, ha_low, ha_close, dates)
    priceActionAnalysis(conn, stock+":"+"Analysis", sopen, shigh, slow, sclose)
}
