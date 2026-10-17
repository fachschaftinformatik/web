// Package files contains file records and filtering rules, independent of HTTP.
package files

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Exam struct {
	Title, Program, Version, Date, Description string
	Pages                                      int
	Degree                                     string
	ModuleID                                   string
	ID                                         string
	Bytes                                      int64
	OwnerID                                    int64
	Number                                     int
	alias                                      string
}

func (e Exam) CanRevise(userID int64, role string) bool {
	return e.ID != "" && userID != 0 && (e.OwnerID == userID || role == "admin" || role == "editor")
}

func (e Exam) URL() string {
	if e.ID == "" {
		return "/assets/example.pdf"
	}
	return "/archive/files/" + e.ID
}

func (e Exam) Size() string {
	if e.Bytes < 1024*1024 {
		return fmt.Sprintf("%d kB", max(1, (e.Bytes+1023)/1024))
	}
	return strings.ReplaceAll(fmt.Sprintf("%.1f MB", float64(e.Bytes)/(1024*1024)), ".", ",")
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
	{Title: "Algorithmen und Datenstrukturen", Program: "Informatik", Version: "2023", Date: "12.02.2026", Description: "Sortierverfahren, Graphen und Laufzeitanalyse. Beispiel für eine Klausur mit Lösungshinweisen.", Pages: 8, Degree: "Bachelor", ModuleID: "algorithmen-und-datenstrukturen"},
	{Title: "Datenbanken", Program: "Wirtschaftsinformatik", Version: "2023", Date: "06.02.2026", Description: "Relationale Modellierung, SQL und Normalisierung. Beispiel für eine schriftliche Prüfung.", Pages: 6, Degree: "Bachelor", ModuleID: "datenbanken"},
	{Title: "Mathematik I", Program: "Informatik", Version: "2023", Date: "29.01.2026", Description: "Lineare Algebra und Analysis. Beispiel für eine Grundlagenklausur.", Pages: 10, Degree: "Bachelor", ModuleID: "mathematik-1"},
	{Title: "Web Engineering", Program: "Medieninformatik", Version: "2023", Date: "15.07.2025", Description: "HTTP, Barrierefreiheit und serverseitige Webanwendungen. Beispiel für eine Modulklausur.", Pages: 5, Degree: "Bachelor", ModuleID: "web-engineering"},
	{Title: "Betriebssysteme", Program: "Informatik", Version: "2017", Date: "10.07.2025", Description: "Prozesse, Speicherverwaltung und Dateisysteme. Beispiel für eine Prüfung ohne Lösung.", Pages: 7, Degree: "Bachelor", ModuleID: "betriebssysteme"},
	{Title: "Software Engineering", Program: "Wirtschaftsinformatik", Version: "2017", Date: "04.02.2025", Description: "Anforderungen, Architektur und Softwaretests. Beispiel für eine Klausur mit Fallstudie.", Pages: 9, Degree: "Bachelor", ModuleID: "software-engineering"},
	{Title: "Algorithmen und Datenstrukturen", Program: "Informatik", Version: "2023", Date: "10.07.2025", Description: "Weitere Beispielrekonstruktion zu Sortierverfahren und Suchbäumen.", Pages: 6, Degree: "Bachelor", ModuleID: "algorithmen-und-datenstrukturen"},
	{Title: "Algorithmen und Datenstrukturen", Program: "Informatik", Version: "2017", Date: "22.01.2025", Description: "Ältere Beispielrekonstruktion zu Datenstrukturen und Rekursion.", Pages: 5, Degree: "Bachelor", ModuleID: "algorithmen-und-datenstrukturen"},
}

func Programs() []string { return slices.Clone(programs) }
func Versions() []string { return slices.Clone(versions) }
func Degrees() []string  { return slices.Clone(degrees) }

type Module struct {
	ID, Title string
	Program   string
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
	return Examples().FindModule(id, filter)
}

type Catalog struct {
	exams                       []Exam
	modules                     []Module
	programs, versions, degrees []string
	programVersions             map[string][]string
}

func (c Catalog) ProgramVersions(program string) []string {
	return slices.Clone(c.programVersions[program])
}

func Examples() Catalog {
	return Catalog{exams: exams, programs: programs, versions: versions, degrees: degrees}
}
func (c Catalog) Programs() []string { return slices.Clone(c.programs) }
func (c Catalog) Versions() []string { return slices.Clone(c.versions) }
func (c Catalog) Degrees() []string  { return slices.Clone(c.degrees) }

func (c Catalog) FindModule(id string, filter Filter) (Module, bool) {
	known := c.modules
	if known == nil {
		known = c.FindModules(Filter{})
	}
	for _, module := range known {
		if module.ID != id {
			continue
		}
		for _, matching := range c.FindModules(filter) {
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
	return Examples().FindModules(filter)
}

func (c Catalog) FindModules(filter Filter) []Module {
	modules := make([]Module, 0)
	indices := make(map[string]int)
	for _, exam := range c.FindExams(filter) {
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
	return Examples().FindExams(filter)
}

func (c Catalog) FindExams(filter Filter) []Exam {
	result := make([]Exam, 0)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, exam := range c.exams {
		if len(filter.Degrees) > 0 && !slices.Contains(filter.Degrees, exam.Degree) {
			continue
		}
		if len(filter.Programs) > 0 && !slices.Contains(filter.Programs, exam.Program) {
			continue
		}
		if len(filter.Versions) > 0 && !slices.Contains(filter.Versions, exam.Version) {
			continue
		}
		if !strings.Contains(strings.ToLower(exam.Title+" "+exam.Description+" "+exam.alias), query) {
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
