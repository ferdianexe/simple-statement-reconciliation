package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRepository(t *testing.T) {
	tests := []struct {
		name string
		want *Repository
	}{
		{
			name: "success",
			want: &Repository{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewRepository()

			assert.NotNil(t, got)
			assert.Equal(t, test.want, got)
		})
	}
}
