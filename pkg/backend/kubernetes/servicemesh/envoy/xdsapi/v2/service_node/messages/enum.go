package messages

import (
	"fmt"

	envoyv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

// convert a string to a Cluster_LbPolicy and panic if not found
func stringToClusterLbPolicy(lbPolicy string) envoyv2.Cluster_LbPolicy {
	_lbPolicy, ok := envoyv2.Cluster_LbPolicy_value[lbPolicy]
	if !ok {
		panic(fmt.Sprintf("unknown cluster load balancer policy <%s>", lbPolicy))
	}
	return envoyv2.Cluster_LbPolicy(_lbPolicy)
}
