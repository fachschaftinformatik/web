package web

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type archiveDate struct {
	Day, Month, Year string
	Invalid          map[string]bool
}

type datePart struct {
	ID, Label, Value, Width string
	Invalid                 bool
}

func (u UploadPage) DateParts() []datePart {
	d := u.DateInput
	parts := []datePart{{ID: "day", Label: "Tag", Value: d.Day, Width: "2"}, {ID: "month", Label: "Monat", Value: d.Month, Width: "2"}, {ID: "year", Label: "Jahr", Value: d.Year, Width: "4"}}
	for i := range parts {
		parts[i].Invalid = u.Errors["date"] != "" && (len(d.Invalid) == 0 || d.Invalid[parts[i].ID])
	}
	return parts
}

func (u UploadPage) DateAnchor() string {
	for _, p := range u.DateParts() {
		if p.Invalid {
			return "date-" + p.ID
		}
	}
	return "date-day"
}

func (d *archiveDate) parse() (string, string) {
	d.Invalid = map[string]bool{}
	var missing []string
	for _, p := range []struct{ id, label, value string }{{"day", "Tag", d.Day}, {"month", "Monat", d.Month}, {"year", "Jahr", d.Year}} {
		if p.value == "" {
			missing = append(missing, p.label)
			d.Invalid[p.id] = true
		}
	}
	if len(missing) == 3 {
		return "", "Gib das Prüfungsdatum ein."
	}
	if len(missing) > 0 {
		return "", "Ergänze beim Prüfungsdatum: " + strings.Join(missing, " und ") + "."
	}
	day, dayErr := strconv.Atoi(d.Day)
	month, monthErr := strconv.Atoi(d.Month)
	if monthErr != nil {
		for i, names := range [][]string{{"januar", "jan"}, {"februar", "feb"}, {"märz", "maerz", "mär", "mrz"}, {"april", "apr"}, {"mai"}, {"juni", "jun"}, {"juli", "jul"}, {"august", "aug"}, {"september", "sep", "sept"}, {"oktober", "okt"}, {"november", "nov"}, {"dezember", "dez"}} {
			for _, name := range names {
				if strings.TrimSuffix(strings.ToLower(d.Month), ".") == name {
					month = i + 1
					monthErr = nil
				}
			}
		}
	}
	year, yearErr := strconv.Atoi(d.Year)
	if dayErr != nil || day < 1 || day > 31 {
		d.Invalid["day"] = true
	}
	if monthErr != nil || month < 1 || month > 12 {
		d.Invalid["month"] = true
	}
	if yearErr != nil || len(d.Year) != 4 {
		d.Invalid["year"] = true
		return "", "Gib beim Prüfungsdatum ein vierstelliges Jahr ein."
	}
	if len(d.Invalid) > 0 {
		return "", "Gib ein gültiges Prüfungsdatum ein."
	}
	date, err := time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", year, month, day))
	if err != nil {
		return "", "Gib ein gültiges Prüfungsdatum ein."
	}
	return date.Format("2006-01-02"), ""
}
