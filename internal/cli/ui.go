package cli

import (
	"fmt"
	"strings"
)

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func displayDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "-"
	}
	return domain
}

func displayTags(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ",")
}

func displayTagsCompact(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	if len(tags) <= 2 {
		return strings.Join(tags, ",")
	}
	return fmt.Sprintf("%s,%s,+%d", tags[0], tags[1], len(tags)-2)
}
