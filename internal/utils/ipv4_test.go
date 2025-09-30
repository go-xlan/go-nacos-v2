package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIPv4(t *testing.T) {
	ipv4, err := GetIPv4()
	require.NoError(t, err)
	require.NotEmpty(t, ipv4)
	t.Log(ipv4)
}
