package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"slices"
	"sync"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
	_ "github.com/guigui-gui/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/ebiten/v2"
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

func (r *Root) SetOnClearTriggered(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, passwordsPanelContentEventClearTriggered, f)
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddChild(&r.background)
	adder.AddChild(&r.form)
	adder.AddChild(&r.passwordsPanel)

	r.countOutputText.SetValue("count of output")
	r.countOutputNumberInput.SetOnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetCountOutputValue(value)
	})
	r.countOutputNumberInput.SetMinimumValue(1)
	r.countOutputNumberInput.SetValue(r.model.CountOutputValue())

	r.numberCharsText.SetValue("number of characters")
	r.numberCharsNumberInput.SetOnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetNumberCharsValue(value)
	})
	r.numberCharsNumberInput.SetMinimumValue(1)
	r.numberCharsNumberInput.SetValue(r.model.NumberCharsValue())

	r.minNumsText.SetValue("minimum count of numbers")
	r.minNumsNumberInput.SetOnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinNumsValue(value)
	})
	r.minNumsNumberInput.SetMinimumValue(-1)
	r.minNumsNumberInput.SetValue(r.model.MinNumsValue())

	r.minSymbolsText.SetValue("minimum count of symbols")
	r.minSymbolsNumberInput.SetOnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinSymbolsValue(value)
	})
	r.minSymbolsNumberInput.SetMinimumValue(-1)
	r.minSymbolsNumberInput.SetValue(r.model.MinSymbolsValue())

	r.once.Do(func() { r.reset() })

	r.resetButton.SetText("Reset")
	r.resetButton.SetOnUp(func(context *guigui.Context) {
		r.reset()
	})

	r.generateButton.SetText("Generate")
	r.generateButton.SetOnUp(func(context *guigui.Context) {
		r.tryGeneratePassword()
	})

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

	r.SetOnClearTriggered(func(context *guigui.Context) {
		r.model.ClearPassword()
	})
	r.passwordsPanel.SetContent(&r.passwordsPanelContent)
	r.passwordsPanel.SetAutoBorder(true)
	r.passwordsPanel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)

	return nil
}

func (r *Root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())

	u := basicwidget.UnitSize(context)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items: []guigui.LinearLayoutItem{
			{
				Widget: &r.form,
				Size:   guigui.FixedSize(r.form.Measure(context, guigui.Constraints{}).Y),
			},
			{
				Widget: &r.passwordsPanel,
				Size:   guigui.FlexibleSize(1),
			},
		},
		Gap: u / 2,
	}).LayoutWidgets(context, widgetBounds.Bounds().Inset(u/2), layouter)
}

func (r *Root) reset() {
	r.model.SetCountOutputValue(64)
	r.model.SetNumberCharsValue(16)
	r.model.SetMinNumsValue(-1)
	r.model.SetMinSymbolsValue(-1)

	guigui.DispatchEvent(r, passwordsPanelContentEventClearTriggered)
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
	guigui.DispatchEvent(r, passwordsPanelContentEventClearTriggered)
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

func (p *passwordWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddChild(&p.copyButton)
	adder.AddChild(&p.text)

	p.copyButton.SetText("Copy")
	p.copyButton.SetOnUp(func(context *guigui.Context) {
		clipboard.WriteAll(p.text.Value())
	})
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	return nil
}

func (p *passwordWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items: []guigui.LinearLayoutItem{
			{
				Widget: &p.copyButton,
				Size:   guigui.FixedSize(3 * u),
			},
			{
				Widget: &p.text,
				Size:   guigui.FlexibleSize(1),
			},
		},
		Gap: u / 2,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *passwordWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), p.copyButton.Measure(context, guigui.Constraints{}).Y)
}

type passwordsPanelContent struct {
	guigui.DefaultWidget

	passwordWidgets []passwordWidget
}

func (p *passwordsPanelContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
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

	for i := range model.PasswordCount() {
		pw := model.PasswordByIndex(i)
		p.passwordWidgets[i].SetText(pw.Text)
	}

	return nil
}

func (p *passwordsPanelContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	layout := guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Gap:       u / 4,
	}
	layout.Items = make([]guigui.LinearLayoutItem, len(p.passwordWidgets))
	for i := range p.passwordWidgets {
		w := widgetBounds.Bounds().Dx()
		h := p.passwordWidgets[i].Measure(context, guigui.FixedWidthConstraints(w)).Y
		layout.Items[i] = guigui.LinearLayoutItem{
			Widget: &p.passwordWidgets[i],
			Size:   guigui.FixedSize(h),
		}
	}
	layout.LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *passwordsPanelContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range p.passwordWidgets {
		h += p.passwordWidgets[i].Measure(context, constraints).Y
		h += int(u / 4)
	}
	w := p.DefaultWidget.Measure(context, constraints).X
	return image.Pt(w, h)
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
