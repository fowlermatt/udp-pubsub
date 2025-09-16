package main

import (
	"encoding/binary"
	"fmt"
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

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:   conn,
		topics: make([]string, 0),
	}
}

func (c *Client) UpdateFilter(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(data) < 8 {
		return fmt.Errorf("invalid filter data: too short")
	}

	bitSize := binary.BigEndian.Uint32(data[0:4])
	hashCount := binary.BigEndian.Uint32(data[4:8])

	c.filter = bloom.New(uint(bitSize), uint(hashCount))

	return nil
}

func (c *Client) TestTopic(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.filter == nil {
		return false
	}

	return c.filter.Test([]byte(topic))
}

func (c *Client) AddTopic(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topics = append(c.topics, topic)
	if c.filter != nil {
		c.filter.Add([]byte(topic))
	}
}
