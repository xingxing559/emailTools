package app

// AccountDTO is a saved mailbox account (no secrets).
type AccountDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Label       string `json:"label"`
	Provider    string `json:"provider"`
	ProviderTag string `json:"providerTag"`
	IsActive    bool   `json:"isActive"`
}

// ProviderDTO describes a supported mailbox provider.
type ProviderDTO struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	AuthType         string `json:"authType"`
	HelpURL          string `json:"helpUrl"`
	EmailPlaceholder string `json:"emailPlaceholder"`
}

// FolderDTO is exposed to the frontend.
type FolderDTO struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Delimiter   string `json:"delimiter"`
}

// MessageSummaryDTO is a list item.
type MessageSummaryDTO struct {
	UID      uint32 `json:"uid"`
	Subject  string `json:"subject"`
	From     string `json:"from"`
	Date     string `json:"date"`
	DateUnix int64  `json:"dateUnix"`
	Seen     bool   `json:"seen"`
}

// AttachmentMetaDTO describes an attachment.
type AttachmentMetaDTO struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

// MessageDetailDTO is full message content.
type MessageDetailDTO struct {
	UID         uint32              `json:"uid"`
	Subject     string              `json:"subject"`
	From        string              `json:"from"`
	To          []string            `json:"to"`
	Date        string              `json:"date"`
	TextPlain   string              `json:"textPlain"`
	TextHtml    string              `json:"textHtml"`
	Attachments []AttachmentMetaDTO `json:"attachments"`
}

// LoginResult is returned after login or auto-login attempt.
type LoginResult struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
	Error   string `json:"error"`
}

// SettingsDTO exposes user preferences.
type SettingsDTO struct {
	FetchDays              int  `json:"fetchDays"`
	RefreshIntervalMinutes int  `json:"refreshIntervalMinutes"` // 0 = disabled
	OpenLinksInBrowser     bool `json:"openLinksInBrowser"`     // true = system default browser
}
