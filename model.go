package main

import (
	"bytes"
	"slices"
)

type Model struct {
	countOutput int64
	numberChars int64
	minNums     int64
	minSymbols  int64

	passwords []Password
}

func (m *Model) CountOutputValue() int64 {
	return m.countOutput
}

func (m *Model) SetCountOutputValue(value int64) {
	m.countOutput = value
}

func (m *Model) NumberCharsValue() int64 {
	return m.numberChars
}

func (m *Model) SetNumberCharsValue(value int64) {
	m.numberChars = value
}

func (m *Model) MinNumsValue() int64 {
	return m.minNums
}

func (m *Model) SetMinNumsValue(value int64) {
	m.minNums = value
}

func (m *Model) MinSymbolsValue() int64 {
	return m.minSymbols
}

func (m *Model) SetMinSymbolsValue(value int64) {
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
