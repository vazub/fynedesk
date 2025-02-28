package ui

import (
	"os"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

const widgetPanelWidth = float32(200)

type widgetRenderer struct {
	panel *widgetPanel
	bg    *canvas.Rectangle

	layout  fyne.Layout
	objects []fyne.CanvasObject
}

func (w *widgetRenderer) MinSize() fyne.Size {
	return w.layout.MinSize(w.objects)
}

func (w *widgetRenderer) Layout(size fyne.Size) {
	w.bg.Resize(size)
	w.layout.Layout(w.objects[1:], size)
}

func (w *widgetRenderer) Refresh() {
	r, _, _, _ := theme.BackgroundColor().RGBA()
	if uint8(r) > 0x99 {
		w.bg.FillColor = wmtheme.WidgetPanelBackgroundLight
	} else {
		w.bg.FillColor = wmtheme.WidgetPanelBackgroundDark
	}
	w.bg.Refresh()

	w.panel.clock.Color = theme.ForegroundColor()
	canvas.Refresh(w.panel.clock)
}

func (w *widgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *widgetRenderer) Destroy() {
}

type widgetPanel struct {
	widget.BaseWidget

	desk            fynedesk.Desktop
	about, settings fyne.Window

	clock         *canvas.Text
	date          *widget.Label
	modules       *fyne.Container
	notifications fyne.CanvasObject
}

func (w *widgetPanel) clockTick() {
	tick := time.NewTicker(time.Second)
	go func() {
		for {
			<-tick.C
			w.clock.Text = w.formattedTime()
			canvas.Refresh(w.clock)

			w.date.SetText(formattedDate())
			canvas.Refresh(w.date)
		}
	}()
}

func (w *widgetPanel) formattedTime() string {
	if w.desk.Settings().ClockFormatting() == "12h" {
		return time.Now().Format("03:04pm")
	}

	return time.Now().Format("15:04")
}

func formattedDate() string {
	return time.Now().Format("2 January")
}

func (w *widgetPanel) createClock() {
	var style fyne.TextStyle
	style.Monospace = true

	w.clock = &canvas.Text{
		Color:     theme.ForegroundColor(),
		Text:      w.formattedTime(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
		TextSize:  3 * theme.TextSize(),
	}
	w.date = &widget.Label{
		Text:      formattedDate(),
		Alignment: fyne.TextAlignCenter,
		TextStyle: style,
	}

	go w.clockTick()
}

func (w *widgetPanel) CreateRenderer() fyne.WidgetRenderer {
	accountLabel := "Account"
	homedir, err := os.UserHomeDir()
	if err == nil {
		accountLabel = path.Base(homedir)
	} else {
		fyne.LogError("Unable to look up user", err)
	}
	var account *widget.Button
	account = widget.NewButtonWithIcon(accountLabel, wmtheme.UserIcon, func() {
		w.showAccountMenu(account)
	})
	appExecButton := widget.NewButtonWithIcon("", theme.SearchIcon(), ShowAppLauncher)

	bg := canvas.NewRectangle(wmtheme.WidgetPanelBackgroundDark)
	objects := []fyne.CanvasObject{
		bg,
		w.clock,
		w.date,
		w.notifications}

	w.modules = container.NewVBox()
	objects = append(objects, layout.NewSpacer(), w.modules,
		container.NewBorder(nil, nil, appExecButton, nil, account))
	w.loadModules(w.desk.Modules())

	return &widgetRenderer{
		panel:   w,
		bg:      bg,
		layout:  layout.NewVBoxLayout(),
		objects: objects,
	}
}

func (w *widgetPanel) MinSize() fyne.Size {
	return fyne.NewSize(widgetPanelWidth, 200)
}

func (w *widgetPanel) reloadModules(mods []fynedesk.Module) {
	w.modules.Objects = nil
	w.loadModules(mods)
	w.modules.Refresh()
}

func (w *widgetPanel) loadModules(mods []fynedesk.Module) {
	for _, m := range mods {
		if statusMod, ok := m.(fynedesk.StatusAreaModule); ok {
			wid := statusMod.StatusAreaWidget()
			if wid == nil {
				continue
			}

			w.modules.Objects = append(w.modules.Objects, wid)
		}
	}
}

func newWidgetPanel(rootDesk fynedesk.Desktop) *widgetPanel {
	w := &widgetPanel{desk: rootDesk}
	w.ExtendBaseWidget(w)
	w.notifications = startNotifications()
	w.createClock()

	return w
}
