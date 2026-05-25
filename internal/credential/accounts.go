package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"EmailTools/internal/provider"

	"github.com/google/uuid"
)

// AccountInfo is exposed metadata (no secrets).
type AccountInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Label     string `json:"label"`
	Provider  string `json:"provider"`
	UpdatedAt string `json:"updatedAt"`
}

type accountEntry struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	Label              string `json:"label"`
	Provider           string `json:"provider"`
	AuthCodeCipher     []byte `json:"authCodeCipher,omitempty"`
	OAuthRefreshCipher []byte `json:"oauthRefreshCipher,omitempty"`
	UpdatedAt          string `json:"updatedAt"`
}

type accountsFile struct {
	Version  int            `json:"version"`
	ActiveID string         `json:"activeId"`
	Accounts []accountEntry `json:"accounts"`
}

// Credentials holds decrypted secrets for an account.
type Credentials struct {
	Provider     string
	Email        string
	AppPassword  string
	RefreshToken string
}

var (
	accountsMu   sync.Mutex
	accountsPath string
	legacyPath   string
)

func initAccounts() {
	if accountsPath != "" {
		return
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	base := filepath.Join(dir, "EmailTools")
	accountsPath = filepath.Join(base, "accounts.json")
	legacyPath = filepath.Join(base, "config.json")
}

func ensureMigrated() error {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	return ensureMigratedLocked()
}

func ensureMigratedLocked() error {
	initAccounts()

	if _, err := os.Stat(accountsPath); err == nil {
		return migrateAccountsV2Locked()
	}
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(legacyPath)
	if err != nil {
		return err
	}
	var sf storedFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}
	if !sf.Remember || sf.Email == "" || len(sf.AuthCodeCipher) == 0 {
		_ = os.Remove(legacyPath)
		return nil
	}

	id := uuid.NewString()
	af := accountsFile{
		Version:  2,
		ActiveID: id,
		Accounts: []accountEntry{{
			ID:             id,
			Email:          sf.Email,
			Label:          defaultLabel(sf.Email),
			Provider:       provider.QQ,
			AuthCodeCipher: sf.AuthCodeCipher,
			UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
		}},
	}
	if err := writeAccountsLocked(&af); err != nil {
		return err
	}
	_ = os.Remove(legacyPath)
	return nil
}

func migrateAccountsV2Locked() error {
	af, err := loadAccountsLockedNoMigrate()
	if err != nil {
		return err
	}
	changed := false
	if af.Version < 2 {
		af.Version = 2
		changed = true
	}
	for i := range af.Accounts {
		if af.Accounts[i].Provider == "" {
			af.Accounts[i].Provider = provider.QQ
			changed = true
		}
	}
	if changed {
		return writeAccountsLocked(af)
	}
	return nil
}

func loadAccountsLockedNoMigrate() (*accountsFile, error) {
	initAccounts()
	data, err := os.ReadFile(accountsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &accountsFile{Version: 2, Accounts: []accountEntry{}}, nil
		}
		return nil, err
	}
	var af accountsFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, err
	}
	if af.Version == 0 {
		af.Version = 1
	}
	return &af, nil
}

func loadAccountsLocked() (*accountsFile, error) {
	if err := ensureMigratedLocked(); err != nil {
		return nil, err
	}
	af, err := loadAccountsLockedNoMigrate()
	if err != nil {
		return nil, err
	}
	if af.Version < 2 {
		_ = migrateAccountsV2Locked()
		return loadAccountsLockedNoMigrate()
	}
	return af, nil
}

func writeAccountsLocked(af *accountsFile) error {
	if err := os.MkdirAll(filepath.Dir(accountsPath), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(af, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(accountsPath, data, 0o600)
}

func defaultLabel(email string) string {
	if email == "" {
		return "未命名"
	}
	return email
}

func normalizeProvider(p string) (string, error) {
	return provider.Normalize(p)
}

// ListAccounts returns all saved accounts (no secrets).
func ListAccounts() ([]AccountInfo, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return nil, err
	}
	out := make([]AccountInfo, len(af.Accounts))
	for i, a := range af.Accounts {
		p, _ := normalizeProvider(a.Provider)
		out[i] = AccountInfo{
			ID:        a.ID,
			Email:     a.Email,
			Label:     a.Label,
			Provider:  p,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return out, nil
}

// HasAccounts reports whether any account is saved.
func HasAccounts() (bool, error) {
	list, err := ListAccounts()
	return len(list) > 0, err
}

// GetActiveID returns the last active account id.
func GetActiveID() (string, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return "", err
	}
	return af.ActiveID, nil
}

// SetActiveID sets the active account without validating IMAP.
func SetActiveID(id string) error {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return err
	}
	if !findAccount(af, id) {
		return fmt.Errorf("账号不存在")
	}
	af.ActiveID = id
	return writeAccountsLocked(af)
}

// GetActiveCredentials returns credentials for the active account.
func GetActiveCredentials() (Credentials, bool, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return Credentials{}, false, err
	}
	if af.ActiveID == "" {
		return Credentials{}, false, nil
	}
	for _, a := range af.Accounts {
		if a.ID == af.ActiveID {
			c, ok, err := decryptEntry(a)
			return c, ok, err
		}
	}
	return Credentials{}, false, nil
}

// GetCredentials returns credentials for a specific account id.
func GetCredentials(id string) (Credentials, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return Credentials{}, err
	}
	for _, a := range af.Accounts {
		if a.ID == id {
			c, ok, err := decryptEntry(a)
			if err != nil {
				return Credentials{}, err
			}
			if !ok {
				return Credentials{}, fmt.Errorf("账号凭据无效")
			}
			return c, nil
		}
	}
	return Credentials{}, fmt.Errorf("账号不存在")
}

func decryptEntry(a accountEntry) (Credentials, bool, error) {
	if a.Email == "" {
		return Credentials{}, false, nil
	}
	p, err := normalizeProvider(a.Provider)
	if err != nil {
		return Credentials{}, false, err
	}
	c := Credentials{Provider: p, Email: a.Email}

	if len(a.AuthCodeCipher) == 0 {
		if len(a.OAuthRefreshCipher) > 0 {
			return Credentials{}, false, fmt.Errorf("该 Outlook 账号使用旧版 OAuth 登录，请在编辑账号中填写应用密码")
		}
		return Credentials{}, false, nil
	}
	plain, err := unprotect(a.AuthCodeCipher)
	if err != nil {
		return Credentials{}, false, fmt.Errorf("解密授权码失败，请重新添加账号：%w", err)
	}
	c.AppPassword = string(plain)
	return c, true, nil
}

// UpsertAccount saves or updates an app-password account; returns account id.
func UpsertAccount(providerID, email, authCode, label string) (string, error) {
	email = strings.TrimSpace(email)
	authCode = strings.TrimSpace(authCode)
	label = strings.TrimSpace(label)
	if email == "" || authCode == "" {
		return "", fmt.Errorf("邮箱地址和授权码不能为空")
	}
	p, err := normalizeProvider(providerID)
	if err != nil {
		return "", err
	}
	if label == "" {
		label = defaultLabel(email)
	}

	cipher, err := protect([]byte(authCode))
	if err != nil {
		return "", fmt.Errorf("加密授权码失败：%w", err)
	}

	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var id string
	for i, a := range af.Accounts {
		if strings.EqualFold(a.Email, email) && accountProvider(a) == p {
			af.Accounts[i].AuthCodeCipher = cipher
			af.Accounts[i].OAuthRefreshCipher = nil
			af.Accounts[i].Label = label
			af.Accounts[i].Provider = p
			af.Accounts[i].UpdatedAt = now
			id = a.ID
			af.ActiveID = id
			return id, writeAccountsLocked(af)
		}
	}

	id = uuid.NewString()
	af.Accounts = append(af.Accounts, accountEntry{
		ID:             id,
		Email:          email,
		Label:          label,
		Provider:       p,
		AuthCodeCipher: cipher,
		UpdatedAt:      now,
	})
	af.ActiveID = id
	if af.Version < 2 {
		af.Version = 2
	}
	return id, writeAccountsLocked(af)
}

// UpdateAccount updates label and/or auth code for an existing account.
func UpdateAccount(id, authCode, label string) error {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return err
	}
	idx := -1
	for i, a := range af.Accounts {
		if a.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("账号不存在")
	}

	if label = strings.TrimSpace(label); label != "" {
		af.Accounts[idx].Label = label
	}
	if authCode = strings.TrimSpace(authCode); authCode != "" {
		cipher, err := protect([]byte(authCode))
		if err != nil {
			return fmt.Errorf("加密授权码失败：%w", err)
		}
		af.Accounts[idx].AuthCodeCipher = cipher
		af.Accounts[idx].OAuthRefreshCipher = nil
	}
	af.Accounts[idx].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return writeAccountsLocked(af)
}

// RemoveAccount deletes a saved account.
func RemoveAccount(id string) error {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	af, err := loadAccountsLocked()
	if err != nil {
		return err
	}
	var kept []accountEntry
	for _, a := range af.Accounts {
		if a.ID != id {
			kept = append(kept, a)
		}
	}
	if len(kept) == len(af.Accounts) {
		return fmt.Errorf("账号不存在")
	}
	af.Accounts = kept
	if af.ActiveID == id {
		if len(kept) > 0 {
			af.ActiveID = kept[len(kept)-1].ID
		} else {
			af.ActiveID = ""
		}
	}
	return writeAccountsLocked(af)
}

func findAccount(af *accountsFile, id string) bool {
	for _, a := range af.Accounts {
		if a.ID == id {
			return true
		}
	}
	return false
}

func accountProvider(a accountEntry) string {
	p, err := normalizeProvider(a.Provider)
	if err != nil {
		return provider.QQ
	}
	return p
}
