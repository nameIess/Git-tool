package gitcfg

import "testing"

func TestValidateEmail(t *testing.T) {
	for _, email := range []string{"user@example.com", "a.b+tag@example.co.in"} {
		if !ValidateEmail(email) { t.Fatalf("expected valid email: %q", email) }
	}
	for _, email := range []string{"", "not-an-email", "@example.com"} {
		if ValidateEmail(email) { t.Fatalf("expected invalid email: %q", email) }
	}
}
