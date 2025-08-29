package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderClause(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "ORDER BY a.name ASC"},
		{"name", "ORDER BY a.name ASC"},
		{"-name", "ORDER BY a.name DESC"},
		{"created_at", "ORDER BY a.created_at ASC"},
		{"-created_at", "ORDER BY a.created_at DESC"},
		{"   Created_At   ", "ORDER BY a.created_at ASC"},
		{"WeIrD", "ORDER BY a.name ASC"},
	}

	for _, tc := range cases {
		got := orderClause(tc.in)
		require.Equal(t, tc.want, got, "input=%q", tc.in)
	}
}

func TestMin(t *testing.T) {
	require.Equal(t, 1, min(1, 2))
	require.Equal(t, 1, min(2, 1))
	require.Equal(t, 0, min(0, 0))
	require.Equal(t, -1, min(-1, 5))
	require.Equal(t, -3, min(-3, -1))
}
