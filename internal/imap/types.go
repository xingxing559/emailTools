package imap

import "time"

// Mailbox is a selectable IMAP folder.
type Mailbox struct {
	Name        string
	DisplayName string
	Delimiter   string
}

// MessageHeader is summary metadata for a message.
type MessageHeader struct {
	UID      uint32
	Subject  string
	From     string
	Date     string
	DateUnix int64
	DateTime time.Time
	Seen     bool
}
