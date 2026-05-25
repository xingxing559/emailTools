package app

import (
	"fmt"
	"strings"

	"EmailTools/internal/credential"
	"EmailTools/internal/imap"
	"EmailTools/internal/provider"
)

func (a *MailApp) dialCredentials(_ string, cred credential.Credentials) (*imap.Repository, error) {
	p, err := provider.Normalize(cred.Provider)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(cred.Email)
	if email == "" {
		return nil, fmt.Errorf("邮箱地址无效")
	}

	secret := strings.TrimSpace(cred.AppPassword)
	if secret == "" {
		if cred.RefreshToken != "" {
			return nil, fmt.Errorf("Outlook 已改为应用密码登录，请在「编辑账号」中填写应用密码后保存")
		}
		return nil, fmt.Errorf("授权码无效，请重新添加账号")
	}
	return imap.DialAppPassword(p, email, secret)
}

func (a *MailApp) connectByID(id string) LoginResult {
	cred, err := credential.GetCredentials(id)
	if err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}
	repo, err := a.dialCredentials(id, cred)
	if err != nil {
		return LoginResult{Success: false, Error: err.Error()}
	}

	a.mu.Lock()
	if a.repo != nil {
		_ = a.repo.Disconnect()
	}
	a.repo = repo
	a.currentEmail = repo.Email()
	a.activeAccountID = id
	a.lastFolder = ""
	a.mu.Unlock()

	_ = credential.SetActiveID(id)
	return LoginResult{Success: true, Email: cred.Email}
}
