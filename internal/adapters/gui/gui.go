package gui

import (
	"fmt"
	"time"

	"streamgogambler/internal/adapters/gui/assets"
	"streamgogambler/internal/ports"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type StatsProvider interface {
	GetStats() ports.BotStats
	IsAutoSlotsEnabled() bool
	SetAutoSlots(enabled bool)
	ExecuteCommand(command string)
}

type GUI struct {
	app           fyne.App
	window        fyne.Window
	statsProvider StatsProvider

	statusLabel   *widget.Label
	channelLabel  *widget.Label
	usernameLabel *widget.Label
	uptimeLabel   *widget.Label
	balanceLabel  *widget.Label
	sentLabel     *widget.Label
	recvLabel     *widget.Label
	reconnLabel   *widget.Label

	logList  *widget.List
	logLines []string
	maxLogs  int

	autoSlotsChk *widget.Check
	commandInput *widget.Entry

	stopChan chan struct{}
}

func New(statsProvider StatsProvider, maxLogs int) *GUI {
	return &GUI{
		statsProvider: statsProvider,
		stopChan:      make(chan struct{}),
		maxLogs:       maxLogs,
		logLines:      make([]string, 0, maxLogs),
	}
}

func (g *GUI) Run() {
	g.app = app.New()
	g.app.SetIcon(assets.AppIcon())
	g.window = g.app.NewWindow("StreamGoGambler Bot")
	g.window.Resize(fyne.NewSize(800, 600))

	g.buildUI()
	g.startUpdateLoop()
	g.setupSystemTray()

	g.window.SetCloseIntercept(func() {
		g.window.Hide()
	})

	g.window.ShowAndRun()
}

func (g *GUI) setupSystemTray() {
	if desk, ok := g.app.(desktop.App); ok {
		menu := fyne.NewMenu("StreamGoGambler",
			fyne.NewMenuItem("Show", func() {
				g.window.Show()
			}),
			fyne.NewMenuItem("Hide", func() {
				g.window.Hide()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				g.app.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}

func (g *GUI) Stop() {
	close(g.stopChan)
	if g.app != nil {
		g.app.Quit()
	}
}

func (g *GUI) AppendLog(message string) {
	g.logLines = append(g.logLines, message)

	if len(g.logLines) > g.maxLogs {
		g.logLines = g.logLines[len(g.logLines)-g.maxLogs:]
	}

	if g.logList != nil {
		g.logList.Refresh()
		g.logList.ScrollToBottom()
	}
}

func (g *GUI) sendCommand() {
	if g.commandInput == nil || g.statsProvider == nil {
		return
	}

	command := g.commandInput.Text
	if command == "" {
		return
	}

	g.statsProvider.ExecuteCommand(command)
	g.commandInput.SetText("")
}

func (g *GUI) buildUI() {
	g.statusLabel = widget.NewLabel("Status: Unknown")
	g.channelLabel = widget.NewLabel("Channel: -")
	g.usernameLabel = widget.NewLabel("Username: -")
	g.uptimeLabel = widget.NewLabel("Uptime: -")

	statusCard := widget.NewCard("Connection", "",
		container.NewVBox(
			g.statusLabel,
			g.channelLabel,
			g.usernameLabel,
			g.uptimeLabel,
		),
	)

	g.balanceLabel = widget.NewLabel("Balance: 0 bombs")
	g.sentLabel = widget.NewLabel("Messages Sent: 0")
	g.recvLabel = widget.NewLabel("Messages Received: 0")
	g.reconnLabel = widget.NewLabel("Reconnects: 0")

	statsCard := widget.NewCard("Statistics", "",
		container.NewVBox(
			g.balanceLabel,
			g.sentLabel,
			g.recvLabel,
			g.reconnLabel,
		),
	)

	g.autoSlotsChk = widget.NewCheck("Auto Slots Enabled", func(checked bool) {
		if g.statsProvider != nil {
			g.statsProvider.SetAutoSlots(checked)
		}
	})
	if g.statsProvider != nil {
		g.autoSlotsChk.Checked = g.statsProvider.IsAutoSlotsEnabled()
	}

	controlsCard := widget.NewCard("Controls", "",
		container.NewVBox(
			g.autoSlotsChk,
		),
	)

	g.logList = widget.NewList(
		func() int {
			return len(g.logLines)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(g.logLines) {
				return
			}
			o.(*widget.Label).SetText(g.logLines[i])
		},
	)

	logCard := widget.NewCard("Activity Log", "", g.logList)

	g.commandInput = widget.NewEntry()
	g.commandInput.SetPlaceHolder("Type a command (e.g., !trust, !ustaw 100) or send a message...")

	sendButton := widget.NewButton("Send", func() {
		g.sendCommand()
	})

	g.commandInput.OnSubmitted = func(_ string) {
		g.sendCommand()
	}

	commandSection := container.NewBorder(
		nil,
		nil,
		nil,
		sendButton,
		g.commandInput,
	)

	topSection := container.NewHBox(
		statusCard,
		statsCard,
		controlsCard,
	)

	content := container.NewBorder(
		topSection,
		commandSection,
		nil,
		nil,
		logCard,
	)

	g.window.SetContent(content)
}

func (g *GUI) startUpdateLoop() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-g.stopChan:
				return
			case <-ticker.C:
				g.updateStats()
			}
		}
	}()
}

func (g *GUI) updateStats() {
	if g.statsProvider == nil {
		return
	}

	stats := g.statsProvider.GetStats()

	g.statusLabel.SetText(fmt.Sprintf("Status: %s", stats.Status))
	g.channelLabel.SetText(fmt.Sprintf("Channel: #%s", stats.Channel))
	g.usernameLabel.SetText(fmt.Sprintf("Username: %s", stats.Username))
	g.uptimeLabel.SetText(fmt.Sprintf("Uptime: %s", stats.Uptime))

	g.balanceLabel.SetText(fmt.Sprintf("Balance: %d bombs", stats.Balance))
	g.sentLabel.SetText(fmt.Sprintf("Messages Sent: %d", stats.MessagesSent))
	g.recvLabel.SetText(fmt.Sprintf("Messages Received: %d", stats.MessagesRecv))
	g.reconnLabel.SetText(fmt.Sprintf("Reconnects: %d", stats.ReconnectCount))

	if g.autoSlotsChk.Checked != g.statsProvider.IsAutoSlotsEnabled() {
		g.autoSlotsChk.Checked = g.statsProvider.IsAutoSlotsEnabled()
		g.autoSlotsChk.Refresh()
	}
}

type SetupResult struct {
	Values    map[string]string
	Completed bool
}

func ShowSetupDialog(missingVars []string, defaults map[string]string) SetupResult {
	result := SetupResult{
		Values:    make(map[string]string),
		Completed: false,
	}

	a := app.New()
	w := a.NewWindow("StreamGoGambler Setup")
	w.Resize(fyne.NewSize(800, 600))

	inputs := make(map[string]*widget.Entry)

	varDescriptions := map[string]string{
		"TWITCH_USERNAME": "Twitch Bot Username",
		"TWITCH_OAUTH":    "Twitch OAuth Token (without 'oauth:' prefix)",
		"TWITCH_CHANNEL":  "Twitch Channel to Join (without #)",
		"COMMAND_PREFIX":  "Command Prefix (e.g., !)",
		"STATUS_COMMAND":  "Status Command Name (e.g., status)",
		"CONNECT_MESSAGE": "Message on Connect (e.g., !pyk)",
		"BOSS_BOT_NAME":   "Boss Bot Name (e.g., demonzzbot)",
	}

	formItems := make([]*widget.FormItem, 0, len(missingVars))

	for _, varName := range missingVars {
		entry := widget.NewEntry()

		if def, ok := defaults[varName]; ok {
			entry.SetText(def)
		}

		if varName == "TWITCH_OAUTH" {
			entry = widget.NewPasswordEntry()
		}

		inputs[varName] = entry

		label := varName
		if desc, ok := varDescriptions[varName]; ok {
			label = desc
		}

		formItems = append(formItems, widget.NewFormItem(label, entry))
	}

	form := widget.NewForm(formItems...)

	done := make(chan bool)

	saveBtn := widget.NewButton("Save & Start Bot", func() {
		allFilled := true
		for _, entry := range inputs {
			if entry.Text == "" {
				allFilled = false
				break
			}
		}

		if !allFilled {
			errDialog := widget.NewLabel("Please fill in all fields!")
			errDialog.Importance = widget.HighImportance
			w.SetContent(container.NewVBox(
				widget.NewLabel("Error"),
				errDialog,
				widget.NewButton("OK", func() {
					w.SetContent(buildSetupContent(form, inputs, done, &result))
				}),
			))
			return
		}

		for varName, entry := range inputs {
			result.Values[varName] = entry.Text
		}
		result.Completed = true
		done <- true
	})

	cancelBtn := widget.NewButton("Exit", func() {
		done <- false
	})

	content := container.NewVBox(
		widget.NewLabel("StreamGoGambler requires configuration before first run."),
		widget.NewLabel("Please enter the following settings:"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(saveBtn, cancelBtn),
	)

	w.SetContent(content)

	w.SetOnClosed(func() {
		select {
		case done <- false:
		default:
		}
	})

	go func() {
		<-done
		w.Close()
		a.Quit()
	}()

	w.ShowAndRun()

	return result
}

func buildSetupContent(form *widget.Form, inputs map[string]*widget.Entry, done chan bool, result *SetupResult) fyne.CanvasObject {
	saveBtn := widget.NewButton("Save & Start Bot", func() {
		allFilled := true
		for _, entry := range inputs {
			if entry.Text == "" {
				allFilled = false
				break
			}
		}

		if !allFilled {
			return
		}

		for varName, entry := range inputs {
			result.Values[varName] = entry.Text
		}
		result.Completed = true
		done <- true
	})

	cancelBtn := widget.NewButton("Exit", func() {
		done <- false
	})

	return container.NewVBox(
		widget.NewLabel("StreamGoGambler requires configuration before first run."),
		widget.NewLabel("Please enter the following settings:"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(saveBtn, cancelBtn),
	)
}

func ShowErrorDialog(title, message string) {
	a := app.New()
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(400, 150))

	okBtn := widget.NewButton("OK", func() {
		a.Quit()
	})

	content := container.NewVBox(
		widget.NewLabel(message),
		widget.NewSeparator(),
		container.NewCenter(okBtn),
	)

	w.SetContent(content)
	w.SetOnClosed(func() {
		a.Quit()
	})

	w.ShowAndRun()
}
