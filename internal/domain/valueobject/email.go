package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	username string
	domain   string
}

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if !emailRegex.MatchString(raw) {
		return Email{}, errors.New("invalid email format")
	}
	parts := strings.SplitN(raw, "@", 2)
	return Email{username: parts[0], domain: parts[1]}, nil
}

func (e Email) String() string {
	return e.username + "@" + e.domain
}

func (e Email) Username() string {
	return e.username
}

func (e Email) Domain() string {
	return e.domain
}

func (e Email) Equals(other Email) bool {
	return e.username == other.username && e.domain == other.domain
}
