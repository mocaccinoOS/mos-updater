package gui

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	process "github.com/mudler/go-processmanager"
)

type dashboard struct {
	window fyne.Window
}

func stateDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(dirname, ".mos-updater")
}

const welcomeMessage string = `
# MocaccinoOS updater
`

const upgradeCommand = "pkexec /bin/bash -c 'luet upgrade --no-spinner -y --color=false'"

func (c *dashboard) Reload(app fyne.App) {
	state := stateDir()
	processStateDir := filepath.Join(state, "process")
	os.MkdirAll(state, os.ModePerm)

	upgrade := process.New(
		process.WithStateDir(processStateDir),
	)

	upgradeButton := widget.NewButton("Upgrade", func() {})

	t := widget.NewLabel("")
	accContent := container.NewVScroll(
		t,
	)

	t.Wrapping = fyne.TextWrapWord

	statusText := widget.NewRichTextFromMarkdown(welcomeMessage)
	upgradeButton.Show()
	t.Hide()

	attachProcess := func() {
		statusText.ParseMarkdown("# Upgrade in progress")
		pr := process.New(
			process.WithStateDir(processStateDir),
		)
		t.Show()
		t.Resize(fyne.NewSize(100, 100))
		accContent.Resize(fyne.NewSize(100, 100))
		ss := newTail(t, accContent)
		tailProcess(context.Background(), pr, ss)
		waitProcess(pr, upgradeButton, statusText)
	}

	if upgrade.IsAlive() {
		// Load upgrade
		attachProcess()
	} else {
		if availableUpgrades() {
			upgradeButton.OnTapped = func() {
				t.Show()
				//t.Resize(fyne.NewSize(100, 100))
				//	accContent.Resize(fyne.NewSize(100, 100))
				pr := process.New(
					process.WithName("/bin/bash"),
					process.WithArgs("-c", upgradeCommand),
					process.WithStateDir(processStateDir),
				)
				pr.Stop()
				os.RemoveAll(processStateDir)
				os.MkdirAll(processStateDir, os.ModePerm)
				if err := pr.Run(); err != nil {
					errorWindow(err, app)
				}
				attachProcess()
			}
			upgradeButton.Importance = widget.HighImportance
			upgradeButton.SetText("Upgrades available")
		} else {
			upgradeButton.Disable()
			upgradeButton.SetText("No available upgrades")
		}
	}

	c.window.SetContent(
		//	container.NewHBox(
		//		layout.NewSpacer(),
		container.NewBorder(
			statusText,
			upgradeButton,
			nil,
			nil,
			accContent,
			// container.NewScroll(container.NewGridWithColumns(
			// 	4,
			// 	cards...,
			// )),
		),
	//	),
	)
}

func (c *dashboard) loadUI(app fyne.App) {
	c.window = app.NewWindow("Updater")
	c.Reload(app)
	c.window.Resize(fyne.NewSize(640, 640))
	c.window.SetFixedSize(true)
	c.window.Show()
}

func newDashboard() *dashboard {
	return &dashboard{}
}
