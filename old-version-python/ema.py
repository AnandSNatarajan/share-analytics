import numpy as np
import pandas as pd
import talib

data = [2,3,4,2,3,4,5,6,7,8,8,2,3,4,2,3,4,5,6,7,8,8,2,3,4,2,3,4,5,6,7,8,8,2,3,4,2,3,4,5,6,7,8,8,2,3,4,2,3,4,5,6,7,8,8,2,3,4,2,3,4,5,6,7,8,8]

float_data = [float(x) for x in data]
print "Exponential Moving Average"
close = np.array(float_data)
a,b,c = talib.MACD(close, fastperiod=3, slowperiod=10, signalperiod=2)
print a
