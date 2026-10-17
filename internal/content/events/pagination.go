package events

import (
	"context"
	"time"
)

// FindPage returns upcoming events first, then history, with stable tie-breaks.
// The home page requests only upcoming records with a smaller limit.
func (s *Service) FindPage(ctx context.Context, now time.Time, page, size int, upcomingOnly bool) ([]Event, int, error) {
	if zone, err := time.LoadLocation("Europe/Berlin"); err == nil {
		now = now.In(zone)
	}
	today := now.Format("2006-01-02")
	date := `COALESCE(NULLIF(event_date,''),NULLIF(source_published_at,''),'')`
	upcoming := `(source_status!='cancelled' AND (` + date + `='' OR substr(` + date + `,1,10)>=?))`
	where := ""
	args := []any{}
	if upcomingOnly {
		where = " AND " + upcoming
		args = append(args, today)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM calendar_events WHERE deleted_at IS NULL`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	size = min(100, max(1, size))
	page = max(1, min(page, max(1, (total+size-1)/size)))
	args = append(args, today, today, size, (page-1)*size)
	rows, err := s.db.QueryContext(ctx, selectEvent+where+` ORDER BY CASE WHEN `+upcoming+` THEN 0 ELSE 1 END,CASE WHEN `+upcoming+` THEN CASE WHEN `+date+`='' THEN '9999' ELSE `+date+` END END ASC,`+date+` DESC,id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []Event
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, e)
	}
	return result, total, rows.Err()
}
