package gremcos

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
)

func TestNew(t *testing.T) {
	// GIVEN
	idleTimeout := time.Second * 12
	maxActiveConnections := 10
	username := "abcd"
	password := "xyz"

	// WHEN
	cosmos, err := New("ws://host",
		ConnectionIdleTimeout(idleTimeout),
		NumMaxActiveConnections(maxActiveConnections),
		WithAuth(username, password),
	)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	assert.Equal(t, idleTimeout, cosmos.connectionIdleTimeout)
	assert.Equal(t, maxActiveConnections, cosmos.numMaxActiveConnections)
	assert.Equal(t, username, cosmos.username)
	assert.Equal(t, password, cosmos.password)
}

func TestStop(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host")
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor
	mockedQueryExecutor.EXPECT().Close().Return(nil)

	// WHEN
	err = cosmos.Stop()

	// THEN
	assert.NoError(t, err)
}

func TestIsHealthy(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host")
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor

	// WHEN -- connected --> healthy
	mockedQueryExecutor.EXPECT().Ping().Return(nil)
	healthyWhenConnected := cosmos.IsHealthy()

	// WHEN -- not connected --> not healthy
	mockedQueryExecutor.EXPECT().Ping().Return(fmt.Errorf("Not connected"))
	healthyWhenNotConnected := cosmos.IsHealthy()

	// THEN
	assert.NoError(t, healthyWhenConnected)
	assert.Error(t, healthyWhenNotConnected)
}

func TestParseAttributeMap(t *testing.T) {
	// GIVEN
	attributeMap := map[string]interface{}{
		"x-ms-status-code":          429,
		"x-ms-substatus-code":       3200,
		"x-ms-request-charge":       1234.56,
		"x-ms-total-request-charge": 78910.11,
		"x-ms-server-time-ms":       11.22,
		"x-ms-total-server-time-ms": 333.444,
		"x-ms-activity-id":          "fdd08592-abcd-efgh-ijkl-97d35c2dda52",
		"x-ms-retry-after-ms":       "00:00:02.345",
		"x-ms-source":               "Microsoft.Azure.Documents.Client",
	}

	// WHEN
	responseInfo, err := parseAttributeMap(attributeMap)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, 429, responseInfo.statusCode)
	assert.NotEmpty(t, responseInfo.statusDescription)
	assert.Equal(t, 3200, responseInfo.subStatusCode)
	assert.Equal(t, float32(1234.56), responseInfo.requestCharge)
	assert.Equal(t, float32(78910.11), responseInfo.requestChargeTotal)
	assert.Equal(t, time.Microsecond*11220, responseInfo.serverTime)
	assert.Equal(t, time.Microsecond*333444, responseInfo.serverTimeTotal)
	assert.Equal(t, "fdd08592-abcd-efgh-ijkl-97d35c2dda52", responseInfo.activityID)
	assert.Equal(t, time.Millisecond*2345, responseInfo.retryAfter)
	assert.Equal(t, "Microsoft.Azure.Documents.Client", responseInfo.source)
}

func TestParseAttributeMapFail(t *testing.T) {
	// GIVEN
	attributeMap := map[string]interface{}{
		"x-ms-status-code": "invalid",
	}

	// WHEN
	responseInfo, err := parseAttributeMap(attributeMap)

	// THEN
	require.Error(t, err)

	// GIVEN
	attributeMap = map[string]interface{}{
		"x-ms-status-code":          429,
		"x-ms-substatus-code":       "invalid",
		"x-ms-request-charge":       "invalid",
		"x-ms-total-request-charge": "invalid",
		"x-ms-server-time-ms":       "invalid",
		"x-ms-total-server-time-ms": "invalid",
		"x-ms-retry-after-ms":       "invalid",
	}

	// WHEN
	responseInfo, err = parseAttributeMap(attributeMap)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, 429, responseInfo.statusCode)
	assert.NotEmpty(t, responseInfo.statusDescription)
	assert.Equal(t, 0, responseInfo.subStatusCode)
	assert.Equal(t, float32(0), responseInfo.requestCharge)
	assert.Equal(t, float32(0), responseInfo.requestChargeTotal)
	assert.Equal(t, time.Microsecond*0, responseInfo.serverTime)
	assert.Equal(t, time.Microsecond*0, responseInfo.serverTimeTotal)
	assert.Equal(t, time.Millisecond*0, responseInfo.retryAfter)
}

func TestStatusCodeToDescription(t *testing.T) {
	// GIVEN
	code := 429

	// WHEN
	desc := statusCodeToDescription(code)

	// THEN
	assert.Contains(t, desc, "throttled")
	assert.NotContains(t, desc, "unknown")

	// GIVEN -- not found
	code = 12345

	// WHEN
	desc = statusCodeToDescription(code)

	// THEN
	assert.Contains(t, desc, "unknown")
}
