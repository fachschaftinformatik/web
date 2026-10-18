package exams

import (
	"slices"
	"time"
)

var programs = []string{"Informatik", "Informatik und Design", "Medieninformatik", "Wirtschaftsinformatik"}
var versions = []string{"2023", "2017"}
var degrees = []string{"Bachelor", "Master"}

// Exams is a read-only fixture. No production exams or accounts are accessed.
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

func Degrees() []string { return slices.Clone(degrees) }

func exampleRecords() Catalog {
	values := slices.Clone(exams)
	for i := range values {
		values[i].ProgramID = values[i].Program
		date, err := time.Parse("02.01.2006", values[i].Date)
		if err != nil {
			panic(err)
		}
		values[i].Date = date.Format(time.DateOnly)
	}
	return Catalog{exams: values, programs: programs, versions: versions, degrees: degrees}
}

func FindModule(id string, filter Filter) (Module, bool) {
	return exampleRecords().FindModule(id, filter)
}

func FindModules(filter Filter) []Module {
	return exampleRecords().FindModules(filter)
}

func FindExams(filter Filter) []Exam {
	return exampleRecords().FindExams(filter)
}
