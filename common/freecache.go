package common

import (
	"github.com/coocood/freecache"
)

type FreeCacheClient struct {
	Client *freecache.Cache
}

func NewFreeCacheClient(cacheMb int) (client *FreeCacheClient) {
	cacheSize := cacheMb * 1024 * 1024
	cache := freecache.NewCache(cacheSize)

	//debug.SetGCPercent(20)
	return &FreeCacheClient{cache}
}


func (client *FreeCacheClient) Get(key []byte) (value []byte, err error){
	return client.Client.Get(key)
}

func (client *FreeCacheClient) Set(key []byte, value []byte, durationSeconds int) (err error){
	return client.Client.Set(key, value, durationSeconds)
}

func (client *FreeCacheClient) Del(key []byte) (affected  bool){
	return client.Client.Del(key)
}


func (client *FreeCacheClient) IsNotFound(err error) bool {
	return err == freecache.ErrNotFound
}
