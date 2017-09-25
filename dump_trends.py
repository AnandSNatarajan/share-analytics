import sys
import subprocess
import json
import re

f = open("dates", "r")

dates = []
cp = []
volume = []

array_index = 1

for line in f:

    # Get a date from the file
    line = line.rstrip('\n')

    # Construct the command for querying share data for a particular date
    sub_cmd = "ovsdb-client transact --pretty '[\"Share_List\",{\"op\":\"select\", \"table\":\"Share_List\", \"where\":[[\"name\",\"==\",\""+sys.argv[1]+"\"],[\"date\",\"==\","+line+"]], \"columns\":[\"name\",\"date\",\"cp\",\"volume\"]}]'"
    out = subprocess.check_output(sub_cmd, shell=True)

    # Create a list with the JSON information
    raw_data = json.loads(out)

    # Append the data to the respective arrays
    dates.append(int(line))
    
    m1 = re.search('cp\': u\'([0-9\.]*)', str(raw_data[0]))
    cp.append(float(m1.group(1)))
            
    m2 = re.search('volume\': ([0-9]*)', str(raw_data[0]))
    volume.append(int(m2.group(1)))

print dates

print cp

print volume
