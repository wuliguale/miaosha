#!/bin/bash

cd /opt/code/miaosha-demo/fronted
rm mq_receive

go build mq_receive.go

ps -aux | grep mq_receive | grep -v grep | grep -v \.sh | awk '{print $2}' | while read pid
do
    echo "kill mq_receive $pid"
    kill -9 $pid
done

./mq_receive &

