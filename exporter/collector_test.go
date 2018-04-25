package exporter

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	existingNamespace = "myns"
)

func setup(t *testing.T) {
	cmd := exec.Command("ip", "netns", "add", existingNamespace)
	err := cmd.Run()
	require.NoError(t, err)
}

func tearDown(t *testing.T) {
	cmd := exec.Command("ip", "netns", "del", existingNamespace)
	err := cmd.Run()
	require.NoError(t, err)
}

func TestCollectorNew(t *testing.T) {
	var testCases = []struct {
		desc        string
		namespace   string
		shouldError bool
	}{
		{
			desc:        "fails if ns doesnt exist",
			namespace:   "something-inexistent",
			shouldError: true,
		},
		{
			desc:        "succeeds if ns exists",
			namespace:   existingNamespace,
			shouldError: true,
		},
	}

	var (
		err       error
		collector Collector
	)

	setup(t)
	defer tearDown(t)

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
		})
	}
}
