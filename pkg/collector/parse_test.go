package collector

import "testing"

func TestParseNumber(t *testing.T) {
	type testCase struct {
		in  string
		out float64
	}

	testCases := []testCase{
		{in: "1", out: 1},
		{in: "1.1", out: 1.1},
	}

	for _, tc := range testCases {
		out, err := parseNumber(tc.in)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if out != tc.out {
			t.Fatalf("wrong output: %f, expected %f", out, tc.out)
		}
	}
}