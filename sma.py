import numpy as np

data = [2,3,4,2,3,4,5,6,7,8,8]

def sma_calculator(values, window):
    weights = np.repeat(1.0,window)/window
    smas = np.convolve(values, weights, 'valid')
    return smas

def ema_calculator(values, window):
    weights = np.exp(np.linspace(-1.,0.,window)) 
    weights = weights.sum()

    emas = np.convolve(values, weights)[:len(values)]
    emas[:window] = emas[window]
    return emas

smas = sma_calculator(data, 3)
emas = ema_calculator(data, 3)

print "Simple Moving Average"
print smas
print "Exponential Moving Average"
print emas[len(data)-1]
