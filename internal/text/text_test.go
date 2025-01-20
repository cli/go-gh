package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeFormatFunc(t *testing.T) {
	_, err := TimeFormatFunc("Mon, 02 Jan 2006 15:04:05 MST", "invalid")
	require.Error(t, err)

	actual, err := TimeFormatFunc("Mon, 02 Jan 2006 15:04:05 MST", "2025-01-20T01:08:15Z")
	require.NoError(t, err)
	assert.Equal(t, "Mon, 20 Jan 2025 01:08:15 UTC", actual)
}

func TestTimeAgoFunc(t *testing.T) {
	const form = "2006-Jan-02 15:04:05"
	now, _ := time.Parse(form, "2020-Nov-22 14:00:00")
	cases := map[string]string{
		"2020-11-22T14:00:00Z": "just now",
		"2020-11-22T13:59:30Z": "just now",
		"2020-11-22T13:59:00Z": "1 minute ago",
		"2020-11-22T13:30:00Z": "30 minutes ago",
		"2020-11-22T13:00:00Z": "1 hour ago",
		"2020-11-22T02:00:00Z": "12 hours ago",
		"2020-11-21T14:00:00Z": "1 day ago",
		"2020-11-07T14:00:00Z": "15 days ago",
		"2020-10-24T14:00:00Z": "29 days ago",
		"2020-10-23T14:00:00Z": "1 month ago",
		"2020-09-23T14:00:00Z": "2 months ago",
		"2019-11-22T14:00:00Z": "1 year ago",
		"2018-11-22T14:00:00Z": "2 years ago",
	}
	for createdAt, expected := range cases {
		relative, err := TimeAgoFunc(now, createdAt)
		require.NoError(t, err)
		assert.Equal(t, expected, relative)
	}

	_, err := TimeAgoFunc(now, "invalid")
	assert.Error(t, err)
}
