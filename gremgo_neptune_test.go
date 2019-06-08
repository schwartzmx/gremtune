package gremgo

import (
	"encoding/json"
	"testing"
)

func init() {
	InitGremlinClients()
}

func truncateData(t *testing.T) {
	t.Logf("Removing all data from gremlin server")
	r, err := g.Execute(`g.V().drop().iterate()`)
	t.Logf("Removed all vertices, response: %v+ \n err: %s", r, err)
	if err != nil {
		t.Fatal(err)
	}
}

func seedData(t *testing.T) {
	truncateData(t)
	t.Logf("Seeding data...")
	r, err := g.Execute(`
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
	t.Logf("Added two vertices and one edge, response: %v+ \n err: %s", r, err)
	if err != nil {
		t.Fatal(err)
	}
}

type nodeLabels struct {
	Label []string `json:"@value"`
}

func TestExecute(t *testing.T) {
	seedData(t)
	r, err := g.Execute(`g.V('1234').label()`)
	t.Logf("Execute get vertex, response: %s \n err: %s", r[0].Result.Data, err)
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

func TestExecuteWithBindings(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteWithBindings(
		`g.V(x).label()`,
		map[string]string{"x": "1234"},
		map[string]string{},
	)
	t.Logf("Execute with bindings get vertex, response: %s \n err: %s", r[0].Result.Data, err)
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

func TestExecuteFile(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteFile("scripts/test.groovy")
	t.Logf("ExecuteFile get vertex, response: %s \n err: %s", r[0].Result.Data, err)
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

func TestExecuteFileWithBindings(t *testing.T) {
	seedData(t)
	r, err := g.ExecuteFileWithBindings(
		"scripts/test-wbindings.groovy",
		map[string]string{"x": "2145"},
		map[string]string{},
	)
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

func TestPoolExecute(t *testing.T) {
	seedData(t)
	r, err := gp.Execute(`g.V('1234').label()`)
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
