package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"time"
)

func UpdateLoad(client redis.Cmdable, nodeName string) error {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			rnd := float64(rand.Intn(100))
			//fmt.Println(rnd)
			client.ZAdd(context.Background(), "node_load", redis.Z{
				Score:  rnd,
				Member: nodeName,
			})
		}
	}

	return nil
}
