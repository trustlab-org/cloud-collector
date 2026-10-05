package googleworkspace

import (
	"fmt"
	"time"
)

// defaultOverlaps follow Google's published lag times per log type, with headroom.
var defaultOverlaps = map[string]time.Duration{
	"login":             30 * time.Minute,
	"admin":             30 * time.Minute,
	"drive":             30 * time.Minute,
	"gmail":             30 * time.Minute,
	"meet":              30 * time.Minute,
	"saml":              30 * time.Minute,
	"rules":             30 * time.Minute,
	"user_accounts":     3 * time.Hour,
	"calendar":          3 * time.Hour,
	"groups":            3 * time.Hour,
	"groups_enterprise": 3 * time.Hour,
	"token":             4 * time.Hour,
	"takeout":           24 * time.Hour,
}

func overlaps(apps []string, overrides map[string]string, fallback time.Duration) (map[string]time.Duration, error) {
	out := make(map[string]time.Duration, len(apps))
	for _, app := range apps {
		d, ok := defaultOverlaps[app]
		if !ok {
			d = fallback
		}
		if raw, ok := overrides[app]; ok {
			var err error
			if d, err = time.ParseDuration(raw); err != nil || d < 0 {
				return nil, fmt.Errorf("overlaps.%s: invalid duration %q", app, raw)
			}
		}
		out[app] = d
	}
	for app := range overrides {
		if _, ok := out[app]; !ok {
			return nil, fmt.Errorf("overlaps.%s: not in applications", app)
		}
	}
	return out, nil
}
