package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterLimit(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		expected int
	}{
		{
			name: "page size default",
			filter: Filter{
				PageSize: 0,
				Page:     1,
			},
			expected: 500,
		},
		{
			name: "page size provided",
			filter: Filter{
				PageSize: 3,
				Page:     1,
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.filter.Limit()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFilterOffset(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		expected int
	}{
		{
			name: "page default",
			filter: Filter{
				PageSize: 0,
				Page:     0,
			},
			expected: 0,
		},
		{
			name: "page provided",
			filter: Filter{
				PageSize: 3,
				Page:     1,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.filter.Offset()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCalculateMetaData(t *testing.T) {
	tests := []struct {
		name     string
		input1   int
		input2   int
		input3   int
		expected Metadata
	}{
		{
			name:     "page default",
			input1:   0,
			input2:   2,
			input3:   1,
			expected: EmptyMetadata,
		},
		{
			name:   "page provided",
			input1: 4,
			input2: 1,
			input3: 2,
			expected: Metadata{
				CurrentPage:  1,
				PageSize:     2,
				FirstPage:    1,
				LastPage:     2,
				TotalRecords: 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CalculateMetadata(tt.input1, tt.input2, tt.input3)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
