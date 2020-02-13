package gremcos

import (
	"sync"
	"testing"

	"github.com/schwartzmx/gremtune/interfaces"
)

var benchmarkClient interfaces.QueryExecutor
var benchmarkPool *pool

var once sync.Once

func initBeforeBenchmark() {
	t := &testing.T{}

	t.Log("Starting the benchmark. In order to run it a local gremlin server has to run and listen on 8182")

	// create the error channels
	clientErrChannel := make(chan error)
	poolErrChannel := make(chan error)

	// create failing readers for those channels
	go failingErrorChannelConsumerFunc(clientErrChannel, t)
	go failingErrorChannelConsumerFunc(poolErrChannel, t)

	benchmarkClient = newTestClient(t, clientErrChannel)
	benchmarkPool = newTestPool(t, poolErrChannel)

	seedData(t, benchmarkClient)
}

func benchmarkPoolExecute(i int, b *testing.B) {
	once.Do(initBeforeBenchmark)

	for n := 0; n < i; n++ {
		go func(p *pool) {
			_, err := p.Execute(`g.V('1234').label()`)
			if err != nil {
				b.Error(err)
			}
		}(benchmarkPool)
	}
}

func BenchmarkPoolExecute1(b *testing.B)   { benchmarkPoolExecute(1, b) }
func BenchmarkPoolExecute5(b *testing.B)   { benchmarkPoolExecute(5, b) }
func BenchmarkPoolExecute10(b *testing.B)  { benchmarkPoolExecute(10, b) }
func BenchmarkPoolExecute20(b *testing.B)  { benchmarkPoolExecute(20, b) }
func BenchmarkPoolExecute40(b *testing.B)  { benchmarkPoolExecute(40, b) }
func BenchmarkPoolExecute80(b *testing.B)  { benchmarkPoolExecute(80, b) }
func BenchmarkPoolExecute160(b *testing.B) { benchmarkPoolExecute(160, b) }
