package scanner

import (
	"reflect"
	"testing"
)

func TestGetUnscannedRanges(t *testing.T) {
	tests := []struct {
		name     string
		covered  []DateRange
		reqStart int64
		reqEnd   int64
		expected []DateRange
	}{
		{
			name:     "no covered ranges",
			covered:  nil,
			reqStart: 100,
			reqEnd:   200,
			expected: []DateRange{{StartTS: 100, EndTS: 200}},
		},
		{
			name:     "fully covered range",
			covered:  []DateRange{{StartTS: 50, EndTS: 250}},
			reqStart: 100,
			reqEnd:   200,
			expected: nil,
		},
		{
			name:     "partially covered - extension to the past",
			covered:  []DateRange{{StartTS: 100, EndTS: 200}},
			reqStart: 50,
			reqEnd:   180,
			expected: []DateRange{{StartTS: 50, EndTS: 99}},
		},
		{
			name:     "partially covered - extension to the future",
			covered:  []DateRange{{StartTS: 100, EndTS: 200}},
			reqStart: 150,
			reqEnd:   250,
			expected: []DateRange{{StartTS: 201, EndTS: 250}},
		},
		{
			name:     "partially covered - extension both sides",
			covered:  []DateRange{{StartTS: 100, EndTS: 200}},
			reqStart: 50,
			reqEnd:   250,
			expected: []DateRange{
				{StartTS: 50, EndTS: 99},
				{StartTS: 201, EndTS: 250},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUnscannedRanges(tt.covered, tt.reqStart, tt.reqEnd)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetUnscannedRanges() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeRanges(t *testing.T) {
	tests := []struct {
		name     string
		covered  []DateRange
		newRange DateRange
		expected []DateRange
	}{
		{
			name:     "merge non-overlapping adjacent ranges",
			covered:  []DateRange{{StartTS: 100, EndTS: 200}},
			newRange: DateRange{StartTS: 50, EndTS: 99},
			expected: []DateRange{{StartTS: 50, EndTS: 200}},
		},
		{
			name:     "merge overlapping ranges",
			covered:  []DateRange{{StartTS: 100, EndTS: 200}},
			newRange: DateRange{StartTS: 150, EndTS: 250},
			expected: []DateRange{{StartTS: 100, EndTS: 250}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeRanges(tt.covered, tt.newRange)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MergeRanges() = %v, want %v", got, tt.expected)
			}
		})
	}
}
