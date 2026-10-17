package archive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/storage"
)

type Exam struct {
	Title, Program, Version, Date, Description string
	Pages                                      int
	Degree                                     string
	ModuleID                                   string
	ID                                         string
	Bytes                                      int64
	OwnerID                                    int64
	Number                                     int
	Assignments                                []Assignment
	alias                                      string
	programs, degrees, versions                []string
}

func (e Exam) CanRevise(userID int64, role string) bool {
	return e.ID != "" && userID != 0 && (e.OwnerID == userID || role == "admin" || role == "editor")
}

type Filter struct {
	Query    string
	Programs []string
	Versions []string
	Degrees  []string
}

type Module struct {
	ID, Title            string
	Program              string
	Exams                []Exam
	StudyProgram, Degree string
}

func (m Module) Qualification() (string, string) {
	if m.StudyProgram != "" && m.Degree != "" {
		return m.StudyProgram, m.Degree
	}
	return ProgramName(m.Program)
}

func (m Module) Programs() []string {
	var programs []string
	for _, exam := range m.Exams {
		if len(exam.Assignments) > 0 {
			for _, assignment := range exam.Assignments {
				if !slices.Contains(programs, assignment.Program) {
					programs = append(programs, assignment.Program)
				}
			}
			continue
		}
		if !slices.Contains(programs, exam.Program) {
			programs = append(programs, exam.Program)
		}
	}
	return programs
}

type Catalog struct {
	exams                       []Exam
	modules                     []Module
	programs, versions, degrees []string
	programVersions             map[string][]string
}

func (c Catalog) ProgramVersions(program string) []string {
	return slices.Clone(c.programVersions[program])
}

func mergeExam(target *Exam, value Exam) {
	for _, a := range value.Assignments {
		if !slices.ContainsFunc(target.Assignments, func(existing Assignment) bool {
			return existing.ModuleID == a.ModuleID && existing.Version == a.Version
		}) {
			target.Assignments = append(target.Assignments, a)
		}
	}
	if target.programs == nil {
		target.programs = []string{target.Program}
		target.degrees = []string{target.Degree}
		target.versions = []string{target.Version}
	}
	target.programs = appendUnique(target.programs, value.Program)
	target.degrees = appendUnique(target.degrees, value.Degree)
	target.versions = appendUnique(target.versions, value.Version)
	target.Program = strings.Join(target.programs, ", ")
	target.Degree = strings.Join(target.degrees, ", ")
	labels := make([]string, len(target.versions))
	for i, value := range target.versions {
		labels[i] = value
	}
	target.Version = strings.Join(labels, ", ")
}

func (c Catalog) matchingExams(filter Filter) []Exam {
	result := make([]Exam, 0)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	for _, exam := range c.exams {
		if len(filter.Degrees) > 0 && !slices.Contains(filter.Degrees, exam.Degree) {
			continue
		}
		if len(filter.Programs) > 0 && !slices.Contains(filter.Programs, exam.Program) {
			continue
		}
		if len(filter.Versions) > 0 && !slices.Contains(filter.Versions, exam.Version) {
			continue
		}
		if !strings.Contains(strings.ToLower(exam.Title+" "+exam.Description+" "+exam.alias), query) {
			continue
		}
		result = append(result, exam)
	}
	slices.SortStableFunc(result, func(a, b Exam) int {
		aDate, _ := time.Parse("02.01.2006", a.Date)
		bDate, _ := time.Parse("02.01.2006", b.Date)
		return bDate.Compare(aDate)
	})
	return result
}

func (c Catalog) Programs() []string { return slices.Clone(c.programs) }

func (c Catalog) Versions() []string { return slices.Clone(c.versions) }

func (c Catalog) Degrees() []string { return slices.Clone(c.degrees) }

func (c Catalog) FindModule(id string, filter Filter) (Module, bool) {
	known := c.modules
	if known == nil {
		known = c.FindModules(Filter{})
	}
	for _, module := range known {
		if module.ID != id {
			continue
		}
		for _, matching := range c.FindModules(filter) {
			if matching.ID == id || c.modules != nil && matching.Title == module.Title {
				return matching, true
			}
		}
		module.Exams = nil
		return module, true
	}
	return Module{}, false
}

func (c Catalog) FindModules(filter Filter) []Module {
	modules := make([]Module, 0)
	indices := make(map[string]int)
	for _, exam := range c.matchingExams(filter) {
		key := exam.ModuleID
		if c.modules != nil {
			key = exam.Title
		}
		index, exists := indices[key]
		if !exists {
			index = len(modules)
			indices[key] = index
			moduleID := exam.ModuleID
			if c.modules != nil {
				for _, m := range c.modules {
					if m.Title == exam.Title {
						moduleID = m.ID
						break
					}
				}
			}
			modules = append(modules, Module{ID: moduleID, Title: exam.Title})
		}
		found := false
		for i := range modules[index].Exams {
			if exam.ID != "" && modules[index].Exams[i].ID == exam.ID {
				mergeExam(&modules[index].Exams[i], exam)
				found = true
				break
			}
		}
		if found {
			continue
		}
		modules[index].Exams = append(modules[index].Exams, exam)
	}
	return modules
}

func (c Catalog) FindExams(filter Filter) []Exam {
	var result []Exam
	indices := map[string]int{}
	for _, exam := range c.matchingExams(filter) {
		if i, ok := indices[exam.ID]; exam.ID != "" && ok {
			mergeExam(&result[i], exam)
			continue
		}
		indices[exam.ID] = len(result)
		result = append(result, exam)
	}
	return result
}

var ErrNotFound = errors.New("Die Datei wurde nicht gefunden.")

var ErrStorage = errors.New("Der Dateispeicher ist nicht erreichbar.")

type Storage interface {
	GetObject(context.Context, string) (*storage.Object, error)
}

type Service struct {
	db      *sql.DB
	storage Storage
}

func NewService(db *sql.DB, storage Storage) *Service { return &Service{db: db, storage: storage} }

// ProgramName separates degree labels used by the independent filter groups.
// Unknown suffixes are retained rather than guessing a degree.
func ProgramName(name string) (string, string) {
	for _, suffix := range []struct{ value, degree string }{
		{" (B. Sc.)", "Bachelor"}, {" (M. Sc.)", "Master"},
		{" (B.Sc.)", "Bachelor"}, {" (M.Sc.)", "Master"},
	} {
		if base, ok := strings.CutSuffix(name, suffix.value); ok {
			return base, suffix.degree
		}
	}
	return name, ""
}

func appendUnique(values []string, value string) []string {
	if value != "" && !slices.Contains(values, value) {
		return append(values, value)
	}
	return values
}

func (s *Service) Catalog(ctx context.Context) (Catalog, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Catalog{}, err
	}
	defer tx.Rollback()
	q := database.New(tx)
	programs, err := q.ListPrograms(ctx)
	if err != nil {
		return Catalog{}, err
	}
	result := Catalog{modules: []Module{}, programVersions: map[string][]string{}}
	versions, err := q.ListProgramVersions(ctx)
	if err != nil {
		return Catalog{}, err
	}
	for _, v := range versions {
		result.versions = appendUnique(result.versions, v.Name)
		for _, p := range programs {
			if p.ID == v.ProgramID {
				result.programVersions[p.Name] = appendUnique(result.programVersions[p.Name], v.Name)
			}
		}
	}
	modules, err := q.ListArchiveModules(ctx)
	if err != nil {
		return Catalog{}, err
	}
	for _, m := range modules {
		module := Module{ID: strconv.FormatInt(m.ID, 10), Title: m.Name, Program: m.ProgramName, StudyProgram: m.StudyProgram, Degree: m.Degree}
		result.modules = append(result.modules, module)
		program, degree := module.Qualification()
		result.programs = appendUnique(result.programs, program)
		result.degrees = appendUnique(result.degrees, degree)
	}
	entries, err := q.ListArchiveCatalog(ctx)
	if err != nil {
		return Catalog{}, err
	}
	for _, entry := range entries {
		program, degree := entry.ProgramName, entry.Degree
		if degree == "" {
			program, degree = ProgramName(program)
		}
		date, err := parseDate(entry.ExamDate)
		if err != nil {
			return Catalog{}, fmt.Errorf("archive entry %d: %w", entry.ID, err)
		}
		exam := Exam{ID: strconv.FormatInt(entry.ID, 10), ModuleID: strconv.FormatInt(entry.ModuleID, 10), Title: entry.ModuleName, Program: program, Degree: degree, Version: entry.Version, Date: date.Format("2006-01-02"), Bytes: entry.Nbytes, OwnerID: entry.OwnerID}
		exam.Assignments = []Assignment{{ModuleID: exam.ModuleID, Version: exam.Version, Title: exam.Title, Program: program, Degree: degree}}
		if entry.Comment != nil {
			exam.Description = *entry.Comment
		}
		if entry.Alias != nil {
			exam.alias = *entry.Alias
		}
		result.exams = append(result.exams, exam)
		result.versions = appendUnique(result.versions, exam.Version)
	}
	slices.Sort(result.programs)
	slices.Sort(result.degrees)
	slices.Sort(result.versions)
	slices.Reverse(result.versions)
	if err := tx.Commit(); err != nil {
		return Catalog{}, err
	}
	return result, nil
}

func (c Catalog) Modules() []Module { return slices.Clone(c.modules) }

func parseDate(value string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", time.RFC3339, "02.01.2006"} {
		if date, err := time.Parse(layout, value); err == nil {
			return date, nil
		}
	}
	return time.Time{}, errors.New("invalid exam date")
}

// Open retains the V1 key scheme and allows explicitly linked older revisions.
func (s *Service) Open(ctx context.Context, userID int64, value string) (*storage.Object, error) {
	role, err := accountRole(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return nil, ErrNotFound
	}
	var size int64
	var ownerID int64
	var status Status
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE archive_blob_cleanup SET blob_key=blob_key WHERE 0"); err != nil {
		return nil, err
	}
	var key, checksum string
	err = tx.QueryRowContext(ctx, `SELECT a.nbytes,a.user_id,COALESCE(r.status,'approved'),COALESCE(o.blob_key,''),a.checksum FROM archive a LEFT JOIN archive_objects o ON o.checksum=a.checksum AND o.nbytes=a.nbytes LEFT JOIN archive_requests r ON r.id=a.request_id WHERE a.id=? AND a.deleted_at IS NULL AND a.mime_type='application/pdf'
    AND EXISTS(SELECT 1 FROM archive_placements b JOIN modules m ON m.id=b.module_id AND m.deleted_at IS NULL JOIN programs p ON p.id=m.program_id AND p.deleted_at IS NULL WHERE b.archive_id=COALESCE((SELECT canonical_id FROM archive_entry_aliases WHERE legacy_id=a.id),a.id))`, id).Scan(&size, &ownerID, &status, &key, &checksum)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != Approved && ownerID != userID && !(status != Draft && canReview(role)) {
		return nil, ErrNotFound
	}
	if s.storage == nil {
		return nil, ErrStorage
	}
	verify := key != ""
	if key == "" {
		key = storage.ArchiveSourceKey(strconv.FormatInt(id, 10))
	}
	object, err := s.storage.GetObject(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if object.Info.Size != size {
		object.Close()
		return nil, fmt.Errorf("%w: stored size differs from archive record %d", ErrStorage, id)
	}
	contents, readErr := io.ReadAll(io.LimitReader(object, size+1))
	readErr = errors.Join(readErr, object.Close())
	if readErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorage, readErr)
	}
	if int64(len(contents)) != size || verify && fmt.Sprintf("%x", sha256.Sum256(contents)) != checksum {
		return nil, fmt.Errorf("%w: stored checksum or size differs", ErrStorage)
	}
	return &storage.Object{ReadCloser: bufferedObject{Reader: bytes.NewReader(contents)}, Info: object.Info}, nil
}

type bufferedObject struct{ *bytes.Reader }

func (bufferedObject) Close() error { return nil }
