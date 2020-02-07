package gremtune

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanicOnMissingAuthCredentials(t *testing.T) {
	c := newClient(nil)
	ws := &websocket{}
	c.conn = ws

	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	c.conn.GetAuth()
}

func TestNewClient(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)

	// WHEN
	client := newClient(mockedDialer)

	// THEN
	require.NotNil(t, client)
	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.requests)
	assert.NotNil(t, client.responses)
	assert.NotNil(t, client.results)
	assert.NotNil(t, client.responseNotifier)
	assert.NotNil(t, client.responseStatusNotifier)
	assert.False(t, client.Errored)
}

func TestDial(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	errorChannel := make(chan error)
	quitChannel := make(chan struct{})

	// WHEN
	mockedDialer.EXPECT().Connect().Return(nil)
	mockedDialer.EXPECT().GetQuitChannel().Return(quitChannel)
	mockedDialer.EXPECT().Ping(errorChannel)
	mockedDialer.EXPECT().Read().Return(-1, nil, fmt.Errorf("Read failed"))
	mockedDialer.EXPECT().Close()
	client, err := Dial(mockedDialer, errorChannel)
	require.NotNil(t, client)
	require.NoError(t, err)
	client.Close()

	// FIXME: Remove the sleep here
	time.Sleep(time.Second * 2)
	// THEN
	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.requests)
	assert.NotNil(t, client.responses)
	assert.NotNil(t, client.results)
	assert.NotNil(t, client.responseNotifier)
	assert.NotNil(t, client.responseStatusNotifier)
	assert.False(t, client.Errored)
}
