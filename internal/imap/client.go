package imap

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

const (
	dialTimeout     = 30 * time.Second
	maxFetchResults = 500
)

type Repository struct {
	mu            sync.Mutex
	email         string
	client        *client.Client
	selected      string
	stopKeepalive chan struct{}
}

func (r *Repository) Email() string {
	return r.email
}

func (r *Repository) SelectedFolder() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.selected
}

// Disconnect closes the connection immediately (for account switch).
func (r *Repository) Disconnect() error {
	r.stopKeepaliveLoop()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil
	}
	_ = r.client.Terminate()
	r.client = nil
	r.selected = ""
	return nil
}

// Close logs out gracefully (for user disconnect / app exit).
func (r *Repository) Close() error {
	r.stopKeepaliveLoop()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil
	}
	err := r.client.Logout()
	r.client = nil
	r.selected = ""
	return err
}

func (r *Repository) ListMailboxes() ([]Mailbox, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil, fmt.Errorf("未连接邮箱")
	}

	ch := make(chan *imap.MailboxInfo, 32)
	done := make(chan error, 1)
	go func() {
		done <- r.client.List("", "*", ch)
	}()

	var boxes []Mailbox
	for info := range ch {
		if !hasAttribute(info.Attributes, imap.NoSelectAttr) {
			boxes = append(boxes, Mailbox{
				Name:        info.Name,
				DisplayName: DisplayName(info.Name),
				Delimiter:   string(info.Delimiter),
			})
		}
	}
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取文件夹失败：%w", err)
	}
	return boxes, nil
}

func (r *Repository) Select(name string) (*imap.MailboxStatus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil, fmt.Errorf("未连接邮箱")
	}
	mbox, err := r.client.Select(name, true)
	if err != nil {
		return nil, fmt.Errorf("打开文件夹「%s」失败：%w", name, err)
	}
	r.selected = name
	return mbox, nil
}

func (r *Repository) FetchHeadersSince(since time.Time, maxResults int) ([]MessageHeader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil, fmt.Errorf("未连接邮箱")
	}
	if r.selected == "" {
		return nil, fmt.Errorf("请先选择文件夹")
	}
	if maxResults <= 0 {
		maxResults = 50
	}
	if maxResults > maxFetchResults {
		maxResults = maxFetchResults
	}

	criteria := imap.NewSearchCriteria()
	if !since.IsZero() {
		criteria.Since = since
	}
	uids, err := r.client.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败：%w", err)
	}
	if len(uids) == 0 {
		return []MessageHeader{}, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid, imap.FetchInternalDate}
	ch := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- r.client.UidFetch(seqSet, items, ch)
	}()

	var headers []MessageHeader
	for msg := range ch {
		if msg.Envelope == nil {
			continue
		}
		subject := msg.Envelope.Subject
		if subject == "" {
			subject = "(无主题)"
		}
		dt := messageTime(msg.Envelope.Date, msg.InternalDate)
		headers = append(headers, MessageHeader{
			UID:      msg.Uid,
			Subject:  subject,
			From:     formatAddresses(msg.Envelope.From),
			Date:     formatDate(dt),
			DateUnix: dt.Unix(),
			DateTime: dt,
			Seen:     hasFlag(msg.Flags, imap.SeenFlag),
		})
	}
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件列表失败：%w", err)
	}

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].DateTime.After(headers[j].DateTime)
	})
	if len(headers) > maxResults {
		headers = headers[:maxResults]
	}
	return headers, nil
}

func (r *Repository) FetchHeaders(offset, limit int) ([]MessageHeader, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil, fmt.Errorf("未连接邮箱")
	}
	if r.selected == "" {
		return nil, fmt.Errorf("请先选择文件夹")
	}

	status, err := r.client.Status(r.selected, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return nil, fmt.Errorf("读取邮件数量失败：%w", err)
	}

	total := int(status.Messages)
	if total == 0 {
		return []MessageHeader{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	end := total - offset
	if end < 1 {
		return []MessageHeader{}, nil
	}
	start := end - limit + 1
	if start < 1 {
		start = 1
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(uint32(start), uint32(end))

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid, imap.FetchInternalDate}
	ch := make(chan *imap.Message, limit)
	done := make(chan error, 1)
	go func() {
		done <- r.client.Fetch(seqSet, items, ch)
	}()

	var headers []MessageHeader
	for msg := range ch {
		if msg.Envelope == nil {
			continue
		}
		subject := msg.Envelope.Subject
		if subject == "" {
			subject = "(无主题)"
		}
		dt := messageTime(msg.Envelope.Date, msg.InternalDate)
		headers = append(headers, MessageHeader{
			UID:      msg.Uid,
			Subject:  subject,
			From:     formatAddresses(msg.Envelope.From),
			Date:     formatDate(dt),
			DateUnix: dt.Unix(),
			DateTime: dt,
			Seen:     hasFlag(msg.Flags, imap.SeenFlag),
		})
	}
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件列表失败：%w", err)
	}

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].DateTime.After(headers[j].DateTime)
	})
	return headers, nil
}

func (r *Repository) FetchBodyByUID(uid uint32) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return nil, fmt.Errorf("未连接邮箱")
	}
	if r.selected == "" {
		return nil, fmt.Errorf("请先选择文件夹")
	}

	seqSet, err := imap.ParseSeqSet(fmt.Sprintf("%d", uid))
	if err != nil {
		return nil, fmt.Errorf("无效的邮件 UID")
	}

	// RFC822 fetches the entire message; empty BODY section often returns only
	// text/plain which may be empty for HTML-only multipart mail.
	items := []imap.FetchItem{imap.FetchRFC822}
	ch := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- r.client.UidFetch(seqSet, items, ch)
	}()

	var body []byte
	section := &imap.BodySectionName{}
	for msg := range ch {
		if msg == nil {
			continue
		}
		if lit := msg.GetBody(section); lit != nil {
			data, err := io.ReadAll(lit)
			if err != nil {
				return nil, fmt.Errorf("读取邮件正文失败：%w", err)
			}
			if len(data) > len(body) {
				body = data
			}
		}
		for _, literal := range msg.Body {
			data, err := io.ReadAll(literal)
			if err != nil {
				return nil, fmt.Errorf("读取邮件正文失败：%w", err)
			}
			if len(data) > len(body) {
				body = data
			}
		}
	}
	if err := <-done; err != nil {
		return nil, fmt.Errorf("获取邮件详情失败：%w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("邮件正文为空")
	}
	return body, nil
}

func hasAttribute(attrs []string, attr string) bool {
	for _, a := range attrs {
		if strings.EqualFold(a, attr) {
			return true
		}
	}
	return false
}

func hasFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if strings.EqualFold(f, flag) {
			return true
		}
	}
	return false
}

func formatAddresses(addrs []*imap.Address) string {
	if len(addrs) == 0 {
		return ""
	}
	var parts []string
	for _, a := range addrs {
		name := a.PersonalName
		addr := a.Address()
		if name != "" {
			parts = append(parts, fmt.Sprintf("%s <%s>", name, addr))
		} else {
			parts = append(parts, addr)
		}
	}
	return strings.Join(parts, ", ")
}

func messageTime(envelope, internal time.Time) time.Time {
	if !envelope.IsZero() {
		return envelope
	}
	return internal
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04")
}
