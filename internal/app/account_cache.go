package app

import (
	"regexp"

	"EmailTools/internal/cache"
)

// AccountCacheDTO is offline cache for instant UI when switching accounts.
type AccountCacheDTO struct {
	Folders        []FolderDTO           `json:"folders"`
	LastFolder     string                `json:"lastFolder"`
	Messages       []MessageSummaryDTO   `json:"messages"`
}

var inboxFolderRe = regexp.MustCompile(`(?i)INBOX|收件箱`)

// PeekAccountCache reads folder/message cache without IMAP (for fast account switch UI).
func (a *MailApp) PeekAccountCache(accountID string) AccountCacheDTO {
	if accountID == "" {
		return AccountCacheDTO{}
	}
	_ = cache.PruneOlderThan(7)

	rawFolders, lastFolder, _ := cache.LoadFolders(accountID)
	folders := make([]FolderDTO, len(rawFolders))
	for i, f := range rawFolders {
		folders[i] = FolderDTO{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Delimiter:   f.Delimiter,
		}
	}
	if lastFolder == "" {
		lastFolder = pickInboxFolder(folders)
	}

	var messages []MessageSummaryDTO
	if lastFolder != "" {
		cached, err := cache.LoadMessages(accountID, lastFolder, a.sinceFromSettings())
		if err == nil && len(cached) > 0 {
			messages = summariesToDTO(cached)
		}
	}

	return AccountCacheDTO{
		Folders:    folders,
		LastFolder: lastFolder,
		Messages:   messages,
	}
}

func pickInboxFolder(folders []FolderDTO) string {
	for _, f := range folders {
		if inboxFolderRe.MatchString(f.Name) || inboxFolderRe.MatchString(f.DisplayName) {
			return f.Name
		}
	}
	if len(folders) > 0 {
		return folders[0].Name
	}
	return ""
}

func folderDTOsToCache(list []FolderDTO) []cache.FolderEntry {
	out := make([]cache.FolderEntry, len(list))
	for i, f := range list {
		out[i] = cache.FolderEntry{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Delimiter:   f.Delimiter,
		}
	}
	return out
}

func (a *MailApp) persistFolderCache(accountID string, folders []FolderDTO) {
	if accountID == "" || len(folders) == 0 {
		return
	}
	a.mu.Lock()
	last := a.lastFolder
	a.mu.Unlock()
	if last == "" {
		last = pickInboxFolder(folders)
	}
	_ = cache.SaveFolders(accountID, folderDTOsToCache(folders), last)
}

func (a *MailApp) updateCachedLastFolder(accountID, folder string) {
	if accountID == "" || folder == "" {
		return
	}
	cached, _, _ := cache.LoadFolders(accountID)
	if len(cached) == 0 {
		return
	}
	_ = cache.SaveFolders(accountID, cached, folder)
}
