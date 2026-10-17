package files

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

	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/data/blob"
)

var ErrNotFound = errors.New("Die Datei wurde nicht gefunden.")
var ErrStorage = errors.New("Der Dateispeicher ist nicht erreichbar.")

type Storage interface {
	GetObject(context.Context, string) (*blob.Object, error)
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
	q := data.New(tx)
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
		exam := Exam{ID: strconv.FormatInt(entry.ID, 10), ModuleID: strconv.FormatInt(entry.ModuleID, 10), Title: entry.ModuleName, Program: program, Degree: degree, Version: entry.Version, Date: date.Format("02.01.2006"), Bytes: entry.Nbytes, OwnerID: entry.OwnerID}
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
func (s *Service) Open(ctx context.Context, userID int64, value string) (*blob.Object, error) {
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
		key = blob.ArchiveSourceKey(strconv.FormatInt(id, 10))
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
	return &blob.Object{ReadCloser: bufferedObject{Reader: bytes.NewReader(contents)}, Info: object.Info}, nil
}

type bufferedObject struct{ *bytes.Reader }

func (bufferedObject) Close() error { return nil }
