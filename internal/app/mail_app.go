package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"EmailTools/internal/cache"
	"EmailTools/internal/credential"
	"EmailTools/internal/imap"
	mimeparse "EmailTools/internal/mime"
	"EmailTools/internal/provider"
	"EmailTools/internal/settings"
)

type MailApp struct {
	ctx context.Context
	mu  sync.Mutex

	repo            *imap.Repository
	activeAccountID string
	currentEmail    string
	lastFolder      string
}

func NewMailApp() *MailApp {
	return &MailApp{}
}

func (a *MailApp) Startup(ctx context.Context) {
	a.ctx = ctx
	_ = credential.MigrateLegacyIfNeeded()
}

func (a *MailApp) GetSettings() SettingsDTO {
	s := settings.Get()
	return SettingsDTO{
		FetchDays:              s.FetchDays,
		RefreshIntervalMinutes: s.RefreshIntervalMinutes,
		OpenLinksInBrowser:     s.OpenLinksInBrowser,
	}
}

func (a *MailApp) SaveSettings(dto SettingsDTO) error {
	return settings.Save(settings.Settings{
		FetchDays:              dto.FetchDays,
		RefreshIntervalMinutes: dto.RefreshIntervalMinutes,
		OpenLinksInBrowser:     dto.OpenLinksInBrowser,
	})
}

func (a *MailApp) ClearCache() error {
	return cache.Clear()
}

func (a *MailApp) ListAccounts() ([]AccountDTO, error) {
	activeID, _ := credential.GetActiveID()
	list, err := credential.ListAccounts()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, len(list))
	for i, acc := range list {
		out[i] = AccountDTO{
			ID:          acc.ID,
			Email:       acc.Email,
			Label:       acc.Label,
			Provider:    acc.Provider,
			ProviderTag: provider.DisplayName(acc.Provider),
			IsActive:    acc.ID == activeID,
		}
	}
	return out, nil
}

func (a *MailApp) HasAccounts() bool {
	ok, _ := credential.HasAccounts()
	return ok
}

// ListProviders returns supported mailbox providers for the add-account UI.
func (a *MailApp) ListProviders() []ProviderDTO {
	list := provider.List()
	out := make([]ProviderDTO, len(list))
	for i, p := range list {
		out[i] = ProviderDTO{
			ID:               p.ID,
			DisplayName:      p.DisplayName,
			AuthType:         p.AuthType,
			HelpURL:          p.HelpURL,
			EmailPlaceholder: p.EmailPlaceholder,
		}
	}
	return out
}

func (a *MailApp) AddAccount(providerID, email, authCode, label string) LoginResult {
	id, err := credential.UpsertAccount(providerID, email, authCode, label)
	if err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	return a.connectByID(id)
}

func (a *MailApp) SwitchAccount(id string) LoginResult {
	a.mu.Lock()
	if a.activeAccountID == id && a.repo != nil {
		email := a.currentEmail
		a.mu.Unlock()
		_ = credential.SetActiveID(id)
		return LoginResult{Success: true, Email: email}
	}
	a.mu.Unlock()

	if _, err := credential.GetCredentials(id); err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	if err := credential.SetActiveID(id); err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	return a.connectByID(id)
}

func (a *MailApp) UpdateAccount(id, authCode, label string) LoginResult {
	if err := credential.UpdateAccount(id, authCode, label); err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	a.mu.Lock()
	isActive := a.activeAccountID == id
	a.mu.Unlock()
	if !isActive {
		return LoginResult{Success: true}
	}
	if authCode != "" {
		return a.connectByID(id)
	}
	return LoginResult{Success: true, Email: a.GetCurrentEmail()}
}

func (a *MailApp) RemoveAccount(id string) error {
	a.mu.Lock()
	wasActive := a.activeAccountID == id
	if wasActive && a.repo != nil {
		_ = a.repo.Disconnect()
		a.repo = nil
		a.currentEmail = ""
		a.activeAccountID = ""
		a.lastFolder = ""
	}
	a.mu.Unlock()

	return credential.RemoveAccount(id)
}

func (a *MailApp) TryAutoLogin() LoginResult {
	_, ok, err := credential.GetActiveCredentials()
	if err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	if !ok {
		return LoginResult{Success: false}
	}
	activeID, _ := credential.GetActiveID()
	if activeID == "" {
		return LoginResult{Success: false}
	}
	return a.connectByID(activeID)
}

func (a *MailApp) Login(providerID, email, authCode string, remember bool) LoginResult {
	if !remember {
		cred := credential.Credentials{
			Provider:    providerID,
			Email:       email,
			AppPassword: authCode,
		}
		if cred.Provider == "" {
			cred.Provider = provider.SuggestFromEmail(email)
		}
		repo, err := a.dialCredentials("", cred)
		if err != nil {
			return LoginResult{Success: false, Error: err.Error()}
		}
		a.mu.Lock()
		if a.repo != nil {
			_ = a.repo.Disconnect()
		}
		a.repo = repo
		a.currentEmail = repo.Email()
		a.activeAccountID = ""
		a.lastFolder = ""
		a.mu.Unlock()
		return LoginResult{Success: true, Email: email}
	}
	return a.AddAccount(providerID, email, authCode, "")
}

func (a *MailApp) Logout() {
	a.mu.Lock()
	if a.repo != nil {
		_ = a.repo.Disconnect()
		a.repo = nil
	}
	a.currentEmail = ""
	a.activeAccountID = ""
	a.lastFolder = ""
	a.mu.Unlock()
}

func (a *MailApp) IsLoggedIn() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.repo != nil
}

func (a *MailApp) GetCurrentEmail() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentEmail
}

func (a *MailApp) GetSavedEmail() string {
	list, err := credential.ListAccounts()
	if err != nil || len(list) == 0 {
		return ""
	}
	return list[0].Email
}

func (a *MailApp) reconnectActive() error {
	a.mu.Lock()
	accountID := a.activeAccountID
	folder := a.lastFolder
	if a.repo != nil {
		_ = a.repo.Disconnect()
		a.repo = nil
	}
	a.mu.Unlock()

	if accountID == "" {
		return fmt.Errorf("连接已断开，请重新登录")
	}
	cred, err := credential.GetCredentials(accountID)
	if err != nil {
		return err
	}
	newRepo, err := a.dialCredentials(accountID, cred)
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.repo = newRepo
	a.currentEmail = newRepo.Email()
	a.mu.Unlock()

	if folder != "" {
		if _, err := newRepo.Select(folder); err != nil {
			return err
		}
	}
	return nil
}

func (a *MailApp) ensureConnected() error {
	a.mu.Lock()
	repo := a.repo
	a.mu.Unlock()
	if repo == nil {
		return fmt.Errorf("请先登录邮箱")
	}
	if err := repo.Ping(); err == nil {
		return nil
	}
	return a.reconnectActive()
}

func (a *MailApp) withRetry(fn func() error) error {
	if err := a.ensureConnected(); err != nil {
		return err
	}
	err := fn()
	if err == nil || !imap.IsConnectionError(err) {
		return err
	}
	if reconn := a.reconnectActive(); reconn != nil {
		return err
	}
	return fn()
}

func (a *MailApp) ListFolders() ([]FolderDTO, error) {
	var out []FolderDTO
	err := a.withRetry(func() error {
		repo, err := a.requireRepo()
		if err != nil {
			return err
		}
		boxes, err := repo.ListMailboxes()
		if err != nil {
			return err
		}
		out = make([]FolderDTO, len(boxes))
		for i, b := range boxes {
			out[i] = FolderDTO{
				Name:        b.Name,
				DisplayName: b.DisplayName,
				Delimiter:   b.Delimiter,
			}
		}
		return nil
	})
	if err == nil && len(out) > 0 {
		a.persistFolderCache(a.getActiveAccountID(), out)
	}
	return out, err
}

func (a *MailApp) sinceFromSettings() time.Time {
	days := settings.GetFetchDays()
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days)
}

func (a *MailApp) getActiveAccountID() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.activeAccountID
}

func summariesToDTO(list []cache.Summary) []MessageSummaryDTO {
	out := make([]MessageSummaryDTO, len(list))
	for i, m := range list {
		out[i] = MessageSummaryDTO{
			UID:      m.UID,
			Subject:  m.Subject,
			From:     m.From,
			Date:     m.Date,
			DateUnix: m.DateUnix,
			Seen:     m.Seen,
		}
	}
	return out
}

func dtoToSummaries(list []MessageSummaryDTO) []cache.Summary {
	out := make([]cache.Summary, len(list))
	for i, m := range list {
		out[i] = cache.Summary{
			UID:      m.UID,
			Subject:  m.Subject,
			From:     m.From,
			Date:     m.Date,
			DateUnix: m.DateUnix,
			Seen:     m.Seen,
		}
	}
	return out
}

// ListMessagesCached returns locally cached list items (may be stale).
func (a *MailApp) ListMessagesCached(folder string) ([]MessageSummaryDTO, error) {
	_ = cache.PruneOlderThan(7)
	accountID := a.getActiveAccountID()
	if accountID == "" {
		return nil, nil
	}
	cached, err := cache.LoadMessages(accountID, folder, a.sinceFromSettings())
	if err != nil {
		return nil, err
	}
	return summariesToDTO(cached), nil
}

func (a *MailApp) ListMessages(folder string, offset, limit int) ([]MessageSummaryDTO, error) {
	_ = offset
	_ = limit
	since := a.sinceFromSettings()
	maxResults := 500
	accountID := a.getActiveAccountID()
	_ = cache.PruneOlderThan(7)

	var out []MessageSummaryDTO
	err := a.withRetry(func() error {
		repo, err := a.requireRepo()
		if err != nil {
			return err
		}
		if _, err := repo.Select(folder); err != nil {
			return err
		}
		a.mu.Lock()
		a.lastFolder = folder
		a.mu.Unlock()

		headers, err := repo.FetchHeadersSince(since, maxResults)
		if err != nil {
			return err
		}
		out = make([]MessageSummaryDTO, len(headers))
		for i, h := range headers {
			out[i] = MessageSummaryDTO{
				UID:      h.UID,
				Subject:  h.Subject,
				From:     h.From,
				Date:     h.Date,
				DateUnix: h.DateUnix,
				Seen:     h.Seen,
			}
		}
		if accountID != "" && len(out) > 0 {
			_ = cache.SaveMessages(accountID, folder, dtoToSummaries(out))
		}
		return nil
	})
	if err == nil && accountID != "" && folder != "" {
		a.updateCachedLastFolder(accountID, folder)
	}
	return out, err
}

func (a *MailApp) GetMessage(folder string, uid uint32) (*MessageDetailDTO, error) {
	var result *MessageDetailDTO
	err := a.withRetry(func() error {
		repo, err := a.requireRepo()
		if err != nil {
			return err
		}
		if _, err := repo.Select(folder); err != nil {
			return err
		}
		a.mu.Lock()
		a.lastFolder = folder
		a.mu.Unlock()

		raw, err := repo.FetchBodyByUID(uid)
		if err != nil {
			return err
		}
		parsed, err := mimeparse.ParseRFC822(raw)
		if err != nil {
			return fmt.Errorf("解析邮件失败：%w", err)
		}

		attachments := make([]AttachmentMetaDTO, len(parsed.Attachments))
		for i, att := range parsed.Attachments {
			attachments[i] = AttachmentMetaDTO{
				Filename:    att.Filename,
				ContentType: att.ContentType,
				Size:        att.Size,
			}
		}

		result = &MessageDetailDTO{
			UID:         uid,
			Subject:     parsed.Subject,
			From:        parsed.From,
			To:          parsed.To,
			Date:        parsed.Date,
			TextPlain:   parsed.TextPlain,
			TextHtml:    parsed.TextHtml,
			Attachments: attachments,
		}
		return nil
	})
	return result, err
}

func (a *MailApp) requireRepo() (*imap.Repository, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.repo == nil {
		return nil, fmt.Errorf("请先登录邮箱")
	}
	return a.repo, nil
}
