package content

import "testing"

func TestPublicationValidationAndMeetingIdentity(t *testing.T) {
	s := NewStore()
	p := Publication{ID: "first", Owner: "owner", Date: "2027-10-28", Number: 45, Kind: "Einladung", Filename: "einladung.pdf", Data: []byte("%PDF-1.4\nexample\n%%EOF")}
	for _, test := range []struct {
		field  string
		mutate func(*Publication)
	}{
		{"date", func(p *Publication) { p.Date = "2027-02-30" }},
		{"number", func(p *Publication) { p.Number = 0 }},
		{"kind", func(p *Publication) { p.Kind = "Other" }},
		{"file", func(p *Publication) { p.Data = []byte("not pdf") }},
		{"file", func(p *Publication) { p.Data = make([]byte, MaxPDFBytes+1) }},
	} {
		copy := p
		test.mutate(&copy)
		if ValidatePublication(copy)[test.field] == "" {
			t.Fatalf("missing %s validation", test.field)
		}
	}
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(p.ID, "other"); err == nil {
		t.Fatal("wrong owner published")
	}
	if err := s.Publish(p.ID, p.Owner); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(p.ID, p.Owner); err != nil {
		t.Fatal("repeat publication should be idempotent")
	}
	if err := s.Save(p); err == nil {
		t.Fatal("published upload changed")
	}
	p.ID, p.Kind = "second", "Protokoll"
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(p.ID, p.Owner); err != nil {
		t.Fatal(err)
	}
	all := s.All()
	if all[0].Date != p.Date || len(all[0].Documents) != 2 {
		t.Fatal("invitation and minutes were not grouped by meeting")
	}
	p.ID, p.Date = "conflict", "2027-10-29"
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(p.ID, p.Owner); err == nil {
		t.Fatal("conflicting date for same meeting number accepted")
	}
	copy, _ := s.Get("first")
	copy.Data[0] = 'x'
	again, _ := s.Get("first")
	if again.Data[0] != '%' {
		t.Fatal("caller mutated stored PDF")
	}
}
