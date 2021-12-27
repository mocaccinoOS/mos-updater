package gui

//go:generate fyne bundle -package gui -o data.go ../Icon.png

import (
	"context"
	"fmt"
	"image/color"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	process "github.com/mudler/go-processmanager"
	"github.com/nxadm/tail"
)

func Run() {
	app := app.New()
	//	app.SetIcon(resourceIconPng)
	app.SetIcon(resourceIconPng)
	c := newDashboard()
	c.loadUI(app)

	app.Run()
}

func errorWindow(err error, app fyne.App) {
	w := app.NewWindow("Error")
	w.SetContent(canvas.NewText("Failed to parse IP:"+err.Error(), color.White))
	w.Show()
}

func availableUpgrades() bool {
	cmd := exec.Command("pkexec", "/bin/bash", "-c", `luet upgrade`)
	b, _ := cmd.CombinedOutput()
	if strings.Contains(string(b), "Nothing to upgrade") {
		return false
	} else {
		return true
	}
}

func tailProcess(ctx context.Context, pr *process.Process, c chan string) {

	go func() {
		t, _ := tail.TailFile(pr.StdoutPath(), tail.Config{Follow: true})

		for {
			select {
			case line := <-t.Lines: // Every 100ms increate number of ticks and update UI
				c <- line.Text
			case <-ctx.Done():
				return
			}
		}

	}()
	go func() {
		t, _ := tail.TailFile(pr.StderrPath(), tail.Config{Follow: true})
		for {
			select {
			case line := <-t.Lines: // Every 100ms increate number of ticks and update UI
				c <- line.Text
			case <-ctx.Done():
				return
			}
		}
	}()
}

func newTail(t *widget.Label, w *container.Scroll) chan string {
	s := make(chan string)

	go func() {
		for ss := range s {
			curr := t.Text
			curr += ss + "\n"
			t.SetText(curr)
			w.ScrollToBottom()
		}
	}()
	return s
}

func waitProcess(pr *process.Process, button *widget.Button, label *widget.RichText) {
	go func() {
		for pr.IsAlive() {
			time.Sleep(1 * time.Second)
			fmt.Println("Process is alive")
		}

		fmt.Println("Process is done")
		c, _ := pr.ExitCode()
		if c == "0" {
			button.Hide()
			label.ParseMarkdown("# Upgrade successfull")
		} else {
			label.ParseMarkdown("# Upgrade failed")
		}
		label.Show()
	}()
}
