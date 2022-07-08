package internal

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPollerAutoScalerOptions_SetAugmentedValue(t *testing.T) {
	tests := []struct {
		name    string
		options PollerAutoScalerOptions
		want    PollerAutoScalerOptions
	}{
		{
			name: "sanitized value",
			options: PollerAutoScalerOptions{
				Enabled:              true,
				MinConcurrentPollers: 1,
			},
			want: PollerAutoScalerOptions{
				Enabled:              true,
				MinConcurrentPollers: 1,
			},
		},
		{
			name: "unsanitized value",
			options: PollerAutoScalerOptions{
				Enabled:              true,
				MinConcurrentPollers: 0,
			},
			want: PollerAutoScalerOptions{
				Enabled:              true,
				MinConcurrentPollers: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.options.SetAugmentedValue()
			require.Equal(t, tt.want, tt.options)
		})
	}
}
