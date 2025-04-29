package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMustGetIpV4(t *testing.T) {
	ipV4 := MustGetIpV4("0.0.0.0:8080")
	require.Equal(t, ipV4, "0.0.0.0")
}

func TestMustGetPort(t *testing.T) {
	port := MustGetPort("0.0.0.0:8080")
	require.Equal(t, port, "8080")
}
