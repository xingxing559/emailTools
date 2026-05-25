package provider

import (
	"fmt"
	"strings"
)

const (
	QQ         = "qq"
	Netease    = "netease"
	Microsoft  = "microsoft"
	AuthAppPwd = "app_password"
	AuthOAuth2 = "oauth2"
)

// Info describes a mailbox provider for UI and connection.
type Info struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	AuthType         string `json:"authType"`
	HelpURL          string `json:"helpUrl"`
	EmailPlaceholder string `json:"emailPlaceholder"`
}

var catalog = []Info{
	{
		ID:               QQ,
		DisplayName:      "QQ",
		AuthType:         AuthAppPwd,
		HelpURL:          "https://help.mail.qq.com/detail/0/985",
		EmailPlaceholder: "123456789@qq.com",
	},
	{
		ID:               Netease,
		DisplayName:      "163",
		AuthType:         AuthAppPwd,
		HelpURL:          "https://help.mail.163.com/faqDetail.do?code=d6af8ee8f1f8e0b0",
		EmailPlaceholder: "name@163.com",
	},
	{
		ID:               Microsoft,
		DisplayName:      "Outlook",
		AuthType:         AuthAppPwd,
		HelpURL:          "https://support.microsoft.com/account-billing/using-app-passwords-with-apps-that-don-t-support-two-step-verification-5896ed9b-4263-6812-ef18-520f4cf8f57c",
		EmailPlaceholder: "name@outlook.com",
	},
}

// List returns all supported providers.
func List() []Info {
	out := make([]Info, len(catalog))
	copy(out, catalog)
	return out
}

// Normalize validates and normalizes a provider id.
func Normalize(id string) (string, error) {
	id = strings.TrimSpace(strings.ToLower(id))
	if id == "" {
		return QQ, nil
	}
	for _, p := range catalog {
		if p.ID == id {
			return id, nil
		}
	}
	return "", fmt.Errorf("不支持的邮箱类型：%s", id)
}

// DisplayName returns the short tag for lists (QQ / 163 / Outlook).
func DisplayName(id string) string {
	id, _ = Normalize(id)
	for _, p := range catalog {
		if p.ID == id {
			return p.DisplayName
		}
	}
	return id
}

// AuthType returns app_password or oauth2 for a provider.
func AuthType(id string) string {
	id, _ = Normalize(id)
	for _, p := range catalog {
		if p.ID == id {
			return p.AuthType
		}
	}
	return AuthAppPwd
}

// SuggestFromEmail guesses provider from email domain.
func SuggestFromEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return QQ
	}
	domain := email[at+1:]
	switch domain {
	case "qq.com", "foxmail.com":
		return QQ
	case "163.com", "126.com", "yeah.net", "188.com":
		return Netease
	case "outlook.com", "hotmail.com", "live.com", "msn.com":
		return Microsoft
	default:
		if strings.HasSuffix(domain, "onmicrosoft.com") || strings.HasSuffix(domain, "outlook.com") {
			return Microsoft
		}
		return QQ
	}
}

// ResolveIMAPServer returns host:port for TLS IMAP.
func ResolveIMAPServer(id, email string) (string, error) {
	id, err := Normalize(id)
	if err != nil {
		return "", err
	}
	switch id {
	case QQ:
		return "imap.qq.com:993", nil
	case Microsoft:
		return "outlook.office365.com:993", nil
	case Netease:
		email = strings.TrimSpace(strings.ToLower(email))
		at := strings.LastIndex(email, "@")
		if at < 0 {
			return "imap.163.com:993", nil
		}
		switch email[at+1:] {
		case "126.com":
			return "imap.126.com:993", nil
		case "yeah.net":
			return "imap.yeah.net:993", nil
		default:
			return "imap.163.com:993", nil
		}
	default:
		return "", fmt.Errorf("未知提供商")
	}
}

// LoginErrorHint returns a localized hint for failed login.
func LoginErrorHint(id string) string {
	id, _ = Normalize(id)
	switch id {
	case QQ:
		return "请检查邮箱地址与授权码（需在 QQ 邮箱开启 IMAP 并生成授权码）"
	case Netease:
		return "请检查邮箱地址与客户端授权密码（需在网易邮箱开启 IMAP/SMTP 并设置授权密码）"
	case Microsoft:
		return "请检查邮箱地址与应用密码（需在微软账户开启 IMAP 并创建应用密码）"
	default:
		return "请检查账号与凭据"
	}
}

// DialErrorPrefix returns a prefix for connection errors.
func DialErrorPrefix(id string) string {
	id, _ = Normalize(id)
	switch id {
	case QQ:
		return "无法连接 QQ 邮箱服务器"
	case Netease:
		return "无法连接网易邮箱服务器"
	case Microsoft:
		return "无法连接 Outlook 邮箱服务器"
	default:
		return "无法连接邮箱服务器"
	}
}
