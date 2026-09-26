package s3api

import (
	"strings"
	"testing"
)

func TestValidBucketName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"abc", true},
		{"photos", true},
		{"my-bucket.2026", true},
		{"a1.b2-c3", true},
		{strings.Repeat("a", 63), true},
		{"1bucket", true},

		{"ab", false},
		{strings.Repeat("a", 64), false},
		{"", false},

		{"aBC", false},
		{"a b", false},
		{"a_b", false},
		{"a/../../etc", false},
		{"a%2fb", false},
		{"şehir", false},

		{"-abc", false},
		{"abc-", false},
		{".abc", false},
		{"abc.", false},
		{"a..b", false},

		{"192.168.1.1", false},
		{"10.0.0.255", false},
		{"192.168.1.1a", true},

		{"xn--abc", false},
		{"abc-s3alias", false},
	}
	for _, tt := range tests {
		if got := validBucketName(tt.name); got != tt.valid {
			t.Errorf("validBucketName(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}
