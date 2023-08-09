package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Printf("JoLoot\n")
		flag.PrintDefaults()
	}
	dateField := flag.String("dt", "date", "A root field used for storing date and time")

	flag.Parse()
	filePath := flag.Arg(0)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	_, height := screen.Size()

	fn, err := NewFileNavigator(filePath, height-1)
	if err != nil {
		log.Fatal("Failed to init file navigator: ", err)
	}

	quit := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	drawer := NewDrawer(fn, screen, *dateField)
	drawer.draw()
	drawer.ShowControlMessage("control bar ready")
}
