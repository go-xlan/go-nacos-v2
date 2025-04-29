package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIpv4(t *testing.T) {
	ipv4, err := GetIpv4()
	require.NoError(t, err)
	t.Log(ipv4)
}
