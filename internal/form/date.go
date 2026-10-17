package form

import (
	"strconv"
	"strings"
	"time"
)

// DateInput retains the submitted parts, including invalid values.
type DateInput struct {
	Day, Month, Year string
	Invalid          map[string]bool
}

func (d *DateInput) Parse(label string) (string, string) {
	d.Invalid = map[string]bool{}
	var missing []string
	for _, p := range []struct{ id, label, value string }{{"day", "Tag", d.Day}, {"month", "Monat", d.Month}, {"year", "Jahr", d.Year}} {
		if p.value == "" {
			missing = append(missing, p.label)
			d.Invalid[p.id] = true
		}
	}
	if len(missing) == 3 {
		return "", "Gib das " + label + " ein."
	}
	if len(missing) > 0 {
		return "", "Ergänze beim " + label + ": " + strings.Join(missing, " und ") + "."
	}
	day, dayErr := strconv.Atoi(d.Day)
	month, monthErr := strconv.Atoi(d.Month)
	if monthErr != nil {
		for i, names := range [][]string{{"januar", "jan"}, {"februar", "feb"}, {"märz", "maerz", "mär", "mrz"}, {"april", "apr"}, {"mai"}, {"juni", "jun"}, {"juli", "jul"}, {"august", "aug"}, {"september", "sep", "sept"}, {"oktober", "okt"}, {"november", "nov"}, {"dezember", "dez"}} {
			for _, name := range names {
				if strings.TrimSuffix(strings.ToLower(d.Month), ".") == name {
					month, monthErr = i+1, nil
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
	if yearErr != nil || len(d.Year) != 4 || year < 1 {
		d.Invalid["year"] = true
		return "", "Gib beim " + label + " ein vierstelliges Jahr ein."
	}
	if len(d.Invalid) > 0 {
		return "", "Gib ein gültiges " + label + " ein."
	}
	// Let the standard library validate the complete Gregorian calendar date.
	// time.Date would normalise invalid dates (such as 31 February) instead.
	date, err := time.Parse("2 1 2006", strings.Join([]string{strconv.Itoa(day), strconv.Itoa(month), d.Year}, " "))
	if err != nil {
		for _, part := range []string{"day", "month", "year"} {
			d.Invalid[part] = true
		}
		return "", "Gib ein gültiges " + label + " ein."
	}
	return date.Format(time.DateOnly), ""
}
