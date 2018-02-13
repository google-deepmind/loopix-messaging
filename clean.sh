#!/usr/bin/env bash

# This script should be run from inside the anonymous-messaging package directory

logDir="$(PWD)/logs"
pkiDir="$(PWD)/pki/database.db"

if [ -d $pkiDir ]
then
    echo "Removing the following directory" $pkiDir
    rm -f $pkiDir
    echo "Removed existing PKI files"
else
    echo "Nothing to remove. The PKI directory does not exists"
fi


if [ -d $logDir ]
then
    echo "Removing existing logs in the following directory" $logDir
    rm -rf $logDir
    echo "Creating a new log folder in directory" $logDir
    mkdir $logDir
else
    echo "Nothing to remove. The logs directory does not exist."
fi

function kill_port() {
    PID=$(lsof -t -i:$1)
    echo "Killing process: $PID"
#    kill -TERM ${PID}
    kill -KILL ${PID}
#    kill -TSTP ${PID}
#    kill -CONT ${PID}
}

for var in "$@"
do
    kill_port ${var}
done
