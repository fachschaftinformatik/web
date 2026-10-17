package meetings

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/fachschaftinformatik/web/internal/storage"
)

type historicalQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (s *Store) historical(ctx context.Context, db historicalQuerier) ([]Meeting, error) {
	rows, err := db.QueryContext(ctx, `SELECT year,number,date,kind,file,note,bytes,sha256,sources FROM historical_meeting_documents ORDER BY year DESC,date DESC,number DESC,file`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []Meeting
	for rows.Next() {
		var m Meeting
		var d Document
		var sources string
		if err := rows.Scan(&m.Year, &m.Number, &m.Date, &d.Kind, &d.File, &d.Note, &d.Bytes, &d.SHA256, &sources); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(sources), &d.Sources); err != nil {
			return nil, err
		}
		if len(records) == 0 || records[len(records)-1].Date != m.Date || records[len(records)-1].Number != m.Number {
			records = append(records, m)
		}
		records[len(records)-1].Documents = append(records[len(records)-1].Documents, d)
	}
	return records, rows.Err()
}

func (s *Store) HistoricalDocument(ctx context.Context, file string) (*storage.Object, error) {
	var key string
	if err := s.db.QueryRowContext(ctx, "SELECT blob_key FROM historical_meeting_documents WHERE file=?", file).Scan(&key); err != nil {
		return nil, err
	}
	return s.blobs.GetObject(ctx, key)
}
