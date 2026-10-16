package account

import "slices"

type Program struct{ ID, Name string }

// Programs is a local snapshot of https://ik.fsv-wh.de/api/v1/programs.
// Rendering the demo does not require a live API connection.
func Programs() []Program {
	return []Program{
		{"140113416076394497", "Informatik (B. Sc.)"},
		{"140113607848361985", "Informatik (M. Sc.)"},
		{"140113628429811713", "Informatik und Design (B. Sc.)"},
		{"140113623300177921", "Internetsicherheit (M. Sc.)"},
		{"140113544057192449", "Medieninformatik (B. Sc.)"},
		{"140113613028327425", "Medieninformatik (M. Sc.)"},
		{"140113479615905793", "Wirtschaftsinformatik (B. Sc.)"},
		{"140113618229264385", "Wirtschaftsinformatik (M. Sc.)"},
	}
}

func ProgramError(id string) string {
	if id == "" {
		return ""
	}
	if !slices.ContainsFunc(Programs(), func(p Program) bool { return p.ID == id }) {
		return "Bitte wähle deinen Studiengang aus."
	}
	return ""
}
