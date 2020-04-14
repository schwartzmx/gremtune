package gremcos

import "sync"

type CloseOnceChannel struct {
	Channel chan error
	once    sync.Once
}

func NewCloseOnceChannel(channel chan error) *CloseOnceChannel {
	return &CloseOnceChannel{
		Channel: channel,
	}
}

func (c *CloseOnceChannel) Close() {
	c.once.Do(func() {
		if c.Channel != nil {
			close(c.Channel)
		}
	})
}
