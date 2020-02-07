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
	//mockedDialer.EXPECT().Ping()
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

//func TestPing(t *testing.T) {
//	// GIVEN
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
//	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
//	errorChannel := make(chan error, 5)
//
//	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory), SetPingInterval(time.Millisecond*100))
//	require.NoError(t, err)
//	require.NotNil(t, dialer)
//
//	// WHEN
//	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
//	err = dialer.Connect()
//	require.NoError(t, err)
//
//	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(nil)
//	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(fmt.Errorf("ERR")).AnyTimes()
//	mockedWebsocketConnection.EXPECT().Close().Return(nil)
//	go dialer.Ping()
//
//	// wait a bit to allow the ping timer to tick
//	time.Sleep(time.Millisecond * 500)
//	dialer.Close()
//
//	// THEN
//	assert.False(t, dialer.IsConnected())
//	assert.NotEmpty(t, errorChannel)
//}
