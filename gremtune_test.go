package gremtune

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SuiteIntegrationTests struct {
	suite.Suite
	client             *Client
	clientErrorChannel chan error
	pool               *Pool
	poolErrorChannel   chan error
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

type nodeLabels []string

var failingErrorChannelConsumerFunc = func(errChan chan error, t *testing.T) {
	err := <-errChan
	if err == nil {
		return
	}
	t.Fatalf("Lost connection to the database: %s", err.Error())
}

func (s *SuiteIntegrationTests) TearDownSuite() {
	s.T().Log("TearDown SuiteIntegrationTests")
	close(s.clientErrorChannel)
	close(s.poolErrorChannel)
}

func (s *SuiteIntegrationTests) SetupSuite() {
	s.T().Log("Initialize SuiteIntegrationTests")
	s.T().Log("In order to run this suite a local gremlin server has to run and listen at port 8182")

	// create the error channels
	s.clientErrorChannel = make(chan error)
	s.poolErrorChannel = make(chan error)

	// create failing readers for those channels
	go failingErrorChannelConsumerFunc(s.clientErrorChannel, s.T())
	go failingErrorChannelConsumerFunc(s.poolErrorChannel, s.T())

	s.client = newTestClient(s.T(), s.clientErrorChannel)
	s.pool = newTestPool(s.T(), s.poolErrorChannel)

	// ensure preconditions
	s.Require().NotNil(s.client)
	s.Require().NotNil(s.pool)
	s.Require().True(s.client.conn.IsConnected())
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func Test_SuiteIntegrationTests(t *testing.T) {
	iTSuite := &SuiteIntegrationTests{}
	suite.Run(t, iTSuite)
}

func (s *SuiteIntegrationTests) truncateData() {
	s.T().Log("Removing all data from gremlin server started...")

	_, err := s.client.Execute(`g.V('1234').drop()`)
	s.Require().NoError(err)

	_, err = s.client.Execute(`g.V('2145').drop()`)
	s.Require().NoError(err)
	s.T().Log("Removing all data from gremlin server completed...")
}

func (s *SuiteIntegrationTests) seedData() {
	s.truncateData()
	s.T().Log("Seeding data started...")

	_, err := s.client.Execute(`
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
	s.Require().NoError(err)
	s.T().Log("Seeding data completed...")
}

func (s *SuiteIntegrationTests) truncateBulkData() {
	s.T().Log("Removing bulk data from gremlin server strated...")
	_, err := s.client.Execute(`g.V().hasLabel('EmployeeBulkData').drop().iterate()`)
	s.Require().NoError(err)

	_, err = s.client.Execute(`g.V().hasLabel('EmployerBulkData').drop()`)
	s.Require().NoError(err)
	s.T().Log("Removing bulk data from gremlin server completed...")
}

func (s *SuiteIntegrationTests) seedBulkData() {
	s.truncateBulkData()
	s.T().Log("Seeding bulk data started...")

	_, err := s.client.Execute("g.addV('EmployerBulkData').property(id, '1234567890').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree')")
	s.Require().NoError(err)

	for i := 9001; i < 9641; i++ {
		_, err = s.client.Execute("g.addV('EmployeeBulkData').property(id, '" + strconv.Itoa(i) + "').property('timestamp', '2018-07-01T13:37:45-05:00').property('source', 'tree').as('y').addE('employes').from(V('1234567890')).to('y')")
		s.Require().NoError(err)
	}
	s.T().Log("Seeding bulk data completed...")
}

func (s *SuiteIntegrationTests) TestExecute_IT() {

	s.seedData()
	r, err := s.client.Execute("g.V('1234').label()")
	s.Require().NoError(err, "Unexpected error from server")
	s.Require().Len(r, 1)

	labels := nodeLabels{}
	err = json.Unmarshal(r[0].Result.Data, &labels)
	s.Require().NoError(err, "Failed to unmarshall")

	s.Assert().Len(labels, 1, "There should be only one node label")
	s.Assert().Equal("Phil", labels[0]) // see seedData()
}

func (s *SuiteIntegrationTests) TestExecuteBulkData_IT() {
	s.seedBulkData()
	defer s.truncateBulkData()

	r, err := s.client.Execute("g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)")
	s.Require().NoError(err, "Unexpected error from server")
	s.Assert().Len(r, 10, "There should only be 10 responses")

	var nl []bulkResponseEntry
	err = json.Unmarshal([]byte(r[0].Result.Data), &nl)
	s.Assert().NoError(err)
	s.Assert().Len(nl, 64, "There should only be 64 values")
}

func (s *SuiteIntegrationTests) TestExecuteBulkDataAsync_IT() {

	s.seedBulkData()
	defer s.truncateBulkData()
	responseChannel := make(chan AsyncResponse, 2)
	err := s.client.ExecuteAsync("g.V().hasLabel('EmployerBulkData').both('employes').hasLabel('EmployeeBulkData').valueMap(true)", responseChannel)
	s.Require().NoError(err, "Unexpected error from server")

	count := 0
	asyncResponse := AsyncResponse{}
	start := time.Now()
	for asyncResponse = range responseChannel {
		s.T().Logf("Time it took to get async response: %s response status: %v (206 means partial and 200 final response)", time.Since(start), asyncResponse.Response.Status.Code)
		count++

		var nl []bulkResponseEntry
		err = json.Unmarshal(asyncResponse.Response.Result.Data, &nl)
		s.Assert().NoError(err)
		s.Assert().Len(nl, 64, "There should only be 64 values")
		start = time.Now()
	}
	s.Assert().Equal(10, count, "There should only be 10 values")
}

func (s *SuiteIntegrationTests) TestExecuteWithBindings_IT() {

	s.seedData()
	r, err := s.client.ExecuteWithBindings(
		"g.V(x).label()",
		map[string]string{"x": "1234"},
		map[string]string{},
	)
	s.Require().NoError(err, "Unexpected error from server")

	s.T().Logf("Execute with bindings get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)
	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	s.Assert().NoError(err)
	s.Assert().Len(nl, 1, "There should only be 1 node label")
	s.Assert().Equal("Phil", nl[0])
}

func (s *SuiteIntegrationTests) TestExecuteFile_IT() {

	s.seedData()

	r, err := s.client.ExecuteFile("scripts/test.groovy")
	s.Require().NoError(err, "Unexpected error from server")

	s.T().Logf("ExecuteFile get vertex, response: %s \n err: %s", string(r[0].Result.Data), err)

	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	s.Assert().NoError(err)
	s.Assert().Len(nl, 1, "There should only be 1 node label")
	s.Assert().Equal("Vincent", nl[0])
}

func (s *SuiteIntegrationTests) TestExecuteFileWithBindings_IT() {

	s.seedData()

	r, err := s.client.ExecuteFileWithBindings(
		"scripts/test-wbindings.groovy",
		map[string]string{"x": "2145"},
		map[string]string{},
	)
	s.Require().NoError(err, "Unexpected error from server")
	s.T().Logf("ExecuteFileWithBindings get vertex, response: %s \n err: %s", r[0].Result.Data, err)

	var nl nodeLabels
	err = json.Unmarshal(r[0].Result.Data, &nl)
	s.Assert().NoError(err)
	s.Assert().Len(nl, 1, "There should only be 1 node label")
	s.Assert().Equal("Vincent", nl[0])
}

func (s *SuiteIntegrationTests) TestPoolExecute_IT() {

	s.seedData()

	r, err := s.pool.Execute(`g.V('1234').label()`)
	s.Require().NoError(err, "Unexpected error from server")
	s.T().Logf("PoolExecute get vertex, response: %s \n err: %s", r[0].Result.Data, err)
	var nl nodeLabels

	err = json.Unmarshal(r[0].Result.Data, &nl)
	s.Assert().NoError(err)
	s.Assert().Len(nl, 1, "There should only be 1 node label")
	s.Assert().Equal("Phil", nl[0])
}
