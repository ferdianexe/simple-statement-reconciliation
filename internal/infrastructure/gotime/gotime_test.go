package gotime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoTime_TruncateToDay(t *testing.T) {
	tests := []struct {
		name string
		arg  time.Time
		want time.Time
	}{
		{
			name: "strips_time_of_day_component",
			arg:  time.Date(2024, 1, 8, 15, 30, 45, 0, time.UTC),
			want: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "already_truncated_input_is_unchanged",
			arg:  time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			want: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Default.TruncateToDay(test.arg)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGoTime_InRange(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	type args struct {
		d, start, end time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "within_range_with_time_of_day_component",
			args: args{d: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), start: start, end: end},
			want: true,
		},
		{
			name: "exactly_on_start_boundary_is_inclusive",
			args: args{d: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC), start: start, end: end},
			want: true,
		},
		{
			name: "exactly_on_end_boundary_is_inclusive",
			args: args{d: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), start: start, end: end},
			want: true,
		},
		{
			name: "before_start_is_excluded",
			args: args{d: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), start: start, end: end},
			want: false,
		},
		{
			name: "after_end_is_excluded",
			args: args{d: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), start: start, end: end},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Default.InRange(test.args.d, test.args.start, test.args.end)
			assert.Equal(t, test.want, got)
		})
	}
}
