package gremtune

import (
	"testing"
)

func init() {
	InitGremlinClients()
	t := testing.T{}
	seedData(&t)

}

func benchmarkPoolExecute(i int, b *testing.B) {
	for n := 0; n < i; n++ {
		go func(p *Pool) {
			_, err := p.Execute(`g.V('1234').label()`)
			if err != nil {
				b.Error(err)
			}
		}(gp)
	}
}

func BenchmarkPoolExecute1(b *testing.B)   { benchmarkPoolExecute(1, b) }
func BenchmarkPoolExecute5(b *testing.B)   { benchmarkPoolExecute(5, b) }
func BenchmarkPoolExecute10(b *testing.B)  { benchmarkPoolExecute(10, b) }
func BenchmarkPoolExecute20(b *testing.B)  { benchmarkPoolExecute(20, b) }
func BenchmarkPoolExecute40(b *testing.B)  { benchmarkPoolExecute(40, b) }
func BenchmarkPoolExecute80(b *testing.B)  { benchmarkPoolExecute(80, b) }
func BenchmarkPoolExecute160(b *testing.B) { benchmarkPoolExecute(160, b) }
