package view

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/archive"
)

type AssignmentChoice struct {
	Value, Label string
	Selected     bool
}

type AssignmentGroup struct {
	Title    string
	Degrees  []AssignmentDegree
	Selected bool
}

type AssignmentDegree struct {
	Name     string
	Programs []AssignmentProgram
}

type AssignmentProgram struct {
	Name    string
	Choices []AssignmentChoice
}

func (g *AssignmentGroup) add(a archive.Assignment, checked bool) {
	d := slices.IndexFunc(g.Degrees, func(d AssignmentDegree) bool { return d.Name == a.Degree })
	if d < 0 {
		d = len(g.Degrees)
		g.Degrees = append(g.Degrees, AssignmentDegree{Name: a.Degree})
	}
	degree := &g.Degrees[d]
	p := slices.IndexFunc(degree.Programs, func(p AssignmentProgram) bool { return p.Name == a.Program })
	if p < 0 {
		p = len(degree.Programs)
		degree.Programs = append(degree.Programs, AssignmentProgram{Name: a.Program})
	}
	degree.Programs[p].Choices = append(degree.Programs[p].Choices, AssignmentChoice{Value: AssignmentKey(a), Label: archive.VersionLabel(a.Version), Selected: checked})
	g.Selected = g.Selected || checked
}

func assignmentSummary(selected []archive.Assignment) []string {
	var titles []string
	for _, a := range selected {
		titles = UniqueChoice(titles, a.Title)
	}
	slices.Sort(titles)
	return titles
}

func RequestScope(request archive.Request) archive.Filter {
	var scope archive.Filter
	for _, document := range request.Documents {
		for _, a := range document.Assignments {
			scope.Degrees = UniqueChoice(scope.Degrees, a.Degree)
			scope.Programs = UniqueChoice(scope.Programs, a.Program)
			scope.Versions = UniqueChoice(scope.Versions, a.Version)
		}
	}
	slices.Sort(scope.Degrees)
	slices.Sort(scope.Programs)
	slices.Sort(scope.Versions)
	return scope
}

func sortAssignmentGroups(groups []AssignmentGroup) {
	for i := range groups {
		slices.SortFunc(groups[i].Degrees, func(a, b AssignmentDegree) int { return strings.Compare(a.Name, b.Name) })
		for j := range groups[i].Degrees {
			for k := range groups[i].Degrees[j].Programs {
				slices.SortFunc(groups[i].Degrees[j].Programs[k].Choices, func(a, b AssignmentChoice) int { return strings.Compare(a.Label, b.Label) })
			}
		}
	}
}

type scopeSelection struct {
	ID, Label, Error string
	Options          []FilterOption
}

func scopeField(id, label string, options []FilterOption, message string) scopeSelection {
	return scopeSelection{ID: id, Label: label, Options: options, Error: message}
}

func UploadOptions(values, selected []string) []FilterOption {
	var options []FilterOption
	for _, value := range values {
		options = append(options, FilterOption{Name: value, Selected: slices.Contains(selected, value)})
	}
	return options
}

func AssignmentKey(a archive.Assignment) string { return a.ModuleID + "|" + a.Version }

func UploadGroups(catalog archive.Catalog, scope archive.Filter, selected []archive.Assignment) ([]AssignmentGroup, map[string]archive.Assignment) {
	var groups []AssignmentGroup
	indices := map[string]int{}
	allowed := map[string]archive.Assignment{}
	for _, module := range catalog.Modules() {
		program, degree := module.Qualification()
		if !slices.Contains(scope.Programs, program) || !slices.Contains(scope.Degrees, degree) {
			continue
		}
		for _, version := range catalog.ProgramVersions(module.Program) {
			if !slices.Contains(scope.Versions, version) {
				continue
			}
			a := archive.Assignment{ModuleID: module.ID, Version: version, Title: module.Title, Program: program, Degree: degree}
			key := AssignmentKey(a)
			allowed[key] = a
			index, ok := indices[module.Title]
			if !ok {
				index = len(groups)
				indices[module.Title] = index
				groups = append(groups, AssignmentGroup{Title: module.Title})
			}
			checked := slices.ContainsFunc(selected, func(value archive.Assignment) bool { return AssignmentKey(value) == key })
			groups[index].add(a, checked)
		}
	}
	sortAssignmentGroups(groups)
	return groups, allowed
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

func (p *ArchivePage) paginate(raw string) {
	p.Total = len(p.Modules)
	p.Pages = max(1, (p.Total+ArchivePageSize-1)/ArchivePageSize)
	p.Page, _ = strconv.Atoi(raw)
	if p.Page < 1 || p.Page > p.Pages {
		p.Page = 1
	}
	start := (p.Page - 1) * ArchivePageSize
	p.Modules = p.Modules[start:min(start+ArchivePageSize, p.Total)]
}

func (p ArchivePage) PageURL(page int) string {
	q, _ := url.ParseQuery(p.querySuffix()[1:])
	if page > 1 {
		q.Set("seite", strconv.Itoa(page))
	}
	return "/archive?" + q.Encode()
}

func (p ArchivePage) Pagination() []PaginationItem {
	return paginationItems(p.Page, p.Pages, p.PageURL)
}

func (p ArchivePage) PreviousURL() string {
	if p.Page > 1 {
		return p.PageURL(p.Page - 1)
	}
	return ""
}

func (p ArchivePage) NextURL() string {
	if p.Page < p.Pages {
		return p.PageURL(p.Page + 1)
	}
	return ""
}

type ArchivePage struct {
	ListPagination     ListPagination
	Module             *archive.Module
	Filter             archive.Filter
	Programs           []FilterOption
	Versions           []FilterOption
	Degrees            []FilterOption
	Modules            []archive.Module
	Page, Pages, Total int
}

func (p ArchivePage) querySuffix() string {
	q := url.Values{}
	q.Set("applied", "1")
	if p.Filter.Query != "" {
		q.Set("q", p.Filter.Query)
	}
	for key, values := range map[string][]string{"degree": p.Filter.Degrees, "program": p.Filter.Programs, "version": p.Filter.Versions} {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func (p ArchivePage) SearchURL() string { return "/archive" + p.querySuffix() }

func (p ArchivePage) ModuleURL(module archive.Module) string {
	return "/archive/" + url.PathEscape(module.ID) + p.querySuffix()
}

func (p ArchivePage) Filtered() bool {
	return p.Filter.Query != "" || len(p.Filter.Programs)+len(p.Filter.Degrees)+len(p.Filter.Versions) > 0
}

func (p ArchivePage) FilterCount() int {
	count := 0
	for _, program := range p.Programs {
		if program.Selected {
			count++
		}
	}
	for _, version := range p.Versions {
		if version.Selected {
			count++
		}
	}
	for _, degree := range p.Degrees {
		if degree.Selected {
			count++
		}
	}
	return count
}

type FilterOption struct {
	Name     string
	Selected bool
}

type RequestListPage struct {
	Requests     []archive.Request
	Reviewing    bool
	CanViewUsers bool
}

type RequestPage struct {
	Entry     archive.Request
	Reviewing bool
	Error     string
	EditURL   string
}

type UploadPage struct {
	RequestID, Date, Comment, Action string
	Existing                         []archive.Exam
	Revision, ChooseScope            bool
	ReviewEditing                    bool
	EditToken                        string
	Modules                          []string
	DateInput                        archive.DateInput
	Scope                            archive.Filter
	Programs, Degrees, Versions      []FilterOption
	Groups                           []AssignmentGroup
	Errors                           map[string]string
}

func (u UploadPage) ErrorFields() []string {
	var fields []string
	for _, Field := range []string{"degree", "program", "version", "module_id", "date", "comment", "files"} {
		if u.Errors[Field] != "" {
			fields = append(fields, Field)
		}
	}
	return fields
}

func UniqueChoice(values []string, value string) []string {
	if value != "" && !slices.Contains(values, value) {
		return append(values, value)
	}
	return values
}
