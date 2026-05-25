package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	minFetchDays     = 1
	maxFetchDays     = 30
	defaultFetchDays = 7

	minRefreshMinutes     = 0
	maxRefreshMinutes     = 120
	defaultRefreshMinutes = 0
)

type Settings struct {
	FetchDays              int  `json:"fetchDays"`
	RefreshIntervalMinutes int  `json:"refreshIntervalMinutes"`
	OpenLinksInBrowser     bool `json:"openLinksInBrowser"`
}

var (
	mu           sync.Mutex
	settingsPath string
)

func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	settingsPath = filepath.Join(dir, "EmailTools", "settings.json")
}

func Get() Settings {
	mu.Lock()
	defer mu.Unlock()
	return loadLocked()
}

func GetFetchDays() int {
	return Get().FetchDays
}

func GetRefreshIntervalMinutes() int {
	return Get().RefreshIntervalMinutes
}

func GetOpenLinksInBrowser() bool {
	return Get().OpenLinksInBrowser
}

func Save(s Settings) error {
	if s.FetchDays < minFetchDays || s.FetchDays > maxFetchDays {
		return fmt.Errorf("读取天数须在 %d～%d 天之间", minFetchDays, maxFetchDays)
	}
	if s.RefreshIntervalMinutes < minRefreshMinutes || s.RefreshIntervalMinutes > maxRefreshMinutes {
		return fmt.Errorf("刷新间隔须在 %d～%d 分钟之间（0 表示关闭）", minRefreshMinutes, maxRefreshMinutes)
	}
	mu.Lock()
	defer mu.Unlock()
	return writeLocked(s)
}

func SaveFetchDays(days int) error {
	s := loadLocked()
	s.FetchDays = days
	return Save(s)
}

func loadLocked() Settings {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaults()
		}
		return defaults()
	}
	var raw struct {
		FetchDays              int   `json:"fetchDays"`
		RefreshIntervalMinutes int   `json:"refreshIntervalMinutes"`
		OpenLinksInBrowser     *bool `json:"openLinksInBrowser"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return defaults()
	}
	s := Settings{
		FetchDays:              raw.FetchDays,
		RefreshIntervalMinutes: raw.RefreshIntervalMinutes,
		OpenLinksInBrowser:     true,
	}
	if raw.OpenLinksInBrowser != nil {
		s.OpenLinksInBrowser = *raw.OpenLinksInBrowser
	}
	if s.FetchDays < minFetchDays || s.FetchDays > maxFetchDays {
		s.FetchDays = defaultFetchDays
	}
	if s.RefreshIntervalMinutes < minRefreshMinutes || s.RefreshIntervalMinutes > maxRefreshMinutes {
		s.RefreshIntervalMinutes = defaultRefreshMinutes
	}
	return s
}

func defaults() Settings {
	return Settings{
		FetchDays:              defaultFetchDays,
		RefreshIntervalMinutes: defaultRefreshMinutes,
		OpenLinksInBrowser:     true,
	}
}

func writeLocked(s Settings) error {
	if s.FetchDays < minFetchDays {
		s.FetchDays = minFetchDays
	}
	if s.FetchDays > maxFetchDays {
		s.FetchDays = maxFetchDays
	}
	if s.RefreshIntervalMinutes < minRefreshMinutes {
		s.RefreshIntervalMinutes = minRefreshMinutes
	}
	if s.RefreshIntervalMinutes > maxRefreshMinutes {
		s.RefreshIntervalMinutes = maxRefreshMinutes
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("保存设置失败：%w", err)
	}
	return nil
}
