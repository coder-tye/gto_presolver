package models

import "testing"

func TestParseCard(t *testing.T) {
	tests := []struct {
		input       string
		expectedStr string
		expectErr   bool
	}{
		{"As", "As", false},
		{"as", "As", false},
		{"10s", "Ts", false},
		{"Ts", "Ts", false},
		{"th", "Th", false},
		{"7d", "7d", false},
		{"2c", "2c", false},
		{"Xz", "", true},
		{"100s", "", true},
		{"A", "", true},
	}

	for _, tc := range tests {
		c, err := ParseCard(tc.input)
		if tc.expectErr {
			if err == nil {
				t.Errorf("ParseCard(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseCard(%q) unexpected error: %v", tc.input, err)
			}
			if c.String() != tc.expectedStr {
				t.Errorf("ParseCard(%q) = %s; want %s", tc.input, c.String(), tc.expectedStr)
			}
		}
	}
}

func TestParseCardsDuplicate(t *testing.T) {
	_, err := ParseCards([]string{"As", "as"})
	if err == nil {
		t.Errorf("expected duplicate card error for As and as")
	}
}
