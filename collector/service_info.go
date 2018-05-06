package collector

import (
	"github.com/mqliang/libipvs"
)

// ServiceInfo wraps libpvs' service with the
// service's real servers (destinations) and
// the destination port that it's mapped to
// via iptables.
type ServiceInfo struct {
	// destinationServers is a list of destination
	// real servers that the service is connected
	// to.
	//
	// This list is retrieved by issuing IPVS_CMD_GET_DEST
	// (list destinations) for a specific service.
	destinationServers []*libipvs.Destination

	// destinationPort represents the port that is
	// used in iptables as the destination port for
	// the fwmark set by docker.
	destinationPort uint16

	// Service makes ServiceInfo act as an "enhanced
	// service" class.
	*libipvs.Service
}
