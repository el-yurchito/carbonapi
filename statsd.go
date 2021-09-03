package main

import (
	"time"

	"gopkg.in/alexcesaro/statsd.v2"
)

const (
	flushPeriod = 200 * time.Millisecond
	poolSize    = 16
)

type StatsdLimiter struct {
	pool chan *boundStatsdClient
}

var statsdLimiter *StatsdLimiter

func NewStatsdLimiter(config statsdConfig) (*StatsdLimiter, error) {
	limiter := &StatsdLimiter{pool: make(chan *boundStatsdClient, poolSize)}
	options := []statsd.Option{
		statsd.Address(config.Address),
		statsd.FlushPeriod(flushPeriod),
		statsd.Prefix(config.Prefix),
		statsd.Network("udp"),
		statsd.Mute(!config.Enabled),
	}

	for i := 0; i < poolSize; i++ {
		client, err := statsd.New(options...)
		if err != nil {
			return nil, err
		}
		limiter.pool <- &boundStatsdClient{
			Client:  client,
			limiter: limiter,
		}
	}

	return limiter, nil
}

func (limiter *StatsdLimiter) Get() *boundStatsdClient {
	return <-limiter.pool
}

type boundStatsdClient struct {
	*statsd.Client
	limiter *StatsdLimiter
}

func (b *boundStatsdClient) Release() {
	b.limiter.pool <- b
}
