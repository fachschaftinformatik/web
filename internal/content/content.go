// Package content manages public documents and their publication.
package content

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed catalog.json
var catalog []byte

type Document struct {
	Kind, File, Note, SHA256 string
	Bytes                    int
	Sources                  []string
	Demo                     bool
}

func (d Document) URL() string {
	if d.Demo {
		return "/documents/" + d.File
	}
	return "/assets/content/" + d.File
}
func (d Document) Size() string { return fmt.Sprintf("%d kB", (d.Bytes+999)/1000) }

type Meeting struct {
	Year, Number int
	Date         string
	Documents    []Document
}

func (m Meeting) DateLabel() string {
	date, err := time.Parse("2006-01-02", m.Date)
	if err != nil {
		return "Datum uneindeutig"
	}
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return fmt.Sprintf("%d. %s %d", date.Day(), months[date.Month()-1], date.Year())
}

// All returns a fresh copy of the curated, newest-first document index.
func All() []Meeting {
	var records []Meeting
	if err := json.Unmarshal(catalog, &records); err != nil {
		panic("invalid embedded meeting catalog: " + err.Error())
	}
	return records
}

func Years() []int {
	var years []int
	for _, meeting := range All() {
		if len(years) == 0 || years[len(years)-1] != meeting.Year {
			years = append(years, meeting.Year)
		}
	}
	return years
}

func ForYear(year int) []Meeting {
	var result []Meeting
	for _, meeting := range All() {
		if meeting.Year == year {
			result = append(result, meeting)
		}
	}
	return result
}
