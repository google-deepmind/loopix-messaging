#!bin/sh

go run main.go -typ=client -id=Client1 -host=localhost -port=9996 -provider=Provider > logs/Client1.log &
go run main.go -typ=client -id=Client2 -host=localhost -port=9995 -provider=Provider > logs/Client2.log ;

trap ctrl_c SIGINT SIGTERM SIGTSTP
function ctrl_c() {
        echo "** Trapped SIGINT, SIGTERM and SIGTSTP"
        kill_port 9996
        kill_port 9995
}

function kill_port() {
    PID=$(lsof -t -i:$1)
    echo "$PID"
    kill -TERM $PID || kill -KILL $PID
}