package main

import (
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"arbok-terminal/internal/input"
	"arbok-terminal/internal/pty"
	"arbok-terminal/internal/terminal"
	"arbok-terminal/internal/ui"
)

func main() {
	fyneApp := app.NewWithID("arbok.terminal")
	window := fyneApp.NewWindow("Arbok Terminal")

	cellW := float32(9)
	cellH := float32(17)

	calcSize := func(s fyne.Size) (cols, rows int) {
		cols = int(s.Width / cellW)
		rows = int(s.Height / cellH)
		if cols < 1 {
			cols = 1
		}
		if rows < 1 {
			rows = 1
		}
		return
	}

	initW, initH := calcSize(fyne.NewSize(800, 600))

	vt := terminal.NewVirtualTerminal(initW, initH)
	parser := terminal.NewParser(vt)
	inputH := input.New()
	termWidget := ui.New(vt, inputH)

	parser.TitleHandler = func(title string) {
		fyne.Do(func() { window.SetTitle(title) })
	}

	ptym, err := pty.New(os.Getenv("SHELL"), initW, initH)
	if err != nil {
		panic(err)
	}
	defer ptym.Close()

	inputH.SetOnInput(func(data []byte) { ptym.Write(data) })

	refreshCh := make(chan struct{}, 1)

	go func() {
		for data := range ptym.OutputCh {
			parser.Parse(data)
			if title := vt.PendingTitle(); title != "" {
				parser.SetTitle(title)
			}
			select {
			case refreshCh <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		var lastCols, lastRows int

		for range ticker.C {
			fyne.Do(func() {
				cols, rows := calcSize(window.Canvas().Size())
				if cols != lastCols || rows != lastRows {
					if cols != vt.Width() || rows != vt.Height() {
						go func(c, r int) {
							vt.Resize(c, r)
							ptym.Resize(c, r)
							fyne.Do(func() {
								termWidget.Refresh()
								window.Canvas().Refresh(termWidget)
							})
						}(cols, rows)
					}
					lastCols, lastRows = cols, rows
				}
			})
		}
	}()

	go func() {
		for range refreshCh {
			fyne.Do(func() {
				termWidget.Refresh()
				window.Canvas().Refresh(termWidget)
			})
		}
	}()

	window.SetContent(termWidget)
	window.Resize(fyne.NewSize(800, 600))
	window.SetMaster()
	window.Canvas().Focus(termWidget)
	window.Show()

	fyneApp.Run()
}
