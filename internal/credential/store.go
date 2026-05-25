package credential

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config holds saved login data (auth code encrypted).
type Config struct {
	Email          string `json:"email"`
	AuthCodeCipher []byte `json:"authCodeCipher"`
	Remember       bool   `json:"remember"`
}

type storedFile struct {
	Email          string `json:"email"`
	AuthCodeCipher []byte `json:"authCodeCipher"`
	Remember       bool   `json:"remember"`
}

var (
	mu       sync.Mutex
	configPath string
)

func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	configPath = filepath.Join(dir, "EmailTools", "config.json")
}

// Save persists credentials when remember is true; clears storage otherwise.
func Save(email, authCode string, remember bool) error {
	mu.Lock()
	defer mu.Unlock()

	if !remember {
		return removeConfig()
	}

	cipher, err := protect([]byte(authCode))
	if err != nil {
		return fmt.Errorf("加密授权码失败：%w", err)
	}

	data, err := json.Marshal(storedFile{
		Email:          email,
		AuthCodeCipher: cipher,
		Remember:       true,
	})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o600)
}

// Load returns saved credentials if present.
func Load() (email, authCode string, ok bool, err error) {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", false, nil
		}
		return "", "", false, err
	}

	var sf storedFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return "", "", false, err
	}
	if !sf.Remember || sf.Email == "" || len(sf.AuthCodeCipher) == 0 {
		return "", "", false, nil
	}

	plain, err := unprotect(sf.AuthCodeCipher)
	if err != nil {
		return "", "", false, fmt.Errorf("解密授权码失败，请重新登录：%w", err)
	}
	return sf.Email, string(plain), true, nil
}

// Clear removes saved credentials.
func Clear() error {
	mu.Lock()
	defer mu.Unlock()
	return removeConfig()
}

func removeConfig() error {
	err := os.Remove(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// SavedEmail returns the stored email without loading the auth code.
func SavedEmail() (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", false
	}
	var sf storedFile
	if json.Unmarshal(data, &sf) != nil || !sf.Remember {
		return "", false
	}
	return sf.Email, sf.Email != ""
}

var errDPAPIUnavailable = errors.New("DPAPI 仅支持 Windows")
