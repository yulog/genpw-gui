package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
	"golang.design/x/clipboard"
)

type Password struct {
	Text string
}

func NewPassword(text string) Password {
	return Password{
		Text: text,
	}
}

type Root struct {
	guigui.RootWidget

	once sync.Once

	background            basicwidget.Background
	form                  basicwidget.Form
	countOutputText       basicwidget.Text
	countOutputTextInput  basicwidget.TextInput
	numberCharsText       basicwidget.Text
	numberCharsTextInput  basicwidget.TextInput
	minNumsText           basicwidget.Text
	minNumsTextInput      basicwidget.TextInput
	minSymbolsText        basicwidget.Text
	minSymbolsTextInput   basicwidget.TextInput
	resetButton           basicwidget.TextButton
	generateButton        basicwidget.TextButton
	passwordsPanel        basicwidget.ScrollablePanel
	passwordsPanelContent passwordsPanelContent

	passwords []Password
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.countOutputText.SetText("count of output")
	r.countOutputTextInput.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.countOutputTextInput.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.numberCharsText.SetText("number of characters")
	r.numberCharsTextInput.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.numberCharsTextInput.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.minNumsText.SetText("minimum count of numbers")
	r.minNumsTextInput.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.minNumsTextInput.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.minSymbolsText.SetText("minimum count of symbols")
	r.minSymbolsTextInput.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.minSymbolsTextInput.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})

	r.once.Do(func() { r.reset() })

	r.resetButton.SetText("Reset")
	r.resetButton.SetOnUp(func() {
		r.reset()
	})

	r.generateButton.SetText("Generate")
	r.generateButton.SetOnUp(func() {
		r.tryGeneratePassword()
	})
	context.SetEnabled(&r.generateButton, r.canGeneratePassword())

	u := basicwidget.UnitSize(context)
	r.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &r.countOutputText,
			SecondaryWidget: &r.countOutputTextInput,
		},
		{
			PrimaryWidget:   &r.numberCharsText,
			SecondaryWidget: &r.numberCharsTextInput,
		},
		{
			PrimaryWidget:   &r.minNumsText,
			SecondaryWidget: &r.minNumsTextInput,
		},
		{
			PrimaryWidget:   &r.minSymbolsText,
			SecondaryWidget: &r.minSymbolsTextInput,
		},
		{
			PrimaryWidget:   &r.resetButton,
			SecondaryWidget: &r.generateButton,
		},
	})

	r.passwordsPanelContent.SetPasswords(r.passwords)
	r.passwordsPanelContent.SetOnClearTriggered(func() {
		r.passwords = []Password{}
	})
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u / 2),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= 1 {
					return 0
				}
				return context.Size(&r.form).Y
			}),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&r.form, bounds)
		case 1:
			context.SetSize(&r.passwordsPanelContent, image.Pt(bounds.Dx(), guigui.DefaultSize)) // Flexibleにならないため
			appender.AppendChildWidgetWithBounds(&r.passwordsPanel, bounds)
		}
	}

	return nil
}

func (r *Root) reset() {
	r.countOutputTextInput.SetText("64")
	r.numberCharsTextInput.SetText("16")
	r.minNumsTextInput.SetText("-1")
	r.minSymbolsTextInput.SetText("-1")

	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
}

func (r *Root) canGeneratePassword() bool {
	o := strings.TrimSpace(r.countOutputTextInput.Text())
	n := strings.TrimSpace(r.numberCharsTextInput.Text())
	nc := strings.TrimSpace(r.minNumsTextInput.Text())
	sc := strings.TrimSpace(r.minSymbolsTextInput.Text())
	return o != "" && n != "" && nc != "" && sc != ""
}

func (r *Root) tryGeneratePassword() {
	o, _ := strconv.Atoi(strings.TrimSpace(r.countOutputTextInput.Text()))
	n, _ := strconv.Atoi(strings.TrimSpace(r.numberCharsTextInput.Text()))
	nc, _ := strconv.Atoi(strings.TrimSpace(r.minNumsTextInput.Text()))
	sc, _ := strconv.Atoi(strings.TrimSpace(r.minSymbolsTextInput.Text()))
	var buf bytes.Buffer
	err := run(&buf, o, n, nc, sc)
	if err != nil {
		return
	}
	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
	for v := range bytes.FieldsSeq(buf.Bytes()) {
		r.passwords = slices.Insert(r.passwords, len(r.passwords), NewPassword(string(v)))
	}
}

type passwordWidget struct {
	guigui.DefaultWidget

	copyButton basicwidget.TextButton
	text       basicwidget.Text
}

func (p *passwordWidget) SetText(text string) {
	p.text.SetText(text)
}

func (p *passwordWidget) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.copyButton.SetText("Copy")
	p.copyButton.SetOnUp(func() {
		clipboard.Write(clipboard.FmtText, []byte(p.text.Text()))
	})
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(p),
		Widths: []layout.Size{
			layout.FixedSize(3 * u),
			layout.FlexibleSize(1),
		},
		ColumnGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&p.copyButton, bounds)
		case 1:
			appender.AppendChildWidgetWithBounds(&p.text, bounds)
		}
	}
	return nil
}

func (p *passwordWidget) DefaultSize(context *guigui.Context) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), context.Size(&p.copyButton).Y)
}

type passwordsPanelContent struct {
	guigui.DefaultWidget

	passwordWidgets  []passwordWidget
	onClearTriggered func()
}

func (p *passwordsPanelContent) SetOnClearTriggered(f func()) {
	p.onClearTriggered = f
}

func (p *passwordsPanelContent) SetPasswords(passwords []Password) {
	if len(passwords) != len(p.passwordWidgets) {
		if len(passwords) > len(p.passwordWidgets) {
			p.passwordWidgets = slices.Grow(p.passwordWidgets, len(passwords)-len(p.passwordWidgets))
			p.passwordWidgets = p.passwordWidgets[:len(passwords)]
		} else {
			p.passwordWidgets = slices.Delete(p.passwordWidgets, len(passwords), len(p.passwordWidgets))
		}
	}
	for i, pw := range passwords {
		p.passwordWidgets[i].SetText(pw.Text)
	}
}

func (p *passwordsPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(p),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= len(p.passwordWidgets) {
					return 0
				}
				return context.Size(&p.passwordWidgets[index]).Y
			}),
		},
		RowGap: u / 4,
	}).RepeatingCellBounds() {
		if i >= len(p.passwordWidgets) {
			break
		}
		appender.AppendChildWidgetWithBounds(&p.passwordWidgets[i], bounds)
	}

	return nil
}

func (c *passwordsPanelContent) DefaultSize(context *guigui.Context) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range c.passwordWidgets {
		h += context.Size(&c.passwordWidgets[i]).Y
		h += int(u / 4)
	}
	return image.Pt(6*int(u), h)
}

func main() {
	if err := clipboard.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	op := &guigui.RunOptions{
		Title:         "Password Generator",
		WindowMinSize: image.Pt(320, 240),
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
