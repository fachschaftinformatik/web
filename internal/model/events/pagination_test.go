package events

import (
	"context"
	"testing"
	"time"
)

func TestEventPagesOrderHistoryAndExcludeCancelledFromHome(t *testing.T) {
	s, db, _ := eventService(t)
	_, err := db.Exec(`INSERT INTO calendar_events(id,title,body,user_id,event_date,source_published_at,source_status) VALUES
 (10,'Future','Text',1,'2026-11-27','',''),
 (11,'Cancelled','Text',1,'2026-12-01','','cancelled'),
 (12,'Recap','Text',1,'','2025-09-01T12:00:00Z','recap'),
 (13,'Past','Text',1,'2024-01-01','',''),
 (14,'Date follows','Text',1,'','','')`)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	want := []string{"10", "14", "11", "12", "13"}
	for page, id := range want {
		records, total, err := s.FindPage(context.Background(), now, page+1, 1, false)
		if err != nil || total != 5 || len(records) != 1 || records[0].ID != id {
			t.Fatalf("page %d: %+v total %d %v", page+1, records, total, err)
		}
	}
	home, total, err := s.FindPage(context.Background(), now, 1, 3, true)
	if err != nil || total != 2 || len(home) != 2 || home[0].ID != "10" || home[1].ID != "14" {
		t.Fatalf("home: %+v %d %v", home, total, err)
	}
}
