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
	guigui.DefaultWidget

	once sync.Once

	background               basicwidget.Background
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
	passwordsPanel           basicwidget.ScrollablePanel
	passwordsPanelContent    passwordsPanelContent

	passwords []Password
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidget(&r.background)

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
	if r.canGeneratePassword() {
		guigui.Enable(&r.generateButton)
	} else {
		guigui.Disable(&r.generateButton)
	}

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
	r.passwordsPanelContent.root = r
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)
	guigui.SetPosition(&r.passwordsPanel, guigui.Position(r).Add(image.Pt(0, int(8.5*u))))
	appender.AppendChildWidget(&r.passwordsPanel)

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

type passwordWidget struct {
	guigui.DefaultWidget

	copyButton basicwidget.TextButton
	text       basicwidget.Text
}

func (p *passwordWidget) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := float64(basicwidget.UnitSize(context))

	pt := guigui.Position(p)
	guigui.SetPosition(&p.copyButton, pt)
	appender.AppendChildWidget(&p.copyButton)

	w, _ := p.Size(context)
	p.text.SetSize(w-int(4.5*u), int(u))
	guigui.SetPosition(&p.text, image.Pt(pt.X+int(3.5*u), pt.Y))
	appender.AppendChildWidget(&p.text)
	return nil
}

func (p *passwordWidget) Size(context *guigui.Context) (int, int) {
	w, _ := guigui.Parent(p).Size(context)
	return w, int(basicwidget.UnitSize(context))
}

type passwordsPanelContent struct {
	guigui.DefaultWidget

	root            *Root
	passwordWidgets []*passwordWidget
}

func (c *passwordsPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := float64(basicwidget.UnitSize(context))

	root := c.root
	p := guigui.Position(c)
	minX := p.X + int(0.5*u)
	y := p.Y
	for i, t := range root.passwords {
		if has := slices.ContainsFunc(c.passwordWidgets, func(pw *passwordWidget) bool {
			return pw.text.Text() == t.Text
		}); !has {
			var pw passwordWidget
			pw.copyButton.SetText("Copy")
			pw.copyButton.SetOnUp(func() {
				clipboard.Write(clipboard.FmtText, []byte(t.Text))
			})
			pw.text.SetText(t.Text)
			pw.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			c.passwordWidgets = slices.Insert(c.passwordWidgets, i, &pw)
		}

		if i > 0 {
			y += int(u / 4)
		}
		guigui.SetPosition(c.passwordWidgets[i], image.Pt(minX, y))
		appender.AppendChildWidget(c.passwordWidgets[i])
		y += int(u)
	}

	c.passwordWidgets = slices.DeleteFunc(c.passwordWidgets, func(pw *passwordWidget) bool {
		return !slices.ContainsFunc(c.root.passwords, func(p Password) bool {
			return p.Text == pw.text.Text()
		})
	})

	return nil
}

func (c *passwordsPanelContent) Size(context *guigui.Context) (int, int) {
	u := basicwidget.UnitSize(context)

	w, _ := guigui.Parent(c).Size(context)
	cnt := len(c.root.passwords)
	h := cnt * (u + u/4)
	return w, h
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
