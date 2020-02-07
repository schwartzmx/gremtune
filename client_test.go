package gremtune

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/schwartzmx/gremtune/interfaces"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCredentials(t *testing.T) {
	assert.Error(t, validateCredentials(interfaces.Auth{}))
	assert.Error(t, validateCredentials(interfaces.Auth{Username: "Hans"}))
	assert.NoError(t, validateCredentials(interfaces.Auth{Username: "Hans", Password: "PW"}))
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
	mockedDialer.EXPECT().Read().Return(-1, nil, fmt.Errorf("Read failed")).AnyTimes()
	mockedDialer.EXPECT().Close().Do(func() {
		close(quitChannel)
	})
	client, err := Dial(mockedDialer, errorChannel)
	require.NotNil(t, client)
	require.NoError(t, err)
	err = client.Close()

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.requests)
	assert.NotNil(t, client.responses)
	assert.NotNil(t, client.results)
	assert.NotNil(t, client.responseNotifier)
	assert.NotNil(t, client.responseStatusNotifier)
	assert.False(t, client.Errored)
}

func TestPingWorker(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := &Client{
		conn:         mockedDialer,
		pingInterval: time.Millisecond * 100,
	}
	errorChannel := make(chan error, 5)
	quitChannel := make(chan struct{})

	// WHEN
	mockedDialer.EXPECT().Ping().Return(nil)
	mockedDialer.EXPECT().Ping().Return(fmt.Errorf("Error")).AnyTimes()
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.pingWorker(errorChannel, quitChannel)

	time.Sleep(time.Millisecond * 500)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
}
