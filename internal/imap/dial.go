package imap

import (
	"fmt"
	"strings"

	"EmailTools/internal/provider"

	"github.com/emersion/go-imap/client"
)

// DialAppPassword connects with IMAP LOGIN (QQ, Netease).
func DialAppPassword(providerID, email, secret string) (*Repository, error) {
	email = strings.TrimSpace(email)
	secret = strings.TrimSpace(secret)
	if email == "" || secret == "" {
		return nil, fmt.Errorf("邮箱地址和授权码不能为空")
	}

	providerID, err := provider.Normalize(providerID)
	if err != nil {
		return nil, err
	}
	host, err := provider.ResolveIMAPServer(providerID, email)
	if err != nil {
		return nil, err
	}

	c, err := clientDialTLS(host)
	if err != nil {
		return nil, fmt.Errorf("%s，请检查网络：%w", provider.DialErrorPrefix(providerID), err)
	}

	if err := c.Login(email, secret); err != nil {
		_ = c.Logout()
		return nil, fmt.Errorf("登录失败，%s", provider.LoginErrorHint(providerID))
	}

	repo := &Repository{email: email, client: c}
	repo.startKeepalive()
	return repo, nil
}

// DialQQ is kept for backward compatibility; delegates to DialAppPassword.
func DialQQ(email, authCode string) (*Repository, error) {
	return DialAppPassword(provider.QQ, email, authCode)
}

func clientDialTLS(host string) (*client.Client, error) {
	c, err := client.DialTLS(host, nil)
	if err != nil {
		return nil, err
	}
	c.Timeout = dialTimeout
	return c, nil
}
