package events

import (
	"context"
	"reflect"
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
	seedReferences(t, db)
	for page, id := range want {
		id = reference(t, db, "events", id)
		records, total, err := s.FindPage(context.Background(), now, page+1, 1, false)
		if err != nil || total != 5 || len(records) != 1 || records[0].ID != id {
			t.Fatalf("page %d: %+v total %d %v", page+1, records, total, err)
		}
	}
	home, total, err := s.FindPage(context.Background(), now, 1, 3, true)
	if err != nil || total != 2 || len(home) != 2 || home[0].ID != reference(t, db, "events", "10") || home[1].ID != reference(t, db, "events", "14") {
		t.Fatalf("home: %+v %d %v", home, total, err)
	}
	all, _, err := s.FindPage(context.Background(), now, 1, 100, false)
	if err != nil {
		t.Fatal(err)
	}
	upcoming, past := Group(all, now)
	if !reflect.DeepEqual(upcoming, home) || len(past) != 3 {
		t.Fatal("calendar grouping and the bounded upcoming query disagree")
	}
}

func TestUpcomingUTCTimestampsAtLocalMidnight(t *testing.T) {
	s, db, _ := eventService(t)
	if _, err := db.Exec(`INSERT INTO calendar_events(id,title,body,user_id,event_date) VALUES
	(20,'Today after midnight','Text',1,'2026-07-17T22:30:00Z'),
	(21,'Yesterday before midnight','Text',1,'2026-07-17T21:30:00Z'),
	(22,'All day today','Text',1,'2026-07-18')`); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 17, 23, 0, 0, 0, time.UTC)
	seedReferences(t, db)
	home, total, err := s.FindPage(context.Background(), now, 1, 6, true)
	if err != nil || total != 2 {
		t.Fatal("UTC timestamps were classified by UTC day instead of local day", home, total, err)
	}
	all, _, err := s.FindPage(context.Background(), now, 1, 100, false)
	if err != nil {
		t.Fatal(err)
	}
	upcoming, past := Group(all, now)
	if len(upcoming) != 2 || len(past) != 1 || past[0].ID != reference(t, db, "events", "21") {
		t.Fatal("UTC calendar grouping disagrees", upcoming, past)
	}
}
