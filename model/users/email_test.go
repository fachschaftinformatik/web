package users

import "testing"

func TestEmailEligibility(t *testing.T) {
	for _, email := range []string{
		"name@studmail.w-hs.de", "name@fsv-wh.de", "Name@STUDMAIL.W-HS.DE",
		"first.last+tag@fsv-wh.de", `"name@elsewhere"@fsv-wh.de`,
	} {
		if err := validateEmail(email, []string{"studmail.w-hs.de", "fsv-wh.de"}); err != "" {
			t.Errorf("%q rejected: %s", email, err)
		}
	}
	for _, email := range []string{
		"", "name", "name@gmail.com", "name@w-hs.de",
		"name@sub.fsv-wh.de", "name@fsv-wh.de.example.org", "name@notfsv-wh.de",
		"name@studmail.w-hs.de.example.org", "name@fsv-wh.de.",
		"Name <name@fsv-wh.de>", "name@fsv-wh.de (Name)",
	} {
		if validateEmail(email, []string{"studmail.w-hs.de", "fsv-wh.de"}) == "" {
			t.Errorf("%q must be rejected", email)
		}
	}
}
