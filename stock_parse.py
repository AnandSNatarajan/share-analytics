import sys
import subprocess
from os import listdir
from os.path import isfile, join, isdir

if (len(sys.argv) == 1):
    print("Usage: python stock_parse.py <info directory>")
    sys.exit()

if (isdir(sys.argv[1]) == 0):
    print("Invalid directory specified")
    sys.exit()

onlyfiles = [f for f in listdir(sys.argv[1]) if isfile(join(sys.argv[1], f))]

for sf in onlyfiles:
    print("Opening file " + sf)
    fd = open(join(sys.argv[1], sf), "r")
    for line in fd:
        line = line.rstrip('\n')
        fields = line.split(",")
        if len(fields) > 5:
            print("Name " + fields[0] + " Date " + fields[1] + " CP " + fields[5] + " Volume " + fields[6]) 
            sub_cmd = "ovsdb-client transact \'[\"Share_List\",{\"op\":\"insert\", \"table\":\"Share_List\", \"row\":{\"name\":\""+fields[0]+"\", \"date\":"+fields[1]+", \"cp\":\""+fields[5]+"\", \"volume\":"+fields[6]+"}}]'"
            print(sub_cmd)
            subprocess.call(sub_cmd, shell=True)
            

