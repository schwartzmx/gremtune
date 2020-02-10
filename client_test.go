package gremtune

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/schwartzmx/gremtune/interfaces"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func packedRequest2Request(packedRequest []byte) (request, error) {

	// the actual request is prepended by the mimetype and its length
	lenMimeType := len(MimeType)

	// remove the mimetype and the byte that specifies the length of the mimetype
	lenToRemove := lenMimeType + 1
	requestData := packedRequest[lenToRemove:]

	// now we have only the bytes of the request
	// --> unmarshal it
	result := request{}
	if err := json.Unmarshal(requestData, &result); err != nil {
		return request{}, err
	}
	return result, nil
}

func TestExecuteAsyncRequest(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	mockedDialer.EXPECT().IsDisposed().Return(false)

	responseChannel := make(chan AsyncResponse)

	err := client.ExecuteAsync("g.V()", responseChannel)
	require.NoError(t, err)

	// catch the request that should be send over the wire
	requestToSend := <-client.requests
	// convert it to a readable request
	req, err := packedRequest2Request(requestToSend)
	require.NoError(t, err)

	// read back the response
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		response := <-responseChannel
		assert.Equal(t, req.RequestID, response.Response.RequestID)
	}()

	// now create the according response
	response := Response{RequestID: req.RequestID, Status: Status{Code: statusSuccess}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// now inject send the response
	err = client.handleResponse(packet)
	require.NoError(t, err)

	// wait until the response was read
	wg.Wait()
}

func TestExecuteRequest(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	mockedDialer.EXPECT().IsDisposed().Return(false)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := client.Execute("g.V()")
		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	}()

	// catch the request that should be send over the wire
	requestToSend := <-client.requests
	// convert it to a readable request
	req, err := packedRequest2Request(requestToSend)
	require.NoError(t, err)

	// now create the according response
	response := Response{RequestID: req.RequestID, Status: Status{Code: statusSuccess}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// now inject send the response
	err = client.handleResponse(packet)
	require.NoError(t, err)

	// wait until the execution has been completed
	wg.Wait()
}

func TestExecuteRequestFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	mockedDialer.EXPECT().IsDisposed().Return(true)

	resp, err := client.Execute("g.V()")
	assert.Empty(t, resp)
	assert.Error(t, err)
}

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
	go func() {
		// just consume the errors to avoid blocking
		for range errorChannel {
		}
	}()
	quitChannel := make(chan struct{})

	// WHEN
	mockedDialer.EXPECT().Connect().Return(nil)
	mockedDialer.EXPECT().GetQuitChannel().Return(quitChannel)
	mockedDialer.EXPECT().Read().Return(1, nil, fmt.Errorf("Read failed")).AnyTimes()
	mockedDialer.EXPECT().Close().Do(func() {
		close(quitChannel)
	}).Return(nil)
	client, err := Dial(mockedDialer, errorChannel)
	require.NotNil(t, client)
	require.NoError(t, err)
	err = client.Close()
	close(errorChannel)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, client.conn)
	assert.NotNil(t, client.requests)
	assert.NotNil(t, client.responses)
	assert.NotNil(t, client.results)
	assert.NotNil(t, client.responseNotifier)
	assert.NotNil(t, client.responseStatusNotifier)
	assert.True(t, client.Errored)
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

func TestWriteWorker(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	dataChannel := make(chan []byte)
	client := &Client{
		conn:     mockedDialer,
		requests: dataChannel,
	}
	errorChannel := make(chan error)
	quitChannel := make(chan struct{})
	wg := sync.WaitGroup{}
	packet := []byte("ABCDEFG")
	numPackets := 10

	// WHEN
	mockedDialer.EXPECT().Write(packet).Return(nil).Times(numPackets)
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.writeWorker(errorChannel, quitChannel)

	// send some data on the channel
	wg.Add(1)
	go func() {
		for i := 0; i < numPackets; i++ {
			dataChannel <- packet
		}
		wg.Done()
	}()

	// wait until data was written and consumed
	wg.Wait()
	client.Close()

	// THEN
	assert.Empty(t, errorChannel)
	assert.False(t, client.Errored)
}

func TestWriteWorkerFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	dataChannel := make(chan []byte, 11)
	client := &Client{
		conn:     mockedDialer,
		requests: dataChannel,
	}
	wg := sync.WaitGroup{}
	errorChannel := make(chan error)
	var errors []error
	wg.Add(1)
	go func() {
		defer wg.Done()
		// just consume the errors to avoid blocking
		for err := range errorChannel {
			errors = append(errors, err)
		}
	}()

	quitChannel := make(chan struct{})
	packet := []byte("ABCDEFG")
	numPackets := 10

	// WHEN
	mockedDialer.EXPECT().Write(packet).Return(fmt.Errorf("Write failed")).Times(numPackets)
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.writeWorker(errorChannel, quitChannel)

	// send some data on the channel
	//wg.Add(1)
	go func() {
		//	defer wg.Done()
		for i := 0; i < numPackets; i++ {
			dataChannel <- packet
		}
	}()

	// wait until data was written and consumed
	time.Sleep(time.Millisecond * 100)
	client.Close()
	close(errorChannel)
	wg.Wait()

	// THEN
	assert.Len(t, errors, numPackets)
	assert.True(t, client.Errored)
}

func TestReadWorker(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := &Client{
		conn:                   mockedDialer,
		responseNotifier:       &sync.Map{},
		responseStatusNotifier: &sync.Map{},
		results:                &sync.Map{},
	}
	errorChannel := make(chan error, 1)
	quitChannel := make(chan struct{})
	response := Response{RequestID: "ABCDEF", Status: Status{Code: statusSuccess}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// WHEN
	mockedDialer.EXPECT().Read().Return(1, packet, nil).AnyTimes()
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, quitChannel)
	client.Close()

	// THEN
	assert.Empty(t, errorChannel)
	assert.False(t, client.Errored)
	assert.NotEmpty(t, client.results)
	_, ok := client.results.Load(response.RequestID)
	assert.True(t, ok)
}

func TestReadWorkerFailOnInvalidResponse(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := &Client{
		conn:                   mockedDialer,
		responseNotifier:       &sync.Map{},
		responseStatusNotifier: &sync.Map{},
		results:                &sync.Map{},
	}
	errorChannel := make(chan error, 1)
	quitChannel := make(chan struct{})
	response := Response{RequestID: "ABCDEF", Status: Status{Code: statusMalformedRequest}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// WHEN
	mockedDialer.EXPECT().Read().Return(1, packet, nil).AnyTimes()
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.True(t, client.Errored)
}

func TestReadWorkerFailOnInvalidFrame(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := &Client{
		conn:                   mockedDialer,
		responseNotifier:       &sync.Map{},
		responseStatusNotifier: &sync.Map{},
		results:                &sync.Map{},
	}
	errorChannel := make(chan error, 1)
	quitChannel := make(chan struct{})

	// WHEN
	mockedDialer.EXPECT().Read().Return(-1, nil, nil).AnyTimes()
	mockedDialer.EXPECT().Close().DoAndReturn(func() {
		close(quitChannel)
	}).Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.True(t, client.Errored)
}
