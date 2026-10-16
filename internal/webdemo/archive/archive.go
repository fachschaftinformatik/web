// Package archive contains exam records and search rules, independent of HTTP.
package archive

import (
	"slices"
	"strings"
	"time"
)

type Exam struct {
	Title, Program, Version, Date, Description string
	Pages                                      int
	Degree                                     string
	ModuleID                                   string
}

type Filter struct {
	Query    string
	Programs []string
	Versions []string
	Degrees  []string
}

var programs = []string{"Informatik", "Informatik und Design", "Medieninformatik", "Wirtschaftsinformatik"}
var versions = []string{"2023", "2017"}
var degrees = []string{"Bachelor", "Master"}

// Archive is a read-only fixture. No production exams or accounts are accessed.
var exams = []Exam{
	{"Algorithmen und Datenstrukturen", "Informatik", "2023", "12.02.2026", "Sortierverfahren, Graphen und Laufzeitanalyse. Beispiel für eine Klausur mit Lösungshinweisen.", 8, "Bachelor", "algorithmen-und-datenstrukturen"},
	{"Datenbanken", "Wirtschaftsinformatik", "2023", "06.02.2026", "Relationale Modellierung, SQL und Normalisierung. Beispiel für eine schriftliche Prüfung.", 6, "Bachelor", "datenbanken"},
	{"Mathematik I", "Informatik", "2023", "29.01.2026", "Lineare Algebra und Analysis. Beispiel für eine Grundlagenklausur.", 10, "Bachelor", "mathematik-1"},
	{"Web Engineering", "Medieninformatik", "2023", "15.07.2025", "HTTP, Barrierefreiheit und serverseitige Webanwendungen. Beispiel für eine Modulklausur.", 5, "Bachelor", "web-engineering"},
	{"Betriebssysteme", "Informatik", "2017", "10.07.2025", "Prozesse, Speicherverwaltung und Dateisysteme. Beispiel für eine Prüfung ohne Lösung.", 7, "Bachelor", "betriebssysteme"},
	{"Software Engineering", "Wirtschaftsinformatik", "2017", "04.02.2025", "Anforderungen, Architektur und Softwaretests. Beispiel für eine Klausur mit Fallstudie.", 9, "Bachelor", "software-engineering"},
	{"Algorithmen und Datenstrukturen", "Informatik", "2023", "10.07.2025", "Weitere Beispielrekonstruktion zu Sortierverfahren und Suchbäumen.", 6, "Bachelor", "algorithmen-und-datenstrukturen"},
	{"Algorithmen und Datenstrukturen", "Informatik", "2017", "22.01.2025", "Ältere Beispielrekonstruktion zu Datenstrukturen und Rekursion.", 5, "Bachelor", "algorithmen-und-datenstrukturen"},
}

func Programs() []string { return slices.Clone(programs) }
func Versions() []string { return slices.Clone(versions) }
func Degrees() []string  { return slices.Clone(degrees) }

type Module struct {
	ID, Title string
	Exams     []Exam
}

func (m Module) Programs() []string {
	var programs []string
	for _, exam := range m.Exams {
		if !slices.Contains(programs, exam.Program) {
			programs = append(programs, exam.Program)
		}
	}
	return programs
}

// FindModule keeps known modules addressable even if the filters match no uploads.
func FindModule(id string, filter Filter) (Module, bool) {
	for _, module := range FindModules(Filter{}) {
		if module.ID != id {
			continue
		}
		for _, matching := range FindModules(filter) {
			if matching.ID == id {
				return matching, true
			}
		}
		module.Exams = nil
		return module, true
	}
	return Module{}, false
}

// FindModules groups matching reconstructions by module, as the existing archive does.
func FindModules(filter Filter) []Module {
	modules := make([]Module, 0)
	indices := make(map[string]int)
	for _, exam := range FindExams(filter) {
		index, exists := indices[exam.ModuleID]
		if !exists {
			index = len(modules)
			indices[exam.ModuleID] = index
			modules = append(modules, Module{ID: exam.ModuleID, Title: exam.Title})
		}
		modules[index].Exams = append(modules[index].Exams, exam)
	}
	return modules
}

func FindExams(filter Filter) []Exam {
	result := make([]Exam, 0)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, exam := range exams {
		if len(filter.Degrees) > 0 && !slices.Contains(filter.Degrees, exam.Degree) {
			continue
		}
		if len(filter.Programs) > 0 && !slices.Contains(filter.Programs, exam.Program) {
			continue
		}
		if len(filter.Versions) > 0 && !slices.Contains(filter.Versions, exam.Version) {
			continue
		}
		if !strings.Contains(strings.ToLower(exam.Title+" "+exam.Description), query) {
			continue
		}
		result = append(result, exam)
	}
	slices.SortStableFunc(result, func(a, b Exam) int {
		aDate, _ := time.Parse("02.01.2006", a.Date)
		bDate, _ := time.Parse("02.01.2006", b.Date)
		return bDate.Compare(aDate)
	})
	return result
}
