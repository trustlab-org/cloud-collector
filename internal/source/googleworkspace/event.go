package googleworkspace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	reports "google.golang.org/api/admin/reports/v1"

	"github.com/HelixY2J/collector/internal/envelope"
)

type event struct {
	id   string
	at   time.Time
	body []byte
}

// split turns one activity into one record per events[] item, keeping Google's raw schema.
func split(a *reports.Activity, app string, meta envelope.Meta) ([]event, error) {
	if a.Id == nil {
		return nil, fmt.Errorf("activity without id")
	}
	at, err := time.Parse(time.RFC3339Nano, a.Id.Time)
	if err != nil {
		return nil, fmt.Errorf("activity time %q: %w", a.Id.Time, err)
	}
	out := make([]event, 0, len(a.Events))
	for i, ev := range a.Events {
		one := *a
		one.Events = []*reports.ActivityEvents{ev}
		raw, err := json.Marshal(&one)
		if err != nil {
			return nil, err
		}
		m := meta
		m.EventID = envelope.ID(a.Id.CustomerId, app, a.Id.Time, strconv.FormatInt(a.Id.UniqueQualifier, 10), strconv.Itoa(i), ev.Type, ev.Name)
		body, err := envelope.Envelope{Meta: m, JSON: raw}.Marshal()
		if err != nil {
			return nil, err
		}
		out = append(out, event{id: m.EventID, at: at, body: body})
	}
	return out, nil
}
