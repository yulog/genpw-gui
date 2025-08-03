package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"slices"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
	"github.com/yulog/genpw-gui/internal/clipboard"
)

type Root struct {
	guigui.DefaultWidget

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
	resetButton            basicwidget.Button
	generateButton         basicwidget.Button
	passwordsPanel         basicwidget.Panel
	passwordsPanelContent  passwordsPanelContent

	model Model
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.countOutputText.SetValue("count of output")
	r.countOutputNumberInput.SetOnValueChangedInt64(func(value int64, committed bool) {
		if !committed {
			return
		}
		r.model.SetCountOutputValue(int(value))
	})
	r.countOutputNumberInput.SetMinimumValueInt64(1)
	r.countOutputNumberInput.SetValueInt64(int64(r.model.CountOutputValue()))

	r.numberCharsText.SetValue("number of characters")
	r.numberCharsNumberInput.SetOnValueChangedInt64(func(value int64, committed bool) {
		if !committed {
			return
		}
		r.model.SetNumberCharsValue(int(value))
	})
	r.numberCharsNumberInput.SetMinimumValueInt64(1)
	r.numberCharsNumberInput.SetValueInt64(int64(r.model.NumberCharsValue()))

	r.minNumsText.SetValue("minimum count of numbers")
	r.minNumsNumberInput.SetOnValueChangedInt64(func(value int64, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinNumsValue(int(value))
	})
	r.minNumsNumberInput.SetMinimumValueInt64(-1)
	r.minNumsNumberInput.SetValueInt64(int64(r.model.MinNumsValue()))

	r.minSymbolsText.SetValue("minimum count of symbols")
	r.minSymbolsNumberInput.SetOnValueChangedInt64(func(value int64, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinSymbolsValue(int(value))
	})
	r.minSymbolsNumberInput.SetMinimumValueInt64(-1)
	r.minSymbolsNumberInput.SetValueInt64(int64(r.model.MinSymbolsValue()))

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
	r.form.SetItems([]basicwidget.FormItem{
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
	r.passwordsPanel.SetAutoBorder(true)

	gl := layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u / 2),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= 1 {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(context.ActualSize(&r.form).Y)
			}),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}
	appender.AppendChildWidgetWithBounds(&r.form, gl.CellBounds(0, 0))
	{
		bounds := gl.CellBounds(0, 1)
		context.SetSize(&r.passwordsPanelContent, image.Pt(bounds.Dx(), guigui.DefaultSize)) // Flexibleにならないため
		appender.AppendChildWidgetWithBounds(&r.passwordsPanel, bounds)
	}

	return nil
}

func (r *Root) reset() {
	r.countOutputNumberInput.SetValueInt64(64)
	r.numberCharsNumberInput.SetValueInt64(16)
	r.minNumsNumberInput.SetValueInt64(-1)
	r.minSymbolsNumberInput.SetValueInt64(-1)

	if r.passwordsPanelContent.onClearTriggered != nil {
		r.passwordsPanelContent.onClearTriggered()
	}
}

func (r *Root) tryGeneratePassword() {
	// TODO: この辺、modelに移す？
	o := r.model.CountOutputValue()
	n := r.model.NumberCharsValue()
	nc := r.model.MinNumsValue()
	sc := r.model.MinSymbolsValue()
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

	copyButton basicwidget.Button
	text       basicwidget.Text
}

func (p *passwordWidget) SetText(text string) {
	p.text.SetValue(text)
}

func (p *passwordWidget) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.copyButton.SetText("Copy")
	p.copyButton.SetOnUp(func() {
		clipboard.WriteAll(p.text.Value())
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
	appender.AppendChildWidgetWithBounds(&p.copyButton, gl.CellBounds(0, 0))
	appender.AppendChildWidgetWithBounds(&p.text, gl.CellBounds(1, 0))

	return nil
}

func (p *passwordWidget) DefaultSize(context *guigui.Context) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), context.ActualSize(&p.copyButton).Y)
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
				return layout.FixedSize(context.ActualSize(&p.passwordWidgets[row]).Y)
			}),
		},
		RowGap: u / 4,
	}
	for i := range p.passwordWidgets {
		bounds := gl.CellBounds(0, i)
		appender.AppendChildWidgetWithBounds(&p.passwordWidgets[i], bounds)
	}

	return nil
}

func (c *passwordsPanelContent) DefaultSize(context *guigui.Context) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range c.passwordWidgets {
		h += context.ActualSize(&c.passwordWidgets[i]).Y
		h += int(u / 4)
	}
	return image.Pt(6*int(u), h)
}

func main() {
	op := &guigui.RunOptions{
		Title:         "Password Generator",
		WindowMinSize: image.Pt(320, 240),
		RunGameOptions: &ebiten.RunGameOptions{
			ApplePressAndHoldEnabled: true,
		},
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
