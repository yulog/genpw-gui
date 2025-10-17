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

type modelKey int

const (
	modelKeyModel modelKey = iota
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

	layout layout.GridLayout
}

func (r *Root) Model(key any) any {
	switch key {
	case modelKeyModel:
		return &r.model
	default:
		return nil
	}
}

const (
	passwordsPanelContentEventClearTriggered = "clearTriggered"
)

func (r *Root) SetOnClearTriggered(f func()) {
	guigui.RegisterEventHandler(r, passwordsPanelContentEventClearTriggered, f)
}

func (r *Root) AddChildren(context *guigui.Context, adder *guigui.ChildAdder) {
	adder.AddChild(&r.background)
	adder.AddChild(&r.form)
	adder.AddChild(&r.passwordsPanel)
}

func (r *Root) Update(context *guigui.Context) error {
	r.countOutputText.SetValue("count of output")
	r.countOutputNumberInput.SetOnValueChanged(func(value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetCountOutputValue(value)
	})
	r.countOutputNumberInput.SetMinimumValue(1)
	r.countOutputNumberInput.SetValue(r.model.CountOutputValue())

	r.numberCharsText.SetValue("number of characters")
	r.numberCharsNumberInput.SetOnValueChanged(func(value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetNumberCharsValue(value)
	})
	r.numberCharsNumberInput.SetMinimumValue(1)
	r.numberCharsNumberInput.SetValue(r.model.NumberCharsValue())

	r.minNumsText.SetValue("minimum count of numbers")
	r.minNumsNumberInput.SetOnValueChanged(func(value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinNumsValue(value)
	})
	r.minNumsNumberInput.SetMinimumValue(-1)
	r.minNumsNumberInput.SetValue(r.model.MinNumsValue())

	r.minSymbolsText.SetValue("minimum count of symbols")
	r.minSymbolsNumberInput.SetOnValueChanged(func(value int, committed bool) {
		if !committed {
			return
		}
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

	r.SetOnClearTriggered(func() {
		r.model.ClearPassword()
	})
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)
	r.passwordsPanel.SetAutoBorder(true)

	r.layout = layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u / 2),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= 1 {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(r.form.Measure(context, guigui.FixedWidthConstraints(context.Bounds(r).Dx()-u)).Y)
			}),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}

	r.passwordsPanelContent.SetWidth(r.layout.CellBounds(0, 1).Dx())

	return nil
}

func (r *Root) Layout(context *guigui.Context, widget guigui.Widget) image.Rectangle {
	switch widget {
	case &r.background:
		return context.Bounds(r)
	case &r.form:
		return r.layout.CellBounds(0, 0)
	case &r.passwordsPanel:
		return r.layout.CellBounds(0, 1)
	}
	return image.Rectangle{}
}

func (r *Root) reset() {
	r.model.SetCountOutputValue(64)
	r.model.SetNumberCharsValue(16)
	r.model.SetMinNumsValue(-1)
	r.model.SetMinSymbolsValue(-1)

	guigui.DispatchEventHandler(r, passwordsPanelContentEventClearTriggered)
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
	guigui.DispatchEventHandler(r, passwordsPanelContentEventClearTriggered)
	r.model.TryAddPassword(&buf)
}

type passwordWidget struct {
	guigui.DefaultWidget

	copyButton basicwidget.Button
	text       basicwidget.Text

	layout layout.GridLayout
}

func (p *passwordWidget) SetText(text string) {
	p.text.SetValue(text)
}

func (p *passwordWidget) AddChildren(context *guigui.Context, adder *guigui.ChildAdder) {
	adder.AddChild(&p.copyButton)
	adder.AddChild(&p.text)
}

func (p *passwordWidget) Update(context *guigui.Context) error {
	p.copyButton.SetText("Copy")
	p.copyButton.SetOnUp(func() {
		clipboard.WriteAll(p.text.Value())
	})
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	u := basicwidget.UnitSize(context)
	p.layout = layout.GridLayout{
		Bounds: context.Bounds(p),
		Widths: []layout.Size{
			layout.FixedSize(3 * u),
			layout.FlexibleSize(1),
		},
		ColumnGap: u / 2,
	}

	return nil
}

func (p *passwordWidget) Layout(context *guigui.Context, widget guigui.Widget) image.Rectangle {
	switch widget {
	case &p.copyButton:
		return p.layout.CellBounds(0, 0)
	case &p.text:
		return p.layout.CellBounds(1, 0)
	}
	return image.Rectangle{}
}

func (p *passwordWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), p.copyButton.Measure(context, guigui.Constraints{}).Y)
}

type passwordsPanelContent struct {
	guigui.DefaultWidget

	passwordWidgets []passwordWidget

	width int

	layout layout.GridLayout
}

func (p *passwordsPanelContent) SetWidth(width int) {
	p.width = width
}

func (p *passwordsPanelContent) AddChildren(context *guigui.Context, adder *guigui.ChildAdder) {
	model := context.Model(p, modelKeyModel).(*Model)
	if model.PasswordCount() > len(p.passwordWidgets) {
		p.passwordWidgets = slices.Grow(p.passwordWidgets, model.PasswordCount()-len(p.passwordWidgets))
		p.passwordWidgets = p.passwordWidgets[:model.PasswordCount()]
	} else {
		p.passwordWidgets = slices.Delete(p.passwordWidgets, model.PasswordCount(), len(p.passwordWidgets))
	}
	for i := range p.passwordWidgets {
		adder.AddChild(&p.passwordWidgets[i])
	}
}

func (p *passwordsPanelContent) Update(context *guigui.Context) error {
	model := context.Model(p, modelKeyModel).(*Model)
	for i := range model.PasswordCount() {
		pw := model.PasswordByIndex(i)
		p.passwordWidgets[i].SetText(pw.Text)
	}

	u := basicwidget.UnitSize(context)
	p.layout = layout.GridLayout{
		Bounds: context.Bounds(p),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= len(p.passwordWidgets) {
					return layout.FixedSize(0)
				}
				w := guigui.FixedWidthConstraints(context.Bounds(p).Dx())
				h := p.passwordWidgets[row].Measure(context, w).Y
				return layout.FixedSize(h)
			}),
		},
		RowGap: u / 4,
	}
	// for i := range p.passwordWidgets {
	// 	bounds := gl.CellBounds(0, i)
	// 	context.SetBounds(&p.passwordWidgets[i], bounds, p)
	// }

	return nil
}

func (p *passwordsPanelContent) Layout(context *guigui.Context, widget guigui.Widget) image.Rectangle {
	for i := range p.passwordWidgets {
		if widget == &p.passwordWidgets[i] {
			return p.layout.CellBounds(0, i)
		}
	}
	return image.Rectangle{}
}

func (p *passwordsPanelContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range p.passwordWidgets {
		h += p.passwordWidgets[i].Measure(context, constraints).Y
		h += int(u / 4)
	}
	return image.Pt(p.width, h)
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
