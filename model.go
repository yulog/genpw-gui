package main

import (
	"bytes"
	"slices"
)

type Model struct {
	countOutput int
	numberChars int
	minNums     int
	minSymbols  int

	passwords []Password
}

func (m *Model) CountOutputValue() int {
	return m.countOutput
}

func (m *Model) SetCountOutputValue(value int) {
	m.countOutput = value
}

func (m *Model) NumberCharsValue() int {
	return m.numberChars
}

func (m *Model) SetNumberCharsValue(value int) {
	m.numberChars = value
}

func (m *Model) MinNumsValue() int {
	return m.minNums
}

func (m *Model) SetMinNumsValue(value int) {
	m.minNums = value
}

func (m *Model) MinSymbolsValue() int {
	return m.minSymbols
}

func (m *Model) SetMinSymbolsValue(value int) {
	m.minSymbols = value
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
