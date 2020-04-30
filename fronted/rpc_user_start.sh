#!/bin/bash

cd /opt/code/miaosha-demo/fronted
rm rpc_user_server

go build rpc_user_server.go

ps -aux | grep rpc_user | awk '{print $2}' | while read pid
do
    echo "kill rpc user $pid"
    kill -9 $pid
done

./rpc_user_server --ip 0.0.0.0 --port 9001 &
./rpc_user_server --ip 0.0.0.0 --port 9002 &
./rpc_user_server --ip 0.0.0.0 --port 9003 &



