package oauth

import "testing"

func TestParseOrigin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{
			input:    "https://app.example.com",
			expected: "https://app.example.com",
			valid:    true,
		},
		{
			input:    "http://127.0.0.1:5173",
			expected: "http://127.0.0.1:5173",
			valid:    true,
		},
		{
			input: "https://app.example.com/path",
			valid: false,
		},
		{
			input: "https://app.example.com?query=true",
			valid: false,
		},
		{
			input: "https://*.example.com",
			valid: false,
		},
		{
			input: "not-an-origin",
			valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := parseOrigin(test.input)

			if test.valid && err != nil {
				t.Fatalf(
					"expected valid origin: %v",
					err,
				)
			}

			if !test.valid && err == nil {
				t.Fatal(
					"expected origin to be rejected",
				)
			}

			if test.valid && actual != test.expected {
				t.Fatalf(
					"expected %q, got %q",
					test.expected,
					actual,
				)
			}
		})
	}
}
