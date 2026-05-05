package grafana

import "testing"

func TestKnownPanels_MinCount(t *testing.T) {
	if len(KnownPanels) < 6 {
		t.Errorf("KnownPanels has %d entries; want at least 6", len(KnownPanels))
	}
}

func TestKnownPanels_NonEmptyFields(t *testing.T) {
	for i, p := range KnownPanels {
		if p.DashUID == "" {
			t.Errorf("KnownPanels[%d].DashUID is empty", i)
		}
		if p.Title == "" {
			t.Errorf("KnownPanels[%d].Title is empty", i)
		}
		if p.Slug == "" {
			t.Errorf("KnownPanels[%d].Slug is empty", i)
		}
		if p.PanelID <= 0 {
			t.Errorf("KnownPanels[%d].PanelID = %d; want > 0", i, p.PanelID)
		}
	}
}
