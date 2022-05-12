package gremcos

import (
	"encoding/json"
	"fmt"
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
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
	assert.Error(t, validateCredentials("", ""))
	assert.Error(t, validateCredentials("Hans", ""))
	assert.NoError(t, validateCredentials("Hans", "PW"))
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
	assert.Nil(t, client.LastError())
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
	once := sync.Once{}

	// WHEN
	mockedDialer.EXPECT().Connect().Return(nil)
	mockedDialer.EXPECT().Read().Return(1, nil, fmt.Errorf("Read failed")).AnyTimes()
	mockedDialer.EXPECT().Close().Do(func() {
		once.Do(func() {
			close(quitChannel)
		})
	}).Return(nil).AnyTimes()

	client, err := Dial(mockedDialer, errorChannel)
	require.NotNil(t, client)
	require.NoError(t, err)
	err = client.Close()
	close(errorChannel)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, client.LastError())
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
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

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
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

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
	assert.Nil(t, client.LastError())
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
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

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
	assert.NotNil(t, client.LastError())
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
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.Empty(t, errorChannel)
	assert.Nil(t, client.LastError())
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
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.NotNil(t, client.LastError())
}

func TestReadWorkerFailOnInvalidFrame(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	errorChannel := make(chan error, 1)

	// WHEN
	mockedDialer.EXPECT().Read().Return(-1, nil, fmt.Errorf("failure")).AnyTimes()
	mockedDialer.EXPECT().Close().Return(nil).AnyTimes()

	client.wg.Add(1)
	go client.readWorker(errorChannel, client.quitChannel)
	client.Close()

	// THEN
	assert.NotEmpty(t, errorChannel)
	assert.NotNil(t, client.LastError())
}

func TestForceCloseOnClosedChannelPanic(t *testing.T) {
	defer goleak.VerifyNone(t)
	// This test was added to reproduce https://github.com/supplyon/gremcos/issues/29

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	mockedDialer.EXPECT().IsConnected().Return(true)
	mockedDialer.EXPECT().Close().Return(nil)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := client.Execute("g.V()")
		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	}()

	// catch the request that should be sent over the wire
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

	// Immediately close the client even while there are still requests ongoing
	// in general that is not a good idea to do that, however this should not result in a
	// panic as described at https://github.com/supplyon/gremcos/issues/29.
	client.Close()

	// wait until the execution has been completed
	wg.Wait()
}

type credProvider struct {
	uname string
	pwd   string
}

func (cp credProvider) Password() (string, error) {
	if cp.pwd == "err" {
		return "", fmt.Errorf("password is missing")
	}
	return cp.pwd, nil
}

func (cp credProvider) Username() (string, error) {
	if cp.uname == "err" {
		return "", fmt.Errorf("username is missing")
	}
	return cp.uname, nil
}

func TestAuthenticate(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)
	client.credentialProvider = credProvider{uname: "username", pwd: "password"}

	// WHEN
	err := client.authenticate("reqID")

	// THEN
	assert.NoError(t, err)
}

func TestAuthenticate_Fail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	client := newClient(mockedDialer)

	// WHEN + THEN uname returned error
	client.credentialProvider = credProvider{uname: "err", pwd: "err"}
	err := client.authenticate("reqID")
	assert.Error(t, err)

	// WHEN + THEN pwd returned error
	client.credentialProvider = credProvider{uname: "username", pwd: "err"}
	err = client.authenticate("reqID")
	assert.Error(t, err)

	// WHEN + THEN uname missing
	client.credentialProvider = credProvider{uname: ""}
	err = client.authenticate("reqID")
	assert.Error(t, err)
}

func TestCloseClient(t *testing.T) {
	defer goleak.VerifyNone(t)

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	mockedDialer.EXPECT().Connect().Return(nil)
	mockedDialer.EXPECT().Read()
	mockedDialer.EXPECT().Close().Return(nil)
	errChan := make(chan error, 100)
	defer close(errChan)

	client, err := Dial(mockedDialer, errChan)
	require.NoError(t, err)

	// WHEN
	closeErr := client.Close()

	// THEN
	assert.NoError(t, closeErr)
}

func TestConcurrentWriteAndClose(t *testing.T){
	defer goleak.VerifyNone(t)

	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	mockedDialer.EXPECT().IsConnected().AnyTimes().Return(true)
	mockedDialer.EXPECT().Connect().Return(nil)
	mockedDialer.EXPECT().Read().AnyTimes().DoAndReturn(func() (int, []byte, error){
		time.Sleep(time.Millisecond*2)

		return 0,[]byte{},nil
	})
	// not synced because the functions write and close should be synced
	sending := false

	mockedDialer.EXPECT().Write(gomock.Any()).MinTimes(1).Do(func(data interface{}) error {
		require.False(t,sending)
		sending = true
		time.Sleep(time.Millisecond*500)
		sending = false
		return nil
	})
	mockedDialer.EXPECT().Close().MinTimes(1).Do(func() error {
		require.False(t,sending)
		sending = true
		time.Sleep(time.Millisecond*1)
		sending = false
		return nil
	})

	errChan := make(chan error,100)
	client,err := Dial(mockedDialer,errChan)
	require.NoError(t, err)


	go func() {
			_, _ = client.Execute("g.V()")
	}()

	time.Sleep(time.Millisecond*50)
	client.Close()

	time.Sleep(time.Millisecond*50)
	close(errChan)
}