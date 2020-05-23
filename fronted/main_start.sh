#!/bin/bash

cd /opt/code/miaosha-demo/fronted
rm main

go build main.go

ps -aux | grep main | grep -v grep | grep -v \.sh | awk '{print $2}' | while read pid
do
    echo "kill main $pid"
    kill -9 $pid
done

./main &


