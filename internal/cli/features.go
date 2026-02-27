package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type templateDef struct {
	Kind         string
	DefaultTitle string
	BodyTemplate string
}

func templateDefinitions() map[string]templateDef {
	return map[string]templateDef{
		"note": {
			Kind:         "note",
			DefaultTitle: "Untitled",
			BodyTemplate: "",
		},
		"adr": {
			Kind:         "adr",
			DefaultTitle: "Architecture decision",
			BodyTemplate: "## Context\n\n## Decision\n\n## Consequences\n",
		},
		"meeting": {
			Kind:         "note",
			DefaultTitle: "Meeting notes",
			BodyTemplate: "## Date\n\n## Attendees\n\n## Notes\n\n## Action items\n- [ ] ",
		},
		"bug": {
			Kind:         "note",
			DefaultTitle: "Bug report",
			BodyTemplate: "## Symptoms\n\n## Steps to reproduce\n\n## Root cause\n\n## Fix\n\n## Validation\n",
		},
	}
}

func templateBody(def templateDef, extraText string) string {
	base := strings.TrimSpace(def.BodyTemplate)
	extra := strings.TrimSpace(extraText)

	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	return base + "\n\n" + extra
}

func resolveDailyDate(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, errors.New("daily --date must use YYYY-MM-DD")
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.UTC), nil
}

func defaultDailyBody(date time.Time) string {
	return strings.TrimSpace(fmt.Sprintf(`
## Plan (%s)

- [ ] Top priority 1
- [ ] Top priority 2

## Notes

## Wins

## Follow-ups
`, date.Format("2006-01-02")))
}
