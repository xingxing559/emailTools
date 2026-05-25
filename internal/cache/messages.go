package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const defaultRetentionDays = 7

// Summary matches list item fields stored on disk.
type Summary struct {
	UID      uint32 `json:"uid"`
	Subject  string `json:"subject"`
	From     string `json:"from"`
	Date     string `json:"date"`
	DateUnix int64  `json:"dateUnix"`
	Seen     bool   `json:"seen"`
}

type dayCache struct {
	Messages []Summary `json:"messages"`
	SavedAt  int64     `json:"savedAt"`
}

func messagesRoot() (string, error) {
	root, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "messages"), nil
}

func folderDir(accountID, folder string) (string, error) {
	root, err := messagesRoot()
	if err != nil {
		return "", err
	}
	safeFolder := sanitizePathSegment(folder)
	return filepath.Join(root, accountID, safeFolder), nil
}

func sanitizePathSegment(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			return '_'
		default:
			return r
		}
	}, s)
	if s == "" {
		return "_"
	}
	return s
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseDayKey(name string) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSuffix(name, ".json"), time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// PruneOlderThan removes cached day files older than retentionDays (default 7).
func PruneOlderThan(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}
	root, err := messagesRoot()
	if err != nil {
		return err
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		day, ok := parseDayKey(d.Name())
		if !ok {
			return nil
		}
		if day.Before(cutoff) {
			_ = os.Remove(path)
		}
		return nil
	})
}

// LoadMessages reads cached summaries for account/folder since the given time.
func LoadMessages(accountID, folder string, since time.Time) ([]Summary, error) {
	if accountID == "" {
		return nil, nil
	}
	dir, err := folderDir(accountID, folder)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	sinceUnix := since.Unix()
	byUID := make(map[uint32]Summary)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		day, ok := parseDayKey(e.Name())
		if !ok || day.Before(since) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var dc dayCache
		if err := json.Unmarshal(data, &dc); err != nil {
			continue
		}
		for _, m := range dc.Messages {
			if m.DateUnix < sinceUnix {
				continue
			}
			if prev, ok := byUID[m.UID]; !ok || m.DateUnix > prev.DateUnix {
				byUID[m.UID] = m
			}
		}
	}

	out := make([]Summary, 0, len(byUID))
	for _, m := range byUID {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DateUnix > out[j].DateUnix })
	return out, nil
}

// SaveMessages writes summaries split by local calendar day.
func SaveMessages(accountID, folder string, messages []Summary) error {
	if accountID == "" || len(messages) == 0 {
		return nil
	}
	dir, err := folderDir(accountID, folder)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("创建缓存目录失败：%w", err)
	}

	byDay := make(map[string][]Summary)
	for _, m := range messages {
		t := time.Unix(m.DateUnix, 0).In(time.Local)
		if m.DateUnix <= 0 {
			t = time.Now()
		}
		key := dayKey(t)
		byDay[key] = append(byDay[key], m)
	}

	now := time.Now().Unix()
	for day, msgs := range byDay {
		path := filepath.Join(dir, day+".json")
		existing := make(map[uint32]Summary)
		if data, err := os.ReadFile(path); err == nil {
			var dc dayCache
			if json.Unmarshal(data, &dc) == nil {
				for _, m := range dc.Messages {
					existing[m.UID] = m
				}
			}
		}
		for _, m := range msgs {
			if prev, ok := existing[m.UID]; !ok || m.DateUnix >= prev.DateUnix {
				existing[m.UID] = m
			}
		}
		merged := make([]Summary, 0, len(existing))
		for _, m := range existing {
			merged = append(merged, m)
		}
		sort.Slice(merged, func(i, j int) bool { return merged[i].DateUnix > merged[j].DateUnix })

		dc := dayCache{Messages: merged, SavedAt: now}
		data, err := json.Marshal(dc)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return fmt.Errorf("写入缓存失败：%w", err)
		}
	}
	return nil
}
