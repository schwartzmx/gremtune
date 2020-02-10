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

	mockedDialer.EXPECT().IsConnected().Return(true)

	responseChannel := make(chan interfaces.AsyncResponse)

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
	response := interfaces.Response{RequestID: req.RequestID, Status: interfaces.Status{Code: interfaces.StatusSuccess}}
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

	mockedDialer.EXPECT().IsConnected().Return(true)

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
	response := interfaces.Response{RequestID: req.RequestID, Status: interfaces.Status{Code: interfaces.StatusSuccess}}
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

	mockedDialer.EXPECT().IsConnected().Return(false)

	resp, err := client.Execute("g.V()")
	assert.Empty(t, resp)
	assert.Error(t, err)
}

func TestValidateCredentials(t *testing.T) {
	assert.Error(t, validateCredentials(auth{}))
	assert.Error(t, validateCredentials(auth{username: "Hans"}))
	assert.NoError(t, validateCredentials(auth{username: "Hans", password: "PW"}))
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
	assert.NotNil(t, client.results)
	assert.NotNil(t, client.responseNotifier)
	assert.NotNil(t, client.responseStatusNotifier)
	assert.False(t, client.HadError())
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
	assert.True(t, client.HadError())
}

func TestPingWorker(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer, PingInterval(time.Millisecond*100))
	errorChannel := make(chan error, 5)

	// WHEN
	mockedDialer.EXPECT().Ping().Return(nil)
	mockedDialer.EXPECT().Ping().Return(fmt.Errorf("Error")).AnyTimes()
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.pingWorker(errorChannel, client.quitChannel)

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
	client := newClient(mockedDialer)
	client.requests = dataChannel
	errorChannel := make(chan error)

	wg := sync.WaitGroup{}
	packet := []byte("ABCDEFG")
	numPackets := 10

	// WHEN
	mockedDialer.EXPECT().Write(packet).Return(nil).Times(numPackets)
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.writeWorker(errorChannel, client.quitChannel)

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
	assert.False(t, client.HadError())
}

func TestWriteWorkerFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	dataChannel := make(chan []byte, 11)
	client := newClient(mockedDialer)
	client.requests = dataChannel

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

	packet := []byte("ABCDEFG")
	numPackets := 10

	// WHEN
	mockedDialer.EXPECT().Write(packet).Return(fmt.Errorf("Write failed")).Times(numPackets)
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.writeWorker(errorChannel, client.quitChannel)

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
	assert.True(t, client.HadError())
}

func TestReadWorker(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	errorChannel := make(chan error, 1)
	response := interfaces.Response{RequestID: "ABCDEF", Status: interfaces.Status{Code: interfaces.StatusSuccess}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// WHEN
	mockedDialer.EXPECT().Read().Return(1, packet, nil).AnyTimes()
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.Empty(t, errorChannel)
	assert.False(t, client.HadError())
	assert.NotEmpty(t, client.results)
	_, ok := client.results.Load(response.RequestID)
	assert.True(t, ok)
}

func TestReadWorkerFailOnInvalidResponse(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	errorChannel := make(chan error, 1)
	response := interfaces.Response{RequestID: "ABCDEF", Status: interfaces.Status{Code: interfaces.StatusMalformedRequest}}
	packet, err := json.Marshal(response)
	require.NoError(t, err)

	// WHEN
	mockedDialer.EXPECT().Read().Return(1, packet, nil).AnyTimes()
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.True(t, client.HadError())
}

func TestReadWorkerFailOnInvalidFrame(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	errorChannel := make(chan error, 1)

	// WHEN
	mockedDialer.EXPECT().Read().Return(-1, nil, nil).AnyTimes()
	mockedDialer.EXPECT().Close().Return(nil)

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.True(t, client.HadError())
}
