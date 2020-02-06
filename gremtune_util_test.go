package gremtune

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var failingErrorChannelConsumerFunc = func(errChan chan error, t *testing.T) {
	err := <-errChan
	if err == nil {
		return
	}
	t.Fatalf("Lost connection to the database: %s", err.Error())
}

func newTestClient(t *testing.T, errChan chan error) *Client {
	dialer := NewWebsocketDialer("ws://127.0.0.1:8182")
	require.NotNil(t, dialer, "Dialer is nil")
	client, err := Dial(dialer, errChan)
	require.NoError(t, err, "Failed to create client")
	return &client
}

func newTestPool(t *testing.T, errChan chan error) *Pool {
	dialFn := func() (*Client, error) {
		dialer := NewWebsocketDialer("ws://127.0.0.1:8182")
		c, err := Dial(dialer, errChan)
		require.NoError(t, err)
		return &c, err
	}

	return &Pool{
		Dial:        dialFn,
		MaxActive:   10,
		IdleTimeout: time.Duration(10 * time.Second),
	}
}

func truncateData(t *testing.T, client *Client) {
	t.Log("Removing all data from gremlin server started...")

	_, err := client.Execute(`g.V('1234').drop()`)
	require.NoError(t, err)

	_, err = client.Execute(`g.V('2145').drop()`)
	require.NoError(t, err)
	t.Log("Removing all data from gremlin server completed...")
}

func seedData(t *testing.T, client *Client) {
	truncateData(t, client)

	t.Log("Seeding data started...")

	_, err := client.Execute(`
		g.addV('Phil').property(id, '1234').
			property('timestamp', '2018-07-01T13:37:45-05:00').
			property('source', 'tree').
			as('x').
		  addV('Vincent').property(id, '2145').
			property('timestamp', '2018-07-01T13:37:45-05:00').
			property('source', 'tree').
			as('y').
		  addE('brother').
			from('x').
			to('y')
	`)
	require.NoError(t, err)
	t.Log("Seeding data completed...")
}
