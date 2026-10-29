package config

import (
	"fmt"
	"net"
	"net/mail"
	"strconv"
	"strings"
)

func validateMailConfig(host, port, from string) error {
	address, err := mail.ParseAddress(from)
	if host == "" || strings.ContainsAny(host, "\r\n/: ") || (net.ParseIP(host) == nil && strings.Contains(host, "\\")) || err != nil || address.Address != from || strings.ContainsAny(from, "\r\n") {
		return fmt.Errorf("email verification requires SMTP_HOST and a bare SMTP_FROM email address")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("SMTP_PORT must be between 1 and 65535")
	}
	return nil
}
