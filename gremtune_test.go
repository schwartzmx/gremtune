package gremtune

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	InitGremlinClients()
}

// One entry returned from gremlin looks like this:
//{"id":{
//	"@type":"g:Int64",
//	"@value":9147
//	},
//	"label":"EmployeeBulkData",
//	"source":["tree"],
//	"timestamp":["2018-07-01T13:37:45-05:00"]
//}
type bulkResponseEntry struct {
	ID        id       `json:"id,omitempty"`
	Label     string   `json:"label,omitempty"`
	Source    []string `json:"source,omitempty"`
	Timestamp []string `json:"timestamp,omitempty"`
}
type id struct {
	Type  string `json:"@type,omitempty"`
	Value int    `json:"@value,omitempty"`
}

func truncateData(t *testing.T) {
	t.Log("Removing all data from gremlin server started...")
	_, err := g.Execute(`g.V('1234').drop()`)
	require.NoError(t, err)

	_, err = g.Execute(`g.V('2145').drop()`)
	require.NoError(t, err)
	t.Log("Removing all data from gremlin server completed...")
}

func seedData(t *testing.T) {
	truncateData(t)
	t.Log("Seeding data started...")
	_, err := g.Execute(`
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

func truncateBulkData(t *testing.T) {
	t.Log("Removing bulk data from gremlin server strated...")
	_, err := g.Execute(`g.V().hasLabel('EmployeeBulkData').drop().iterate()`)
	require.NoError(t, err)

	_, err = g.Execute(`g.V().hasLabel('EmployerBulkData').drop()`)
	require.NoError(t, err)
	t.Log("Removing bulk data from gremlin server completed...")
}

func seedBulkData(t *testing.T) {
	truncateBulkData(t)
	t.Log("Seeding bulk data started...")

	_, err := g.Execute("g.addV('EmployerBulkData').property(id, '1234567890').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree')")
	require.NoError(t, err)

	for i := 9001; i < 9641; i++ {
		_, err = g.Execute("g.addV('EmployeeBulkData').property(id, '" + strconv.Itoa(i) + "').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree').as('y').addE('employes').from(V('1234567890')).to('y')")
		require.NoError(t, err)
	}
	t.Log("Seeding bulk data completed...")
}

type nodeLabels []string

func TestExecute_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())

	seedData(t)
	r, err := g.Execute("g.V('1234').label()")
	require.NoError(t, err, "Unexpected error from server")
	require.Len(t, r, 1)

	nl := nodeLabels{}
	err = json.Unmarshal(r[0].Result.Data, &nl)
	require.NoError(t, err, "Failed to unmarshall")

	assert.Len(t, nl, 1, "There should be only one node label")
	assert.Equal(t, "Phil", nl[0]) // see seedData()
}

func TestExecuteBulkData_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())
	seedBulkData(t)
	defer truncateBulkData(t)

	r, err := g.Execute("g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)")
	require.NoError(t, err, "Unexpected error from server")
	assert.Len(t, r, 10, "There should only be 10 responses")

	var nl []bulkResponseEntry
	err = json.Unmarshal([]byte(r[0].Result.Data), &nl)
	assert.NoError(t, err)
	assert.Len(t, nl, 64, "There should only be 64 values")
}

func TestExecuteBulkDataAsync_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())

	seedBulkData(t)
	defer truncateBulkData(t)
	responseChannel := make(chan AsyncResponse, 2)
	err := g.ExecuteAsync("g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)", responseChannel)
	require.NoError(t, err, "Unexpected error from server")

	count := 0
	asyncResponse := AsyncResponse{}
	start := time.Now()
	for asyncResponse = range responseChannel {
		t.Logf("Time it took to get async response: %s response status: %v (206 means partial and 200 final response)", time.Since(start), asyncResponse.Response.Status.Code)
		count++

		var nl []bulkResponseEntry
		err = json.Unmarshal(asyncResponse.Response.Result.Data, &nl)
		assert.NoError(t, err)
		assert.Len(t, nl, 64, "There should only be 64 values")
		start = time.Now()
	}
	assert.Equal(t, 10, count, "There should only be 10 values")
}

func TestExecuteWithBindings_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())

	seedData(t)
	r, err := g.ExecuteWithBindings(
		"g.V(x).label()",
		map[string]string{"x": "1234"},
		map[string]string{},
	)
	require.NoError(t, err, "Unexpected error from server")

	t.Logf("Execute with bindings get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	assert.NoError(t, err)
	assert.Len(t, nl, 1, "There should only be 1 node label")
	assert.Equal(t, "Phil", nl[0])
}

func TestExecuteFile_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())
	seedData(t)

	r, err := g.ExecuteFile("scripts/test.groovy")
	require.NoError(t, err, "Unexpected error from server")

	t.Logf("ExecuteFile get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)

	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	assert.NoError(t, err)
	assert.Len(t, nl, 1, "There should only be 1 node label")
	assert.Equal(t, "Vincent", nl[0])
}

func TestExecuteFileWithBindings_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())
	seedData(t)

	r, err := g.ExecuteFileWithBindings(
		"scripts/test-wbindings.groovy",
		map[string]string{"x": "2145"},
		map[string]string{},
	)
	require.NoError(t, err, "Unexpected error from server")
	t.Logf("ExecuteFileWithBindings get vertex, response: %s \n err: %s", r[0].Result.Data, err)

	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	assert.NoError(t, err)
	assert.Len(t, nl, 1, "There should only be 1 node label")
	assert.Equal(t, "Vincent", nl[0])
}

func TestPoolExecute_IT(t *testing.T) {
	// This is an integration test and belongs on data filled in
	// via seedBulkData()
	// As precondition a local gremlin-server has to run listening on port 8182

	// ensure that the used gremlin client instance is available
	require.NotNil(t, g)
	require.True(t, g.conn.IsConnected())
	seedData(t)

	r, err := gp.Execute(`g.V('1234').label()`)
	require.NoError(t, err, "Unexpected error from server")
	t.Logf("PoolExecute get vertex, response: %s \n err: %s", r[0].Result.Data, err)
	var nl nodeLabels

	err = json.Unmarshal(r[0].Result.Data, &nl)
	assert.NoError(t, err)
	assert.Len(t, nl, 1, "There should only be 1 node label")
	assert.Equal(t, "Phil", nl[0])
}
