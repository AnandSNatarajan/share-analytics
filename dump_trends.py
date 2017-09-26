###############################################################################
# Given a particular share name, get the share price and volume information for
# all dates specified in the "dates" file and store them into an array. This 
# array can then be manipulated to get analytics information.
###############################################################################

import sys
import subprocess
import json
import re
import numpy as np

# Initialize the dates, closing price and volume arrays.
dates = []
cp = []
volume = []

def sma_calculator(values, window):
    weights = np.repeat(1.0,window)/window
    smas = np.convolve(values, weights, 'valid')
    return smas

def parse_info(scrip):

    del dates[:]
    del cp[:]
    del volume[:]

    # Open the dates file.
    f = open("dates", "r")

    for line in f:

        # Get a date from the file.
        line = line.rstrip('\n')

        # Construct the command for querying share data for a particular date.
        sub_cmd = "ovsdb-client transact --pretty '[\"Share_List\",{\"op\":\"select\", \"table\":\"Share_List\", \"where\":[[\"name\",\"==\",\""+scrip+"\"],[\"date\",\"==\","+line+"]], \"columns\":[\"name\",\"date\",\"cp\",\"volume\"]}]'"
        out = subprocess.check_output(sub_cmd, shell=True)

        # Create a list with the JSON information.
        raw_data = json.loads(out)

        # Append the data to date array.
        dates.append(int(line))
    
        # Append the data to the closing price array.
        m1 = re.search('cp\': u\'([0-9\.]*)', str(raw_data[0]))
        cp.append(float(m1.group(1)))
            
        # Append the data to the volume array.
        m2 = re.search('volume\': ([0-9]*)', str(raw_data[0]))
        volume.append(int(m2.group(1)))

def dump_stock_data():
    for i in range (0, len(dates) - 1):
        print(str(dates[i]) + " " + str(cp[i]) + " " + str(volume[i]))

def strategy_1 (scrip):
    success = 0
    failure = 0

    # Calculate SMAs (it results in an array that is less in size of window_size - 1 because the first elements
    # cannot have an SMA. So this needs to be adjusted while looking up for trigger.
    window_size = 10
    smas = sma_calculator(cp, window_size)

    for day in range (0, len(dates) - 1):
        if (day > 0):
            if ((volume[day] > prev_volume) & (((volume[day] - prev_volume) * 100)/prev_volume > 50) \
                    & (cp[day] < prev_price)):
                #print("Volume and price increased on " + str(dates[day]))
                if ((cp[day] < cp[day+1])):
                    success += 1 
                    print("Success date : " + str(dates[day]))
                        #print("Price increased next day" + str(cp[day]) + " " + str(cp[day+1]))
                else:
                    failure += 1 
                    print("Failure date : " + str(dates[day]))
                    #print("Price decreased next day " + str(cp[day]) + " " + str(cp[day+1]))
            prev_volume = volume[day]
        prev_price = cp[day]

    success_percentile = (success * 100)/(success + failure)
    print("%s Success percentile : %d%% Sucess %d Failure %d" (scrip, success_percentile, success, failure))


def strategy_2 (scrip):
    success = 0
    failure = 0
    down_day = 0


    # Check if the price is dropping for 3rd consecutive day and check price on 4th day.
    for day in range (1, len(dates) - 1):
        if (cp[day-1] > cp[day]):
            down_day += 1
        else:
            down_day = 0

        if (down_day == 3):
            down_day = 0
            if ((day+6) < (len(cp) - 1)): 
                if ((cp[day+4] > cp[day]) | (cp[day+5] > cp[day]) | (cp[day+6] > cp[day])):
                    success += 1
                    #print("Success date : %s" % dates[day])
                    #print("Price increased next day" + str(cp[day]) + " " + str(cp[day+1]))
                else:
                    failure += 1
                    #print("Failure date : %s" % dates[day])
                    #print("Price decreased next day " + str(cp[day]) + " " + str(cp[day+1]))

    success_percentile = (success * 100)/(success + failure)
    print("%s Success percentile : %d%% Sucess %d Failure %d" % (scrip, success_percentile, success, failure))


if (len(sys.argv) < 2):
    print("Usage: python dump_trends.py <stock_list>")
    sys.exit()
    
# Open the stocks file.
f = open(sys.argv[1], "r")

for line in f:
    # Get a date from the file.
    line = line.rstrip('\n')

    parse_info(line)

    # Execute the strategy of choice
    strategy_2(line)
