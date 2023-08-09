package main

import (
	"github.com/czcorpus/cnc-gokit/util"
	"github.com/gdamore/tcell/v2"
)

var (
	styleGreen  = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	styleWhite  = tcell.StyleDefault.Foreground(tcell.ColorWhite)
	styleOrange = tcell.StyleDefault.Foreground(tcell.ColorOrange)
	stylePurple = tcell.StyleDefault.Foreground(tcell.ColorPurple)
	styleGrey   = tcell.StyleDefault.Foreground(tcell.ColorGrey)
)

type Drawer struct {
	fn          *FileNavigator
	controlLine string
	screen      tcell.Screen
	dtField     string
}

func (drw *Drawer) ShowControlMessage(msg string) {
	drw.controlLine = msg
}

func (drw *Drawer) drawBuffer() {
	var row int
	drw.fn.ForItemsBuffer(func(_ int, line *Line) bool {
		var col int
		for _, chunk := range line.Data.Chunks(drw.dtField) {
			sr := chunk.StringValue()
			currStyle := tcell.StyleDefault
			if chunk.IsKey() {
				currStyle = styleGreen

			} else if chunk.IsDatetime() {
				currStyle = styleOrange

			} else if chunk.IsMapValue() {
				currStyle = stylePurple

			} else if chunk.IsNil() {
				currStyle = styleGrey
			}
			for _, r := range sr {
				drw.screen.SetContent(col, row, r, nil, currStyle)
				col++
			}
			if chunk.IsKey() {
				drw.screen.SetContent(col, row, ':', nil, tcell.StyleDefault)
				col++
			}
			drw.screen.SetContent(col, row, ' ', nil, tcell.StyleDefault)
			col++
			if col >= drw.getWidth() {
				break
			}

		}
		row++
		return true
	})
}

func (drw *Drawer) getWidth() int {
	w, _ := drw.screen.Size()
	return w
}

func (drw *Drawer) draw() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	drw.screen.SetStyle(defStyle)
	drw.screen.Clear()

	width, height := drw.screen.Size()
	mainHeight := height - 1 // Reserve one line for controls

	// Draw main area
	/*
		for row := 0; row < mainHeight; row++ {
			for col := 0; col < width; col++ {
				drw.screen.SetContent(col, row, ' ', nil, tcell.StyleDefault)
			}
		}
	*/
	drw.fn.Init(mainHeight - 1)
	drw.drawBuffer()

	// Draw control line
	clRunes := []rune(drw.controlLine)
	for col := 0; col < util.Min(width, len(clRunes)); col++ {
		drw.screen.SetContent(col, mainHeight, clRunes[col], nil, tcell.StyleDefault.Background(tcell.ColorDarkCyan))
	}

	drw.screen.Show()

	for {
		// Update screen
		drw.screen.Show()

		// Poll event
		ev := drw.screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			drw.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return
			case 258:
				drw.fn.NextLine()
				drw.drawBuffer()
			case 257:
				drw.fn.PreviousLine()
				drw.drawBuffer()
			}
		}
	}
}

func NewDrawer(
	fn *FileNavigator,
	screen tcell.Screen,
	dtField string,
) *Drawer {

	return &Drawer{
		fn:      fn,
		screen:  screen,
		dtField: dtField,
	}
}
