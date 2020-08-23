package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"121.36.61.156:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		// handle error!
		fmt.Println(err)
		return
	}

	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), time.Second * 5)
	resp, err := cli.Put(ctx, "sample_key", "sample_value")
	//cancel()
	if err != nil {
		fmt.Println(err)
		return
		// handle error!
	}

	a, err := cli.Get(ctx, "aaa")
	b, err := cli.Get(ctx, "sample_key")

	fmt.Println(resp, a, b)

}
