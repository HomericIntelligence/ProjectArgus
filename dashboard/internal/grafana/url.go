package grafana

import "fmt"

// PanelURL builds a Grafana d-solo embed URL for the given Panel with kiosk
// mode enabled and the supplied time-range bounds.
func PanelURL(grafanaBase string, p Panel, from, to string) string {
	return fmt.Sprintf("%s/d-solo/%s/%s?panelId=%d&kiosk=tv&theme=dark&from=%s&to=%s",
		grafanaBase, p.DashUID, p.Slug, p.PanelID, from, to)
}
