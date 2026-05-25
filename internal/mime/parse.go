package mimeparse

import (
	"bytes"
	"io"
	"strings"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func init() {
	message.CharsetReader = charsetReader
}

func charsetReader(charsetName string, input io.Reader) (io.Reader, error) {
	charsetName = strings.ToLower(strings.TrimSpace(charsetName))
	switch charsetName {
	case "gb2312", "gbk", "gb18030", "cp936", "windows-936":
		return transform.NewReader(input, simplifiedchinese.GBK.NewDecoder()), nil
	default:
		e, name := charset.Lookup(charsetName)
		if e == nil || name == "" {
			return input, nil
		}
		return e.NewDecoder().Reader(input), nil
	}
}

// AttachmentMeta describes a non-inline attachment.
type AttachmentMeta struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

// ParsedBody is extracted mail content.
type ParsedBody struct {
	Subject     string
	From        string
	To          []string
	Date        string
	TextPlain   string
	TextHtml    string
	Attachments []AttachmentMeta
}

var htmlPolicy = newEmailHTMLPolicy()

// ParseRFC822 parses a raw email message.
func ParseRFC822(raw []byte) (*ParsedBody, error) {
	reader, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	h := reader.Header
	result := &ParsedBody{
		Subject: headerValue(h, "Subject"),
		From:    headerValue(h, "From"),
		To:      splitAddresses(headerValue(h, "To")),
		Date:    headerValue(h, "Date"),
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			ct, _, _ := h.ContentType()
			body, _ := io.ReadAll(part.Body)
			text := string(body)
			if strings.HasPrefix(strings.ToLower(ct), "text/html") {
				if result.TextHtml == "" {
					result.TextHtml = sanitizeHTML(text)
				}
			} else if strings.HasPrefix(strings.ToLower(ct), "text/plain") {
				if result.TextPlain == "" {
					result.TextPlain = text
				}
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			ct, _, _ := h.ContentType()
			body, _ := io.ReadAll(part.Body)
			ctLower := strings.ToLower(ct)
			text := string(body)
			if strings.HasPrefix(ctLower, "text/html") {
				if result.TextHtml == "" {
					result.TextHtml = sanitizeHTML(text)
				}
			} else if strings.HasPrefix(ctLower, "text/plain") {
				if result.TextPlain == "" {
					result.TextPlain = text
				}
			} else {
				result.Attachments = append(result.Attachments, AttachmentMeta{
					Filename:    filename,
					ContentType: ct,
					Size:        len(body),
				})
			}
		}
	}

	if result.Subject == "" {
		result.Subject = "(无主题)"
	}
	return result, nil
}

func headerValue(h mail.Header, key string) string {
	v, err := h.Text(key)
	if err != nil {
		return h.Get(key)
	}
	return v
}

func splitAddresses(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sanitizeHTML(html string) string {
	if html == "" {
		return ""
	}
	return htmlPolicy.Sanitize(html)
}
