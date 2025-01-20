package text

import (
	"time"

	"github.com/cli/go-gh/v2/pkg/text"
)

func TimeFormatFunc(format, input string) (string, error) {
	t, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

func TimeAgoFunc(now time.Time, input string) (string, error) {
	t, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return "", err
	}
	return timeAgo(now.Sub(t)), nil
}

func timeAgo(ago time.Duration) string {
	if ago < time.Minute {
		return "just now"
	}
	if ago < time.Hour {
		return text.Pluralize(int(ago.Minutes()), "minute") + " ago"
	}
	if ago < 24*time.Hour {
		return text.Pluralize(int(ago.Hours()), "hour") + " ago"
	}
	if ago < 30*24*time.Hour {
		return text.Pluralize(int(ago.Hours())/24, "day") + " ago"
	}
	if ago < 365*24*time.Hour {
		return text.Pluralize(int(ago.Hours())/24/30, "month") + " ago"
	}
	return text.Pluralize(int(ago.Hours()/24/365), "year") + " ago"
}
