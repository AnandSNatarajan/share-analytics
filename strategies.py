import sys
import stockramdb as srb

dates = []
stocklist = []
total_success = 0
total_failure = 0

# In this strategy, we look at a volume increase of above 50% of previous day and a price drop. This means there is off-loading
# but at the same time there is an investment interest. So the stock can potentially go up very soon.

def strategy_1 (s):
    success = 0
    failure = 0
    prev_volume = s.volume[0]
    prev_price = s.cp[0]
    global total_success
    global total_failure

    for day in range (1, len(s.dates) - 1):

        if ((s.volume[day] > prev_volume) & (((s.volume[day] - prev_volume) * 100)/prev_volume > 50) & (s.cp[day] < prev_price)):

            if ((s.cp[day] < s.cp[day+1])):
                success += 1 
                print("Success date : " + str(s.dates[day]))
            else:
                failure += 1 
                print("Failure date : " + str(s.dates[day]))

            prev_volume = s.volume[day]
            prev_price = s.cp[day]

    if (success + failure > 0):
        success_percentile = (success * 100)/(success + failure)
        print("%s Success percentile : %d%% Sucess %d Failure %d" % (s.name, success_percentile, success, failure))
    else:
        print("%s No hits" % scrip)


# In this strategy, we look at stocks that have gone down for 3 consecutive days. If the price of the stock went up after the 4th day
# in any of the next 3 days, the strategy is considered as a success.

def strategy_2 (s):
    success = 0
    failure = 0
    down_day = 0
    global total_success
    global total_failure

    # Check if the price is dropping for 3rd consecutive day and check price on 4th day.
    for day in range (1, len(s.dates) - 1):

        if (s.cp[day-1] > s.cp[day]):
            down_day += 1
        else:
            down_day = 0

        if (down_day == 3):

            down_day = 0

            if ((day+6) < (len(s.cp) - 1)): 
                if ((s.cp[day+4] > s.cp[day]) | (s.cp[day+5] > s.cp[day]) | (s.cp[day+6] > s.cp[day])):
                    success += 1
                    total_success += 1
                    print("Success date : %s" % dates[day])
                else:
                    failure += 1
                    total_failure += 1
                    print("Failure date : %s" % dates[day])

    if (success + failure > 0):
        success_percentile = (success * 100)/(success + failure)
        print("%s Success percentile : %d%% Sucess %d Failure %d" % (s.name, success_percentile, success, failure))
    else:
        print("%s No hits" % scrip)

# Open the dates file.
f = open(sys.argv[1], "r")

for line in f:

    # Get a date from the file.
    line = line.rstrip('\n')

    dates.append(line)

# Open the stocks file.
f = open(sys.argv[2], "r")

for line in f:
    # Get a stock name from the file.
    line = line.rstrip('\n')

    s = srb.stockdata(line)

    s = srb.parse_stocks(s, dates)

    stocklist.append(s)

for s in stocklist:
    strategy_2(s)

print("Total success : %d, Total failure : %d" % (total_success, total_failure))
