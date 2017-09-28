###############################################################################
# Open multiple bhavcopy files available in a directory and store the closing 
# share price and volume information for all the stocks.
###############################################################################

import sys
import subprocess
from os import listdir
from os.path import isfile, join, isdir

# Help check.
if (len(sys.argv) == 1):
    print("Usage: python eodparser.py <info directory>")
    sys.exit()

# Ensure that a directory name is specified for retreiving bhavcopy files.
if (isdir(sys.argv[1]) == 0):
    print("Invalid directory specified")
    sys.exit()

# Retreive all file names from the directory. This means all files should only be share data.
onlyfiles = [f for f in listdir(sys.argv[1]) if isfile(join(sys.argv[1], f))]

for sf in onlyfiles:
    # Open the bhavcopy file
    print("Opening file " + sf)
    fd = open(join(sys.argv[1], sf), "r")

    for line in fd:
        line = line.rstrip('\n')
        fields = line.split(",")

        # Parse each line and obtain the closing price and volume data for the stock, push it into ovsdb.
        if len(fields) > 5:
            #sub_cmd = "ovsdb-client transact \'[\"Share_List\",{\"op\":\"insert\", \"table\":\"Share_List\", \"row\":{\"name\":\""+fields[0]+"\", \"date\":"+fields[1]+", \"cp\":\""+fields[5]+"\", \"volume\":"+fields[6]+"}}]'"
            #subprocess.call(sub_cmd, shell=True)
        else:   
            sub_cmd = "ovsdb-client transact \'[\"Share_List\",{\"op\":\"insert\", \"table\":\"Share_List\", \"row\":{\"name\":\""+fields[0]+"\", \"date\":"+fields[1]+", \"cp\":\""+fields[5]+"\"}}]'"
            subprocess.call(sub_cmd, shell=True)
