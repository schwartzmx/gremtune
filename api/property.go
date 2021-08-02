package api

import (
	"github.com/supplyon/gremcos/interfaces"
)

type property struct {
	builders []interfaces.QueryBuilder
}

func NewPropertyV(v interfaces.Vertex) interfaces.Property {
	queryBuilders := make([]interfaces.QueryBuilder, 0)
	queryBuilders = append(queryBuilders, v)

	return &property{
		builders: queryBuilders,
	}
}

func (p *property) String() string {
	queryString := ""
	for _, queryBuilder := range p.builders {
		queryString += queryBuilder.String()
	}
	return queryString
}

// Add can be used to add a custom QueryBuilder
// e.g. g.V().Add(NewSimpleQB(".myCustomCall("%s")",label))
func (p *property) Add(builder interfaces.QueryBuilder) interfaces.Property {
	p.builders = append(p.builders, builder)
	return p
}

// Drop adds .drop(), to the query. The query call will drop/ delete all referenced entities
func (p *property) Drop() interfaces.QueryBuilder {
	return p.Add(NewSimpleQB(".drop()"))
}

// Profile adds .executionProfile(), to the query. The query call will return profiling information of the executed query
func (p *property) Profile() interfaces.QueryBuilder {
	if !gUSE_COSMOS_DB_QUERY_LANGUAGE {
		return p.Add(NewSimpleQB(".profile()"))
	}
	return p.Add(NewSimpleQB(".executionProfile()"))
}

// Count adds .count(), to the query. The query call will return the number of entities found in the query.
func (p *property) Count() interfaces.QueryBuilder {
	return p.Add(NewSimpleQB(".count()"))
}

// Limit adds .limit(<num>), to the query. The query call will limit the results of the query to the given number.
func (p *property) Limit(maxElements int) interfaces.Property {
	return p.Add(NewSimpleQB(".limit(%d)", maxElements))
}

// As adds .as([<label_1>,<label_2>,..,<label_n>]), to the query to label that query step for later access.
func (p *property) As(labels ...string) interfaces.Property {
	query := multiParamQuery(".as", labels...)
	return p.Add(query)
}
