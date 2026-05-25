package imap

import "strings"

var folderDisplayNames = map[string]string{
	"INBOX":            "收件箱",
	"inbox":            "收件箱",
	"Sent Messages":    "已发送",
	"Sent":             "已发送",
	"Drafts":           "草稿箱",
	"Deleted Messages": "已删除",
	"Trash":            "垃圾箱",
	"Junk":             "垃圾邮件",
	"Spam":             "垃圾邮件",
	"Starred":          "星标邮件",
	"Archive":          "归档",
}

// DisplayName returns a Chinese-friendly folder label.
func DisplayName(imapName string) string {
	if imapName == "" {
		return imapName
	}
	if cn, ok := folderDisplayNames[imapName]; ok {
		return cn
	}
	lower := strings.ToLower(imapName)
	for k, v := range folderDisplayNames {
		if strings.ToLower(k) == lower {
			return v
		}
	}
	return imapName
}
