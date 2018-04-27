package exporter

import (
	"os/exec"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyNamespace = "empty-ns"
	ipvsNamespace  = "ipvs-ns"
)

func createNamespace(namespace string) (err error) {
	var (
		createNetworkNamespaceCommand = exec.Command(
			"ip", "netns", "add", namespace)
		turnLoopbackUpInNamespaceCommand = exec.Command(
			"ip", "netns", "exec", namespace,
			"ip", "link", "set", "dev", "lo", "up")
	)

	err = createNetworkNamespaceCommand.Run()
	if err != nil {
		return
	}

	err = turnLoopbackUpInNamespaceCommand.Run()
	return
}

func deleteNamespace(namespace string) (err error) {
	var (
		deleteNetworkNamespaceCommand = exec.Command(
			"ip", "netns", "del", namespace)
	)

	err = deleteNetworkNamespaceCommand.Run()
	return
}

func setupIPVSInNamespace(namespace string) (err error) {
	var (
		createVirtualServer = exec.Command(
			"ip", "netns", "exec", namespace,
			"ipvsadm", "-A",
			"-t", "127.0.0.1:80",
			"-s", "rr")
		addRealServer = exec.Command(
			"ip", "netns", "exec", namespace,
			"ipvsadm", "-a",
			"-t", "127.0.0.1:80",
			"-r", "127.0.0.2")
	)

	err = createVirtualServer.Run()
	if err != nil {
		return
	}

	err = addRealServer.Run()
	return
}

func TestCollectorNew(t *testing.T) {
	var (
		testCases = []struct {
			desc        string
			namespace   string
			shouldError bool
		}{
			{
				desc:        "succeeds with no namespace path",
				namespace:   "",
				shouldError: false,
			},
			{
				desc:        "fails if ns doesnt exist",
				namespace:   "something-inexistent",
				shouldError: true,
			},
			{
				desc:        "succeeds if ns exists",
				namespace:   "/var/run/netns/" + emptyNamespace,
				shouldError: false,
			},
		}
		err       error
		collector Collector
	)

	createNamespace(emptyNamespace)
	defer func() {
		deleteNamespace(emptyNamespace)
	}()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			collector, err = NewCollector(CollectorConfig{
				NamespacePath: tc.namespace,
			})
			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, collector)
		})
	}
}

func TestCollectorGetStats(t *testing.T) {
	var (
		testCases = []struct {
			desc            string
			namespace       string
			numberOfMetrics int
		}{
			{
				desc:            "empty stats in brand new ns",
				namespace:       "/var/run/netns/" + emptyNamespace,
				numberOfMetrics: 0,
			},
			{
				desc:            "zero-ed single stat if single service created",
				namespace:       "/var/run/netns/" + ipvsNamespace,
				numberOfMetrics: 1,
			},
		}
		metricsChan chan prometheus.Metric
	)

	createNamespace(emptyNamespace)
	createNamespace(ipvsNamespace)
	setupIPVSInNamespace(ipvsNamespace)
	defer func() {
		deleteNamespace(emptyNamespace)
		deleteNamespace(ipvsNamespace)
	}()

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			collector, err := NewCollector(CollectorConfig{
				NamespacePath: tc.namespace,
			})
			require.NoError(t, err)
			require.NotNil(t, collector)

			metricsChan = make(chan prometheus.Metric, tc.numberOfMetrics+1)
			collector.Collect(metricsChan)
			assert.Len(t, metricsChan, tc.numberOfMetrics)
		})
	}
}
