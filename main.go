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

var (
	modelKeyModel = guigui.GenerateEnvKey()
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

	layoutItems []guigui.LinearLayoutItem
}

func (r *Root) Env(context *guigui.Context, key guigui.EnvKey, source *guigui.EnvSource) (any, bool) {
	switch key {
	case modelKeyModel:
		return &r.model, true
	default:
		return nil, false
	}
}

var (
	passwordsPanelContentEventClearTriggered guigui.EventKey = guigui.GenerateEventKey()
)

func (r *Root) OnClearTriggered(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, passwordsPanelContentEventClearTriggered, f)
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.background)
	adder.AddWidget(&r.form)
	adder.AddWidget(&r.passwordsPanel)

	r.countOutputText.SetValue("count of output")
	r.countOutputNumberInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetCountOutputValue(value)
	})
	r.countOutputNumberInput.SetMinimumValue(1)
	r.countOutputNumberInput.SetValue(r.model.CountOutputValue())

	r.numberCharsText.SetValue("number of characters")
	r.numberCharsNumberInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetNumberCharsValue(value)
	})
	r.numberCharsNumberInput.SetMinimumValue(1)
	r.numberCharsNumberInput.SetValue(r.model.NumberCharsValue())

	r.minNumsText.SetValue("minimum count of numbers")
	r.minNumsNumberInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinNumsValue(value)
	})
	r.minNumsNumberInput.SetMinimumValue(-1)
	r.minNumsNumberInput.SetValue(r.model.MinNumsValue())

	r.minSymbolsText.SetValue("minimum count of symbols")
	r.minSymbolsNumberInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		r.model.SetMinSymbolsValue(value)
	})
	r.minSymbolsNumberInput.SetMinimumValue(-1)
	r.minSymbolsNumberInput.SetValue(r.model.MinSymbolsValue())

	r.once.Do(func() { r.reset() })

	r.resetButton.SetText("Reset")
	r.resetButton.OnUp(func(context *guigui.Context) {
		r.reset()
	})

	r.generateButton.SetText("Generate")
	r.generateButton.SetType(basicwidget.ButtonTypePrimary)
	r.generateButton.OnUp(func(context *guigui.Context) {
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

	r.OnClearTriggered(func(context *guigui.Context) {
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
	r.layoutItems = slices.Delete(r.layoutItems, 0, len(r.layoutItems))
	r.layoutItems = append(r.layoutItems,
		guigui.LinearLayoutItem{
			Widget: &r.form,
			Size:   guigui.FixedSize(r.form.Measure(context, guigui.Constraints{}).Y),
		},
		guigui.LinearLayoutItem{
			Widget: &r.passwordsPanel,
			Size:   guigui.FlexibleSize(1),
		},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     r.layoutItems,
		Gap:       u / 2,
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

	layoutItems []guigui.LinearLayoutItem
}

func (p *passwordWidget) SetText(text string) {
	p.text.SetValue(text)
}

func (p *passwordWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&p.copyButton)
	adder.AddWidget(&p.text)

	p.copyButton.SetText("Copy")
	p.copyButton.OnUp(func(context *guigui.Context) {
		clipboard.WriteAll(p.text.Value())
	})
	p.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	return nil
}

func (p *passwordWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	p.layoutItems = slices.Delete(p.layoutItems, 0, len(p.layoutItems))
	p.layoutItems = append(p.layoutItems,
		guigui.LinearLayoutItem{
			Widget: &p.copyButton,
			Size:   guigui.FixedSize(3 * u),
		},
		guigui.LinearLayoutItem{
			Widget: &p.text,
			Size:   guigui.FlexibleSize(1),
		},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     p.layoutItems,
		Gap:       u / 2,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *passwordWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), p.copyButton.Measure(context, guigui.Constraints{}).Y)
}

type passwordsPanelContent struct {
	guigui.DefaultWidget

	passwordWidgets guigui.WidgetSlice[*passwordWidget]

	layoutItems []guigui.LinearLayoutItem
}

func (p *passwordsPanelContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	v, ok := context.Env(p, modelKeyModel)
	if !ok {
		return nil
	}
	model := v.(*Model)

	p.passwordWidgets.SetLen(model.PasswordCount())
	for i := range p.passwordWidgets.Len() {
		adder.AddWidget(p.passwordWidgets.At(i))
	}

	for i := range model.PasswordCount() {
		pw := model.PasswordByIndex(i)
		p.passwordWidgets.At(i).SetText(pw.Text)
	}

	return nil
}

func (p *passwordsPanelContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	p.layoutItems = slices.Delete(p.layoutItems, 0, len(p.layoutItems))
	for i := range p.passwordWidgets.Len() {
		w := widgetBounds.Bounds().Dx()
		h := p.passwordWidgets.At(i).Measure(context, guigui.FixedWidthConstraints(w)).Y
		p.layoutItems = append(p.layoutItems, guigui.LinearLayoutItem{
			Widget: p.passwordWidgets.At(i),
			Size:   guigui.FixedSize(h),
		})
	}
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     p.layoutItems,
		Gap:       u / 4,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *passwordsPanelContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range p.passwordWidgets.Len() {
		h += p.passwordWidgets.At(i).Measure(context, constraints).Y
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
