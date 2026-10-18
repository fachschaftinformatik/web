package media

import "context"

// Carousel returns one slide per explicitly selected photograph. Public albums
// and their contents remain independent of this homepage-only selection.
func (s *Service) Carousel(ctx context.Context) ([]Album, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT CAST(e.id AS TEXT),e.title,CAST(m.id AS TEXT),COALESCE(m.title,''),COALESCE(m.description,'')
FROM media m JOIN events e ON e.id=m.event_id
WHERE e.deleted_at IS NULL AND m.deleted_at IS NULL AND m.homepage_visible=1 AND m.mime_type IN `+imageTypes+`
ORDER BY m.created_at DESC,m.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slides []Album
	for rows.Next() {
		var a Album
		var id, title, description string
		if err := rows.Scan(&a.ID, &a.Title, &id, &title, &description); err != nil {
			return nil, err
		}
		a.InternalID = a.ID
		a.Cover = albumPhoto(id, title, description, len(slides)+1, a.Title)
		slides = append(slides, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range slides {
		if err := s.albumReferences(ctx, &slides[i]); err != nil {
			return nil, err
		}
	}
	return slides, nil
}
