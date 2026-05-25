package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "EmailTools", "cache"), nil
}

func Clear() error {
	cacheDir, err := Dir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取缓存目录失败：%w", err)
	}
	for _, e := range entries {
		p := filepath.Join(cacheDir, e.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("清理缓存失败：%w", err)
		}
	}
	return nil
}
