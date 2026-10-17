package meetings

import (
	"bytes"
	"database/sql"
	_ "embed"
	"slices"
	"time"

	"github.com/fachschaftinformatik/web/internal/storage"
)

type Document struct {
	Kind, File, Note, SHA256 string
	Bytes                    int
	Sources                  []string
	Stored                   bool
}

type Meeting struct {
	Year, Number int
	Date         string
	Documents    []Document
}

const MaxPDFBytes = 8 << 20

type Publication struct {
	ID, Owner, Date, Kind, Filename string
	Number                          int
	Data                            []byte
	Published                       bool
}

type Store struct {
	db    *sql.DB
	blobs storage.Provider
}

func ValidatePublication(p Publication) map[string]string {
	errors := map[string]string{}
	date, err := time.Parse("2006-01-02", p.Date)
	if err != nil || date.Year() < 2000 || date.Year() > 2100 {
		errors["date"] = "Gib ein gültiges Sitzungsdatum zwischen 2000 und 2100 ein."
	}
	if p.Number < 1 || p.Number > 999 {
		errors["number"] = "Gib eine Sitzungsnummer zwischen 1 und 999 ein."
	}
	if p.Kind != "Einladung" && p.Kind != "Protokoll" {
		errors["kind"] = "Wähle Einladung zur Sitzung oder Sitzungsprotokoll aus."
	}
	if message := PDFError(p.Data); message != "" {
		errors["file"] = message
	}
	return errors
}

func PDFError(data []byte) string {
	if len(data) == 0 {
		return "Wähle eine Datei aus."
	}
	if len(data) > MaxPDFBytes {
		return "Die Datei darf höchstens 8 MB groß sein."
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("%%EOF")) {
		return "Die ausgewählte Datei muss im PDF-Format vorliegen."
	}
	return ""
}

func (s *Store) Get(id string) (Publication, bool) {
	p, ok, _ := s.GetWithError(id)
	return p, ok
}

func appendPublication(records []Meeting, p Publication, size int) []Meeting {
	date, _ := time.Parse("2006-01-02", p.Date)
	doc := Document{Kind: p.Kind, File: p.ID, Bytes: size, Stored: true}
	index := slices.IndexFunc(records, func(m Meeting) bool { return m.Date == p.Date && m.Number == p.Number })
	if index < 0 {
		records = append(records, Meeting{Year: date.Year(), Number: p.Number, Date: p.Date, Documents: []Document{doc}})
	} else {
		records[index].Documents = append(records[index].Documents, doc)
	}
	return records
}

func sortMeetings(records []Meeting) {
	slices.SortFunc(records, func(a, b Meeting) int {
		if a.Year != b.Year {
			return b.Year - a.Year
		}
		if a.Date != b.Date {
			if a.Date > b.Date {
				return -1
			}
			return 1
		}
		return b.Number - a.Number
	})
}

func (s *Store) All() []Meeting { records, _ := s.AllWithError(); return records }
