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
	guigui.RootWidget

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
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

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
		context.Enable(&r.generateButton)
	} else {
		context.Disable(&r.generateButton)
	}

	u := float64(basicwidget.UnitSize(context))
	context.SetSize(&r.form, image.Pt(context.Size(r).X-int(1*u), guigui.DefaultSize))
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
		p := context.Position(r).Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidgetWithPosition(&r.form, p)
	}

	context.SetSize(&r.passwordsPanel, context.Size(r).Add(image.Pt(0, -int(8.5*u))))
	r.passwordsPanelContent.SetPasswords(r.passwords)
	r.passwordsPanelContent.SetOnClearTriggered(func() {
		r.passwords = []Password{}
	})
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)
	context.SetSize(&r.passwordsPanelContent, image.Pt(context.Size(r).X, guigui.DefaultSize))
	appender.AppendChildWidgetWithPosition(&r.passwordsPanel, context.Position(r).Add(image.Pt(0, int(8.5*u))))

	return nil
}

func (r *Root) reset() {
	r.countOutputTextField.SetText("64")
	r.numberCharsTextField.SetText("16")
	r.minNumsTextField.SetText("-1")
	r.minSymbolsTextField.SetText("-1")

	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
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
	u := float64(basicwidget.UnitSize(context))

	pt := context.Position(p)
	p.copyButton.SetText("Copy")
	p.copyButton.SetOnUp(func() {
		clipboard.Write(clipboard.FmtText, []byte(p.text.Text()))
	})
	context.SetSize(&p.copyButton, image.Pt(int(3*u), guigui.DefaultSize))
	appender.AppendChildWidgetWithPosition(&p.copyButton, pt)

	// w, h := context.Size(p)
	context.SetSize(&p.text, context.Size(p).Add(image.Pt(-int(4.5*u), 0)))
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	appender.AppendChildWidgetWithPosition(&p.text, image.Pt(pt.X+int(3.5*u), pt.Y))
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
	u := float64(basicwidget.UnitSize(context))

	pt := context.Position(p)
	x := pt.X + int(0.5*u)
	y := pt.Y
	for i := range p.passwordWidgets {
		if i > 0 {
			y += int(u / 4)
		}
		appender.AppendChildWidgetWithBounds(&p.passwordWidgets[i],
			image.Rectangle{
				Min: image.Pt(x, y),
				Max: image.Pt(x, y).Add(image.Pt(context.Size(p).X, int(u))),
			})
		y += int(u)
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
		Title:           "Password Generator",
		WindowMinWidth:  320,
		WindowMinHeight: 240,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
