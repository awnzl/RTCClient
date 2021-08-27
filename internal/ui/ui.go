package ui

import (
	"log"

	"github.com/jroimartin/gocui"
)

const (
	outputViewName = "output_view"
	inputViewName  = "input_view"
)

type GUI struct {
	gui      *gocui.Gui
	Sender   chan string
	Receiver chan string
}

func NewGUI() (*GUI, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}

	return &GUI{
		gui:      g,
		Sender:   make(chan string),
		Receiver: make(chan string),
	}, nil
}

func (ui *GUI) Close() {
	ui.gui.Close()
}

func (ui *GUI) Init() error {
	ui.gui.Highlight = true
	ui.gui.Cursor = true
	ui.gui.SelFgColor = gocui.ColorGreen

	ui.gui.SetManagerFunc(layout)

	return ui.initKeybindings(ui.gui)
}

func (ui *GUI) MainLoop() error {
	go ui.updateMessageViewContent()

	if err := ui.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

func (ui *GUI) initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding(inputViewName, gocui.KeyEnter, gocui.ModNone, ui.copyViewBuf()); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	outputViewSizeY := maxY - 6

	if v, err := g.SetView(outputViewName, 0, 0, maxX-1, outputViewSizeY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Autoscroll = true
		v.Title = " Real Time Chat "
		v.Wrap = true
	}

	if v, err := g.SetView(inputViewName, 0, outputViewSizeY+1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		if _, err := g.SetCurrentView(inputViewName); err != nil {
			return err
		}

		v.Autoscroll = true
		v.Editable = true
		v.Wrap = true
	}

	return nil
}

func (ui *GUI) copyViewBuf() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		ui.Sender <- v.Buffer()

		if err := v.SetCursor(0, 0); err != nil {
			return err
		}

		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}

		v.Clear()

		return nil
	}
}

func (ui *GUI) updateOutputViewBuf(msg string) error {
	dstView, err := ui.gui.View(outputViewName)
	if err != nil {
		return err
	}

	content := dstView.Buffer() + msg

	ui.gui.Update(func(g *gocui.Gui) error {
		v, err := g.View(outputViewName)
		if err != nil {
			return err
		}

		v.Clear()
		if _, err := v.Write([]byte(content)); err != nil {
			return err
		}

		return nil
	})

	return nil
}

func (ui *GUI) updateMessageViewContent() {
	for {
		err := ui.updateOutputViewBuf(<-ui.Receiver)
		if err != nil {
			log.Println("[g.updateViewBuf()]", err)
		}
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
