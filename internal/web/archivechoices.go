package web

import (
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content/files"
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

func (g *AssignmentGroup) add(a files.Assignment, checked bool) {
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
	degree.Programs[p].Choices = append(degree.Programs[p].Choices, AssignmentChoice{Value: assignmentKey(a), Label: files.VersionLabel(a.Version), Selected: checked})
	g.Selected = g.Selected || checked
}

func assignmentSummary(selected []files.Assignment) []string {
	var titles []string
	for _, a := range selected {
		titles = uniqueChoice(titles, a.Title)
	}
	slices.Sort(titles)
	return titles
}

func requestScope(request files.Request) files.Filter {
	var scope files.Filter
	for _, document := range request.Documents {
		for _, a := range document.Assignments {
			scope.Degrees = uniqueChoice(scope.Degrees, a.Degree)
			scope.Programs = uniqueChoice(scope.Programs, a.Program)
			scope.Versions = uniqueChoice(scope.Versions, a.Version)
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

func uploadOptions(values, selected []string) []FilterOption {
	var options []FilterOption
	for _, value := range values {
		options = append(options, FilterOption{Name: value, Selected: slices.Contains(selected, value)})
	}
	return options
}

func assignmentKey(a files.Assignment) string { return a.ModuleID + "|" + a.Version }

func selectedAssignments(values []string) []files.Assignment {
	var result []files.Assignment
	for _, value := range values {
		module, version, ok := strings.Cut(value, "|")
		if ok {
			result = append(result, files.Assignment{ModuleID: module, Version: version})
		}
	}
	return result
}

func uploadGroups(catalog files.Catalog, scope files.Filter, selected []files.Assignment) ([]AssignmentGroup, map[string]files.Assignment) {
	var groups []AssignmentGroup
	indices := map[string]int{}
	allowed := map[string]files.Assignment{}
	for _, module := range catalog.Modules() {
		program, degree := module.Qualification()
		if !slices.Contains(scope.Programs, program) || !slices.Contains(scope.Degrees, degree) {
			continue
		}
		for _, version := range catalog.ProgramVersions(module.Program) {
			if !slices.Contains(scope.Versions, version) {
				continue
			}
			a := files.Assignment{ModuleID: module.ID, Version: version, Title: module.Title, Program: program, Degree: degree}
			key := assignmentKey(a)
			allowed[key] = a
			index, ok := indices[module.Title]
			if !ok {
				index = len(groups)
				indices[module.Title] = index
				groups = append(groups, AssignmentGroup{Title: module.Title})
			}
			checked := slices.ContainsFunc(selected, func(value files.Assignment) bool { return assignmentKey(value) == key })
			groups[index].add(a, checked)
		}
	}
	sortAssignmentGroups(groups)
	return groups, allowed
}
