// Package team describes the Fachschaft's members and responsibilities.
package team

type Member struct {
	Name, Initials, Role, Email string
}

type TeamSection struct {
	Title, Group string
	Members      []Member
}

type Group struct {
	ID, Title string
	Sections  []TeamSection
}

func Groups() []Group {
	sections := Team()
	return []Group{
		{ID: "board", Title: "Vorstand", Sections: sections[:1]},
		{ID: "departments", Title: "Referate", Sections: sections[1:6]},
		{ID: "project", Title: "Projektstellen", Sections: []TeamSection{
			{Title: "Öffentlichkeit", Members: sections[6].Members[:1]},
			{Title: "Infrastruktur", Members: sections[6].Members[1:]},
		}},
	}
}

// SharedEmail avoids repeating a responsibility's shared address for each member.
func (s TeamSection) SharedEmail() string {
	if len(s.Members) < 2 || s.Members[0].Email == "" {
		return ""
	}
	email := s.Members[0].Email
	for _, member := range s.Members[1:] {
		if member.Email != email {
			return ""
		}
	}
	return email
}

// Team mirrors the current public team listing; it is demo data, not a new source of truth.
func Team() []TeamSection {
	return []TeamSection{
		{"Vorstand", "Gewählte Mitglieder", []Member{
			{"Toni Gudaqi", "TG", "Vorsitz", "vorsitz.informatik@fsv-whs.de"},
			{"Malte Gaelings", "MG", "Stellvertretender Vorsitz", "stellv.vorsitz.informatik@fsv-whs.de"},
			{"Leon Pearse", "LP", "Finanzen", "finanzen.informatik@fsv-whs.de"},
		}},
		{"Kultur", "Gewählte Mitglieder", []Member{
			{"Joel Enrico Soriano", "JS", "Kultur", "kultur.informatik@fsv-whs.de"},
			{"Carlo Niermann", "CN", "Kultur", "kultur.informatik@fsv-whs.de"},
			{"Leon Weihrauch", "LW", "Kultur", "kultur.informatik@fsv-whs.de"},
		}},
		{"Infrastruktur", "Gewählte Mitglieder", []Member{
			{"Malte Gaelings", "MG", "Infrastruktur", "edv.informatik@fsv-whs.de"},
			{"Birhat Alessandro Moffa", "BM", "Infrastruktur", "edv.informatik@fsv-whs.de"},
			{"Zafer Elemen", "ZE", "Infrastruktur", "edv.informatik@fsv-whs.de"},
		}},
		{"Öffentlichkeit", "Gewählte Mitglieder", []Member{
			{"Toni Gudaqi", "TG", "Öffentlichkeit", "oeffentlichkeit.informatik@fsv-whs.de"},
		}},
		{"Beschaffung", "Gewählte Mitglieder", []Member{
			{"Leon Weihrauch", "LW", "Beschaffung", "beschaffung.informatik@fsv-whs.de"},
			{"Marcel Roith", "MR", "Beschaffung", "beschaffung.informatik@fsv-whs.de"},
			{"Birhat Alessandro Moffa", "BM", "Beschaffung", "beschaffung.informatik@fsv-whs.de"},
		}},
		{"Sport", "Gewählte Mitglieder", []Member{
			{"Caner Bayram", "CB", "Sport", ""},
		}},
		{"Projektstellen", "Nicht gewählte Mitglieder", []Member{
			{"Kubilay Cakmakkaya", "KC", "Öffentlichkeit", ""},
			{"Francesco La Camera", "FL", "Infrastruktur", ""},
		}},
	}
}
