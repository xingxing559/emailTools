package imap

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, sub := range []string{
		"connection closed",
		"connection reset",
		"broken pipe",
		"use of closed network",
		"not connected",
		"timeout",
		"i/o timeout",
		"eof",
		"未连接邮箱",
	} {
		if strings.Contains(msg, sub) {
			return true
		}
	}
	return false
}
