package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMustGetIPv4(t *testing.T) {
	require.Equal(t, "0.0.0.0", MustGetIPv4("0.0.0.0:8080"))
}

func TestMustGetPort(t *testing.T) {
	require.Equal(t, "8080", MustGetPort("0.0.0.0:8080"))
}
