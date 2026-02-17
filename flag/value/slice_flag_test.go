package value

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceFlag(t *testing.T) {
	params := []string{
		"1",
		"2",
		"3",
	}

	sf := new(SliceFlag)
	for _, p := range params {
		sf.Set(p)
	}

	require.Equal(t, params, []string(*sf))
}
