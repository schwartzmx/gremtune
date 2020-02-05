package gremtune

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

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
	log.Println("Removing all data from gremlin server started...")
	_, err := g.Execute(`g.V('1234').drop()`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Execute(`g.V('2145').drop()`)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Removing all data from gremlin server completed...")
}

func seedData(t *testing.T) {
	truncateData(t)
	log.Println("Seeding data started...")
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
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Seeding data completed...")
}

func truncateBulkData(t *testing.T) {
	log.Println("Removing bulk data from gremlin server strated...")
	_, err := g.Execute(`g.V().hasLabel('EmployeeBulkData').drop().iterate()`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Execute(`g.V().hasLabel('EmployerBulkData').drop()`)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Removing bulk data from gremlin server completed...")
}

func seedBulkData(t *testing.T) {
	truncateBulkData(t)
	log.Println("Seeding bulk data started...")

	_, err := g.Execute(`
		g.addV('EmployerBulkData').property(id, '1234567890').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree')
	`)
	if err != nil {
		t.Fatal(err)
	}

	for i := 9001; i < 9641; i++ {
		_, err = g.Execute("g.addV('EmployeeBulkData').property(id, '" + strconv.Itoa(i) + "').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree').as('y').addE('employes').from(V('1234567890')).to('y')")
		if err != nil {
			t.Fatal(err)
		}
	}
	log.Println("Seeding bulk data completed...")
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

	t.Logf("Execute get vertex, response: %v \n err: %v", string(r[0].Result.Data), err)

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

//
//func TestExecuteBulkDataAsync(t *testing.T) {
//	seedBulkData(t)
//	start := time.Now()
//	responseChannel := make(chan AsyncResponse, 2)
//	err := g.ExecuteAsync(`g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)`, responseChannel)
//	log.Println(fmt.Sprintf("Time it took to execute query %s", time.Since(start)))
//	if err != nil {
//		t.Errorf("Unexpected error returned from server err: %v", err.Error())
//	} else {
//		count := 0
//		asyncResponse := AsyncResponse{}
//		start = time.Now()
//		for asyncResponse = range responseChannel {
//			log.Println(fmt.Sprintf("Time it took to get async response: %s response status: %v (206 means partial and 200 final response)", time.Since(start), asyncResponse.Response.Status.Code))
//			count++
//			nl := new(BulkResponse)
//			datastr := strings.Replace(string(asyncResponse.Response.Result.Data), "@type", "type", -1)
//			datastr = strings.Replace(datastr, "@value", "value", -1)
//			err = json.Unmarshal([]byte(datastr), &nl)
//			if len(nl.Value) != 64 {
//				t.Errorf("There should only be 64 value, got: %v+", len(nl.Value))
//			}
//			start = time.Now()
//		}
//		if count != 10 {
//			t.Errorf("There should only be 10 value, got: %v+", count)
//		}
//	}
//}
//
//func TestExecuteWithBindings(t *testing.T) {
//	seedData(t)
//	r, err := g.ExecuteWithBindings(
//		"g.V(x).label()",
//		map[string]string{"x": "1234"},
//		map[string]string{},
//	)
//	if err != nil {
//		t.Errorf("Unexpected error returned from server err: %v", err.Error())
//	} else {
//		t.Logf("Execute with bindings get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
//		nl := new(nodeLabels)
//		err = json.Unmarshal(r[0].Result.Data, &nl)
//		if len(nl.Label) != 1 {
//			t.Errorf("There should only be 1 node label, got: %v+", nl)
//		}
//		expected := "Phil"
//		got := nl.Label[0]
//		if nl.Label[0] != expected {
//			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
//		}
//	}
//}
//
//func TestExecuteFile(t *testing.T) {
//	seedData(t)
//	r, err := g.ExecuteFile("scripts/test.groovy")
//	if err != nil {
//		t.Errorf("Unexpected error returned from server err: %v", err.Error())
//	} else {
//		t.Logf("ExecuteFile get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
//		nl := new(nodeLabels)
//		err = json.Unmarshal(r[0].Result.Data, &nl)
//		if len(nl.Label) != 1 {
//			t.Errorf("There should only be 1 node label, got: %v+", nl)
//		}
//		expected := "Vincent"
//		got := nl.Label[0]
//		if nl.Label[0] != expected {
//			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
//		}
//	}
//}
//
//func TestExecuteFileWithBindings(t *testing.T) {
//	seedData(t)
//	r, err := g.ExecuteFileWithBindings(
//		"scripts/test-wbindings.groovy",
//		map[string]string{"x": "2145"},
//		map[string]string{},
//	)
//	if err != nil {
//		t.Errorf("Unexpected error returned from server err: %v", err.Error())
//	} else {
//		t.Logf("ExecuteFileWithBindings get vertex, response: %s \n err: %s", r[0].Result.Data, err)
//		nl := new(nodeLabels)
//		err = json.Unmarshal(r[0].Result.Data, &nl)
//		if len(nl.Label) != 1 {
//			t.Errorf("There should only be 1 node label, got: %v+", nl)
//		}
//		expected := "Vincent"
//		got := nl.Label[0]
//		if nl.Label[0] != expected {
//			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
//		}
//	}
//}
//
//func TestPoolExecute(t *testing.T) {
//	seedData(t)
//	r, err := gp.Execute(`g.V('1234').label()`)
//	if err != nil {
//		t.Errorf("Unexpected error returned from server err: %v", err.Error())
//	} else {
//		t.Logf("PoolExecute get vertex, response: %s \n err: %s", r[0].Result.Data, err)
//		nl := new(nodeLabels)
//		err = json.Unmarshal(r[0].Result.Data, &nl)
//		if len(nl.Label) != 1 {
//			t.Errorf("There should only be 1 node label, got: %v+", nl)
//		}
//		expected := "Phil"
//		got := nl.Label[0]
//		if nl.Label[0] != expected {
//			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
//		}
//	}
//}
//
