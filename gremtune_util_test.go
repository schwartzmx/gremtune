package gremtune

import (
	"testing"
	"time"

	"github.com/schwartzmx/gremtune/interfaces"
	"github.com/stretchr/testify/require"
)

var failingErrorChannelConsumerFunc = func(errChan chan error, t *testing.T) {
	err := <-errChan
	if err == nil {
		return
	}
	t.Fatalf("Lost connection to the database: %s", err.Error())
}

func newTestClient(t *testing.T, errChan chan error) interfaces.QueryExecutor {
	websocket, err := NewWebsocket("ws://127.0.0.1:8182/gremlin")
	require.NotNil(t, websocket, "Dialer is nil")
	require.NoError(t, err)
	client, err := Dial(websocket, errChan)
	require.NoError(t, err, "Failed to create client")
	return client
}

func newTestPool(t *testing.T, errChan chan error) *pool {
	createQueryExecutorFn := func() (interfaces.QueryExecutor, error) {
		websocket, err := NewWebsocket("ws://127.0.0.1:8182/gremlin")
		require.NoError(t, err)
		c, err := Dial(websocket, errChan)
		require.NoError(t, err)

		return c, err
	}

	return &pool{
		createQueryExecutor: createQueryExecutorFn,
		maxActive:           10,
		idleTimeout:         time.Duration(10 * time.Second),
	}
}

func truncateData(t *testing.T, client interfaces.QueryExecutor) {
	t.Log("Removing all data from gremlin server started...")

	_, err := client.Execute(`g.V('1234').drop()`)
	require.NoError(t, err)

	_, err = client.Execute(`g.V('2145').drop()`)
	require.NoError(t, err)
	t.Log("Removing all data from gremlin server completed...")
}

func seedData(t *testing.T, client interfaces.QueryExecutor) {
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
