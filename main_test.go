package main

import (
	"slices"
	"testing"
)

func TestExpandCategoryFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no flags untouched",
			args: []string{"azproviderlint", "./..."},
			want: []string{"azproviderlint", "./..."},
		},
		{
			name: "rule flags untouched",
			args: []string{"azproviderlint", "-AZG001", "./..."},
			want: []string{"azproviderlint", "-AZG001", "./..."},
		},
		{
			name: "category flag expands to its rules",
			args: []string{"azproviderlint", "-AZG", "./..."},
			want: []string{"azproviderlint", "-AZG001", "-AZG002", "-AZG003", "-AZG004", "-AZG005", "-AZG006", "-AZG007", "-AZG008", "-AZG009", "-AZG000", "./..."},
		},
		{
			name: "lowercase and double-dash accepted",
			args: []string{"azproviderlint", "--azd", "./..."},
			want: []string{"azproviderlint", "-AZD001", "-AZD002", "./..."},
		},
		{
			name: "category and rule flags combine",
			args: []string{"azproviderlint", "-AZG", "-AZR001", "./..."},
			want: []string{"azproviderlint", "-AZG001", "-AZG002", "-AZG003", "-AZG004", "-AZG005", "-AZG006", "-AZG007", "-AZG008", "-AZG009", "-AZG000", "-AZR001", "./..."},
		},
		{
			name: "empty category is not expanded",
			args: []string{"azproviderlint", "-AZN", "./..."},
			want: []string{"azproviderlint", "-AZN", "./..."},
		},
		{
			name: "non-category flags untouched",
			args: []string{"azproviderlint", "-json", "./..."},
			want: []string{"azproviderlint", "-json", "./..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := expandCategoryFlags(tt.args)
			if !slices.Equal(got, tt.want) {
				t.Errorf("expandCategoryFlags(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
