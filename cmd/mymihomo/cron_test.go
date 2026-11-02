package main

import (
	"testing"
)

func TestValidateCronExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"empty string", "", false},
		{"whitespace string", "   ", false},
		{"valid standard 5 fields", "1 * * * *", false},
		{"valid daily 2am", "0 2 * * *", false},
		{"valid every 5 minutes", "*/5 * * * *", false},
		{"valid complex range", "0 2,14 * * 1-5", false},
		{"invalid 4 fields", "* * * *", true},
		{"invalid 6 fields with seconds", "0 1 * * * *", true},
		{"invalid single word", "daily", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCronExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCronExpression(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}
