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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
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

type PasswordWidgets struct {
	copyButton basicwidget.TextButton
	text       basicwidget.Text
}

type Root struct {
	guigui.RootWidget

	once sync.Once

	form                     basicwidget.Form
	countOutputTextFieldText basicwidget.Text
	countOutputTextField     basicwidget.TextField
	numberCharsTextFieldText basicwidget.Text
	numberCharsTextField     basicwidget.TextField
	minNumsTextFieldText     basicwidget.Text
	minNumsTextField         basicwidget.TextField
	minSymbolsTextFieldText  basicwidget.Text
	minSymbolsTextField      basicwidget.TextField
	resetButton              basicwidget.TextButton
	generateButton           basicwidget.TextButton
	passwordWidgets          []*PasswordWidgets
	passwordsPanel           basicwidget.ScrollablePanel

	passwords []Password
}

func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	r.countOutputTextFieldText.SetText("count of output")
	r.countOutputTextField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.countOutputTextField.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.numberCharsTextFieldText.SetText("number of characters")
	r.numberCharsTextField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.numberCharsTextField.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.minNumsTextFieldText.SetText("minimum count of numbers")
	r.minNumsTextField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.minNumsTextField.SetOnEnterPressed(func(text string) {
		r.tryGeneratePassword()
	})
	r.minSymbolsTextFieldText.SetText("minimum count of symbols")
	r.minSymbolsTextField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	r.minSymbolsTextField.SetOnEnterPressed(func(text string) {
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

	u := float64(basicwidget.UnitSize(context))
	w, h := r.Size(context)
	r.form.SetWidth(context, w-int(1*u))
	r.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &r.countOutputTextFieldText,
			SecondaryWidget: &r.countOutputTextField,
		},
		{
			PrimaryWidget:   &r.numberCharsTextFieldText,
			SecondaryWidget: &r.numberCharsTextField,
		},
		{
			PrimaryWidget:   &r.minNumsTextFieldText,
			SecondaryWidget: &r.minNumsTextField,
		},
		{
			PrimaryWidget:   &r.minSymbolsTextFieldText,
			SecondaryWidget: &r.minSymbolsTextField,
		},
		{
			PrimaryWidget:   &r.resetButton,
			SecondaryWidget: &r.generateButton,
		},
	})
	{
		p := guigui.Position(r).Add(image.Pt(int(0.5*u), int(0.5*u)))
		guigui.SetPosition(&r.form, p)
		appender.AppendChildWidget(&r.form)
	}

	r.passwordsPanel.SetSize(context, w, h-int(8.5*u))
	r.passwordsPanel.SetContent(func(context *guigui.Context, childAppender *basicwidget.ContainerChildWidgetAppender, offsetX, offsetY float64) {
		p := guigui.Position(&r.passwordsPanel).Add(image.Pt(int(offsetX), int(offsetY)))
		minX := p.X + int(0.5*u)
		y := p.Y
		for i, t := range r.passwords {
			if has := slices.ContainsFunc(r.passwordWidgets, func(pw *PasswordWidgets) bool {
				return pw.text.Text() == t.Text
			}); !has {
				var pw PasswordWidgets
				pw.copyButton.SetText("Copy")
				pw.copyButton.SetOnUp(func() {
					clipboard.Write(clipboard.FmtText, []byte(t.Text))
				})
				pw.text.SetSelectable(true)
				pw.text.SetText(t.Text)
				pw.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
				r.passwordWidgets = slices.Insert(r.passwordWidgets, i, &pw)
			}

			if i > 0 {
				y += int(u / 4)
			}
			guigui.SetPosition(&r.passwordWidgets[i].copyButton, image.Pt(minX, y))
			childAppender.AppendChildWidget(&r.passwordWidgets[i].copyButton)
			r.passwordWidgets[i].text.SetSize(w-int(4.5*u), int(u))
			guigui.SetPosition(&r.passwordWidgets[i].text, image.Pt(minX+int(3.5*u), y))
			childAppender.AppendChildWidget(&r.passwordWidgets[i].text)
			y += int(u)
		}
	})
	r.passwordsPanel.SetPadding(0, int(0.5*u))
	guigui.SetPosition(&r.passwordsPanel, guigui.Position(r).Add(image.Pt(0, int(8.5*u))))
	appender.AppendChildWidget(&r.passwordsPanel)

	r.passwordWidgets = slices.DeleteFunc(r.passwordWidgets, func(pw *PasswordWidgets) bool {
		return !slices.ContainsFunc(r.passwords, func(p Password) bool {
			return p.Text == pw.text.Text()
		})
	})
}

func (r *Root) Update(context *guigui.Context) error {
	if r.canGeneratePassword() {
		guigui.Enable(&r.generateButton)
	} else {
		guigui.Disable(&r.generateButton)
	}
	return nil
}

func (r *Root) reset() {
	r.countOutputTextField.SetText("64")
	r.numberCharsTextField.SetText("16")
	r.minNumsTextField.SetText("-1")
	r.minSymbolsTextField.SetText("-1")

	r.passwords = []Password{}
}

func (r *Root) canGeneratePassword() bool {
	o := strings.TrimSpace(r.countOutputTextField.Text())
	n := strings.TrimSpace(r.numberCharsTextField.Text())
	nc := strings.TrimSpace(r.minNumsTextField.Text())
	sc := strings.TrimSpace(r.minSymbolsTextField.Text())
	return o != "" && n != "" && nc != "" && sc != ""
}

func (r *Root) tryGeneratePassword() {
	o, _ := strconv.Atoi(strings.TrimSpace(r.countOutputTextField.Text()))
	n, _ := strconv.Atoi(strings.TrimSpace(r.numberCharsTextField.Text()))
	nc, _ := strconv.Atoi(strings.TrimSpace(r.minNumsTextField.Text()))
	sc, _ := strconv.Atoi(strings.TrimSpace(r.minSymbolsTextField.Text()))
	var buf bytes.Buffer
	err := run(&buf, o, n, nc, sc)
	if err != nil {
		return
	}
	r.passwords = []Password{}
	for v := range bytes.FieldsSeq(buf.Bytes()) {
		r.passwords = slices.Insert(r.passwords, len(r.passwords), NewPassword(string(v)))
	}
}

func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	if err := clipboard.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	op := &guigui.RunOptions{
		Title:           "Password Generator",
		WindowMinWidth:  320,
		WindowMinHeight: 240,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
