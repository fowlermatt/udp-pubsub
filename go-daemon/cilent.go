package main

import (
	"net"
	"sync"

	"github.com/willf/bloom"
)

type Client struct {
	conn   net.Conn
	filter *bloom.BloomFilter
	topics []string
	mu     sync.RWMutex
}
