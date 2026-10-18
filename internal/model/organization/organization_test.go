package organization

import (
	"encoding/json"
	"testing"
)

func TestRoleFieldsPreservePersistedJSONKeys(t *testing.T) {
	var details Details
	if err := json.Unmarshal([]byte(`{"Name":"Original organization","ChairResponsibility":"chair-17","PublicityResponsibility":"publicity-29"}`), &details); err != nil {
		t.Fatal(err)
	}
	if details.ChairRole != "chair-17" || details.PublicityRole != "publicity-29" {
		t.Fatalf("persisted contacts lost: %+v", details)
	}
	details.ChairRole = "chair-18"
	raw, err := json.Marshal(details)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"ChairResponsibility": `"chair-18"`, "PublicityResponsibility": `"publicity-29"`, "Name": `"Original organization"`} {
		if string(fields[key]) != want {
			t.Errorf("persisted %s = %s, want %s", key, fields[key], want)
		}
	}
	for _, key := range []string{"ChairRole", "PublicityRole"} {
		if _, exists := fields[key]; exists {
			t.Errorf("introduced incompatible JSON key %s", key)
		}
	}
}
