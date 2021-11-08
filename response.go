package gremcos

import (
	"encoding/json"
	"fmt"

	"github.com/supplyon/gremcos/interfaces"
)

func (c *client) handleResponse(msg []byte) error {
	resp, err := marshalResponse(msg)

	// ignore the error here in case the response status code tells that an authentication is needed
	if resp.Status.Code == interfaces.StatusAuthenticate { //Server request authentication
		return c.authenticate(resp.RequestID)
	}

	c.saveResponse(resp, err)
	return err
}

// marshalResponse creates a response struct for every incoming response for further manipulation
func marshalResponse(msg []byte) (interfaces.Response, error) {
	resp := interfaces.Response{}
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return resp, err
	}

	err = extractError(resp)
	return resp, err
}

// saveResponse makes the response available for retrieval by the requester. Mutexes are used for thread safety.
func (c *client) saveResponse(resp interfaces.Response, err error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	var container []interface{}
	existingData, ok := c.results.Load(resp.RequestID) // Retrieve old data container (for requests with multiple responses)
	if ok {
		container = existingData.([]interface{})
	}
	newdata := append(container, resp)       // Create new data container with new data
	c.results.Store(resp.RequestID, newdata) // Add new data to buffer for future retrieval

	// obtain or create (if needed) the error notification channel for the currently active response
	respNotifier, _ := c.responseNotifier.LoadOrStore(resp.RequestID, newSafeCloseErrorChannel(1))
	respNotifierChannel := respNotifier.(*safeCloseErrorChannel)

	// obtain or create (if needed) the status notification channel for the currently active response
	responseStatusNotifier, _ := c.responseStatusNotifier.LoadOrStore(resp.RequestID, newSafeCloseIntChannel(1))
	responseStatusNotifierChannel := responseStatusNotifier.(*safeCloseIntChannel)

	// FIXME: This looks weird. the status code of the current response is only posted to the responseStatusNotifier channel
	// if there is space left on the channel. If not then the status is just silently not posted (ignored).
	if cap(responseStatusNotifierChannel.c) > len(responseStatusNotifierChannel.c) {
		// Channel is not full so adding the response status to the channel else it will cause the method to wait till the response is read by requester
		responseStatusNotifierChannel.c <- resp.Status.Code
	}

	// post an error in case it is not a partial messsage.
	// note that here the given error can be nil.
	// this is the good case that just completes the retrieval of the response
	if resp.Status.Code != interfaces.StatusPartialContent {
		respNotifierChannel.c <- err
	}
}

// retrieveResponseAsync retrieves the response saved by saveResponse and send the retrieved repose to the channel .
func (c *client) retrieveResponseAsync(id string, responseChannel chan interfaces.AsyncResponse) {
	var responseProcessedIndex int
	responseNotifier, _ := c.responseNotifier.Load(id)
	responseNotifierChannel := responseNotifier.(*safeCloseErrorChannel)
	responseStatusNotifier, _ := c.responseStatusNotifier.Load(id)
	responseStatusNotifierChannel := responseStatusNotifier.(*safeCloseIntChannel)

	for status := range responseStatusNotifierChannel.c {
		_ = status

		// this block retrieves all but the last of the partial responses
		// and sends it to the response channel
		if dataI, ok := c.results.Load(id); ok {
			d := dataI.([]interface{})
			// Only retrieve all but one from the partial responses saved in results Map that are not sent to responseChannel
			for i := responseProcessedIndex; i < len(d)-1; i++ {
				responseProcessedIndex++
				var asyncResponse interfaces.AsyncResponse = interfaces.AsyncResponse{}
				asyncResponse.Response = d[i].(interfaces.Response)
				// Send the Partial response object to the responseChannel
				responseChannel <- asyncResponse
			}
		}

		// Checks to see If there was an Error or full response that has been provided by cosmos
		// If not, then continue with consuming the other partial messages
		if len(responseNotifierChannel.c) <= 0 {
			continue
		}

		//Checks to see If there was an Error or will get nil when final response has been provided by cosmos
		err := <-responseNotifierChannel.c

		if dataI, ok := c.results.Load(id); ok {
			d := dataI.([]interface{})
			// Retrieve all the partial responses that are not sent to responseChannel
			for i := responseProcessedIndex; i < len(d); i++ {
				responseProcessedIndex++
				asyncResponse := interfaces.AsyncResponse{}
				asyncResponse.Response = d[i].(interfaces.Response)
				//when final partial response it sent it also sends the error message if there was an error on the last partial response retrival
				if responseProcessedIndex == len(d) && err != nil {
					asyncResponse.ErrorMessage = err.Error()
				}
				// Send the Partial response object to the responseChannel
				responseChannel <- asyncResponse
			}
		}
		// All the Partial response object including the final one has been sent to the responseChannel
		break
	}

	// All the Partial response object including the final one has been sent to the responseChannel
	// so closing responseStatusNotifierChannel, responseNotifierChannel, responseChannel and removing all the repose stored
	responseStatusNotifierChannel.Close()
	responseNotifierChannel.Close()
	c.responseNotifier.Delete(id)
	c.responseStatusNotifier.Delete(id)
	c.deleteResponse(id)
	close(responseChannel)
}

func emptyIfNilOrError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// retrieveResponse retrieves the response saved by saveResponse.
func (c *client) retrieveResponse(id string) ([]interfaces.Response, error) {

	var responseErrorChannel *safeCloseErrorChannel
	var responseStatusNotifierChannel *safeCloseIntChannel

	// ensure that the cleanup is done in any case
	defer func() {
		if responseErrorChannel != nil {
			responseErrorChannel.Close()
		}
		if responseStatusNotifierChannel != nil {
			responseStatusNotifierChannel.Close()
		}
		c.responseNotifier.Delete(id)
		c.responseStatusNotifier.Delete(id)
		c.deleteResponse(id)
	}()

	responseErrorChannelUntyped, ok := c.responseNotifier.Load(id)
	if !ok {
		return nil, fmt.Errorf("Response with id %s not found", id)
	}
	responseErrorChannel = responseErrorChannelUntyped.(*safeCloseErrorChannel)

	responseStatusNotifierUntyped, ok := c.responseStatusNotifier.Load(id)
	if !ok {
		return nil, fmt.Errorf("Response with id %s not found", id)
	}
	responseStatusNotifierChannel = responseStatusNotifierUntyped.(*safeCloseIntChannel)

	err := <-responseErrorChannel.c
	// Hint: Don't return here immediately in case the obtained error is != nil.
	// We don't want to loose the responses obtained so far, especially the
	// data stored in the attribute map of each response is useful.
	// For example the response contains the request charge for this request.

	dataI, ok := c.results.Load(id)
	if !ok {
		lastErr := c.LastError() // add more information to find out why there was no result
		return nil, fmt.Errorf("no result for response with id %s found, err='%s'", id, emptyIfNilOrError(lastErr))
	}

	// cast the given data into an array of Responses
	d := dataI.([]interface{})
	data := make([]interfaces.Response, len(d))
	for i := range d {
		data[i] = d[i].(interfaces.Response)
	}

	return data, err
}

// deleteRespones deletes the response from the container. Used for cleanup purposes by requester.
func (c *client) deleteResponse(id string) {
	c.results.Delete(id)
}
