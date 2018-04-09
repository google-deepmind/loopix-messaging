#// Copyright 2018 The Loopix-Messaging Authors
#//
#// Licensed under the Apache License, Version 2.0 (the "License");
#// you may not use this file except in compliance with the License.
#// You may obtain a copy of the License at
#//
#//      http://www.apache.org/licenses/LICENSE-2.0
#//
#// Unless required by applicable law or agreed to in writing, software
#// distributed under the License is distributed on an "AS IS" BASIS,
#// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#// See the License for the specific language governing permissions and
#// limitations under the License.

#!bin/sh

echo "Press CTRL-C to stop."

logDir="$(PWD)/logs"

if [ -d $logDir ]
then
    echo "Logging directory already exists"
else
    mkdir $logDir
    echo "Created logging directory"
fi

go run main.go -typ=client -id=Client1 -host=localhost -port=9996 -provider=Provider >> logs/bash.log ;

#NUMCLIENTS=$1
#
#for (( j=0; j<NUMCLIENTS; j++ ));
#do
#    go run main.go -typ=mix -id=Client1 -host=localhost -port=$((9990+$j)) -provider=Provider > logs/bash.log &
#    sleep 1
#done
#
#sleep 1


trap ctrl_c SIGINT SIGTERM SIGTSTP
function ctrl_c() {
        echo "** Trapped SIGINT, SIGTERM and SIGTSTP"
        kill_port 9996
#        for (( j=0; j<NUMCLIENTS; j++ ));
#        do
#            kill_port $((9980+$j))
#        done
}

function kill_port() {
    PID=$(lsof -t -i:$1)
    echo "$PID"
    kill -TERM $PID || kill -KILL $PID
}