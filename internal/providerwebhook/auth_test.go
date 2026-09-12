package providerwebhook

import "testing"

func TestValidSharedSecret(t *testing.T) {
	const fixture = "synthetic-test-fixture"
	tests := []struct {
		name     string
		expected string
		supplied string
		valid    bool
	}{
		{name: "both missing"},
		{name: "configuration missing", supplied: fixture},
		{name: "credential missing", expected: fixture},
		{name: "exact match", expected: fixture, supplied: fixture, valid: true},
		{name: "different value", expected: fixture, supplied: "synthetic-test-fixturE"},
		{name: "different length", expected: fixture, supplied: "short"},
		{name: "extra whitespace", expected: fixture, supplied: fixture + " "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ValidSharedSecret(test.expected, test.supplied); got != test.valid {
				t.Errorf("authentication result = %t, want %t", got, test.valid)
			}
		})
	}
}
