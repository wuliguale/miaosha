#! /bin/bash

#判断port是否监听

port=$1
portstr=$(netstat -antlup | grep "$port")

if [ -z "$portstr" ]
then
    res=1
else
    res=0
fi

exit $res

