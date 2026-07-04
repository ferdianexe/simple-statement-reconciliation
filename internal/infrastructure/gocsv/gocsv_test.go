package gocsv

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoCsv_NewReader(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "returns_a_usable_csv_reader",
			content: "a,b,c\n1,2,3\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Default.NewReader(strings.NewReader(test.content))

			assert.NotNil(t, got)
			record, err := got.Read()
			assert.NoError(t, err)
			assert.Equal(t, []string{"a", "b", "c"}, record)
		})
	}
}

func TestGoCsv_Read(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr error
	}{
		{
			name:    "success",
			content: "a,b,c\n",
			want:    []string{"a", "b", "c"},
			wantErr: nil,
		},
		{
			name:    "empty_input_returns_EOF",
			content: "",
			want:    nil,
			wantErr: io.EOF,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := Default.NewReader(strings.NewReader(test.content))

			got, err := Default.Read(r)

			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestGoCsv_ReadAll(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    [][]string
		wantErr bool
	}{
		{
			name:    "success",
			content: "a,b,c\n1,2,3\n4,5,6\n",
			want:    [][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}},
			wantErr: false,
		},
		{
			name:    "empty_input_returns_nil_no_error",
			content: "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "ragged_rows_return_error",
			content: "a,b,c\n1,2\n",
			want:    nil,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := Default.NewReader(strings.NewReader(test.content))

			got, err := Default.ReadAll(r)

			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
