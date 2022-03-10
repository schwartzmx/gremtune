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
	// incrementConnectivityErrorCount increments the counter for connectivity errors
	incrementConnectivityErrorCount()

	// incrementConnectionUsageCount increments the counter for using a connection
	incrementConnectionUsageCount(kindOfUsage connectionUsageKind, wasAnError bool)
}

// clientMetricsNop implements clientMetrics and can be used when metrics should be disabled
type clientMetricsNop struct{}

func (c *clientMetricsNop) incrementConnectivityErrorCount()                            {}
func (c *clientMetricsNop) incrementConnectionUsageCount(_ connectionUsageKind, _ bool) {}
