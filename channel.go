package gremcos

import "sync"

type safeCloseErrorChannel struct {
	c    chan error
	once sync.Once
}

func newSafeCloseErrorChannel(channelBuffer int) *safeCloseErrorChannel {
	return &safeCloseErrorChannel{
		c: make(chan error, channelBuffer),
	}
}

func (c *safeCloseErrorChannel) Close() {
	c.once.Do(func() {
		if c.c != nil {
			close(c.c)
		}
	})
}
