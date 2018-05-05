// +build linux
package libipvs

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	schedMethods = []string{
		RoundRobin,
		LeastConnection,
		DestinationHashing,
		SourceHashing,
	}

	schedFlags     = Flags{Flags: IP_VS_SVC_F_SCHED1 | IP_VS_SVC_F_SCHED2 | IP_VS_SVC_F_SCHED3, Mask: ^uint32(0)}
	schedFlagNames = "(flag-1,flag-2,flag-3)"
	shFlagNames    = "(sh-fallback,sh-port,flag-3)"

	protocols = []string{
		"TCP",
		"UDP",
		"FWM",
	}

	fwdMethods = []FwdMethod{
		IP_VS_CONN_F_MASQ,
		IP_VS_CONN_F_TUNNEL,
		IP_VS_CONN_F_DROUTE,
	}

	fwdMethodStrings = []string{
		"Masq",
		"Tunnel",
		"Route",
	}
)

func checkDestination(t *testing.T, checkPresent bool, protocol, serviceAddress, realAddress, fwdMethod string) {
	var (
		realServerStart bool
		realServers     []string
	)

	out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	require.NoError(t, err)

	for _, o := range strings.Split(string(out), "\n") {
		cmpStr := serviceAddress
		if protocol == "FWM" {
			cmpStr = " " + cmpStr
		}

		if strings.Contains(o, cmpStr) {
			realServerStart = true
			continue
		}

		if realServerStart {
			if !strings.Contains(o, "->") {
				break
			}

			realServers = append(realServers, o)
		}
	}

	for _, r := range realServers {
		if strings.Contains(r, realAddress) {
			parts := strings.Fields(r)
			assert.Equal(t, fwdMethod, parts[2])
			return
		}
	}

	if checkPresent {
		t.Fatalf("Did not find the destination %s fwdMethod %s in ipvs output", realAddress, fwdMethod)
	}
}

func checkService(t *testing.T, checkPresent bool, protocol, schedMethod, serviceAddress string) {
	out, err := exec.Command("ipvsadm", "-Ln").CombinedOutput()
	require.NoError(t, err)

	for _, o := range strings.Split(string(out), "\n") {
		cmpStr := serviceAddress
		if protocol == "FWM" {
			cmpStr = " " + cmpStr
		}

		if strings.Contains(o, cmpStr) {
			parts := strings.Split(o, " ")
			assert.Equal(t, protocol, parts[0])
			assert.Equal(t, serviceAddress, parts[2])
			assert.Equal(t, schedMethod, parts[3])
			if schedMethod == SourceHashing {
				assert.Equal(t, shFlagNames, parts[4])
			} else {
				assert.Equal(t, schedFlagNames, parts[4])
			}

			if !checkPresent {
				t.Fatalf("Did not expect the service %s in ipvs output", serviceAddress)
			}

			return
		}
	}

	if checkPresent {
		t.Fatalf("Did not find the service %s in ipvs output", serviceAddress)
	}
}

func clearIpvsState(t *testing.T) {
	_, err := exec.Command("ipvsadm", "--clear").CombinedOutput()
	require.NoError(t, err)
}

func TestService(t *testing.T) {
	defer clearIpvsState(t)

	i, err := New()
	require.NoError(t, err)

	for _, protocol := range protocols {
		for _, schedMethod := range schedMethods {
			var serviceAddress string

			s := Service{
				AddressFamily: syscall.AF_INET,
				SchedName:     schedMethod,
				Flags:         schedFlags,
			}

			switch protocol {
			case "FWM":
				s.FWMark = 1234
				serviceAddress = fmt.Sprintf("%d", 1234)
			case "TCP":
				s.Protocol = syscall.IPPROTO_TCP
				s.Port = 80
				s.Address = net.ParseIP("1.2.3.4")
				s.Netmask = 0xFFFFFFFF
				serviceAddress = "1.2.3.4:80"
			case "UDP":
				s.Protocol = syscall.IPPROTO_UDP
				s.Port = 53
				s.Address = net.ParseIP("2.3.4.5")
				serviceAddress = "2.3.4.5:53"
			}

			err := i.NewService(&s)
			assert.NoError(t, err)
			checkService(t, true, protocol, schedMethod, serviceAddress)

			svcs, err := i.ListServices()
			assert.NoError(t, err)
			var listed *Service
			assert.NotEmpty(t, svcs, "should list at least one virtual service")
			for _, svc := range svcs {
				if svc.FWMark > 0 && svc.FWMark == s.FWMark {
					listed = svc
					break
				}
				if svc.Address.Equal(s.Address) && svc.Port == s.Port && svc.Protocol == s.Protocol {
					listed = svc
					break
				}
			}
			assert.NotNil(t, listed, "could not list %v", s)
			if listed != nil {
				assert.Equal(t, s.AddressFamily, listed.AddressFamily)
				assert.Equal(t, s.SchedName, listed.SchedName)
				expectedFlags := s.Flags
				expectedFlags.Flags |= IP_VS_SVC_F_HASHED
				assert.Equal(t, expectedFlags, listed.Flags)
				assert.Equal(t, s.Netmask, listed.Netmask)
			}

			var lastMethod string
			for _, updateSchedMethod := range schedMethods {
				if updateSchedMethod == schedMethod {
					continue
				}

				s.SchedName = updateSchedMethod
				err = i.UpdateService(&s)
				assert.NoError(t, err)
				checkService(t, true, protocol, updateSchedMethod, serviceAddress)
				lastMethod = updateSchedMethod
			}

			err = i.DelService(&s)
			checkService(t, false, protocol, lastMethod, serviceAddress)
		}
	}
}

func TestDestination(t *testing.T) {
	defer clearIpvsState(t)

	i, err := New()
	require.NoError(t, err)

	for _, protocol := range protocols {
		var serviceAddress string

		s := Service{
			AddressFamily: syscall.AF_INET,
			SchedName:     RoundRobin,
			Flags:         schedFlags,
		}

		switch protocol {
		case "FWM":
			s.FWMark = 1234
			serviceAddress = fmt.Sprintf("%d", 1234)
		case "TCP":
			s.Protocol = syscall.IPPROTO_TCP
			s.Port = 80
			s.Address = net.ParseIP("1.2.3.4")
			s.Netmask = 0xFFFFFFFF
			serviceAddress = "1.2.3.4:80"
		case "UDP":
			s.Protocol = syscall.IPPROTO_UDP
			s.Port = 53
			s.Address = net.ParseIP("2.3.4.5")
			serviceAddress = "2.3.4.5:53"
		}

		err := i.NewService(&s)
		assert.NoError(t, err)
		checkService(t, true, protocol, RoundRobin, serviceAddress)

		s.SchedName = ""
		for j, fwdMethod := range fwdMethods {
			d1 := Destination{
				AddressFamily: syscall.AF_INET,
				Address:       net.ParseIP("10.1.1.2"),
				Port:          5000,
				Weight:        1,
				FwdMethod:     fwdMethod,
			}

			realAddress := "10.1.1.2:5000"
			err := i.NewDestination(&s, &d1)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])
			d2 := Destination{
				AddressFamily: syscall.AF_INET,
				Address:       net.ParseIP("10.1.1.3"),
				Port:          5000,
				Weight:        1,
				FwdMethod:     fwdMethod,
			}

			realAddress = "10.1.1.3:5000"
			err = i.NewDestination(&s, &d2)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])

			d3 := Destination{
				AddressFamily: syscall.AF_INET,
				Address:       net.ParseIP("10.1.1.4"),
				Port:          5000,
				Weight:        1,
				FwdMethod:     fwdMethod,
			}

			realAddress = "10.1.1.4:5000"
			err = i.NewDestination(&s, &d3)
			assert.NoError(t, err)
			checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[j])

			for m, updateFwdMethod := range fwdMethods {
				if updateFwdMethod == fwdMethod {
					continue
				}
				d1.FwdMethod = updateFwdMethod
				realAddress = "10.1.1.2:5000"
				err = i.UpdateDestination(&s, &d1)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])

				d2.FwdMethod = updateFwdMethod
				realAddress = "10.1.1.3:5000"
				err = i.UpdateDestination(&s, &d2)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])

				d3.FwdMethod = updateFwdMethod
				realAddress = "10.1.1.4:5000"
				err = i.UpdateDestination(&s, &d3)
				assert.NoError(t, err)
				checkDestination(t, true, protocol, serviceAddress, realAddress, fwdMethodStrings[m])
			}

			err = i.DelDestination(&s, &d1)
			assert.NoError(t, err)
			err = i.DelDestination(&s, &d2)
			assert.NoError(t, err)
			err = i.DelDestination(&s, &d3)
			assert.NoError(t, err)
		}
	}
}
