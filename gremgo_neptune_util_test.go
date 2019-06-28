package gremgo

import (
	"fmt"
	"log"
	"time"
)

var g *Client
var errs = make(chan error)
var gp *Pool
var gperrs = make(chan error)

// InitGremlinClients intializes gremlin client and pool for use by tests
func InitGremlinClients() {
	go func(chan error) {
		err := <-errs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs)
	go func(chan error) {
		err := <-gperrs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs)
	initClient()
	initPool()
}

func initClient() {
	if g != nil {
		return
	}
	var err error
	dialer := NewDialer("ws://127.0.0.1:8182")
	r, err := Dial(dialer, errs)
	if err != nil {
		fmt.Println(err)
	}
	g = &r
}

func initPool() {
	if gp != nil {
		return
	}
	dialFn := func() (*Client, error) {
		dialer := NewDialer("ws://127.0.0.1:8182")
		c, err := Dial(dialer, gperrs)
		if err != nil {
			log.Fatal(err)
		}
		return &c, err
	}
	pool := Pool{
		Dial:        dialFn,
		MaxActive:   10,
		IdleTimeout: time.Duration(10 * time.Second),
	}
	gp = &pool
}
