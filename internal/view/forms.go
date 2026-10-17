package view

type AdminOption struct {
	Value, Label string
	Selected     bool
}

type AdminField struct {
	ID, Label, Value, Kind, Hint, Error string
	Day, Month, Year                    string
	Options                             []AdminOption
}

func Field(id, label, value, kind string) AdminField {
	return AdminField{ID: id, Label: label, Value: value, Kind: kind}
}

func AttachErrors(fields []AdminField, errors map[string]string) {
	for i := range fields {
		fields[i].Error = errors[fields[i].ID]
	}
}
