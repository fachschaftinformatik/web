package content

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

const MaxPDFBytes = 8 << 20

type Publication struct {
	ID, Owner, Date, Kind, Filename string
	Number                          int
	Data                            []byte
	Published                       bool
}

type Store struct {
	mu           sync.RWMutex
	publications map[string]Publication
}

func NewStore() *Store { return &Store{publications: make(map[string]Publication)} }

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
		return "Wähle eine PDF-Datei aus."
	}
	if len(data) > MaxPDFBytes {
		return "Die PDF-Datei darf höchstens 8 MB groß sein."
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data, []byte("%%EOF")) {
		return "Die ausgewählte Datei muss eine PDF-Datei sein."
	}
	return ""
}

func (s *Store) Save(p Publication) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.publications[p.ID]; ok && (old.Owner != p.Owner || old.Published) {
		return errors.New("Dieser Entwurf kann nicht geändert werden.")
	}
	total := len(p.Data)
	for id, existing := range s.publications {
		if id != p.ID {
			total += len(existing.Data)
		}
	}
	if total > 32<<20 {
		return errors.New("Die Dateien konnten nicht gespeichert werden. Versuche es später erneut.")
	}
	p.Data = slices.Clone(p.Data)
	s.publications[p.ID] = p
	return nil
}

func (s *Store) Get(id string) (Publication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.publications[id]
	p.Data = slices.Clone(p.Data)
	return p, ok
}

func (s *Store) Publish(id, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.publications[id]
	if !ok || p.Owner != owner {
		return errors.New("Dieser Entwurf wurde nicht gefunden.")
	}
	if p.Published {
		return nil
	}
	if len(ValidatePublication(p)) != 0 {
		return errors.New("Prüfe die Angaben und die PDF-Datei.")
	}
	date, _ := time.Parse("2006-01-02", p.Date)
	for _, meeting := range s.allLocked() {
		if meeting.Year != date.Year() {
			continue
		}
		if meeting.Date == p.Date && meeting.Number != p.Number || meeting.Number == p.Number && meeting.Date != p.Date {
			return errors.New("Datum und Sitzungsnummer stimmen nicht mit der vorhandenen Sitzung überein.")
		}
	}
	p.Published = true
	s.publications[id] = p
	return nil
}

func (s *Store) allLocked() []Meeting {
	records := All()
	for _, p := range s.publications {
		if !p.Published {
			continue
		}
		date, _ := time.Parse("2006-01-02", p.Date)
		doc := Document{Kind: p.Kind, File: p.ID, Bytes: len(p.Data), Demo: true}
		index := slices.IndexFunc(records, func(m Meeting) bool { return m.Date == p.Date && m.Number == p.Number })
		if index < 0 {
			records = append(records, Meeting{Year: date.Year(), Number: p.Number, Date: p.Date, Documents: []Document{doc}})
		} else {
			records[index].Documents = append(records[index].Documents, doc)
		}
	}
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
	return records
}

func (s *Store) All() []Meeting { s.mu.RLock(); defer s.mu.RUnlock(); return s.allLocked() }

func (p Publication) Size() string      { return fmt.Sprintf("%d kB", (len(p.Data)+999)/1000) }
func (p Publication) KindLabel() string { return kindLabel(p.Kind) }
func (p Publication) DateLabel() string { return (Meeting{Date: p.Date}).DateLabel() }
