package view

import "github.com/fachschaftinformatik/web/internal/model/organization"

type AdminOption struct {
	Value, Label string
	Selected     bool
}

type AdminField struct {
	ID, Label, Value, Kind, Hint, Error string
	Day, Month, Year                    string
	InvalidDateParts                    map[string]bool
	Options                             []AdminOption
	Links                               []organization.Link
	FixedLinkTitles                     bool
}

func (f AdminField) DatePartInvalid(part string) bool {
	return f.Error != "" && (f.InvalidDateParts == nil || f.InvalidDateParts[part])
}

func (f AdminField) ErrorAnchor() string {
	if f.Kind == "date" {
		for _, part := range []string{"day", "month", "year"} {
			if f.DatePartInvalid(part) {
				return f.ID + "-" + part
			}
		}
		return f.ID + "-day"
	}
	return f.ID
}

func Field(id, label, value, kind string) AdminField {
	return AdminField{ID: id, Label: label, Value: value, Kind: kind}
}

func AttachErrors(fields []AdminField, errors map[string]string) {
	for i := range fields {
		fields[i].Error = errors[fields[i].ID]
	}
}
