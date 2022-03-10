package gremcos

type connectionUsageKind string

const (
	connectionUsageKindWrite connectionUsageKind = "WRITE"
	connectionUsageKindRead  connectionUsageKind = "READ"
	connectionUsageKindPing  connectionUsageKind = "PING"
)

func (uk connectionUsageKind) String() string {
	switch uk {
	case connectionUsageKindWrite, connectionUsageKindRead, connectionUsageKindPing:
		return string(uk)
	default:
		return "UNKNOWN"
	}
}

type clientMetrics interface {
	// incConnectivityErrorCount increments the counter for connectivity errors
	incConnectivityErrorCount()

	// incConnectionUsageCount increments the counter for using a connection
	incConnectionUsageCount(kindOfUsage connectionUsageKind, wasAnError bool)
}

// clientMetricsNop implements clientMetrics and can be used when metrics should be disabled
type clientMetricsNop struct{}

func (c *clientMetricsNop) incConnectivityErrorCount()                            {}
func (c *clientMetricsNop) incConnectionUsageCount(_ connectionUsageKind, _ bool) {}
