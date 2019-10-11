package gremtune

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	InitGremlinClients()
}

type BulkResponse struct {
	Type  string `json:"type"`
	Value []struct {
		Type  string        `json:"type"`
		Value []interface{} `json:"value"`
	} `json:"value"`
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

type nodeLabels struct {
	Label []string `json:"@value"`
}

func TestExecute(t *testing.T) {
	seedData(t)
	r, err := g.Execute(`g.V('1234').label()`)
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		t.Logf("Execute get vertex, response: %v \n err: %v", string(r[0].Result.Data), err)
		nl := new(nodeLabels)
		err = json.Unmarshal(r[0].Result.Data, &nl)
		if len(nl.Label) != 1 {
			t.Errorf("There should only be 1 node label, got: %v+", nl)
		}
		expected := "Phil"
		got := nl.Label[0]
		if nl.Label[0] != expected {
			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
		}
	}
}

func TestExecuteBulkData(t *testing.T) {
	seedBulkData(t)
	defer truncateBulkData(t)
	start := time.Now()
	r, err := g.Execute(`g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)`)
	log.Println(fmt.Sprintf("Execution time it took to execute query %s", time.Since(start)))
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		nl := new(BulkResponse)
		datastr := strings.Replace(string(r[0].Result.Data), "@type", "type", -1)
		datastr = strings.Replace(datastr, "@value", "value", -1)
		err = json.Unmarshal([]byte(datastr), &nl)
		if len(nl.Value) != 64 {
			t.Errorf("There should only be 64 value, got: %v+", len(nl.Value))
		}
		if len(r) != 10 {
			t.Errorf("There should only be 10 value, got: %v+", len(r))
		}
	}
}

func TestExecuteBulkDataAsync(t *testing.T) {
	seedBulkData(t)
	start := time.Now()
	responseChannel := make(chan AsyncResponse, 2)
	err := g.ExecuteAsync(`g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)`, responseChannel)
	log.Println(fmt.Sprintf("Time it took to execute query %s", time.Since(start)))
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		count := 0
		asyncResponse := AsyncResponse{}
		start = time.Now()
		for asyncResponse = range responseChannel {
			log.Println(fmt.Sprintf("Time it took to get async response: %s response status: %v (206 means partial and 200 final response)", time.Since(start), asyncResponse.Response.Status.Code))
			count++
			nl := new(BulkResponse)
			datastr := strings.Replace(string(asyncResponse.Response.Result.Data), "@type", "type", -1)
			datastr = strings.Replace(datastr, "@value", "value", -1)
			err = json.Unmarshal([]byte(datastr), &nl)
			if len(nl.Value) != 64 {
				t.Errorf("There should only be 64 value, got: %v+", len(nl.Value))
			}
			start = time.Now()
		}
		if count != 10 {
			t.Errorf("There should only be 10 value, got: %v+", count)
		}
	}
}

func TestExecuteWithBindings(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteWithBindings(
		"g.V(x).label()",
		map[string]string{"x": "1234"},
		map[string]string{},
	)
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		t.Logf("Execute with bindings get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
		nl := new(nodeLabels)
		err = json.Unmarshal(r[0].Result.Data, &nl)
		if len(nl.Label) != 1 {
			t.Errorf("There should only be 1 node label, got: %v+", nl)
		}
		expected := "Phil"
		got := nl.Label[0]
		if nl.Label[0] != expected {
			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
		}
	}
}

func TestExecuteFile(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteFile("scripts/test.groovy")
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		t.Logf("ExecuteFile get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
		nl := new(nodeLabels)
		err = json.Unmarshal(r[0].Result.Data, &nl)
		if len(nl.Label) != 1 {
			t.Errorf("There should only be 1 node label, got: %v+", nl)
		}
		expected := "Vincent"
		got := nl.Label[0]
		if nl.Label[0] != expected {
			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
		}
	}
}

func TestExecuteFileWithBindings(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteFileWithBindings(
		"scripts/test-wbindings.groovy",
		map[string]string{"x": "2145"},
		map[string]string{},
	)
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		t.Logf("ExecuteFileWithBindings get vertex, response: %s \n err: %s", r[0].Result.Data, err)
		nl := new(nodeLabels)
		err = json.Unmarshal(r[0].Result.Data, &nl)
		if len(nl.Label) != 1 {
			t.Errorf("There should only be 1 node label, got: %v+", nl)
		}
		expected := "Vincent"
		got := nl.Label[0]
		if nl.Label[0] != expected {
			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
		}
	}
}

func TestPoolExecute(t *testing.T) {
	seedData(t)
	r, err := gp.Execute(`g.V('1234').label()`)
	if err != nil {
		t.Errorf("Unexpected error returned from server err: %v", err.Error())
	} else {
		t.Logf("PoolExecute get vertex, response: %s \n err: %s", r[0].Result.Data, err)
		nl := new(nodeLabels)
		err = json.Unmarshal(r[0].Result.Data, &nl)
		if len(nl.Label) != 1 {
			t.Errorf("There should only be 1 node label, got: %v+", nl)
		}
		expected := "Phil"
		got := nl.Label[0]
		if nl.Label[0] != expected {
			t.Errorf("Unexpected label returned,  expected: %s got: %s", expected, got)
		}
	}
}
