package imap

import (
	"fmt"
	"time"
)

const keepaliveInterval = 9 * time.Minute

func (r *Repository) startKeepalive() {
	r.mu.Lock()
	if r.stopKeepalive != nil {
		select {
		case <-r.stopKeepalive:
		default:
			close(r.stopKeepalive)
		}
	}
	r.stopKeepalive = make(chan struct{})
	stop := r.stopKeepalive
	c := r.client
	r.mu.Unlock()

	if c == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(keepaliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = r.Ping()
			}
		}
	}()
}

func (r *Repository) stopKeepaliveLoop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopKeepalive != nil {
		select {
		case <-r.stopKeepalive:
		default:
			close(r.stopKeepalive)
		}
		r.stopKeepalive = nil
	}
}

func (r *Repository) Ping() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		return fmt.Errorf("未连接邮箱")
	}
	return r.client.Noop()
}
