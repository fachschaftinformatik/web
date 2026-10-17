package config

import "testing"

func TestDomainConfiguration(t *testing.T) {
	t.Setenv("ACCOUNT_DOMAINS", " Students.Example.org , council.example.org ")
	t.Setenv("REGISTRATION_DOMAINS", "students.example.org")
	c := New()
	if err := c.ValidateDomains(); err != nil || c.AccountDomains[0] != "students.example.org" {
		t.Fatalf("invalid normalized configuration: %+v %v", c, err)
	}
	c.RegistrationDomains = []string{"unlisted.example.org"}
	if c.ValidateDomains() == nil {
		t.Fatal("registration domain outside account policy accepted")
	}
	c.AccountDomains = []string{"https://example.org"}
	if c.ValidateDomains() == nil {
		t.Fatal("URL accepted as an email domain")
	}
	c.AccountDomains = nil
	c.RegistrationDomains = nil
	if c.ValidateDomains() == nil {
		t.Fatal("missing account policy silently accepted")
	}
}
