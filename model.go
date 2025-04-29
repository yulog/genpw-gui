package main

import (
	"bytes"
	"slices"
	"strings"
)

type Model struct {
	passwords []Password
}

func (m *Model) CanGeneratePassword(o, n, nc, sc string) bool {
	return strings.TrimSpace(o) != "" &&
		strings.TrimSpace(n) != "" &&
		strings.TrimSpace(nc) != "" &&
		strings.TrimSpace(sc) != ""
}

func (m *Model) TryAddPassword(buf *bytes.Buffer) bool {
	for v := range bytes.FieldsSeq(buf.Bytes()) {
		m.passwords = slices.Insert(m.passwords, len(m.passwords), NewPassword(string(v)))
	}
	return true
}

func (m *Model) ClearPassword() {
	m.passwords = []Password{}
}

func (m *Model) PasswordCount() int {
	return len(m.passwords)
}

func (m *Model) PasswordByIndex(i int) Password {
	return m.passwords[i]
}

type Password struct {
	Text string
}

func NewPassword(text string) Password {
	return Password{
		Text: text,
	}
}
