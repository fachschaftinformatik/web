package view

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/urls"
)

type Assignment struct {
	ModuleID, ProgramID, Version, Title, Program, Degree, Value string
}
type ExamsFilter struct {
	Query                       string
	Programs, Versions, Degrees []string
}
type DateInput struct {
	Day, Month, Year string
	Invalid          map[string]bool
}
type Exam struct {
	ID, Title, Program, Version, ExamDate, Description, Degree, ModuleID, ProgramID string
	Pages, Number                                                                   int
	Bytes                                                                           int64
	CanRevise                                                                       bool
	Assignments                                                                     []Assignment
}
type Module struct {
	ID, Title, ProgramID, Program, StudyProgram, Degree string
	Exams                                               []Exam
}
type SubmissionStatus string
type Submission struct {
	ID, Name, Title, Program, ModuleID, Version, Date, Comment, RevisionOf, UserPublicID string
	UserID                                                                               int64
	Status                                                                               SubmissionStatus
	CreatedAt                                                                            string
	SubmittedAt, ReviewedAt                                                              *string
	Documents                                                                            []Exam
	Count, ModuleCount                                                                   int
	Assignments                                                                          []Assignment
	SharedAssignments                                                                    bool
	EditToken                                                                            string
}

func (a Assignment) VersionLabel() string { return VersionLabel(a.Version) }
func VersionLabel(value string) string {
	versions := strings.Split(value, ", ")
	for i, value := range versions {
		version := strings.TrimSpace(strings.TrimPrefix(value, "PO"))
		if len(version) == 4 && strings.Trim(version, "0123456789") == "" {
			versions[i] = version
		}
	}
	return strings.Join(versions, ", ")
}
func (e Exam) URL() string { return urls.ExamDocument(e.ID) }
func (e Exam) Size() string {
	if e.Bytes < 1024*1024 {
		return fmt.Sprintf("%d kB", max(1, (e.Bytes+1023)/1024))
	}
	return strings.ReplaceAll(fmt.Sprintf("%.1f MB", float64(e.Bytes)/(1024*1024)), ".", ",")
}
func (e Exam) VersionLabel() string { return VersionLabel(e.Version) }
func (e Exam) Date() string {
	if date, err := parseDate(e.ExamDate); err == nil {
		return dateLabel(date.Format(time.DateOnly))
	}
	return e.ExamDate
}
func (m Module) URL() string { return urls.ExamModule(m.ID) }
func (r Submission) ListTitle() string {
	if r.ModuleCount > 1 {
		return fmt.Sprintf("%s und %d weitere", r.Title, r.ModuleCount-1)
	}
	return r.Title
}
func (s SubmissionStatus) Label() string {
	switch s {
	case "draft":
		return "Entwurf"
	case "pending":
		return "In Prüfung"
	case "approved":
		return "Freigegeben"
	case "rejected":
		return "Abgelehnt"
	}
	return ""
}
func (r Submission) DateLabel() string {
	if date, err := parseDate(r.Date); err == nil {
		return dateLabel(date.Format(time.DateOnly))
	}
	return r.Date
}
func (r Submission) SubmittedLabel() string {
	value := r.CreatedAt
	if r.SubmittedAt != nil {
		value = *r.SubmittedAt
	}
	if label := dateLabel(value); label != "" {
		return label
	}
	return value
}
func (r Submission) URL() string       { return urls.Submission(r.ID) }
func (r Submission) CheckURL() string  { return urls.CheckSubmission(r.ID) }
func (r Submission) ReviewURL() string { return urls.ReviewSubmission(r.ID) }
func (r Submission) AdminURL() string  { return urls.AdminSubmission(r.ID) }
func (r Submission) UserURL() string   { return urls.AdminUser(r.UserPublicID) }
func (e Exam) RevisionURL() string {
	return "/exams/submissions/create?" + url.Values{"revision": {e.ID}}.Encode()
}

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

func (g *AssignmentGroup) add(a Assignment, checked bool) {
	label := a.Degree
	d := slices.IndexFunc(g.Degrees, func(d AssignmentDegree) bool { return d.Name == label })
	if d < 0 {
		d = len(g.Degrees)
		g.Degrees = append(g.Degrees, AssignmentDegree{Name: label})
	}
	degree := &g.Degrees[d]
	p := slices.IndexFunc(degree.Programs, func(p AssignmentProgram) bool { return p.Name == a.Program })
	if p < 0 {
		p = len(degree.Programs)
		degree.Programs = append(degree.Programs, AssignmentProgram{Name: a.Program})
	}
	degree.Programs[p].Choices = append(degree.Programs[p].Choices, AssignmentChoice{Value: a.Value, Label: VersionLabel(a.Version), Selected: checked})
	g.Selected = g.Selected || checked
}

func assignmentSummary(selected []Assignment) []string {
	var titles []string
	for _, a := range selected {
		titles = UniqueChoice(titles, a.Title)
	}
	slices.Sort(titles)
	return titles
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

// UploadGroups organizes resolved assignments for display; it does not select
// which combinations are valid for a submission.
func UploadGroups(available, selected []Assignment) []AssignmentGroup {
	var groups []AssignmentGroup
	indices := map[string]int{}
	for _, a := range available {
		index, ok := indices[a.Title]
		if !ok {
			index = len(groups)
			indices[a.Title] = index
			groups = append(groups, AssignmentGroup{Title: a.Title})
		}
		checked := slices.ContainsFunc(selected, func(value Assignment) bool { return value.Value == a.Value })
		groups[index].add(a, checked)
	}
	sortAssignmentGroups(groups)
	return groups
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

func (p *ExamsPage) paginate(raw string) {
	p.Total = len(p.Modules)
	p.Pages = max(1, (p.Total+ExamsPageSize-1)/ExamsPageSize)
	p.Page, _ = strconv.Atoi(raw)
	if p.Page < 1 || p.Page > p.Pages {
		p.Page = 1
	}
	start := (p.Page - 1) * ExamsPageSize
	p.Modules = p.Modules[start:min(start+ExamsPageSize, p.Total)]
}

func (p ExamsPage) PageURL(page int) string {
	q, _ := url.ParseQuery(strings.TrimPrefix(p.querySuffix(), "?"))
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/exams"
	}
	return "/exams?" + q.Encode()
}

func (p ExamsPage) Pagination() []PaginationItem {
	return paginationItems(p.Page, p.Pages, p.PageURL)
}

func (p ExamsPage) PreviousURL() string {
	if p.Page > 1 {
		return p.PageURL(p.Page - 1)
	}
	return ""
}

func (p ExamsPage) NextURL() string {
	if p.Page < p.Pages {
		return p.PageURL(p.Page + 1)
	}
	return ""
}

type ExamsPage struct {
	ListPagination     ListPagination
	Module             *Module
	Filter             ExamsFilter
	Programs           []FilterOption
	Versions           []FilterOption
	Degrees            []FilterOption
	Modules            []Module
	Page, Pages, Total int
}

func (p ExamsPage) querySuffix() string {
	q := url.Values{}
	if p.Filter.Query != "" {
		q.Set("search", p.Filter.Query)
	}
	for key, values := range map[string][]string{"degree": p.Filter.Degrees, "program": p.Filter.Programs, "version": p.Filter.Versions} {
		for _, value := range values {
			if value != "" {
				q.Add(key, value)
			}
		}
	}
	// An explicit all-programs selection must not restore saved defaults when
	// following a module or pagination link after clearing the filters.
	if len(p.Filter.Programs)+len(p.Filter.Degrees)+len(p.Filter.Versions) == 0 {
		q.Set("program", "all")
	}
	return "?" + q.Encode()
}

func (p ExamsPage) SearchURL() string { return "/exams" + p.querySuffix() }

func (p ExamsPage) ModuleURL(module Module) string {
	return urls.ExamModule(module.ID) + p.querySuffix()
}

func (p ExamsPage) Filtered() bool {
	return p.Filter.Query != "" || len(p.Filter.Programs)+len(p.Filter.Degrees)+len(p.Filter.Versions) > 0
}

func (p ExamsPage) FilterCount() int {
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
	Label    string
	Selected bool
}

func (o FilterOption) DisplayLabel() string {
	if o.Label != "" {
		return o.Label
	}
	return o.Name
}

type SubmissionListPage struct {
	Submissions  []Submission
	Reviewing    bool
	CanViewUsers bool
}

type SubmissionPage struct {
	Entry     Submission
	Reviewing bool
	Error     string
	EditURL   string
}

// DisplayScope keeps submitted filter identities out of the human-readable
// check-answers summary. Each label comes from the resolved assignments.
func (s Submission) DisplayScope() ExamsFilter {
	var scope ExamsFilter
	for _, document := range s.Documents {
		for _, assignment := range document.Assignments {
			scope.Programs = UniqueChoice(scope.Programs, assignment.Program)
			scope.Degrees = UniqueChoice(scope.Degrees, assignment.Degree)
			scope.Versions = UniqueChoice(scope.Versions, assignment.Version)
		}
	}
	slices.Sort(scope.Programs)
	slices.Sort(scope.Degrees)
	slices.Sort(scope.Versions)
	return scope
}

type UploadPage struct {
	SubmissionID, Date, Comment, Action      string
	RevisionID                               string
	Existing                                 []Exam
	Revision, ChooseScope                    bool
	ReviewEditing                            bool
	SubmissionMail, CanReceiveSubmissionMail bool
	EditToken                                string
	Modules                                  []string
	DateInput                                DateInput
	Scope                                    ExamsFilter
	Programs, Degrees, Versions              []FilterOption
	Groups                                   []AssignmentGroup
	Errors                                   map[string]string
}

func (u UploadPage) CancelURL() string {
	if u.ReviewEditing {
		return urls.ReviewSubmission(u.SubmissionID)
	}
	return urls.ExamsPath
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
