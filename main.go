package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"slices"
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
	"golang.design/x/clipboard"
)

type Root struct {
	guigui.RootWidget

	once sync.Once

	background             basicwidget.Background
	form                   basicwidget.Form
	countOutputText        basicwidget.Text
	countOutputNumberInput basicwidget.NumberInput
	numberCharsText        basicwidget.Text
	numberCharsNumberInput basicwidget.NumberInput
	minNumsText            basicwidget.Text
	minNumsNumberInput     basicwidget.NumberInput
	minSymbolsText         basicwidget.Text
	minSymbolsNumberInput  basicwidget.NumberInput
	resetButton            basicwidget.TextButton
	generateButton         basicwidget.TextButton
	passwordsPanel         basicwidget.ScrollablePanel
	passwordsPanelContent  passwordsPanelContent

	model Model
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.countOutputText.SetText("count of output")
	r.countOutputNumberInput.SetOnValueChanged(func(value int64) {
		r.model.SetCountOutputValue(value)
	})
	r.countOutputNumberInput.SetMinimumValue(1)
	r.countOutputNumberInput.SetValue(r.model.CountOutputValue())

	r.numberCharsText.SetText("number of characters")
	r.numberCharsNumberInput.SetOnValueChanged(func(value int64) {
		r.model.SetNumberCharsValue(value)
	})
	r.numberCharsNumberInput.SetMinimumValue(1)
	r.numberCharsNumberInput.SetValue(r.model.NumberCharsValue())

	r.minNumsText.SetText("minimum count of numbers")
	r.minNumsNumberInput.SetOnValueChanged(func(value int64) {
		r.model.SetMinNumsValue(value)
	})
	r.minNumsNumberInput.SetMinimumValue(-1)
	r.minNumsNumberInput.SetValue(r.model.MinNumsValue())

	r.minSymbolsText.SetText("minimum count of symbols")
	r.minSymbolsNumberInput.SetOnValueChanged(func(value int64) {
		r.model.SetMinSymbolsValue(value)
	})
	r.minSymbolsNumberInput.SetMinimumValue(-1)
	r.minSymbolsNumberInput.SetValue(r.model.MinSymbolsValue())

	r.once.Do(func() { r.reset() })

	r.resetButton.SetText("Reset")
	r.resetButton.SetOnUp(func() {
		r.reset()
	})

	r.generateButton.SetText("Generate")
	r.generateButton.SetOnUp(func() {
		r.tryGeneratePassword()
	})

	u := basicwidget.UnitSize(context)
	r.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &r.countOutputText,
			SecondaryWidget: &r.countOutputNumberInput,
		},
		{
			PrimaryWidget:   &r.numberCharsText,
			SecondaryWidget: &r.numberCharsNumberInput,
		},
		{
			PrimaryWidget:   &r.minNumsText,
			SecondaryWidget: &r.minNumsNumberInput,
		},
		{
			PrimaryWidget:   &r.minSymbolsText,
			SecondaryWidget: &r.minSymbolsNumberInput,
		},
		{
			PrimaryWidget:   &r.resetButton,
			SecondaryWidget: &r.generateButton,
		},
	})

	r.passwordsPanelContent.SetModel(&r.model)
	r.passwordsPanelContent.SetOnClearTriggered(func() {
		r.model.ClearPassword()
	})
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)

	gl := layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u / 2),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= 1 {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(context.Size(&r.form).Y)
			}),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}
	for i, bounds := range gl.CellBounds() {
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
	r.countOutputNumberInput.SetValue(64)
	r.numberCharsNumberInput.SetValue(16)
	r.minNumsNumberInput.SetValue(-1)
	r.minSymbolsNumberInput.SetValue(-1)

	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
}

func (r *Root) tryGeneratePassword() {
	// TODO: この辺、modelに移す？
	o := int(r.model.CountOutputValue())
	n := int(r.model.NumberCharsValue())
	nc := int(r.model.MinNumsValue())
	sc := int(r.model.MinSymbolsValue())
	var buf bytes.Buffer
	err := run(&buf, o, n, nc, sc)
	if err != nil {
		return
	}
	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
	r.model.TryAddPassword(&buf)
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
	gl := layout.GridLayout{
		Bounds: context.Bounds(p),
		Widths: []layout.Size{
			layout.FixedSize(3 * u),
			layout.FlexibleSize(1),
		},
		ColumnGap: u / 2,
	}
	for i, bounds := range gl.CellBounds() {
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

	model *Model
}

func (p *passwordsPanelContent) SetOnClearTriggered(f func()) {
	p.onClearTriggered = f
}

func (p *passwordsPanelContent) SetModel(model *Model) {
	p.model = model
}

func (p *passwordsPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if p.model.PasswordCount() > len(p.passwordWidgets) {
		p.passwordWidgets = slices.Grow(p.passwordWidgets, p.model.PasswordCount()-len(p.passwordWidgets))
		p.passwordWidgets = p.passwordWidgets[:p.model.PasswordCount()]
	} else {
		p.passwordWidgets = slices.Delete(p.passwordWidgets, p.model.PasswordCount(), len(p.passwordWidgets))
	}

	for i := range p.model.PasswordCount() {
		pw := p.model.PasswordByIndex(i)
		p.passwordWidgets[i].SetText(pw.Text)
	}

	u := basicwidget.UnitSize(context)
	gl := layout.GridLayout{
		Bounds: context.Bounds(p),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= len(p.passwordWidgets) {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(context.Size(&p.passwordWidgets[row]).Y)
			}),
		},
		RowGap: u / 4,
	}
	for i, bounds := range gl.RepeatingCellBounds() {
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
