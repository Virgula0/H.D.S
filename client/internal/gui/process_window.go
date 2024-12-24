package gui

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strings"
)

const (
	WindowWidth      = 1280
	WindowHeight     = 900
	DefaultLogHeight = 400
	LogScrollSpeed   = 20
	MinLogHeight     = 200

	LabelSpacing  = 250
	LabelFontSize = 20
	LogFontSize   = 18

	// Point this to where your Roboto-Black.ttf (or any TTF) is located
	FontPath = "/home/angelo/Universita/DP/H.D.S/client/internal/resources/fonts/Roboto-Black.ttf"
)

// ProcessWindowInfo now stores just a *single string* for logs.
type ProcessWindowInfo struct {
	isConnected   bool
	statusLabel   string
	pcapFile      string
	hashcatFile   string
	hashcatStatus string

	// One big string that may contain multiple lines
	logs string
}

// StateUpdate describes how to update the GUI state in one shot.
type StateUpdate struct {
	StatusLabel   string
	IsConnected   bool
	PCAPFile      string
	HashcatFile   string
	HashcatStatus string
	// This entire string replaces the GUI's logs
	LogContent string
}

// applyStateUpdate overwrites the GUI state with incoming data.
func applyStateUpdate(state *ProcessWindowInfo, update StateUpdate) {
	state.isConnected = update.IsConnected
	state.statusLabel = update.StatusLabel
	state.pcapFile = update.PCAPFile
	state.hashcatFile = update.HashcatFile
	state.hashcatStatus = update.HashcatStatus
	// Replace the entire logs string
	state.logs = update.LogContent
}

// RunGUI starts the Raylib window and listens for updates via the channel.
func RunGUI(stateUpdateCh <-chan StateUpdate) {
	// 1) Initialize Raylib
	rl.InitWindow(WindowWidth, WindowHeight, "Process Window - Single Log String")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// 2) Load font (use the default font if the custom one isn't found)
	font := rl.LoadFont(FontPath)
	defer rl.UnloadFont(font)

	// 3) Initial GUI state
	state := &ProcessWindowInfo{
		isConnected:   false,
		statusLabel:   "Initializing...",
		pcapFile:      "N/A",
		hashcatFile:   "N/A",
		hashcatStatus: "Idle",
		logs:          "", // initially empty
	}

	// 4) Variables for log box scrolling/resizing
	var (
		logOffsetY float32 = 0
		logHeight  float32 = DefaultLogHeight
		isResizing bool    = false
	)

	// 5) Main loop
	for !rl.WindowShouldClose() {
		// -- Process incoming updates (non-blocking) --
		select {
		case update := <-stateUpdateCh:
			applyStateUpdate(state, update)
		default:
			// No updates pending
		}

		// -- Handle user input (scroll, resize) --
		updateGUIState(&logOffsetY, &logHeight, &isResizing, state)

		// -- Draw everything --
		drawGUI(state, font, logOffsetY, logHeight)
	}
}

// updateGUIState handles scrolling and resizing for the log box.
func updateGUIState(
	logOffsetY *float32,
	logHeight *float32,
	isResizing *bool,
	state *ProcessWindowInfo,
) {
	mousePos := rl.GetMousePosition()

	// The top of the log box
	logBoxY := float32(WindowHeight) - *logHeight

	// 1) Scroll if hovering over the log box
	if mousePos.Y > logBoxY && mousePos.Y < float32(WindowHeight) {
		if rl.IsKeyDown(rl.KeyUp) {
			*logOffsetY += LogScrollSpeed
		}
		if rl.IsKeyDown(rl.KeyDown) {
			*logOffsetY -= LogScrollSpeed
		}
	}

	// We need to clamp scrolling. We'll estimate the total text height
	// by splitting into lines and counting them. Each line ~ LogFontSize + 2
	lines := strings.Split(state.logs, "\n")
	totalTextHeight := float32(len(lines)) * (LogFontSize + 2)
	maxOffsetY := totalTextHeight - *logHeight

	if *logOffsetY > 0 {
		*logOffsetY = 0
	}
	if *logOffsetY < -maxOffsetY {
		*logOffsetY = -maxOffsetY
	}

	// 2) Handle resizing of the log box
	resizeBarY := float32(WindowHeight) - *logHeight - 5
	if mousePos.Y > resizeBarY && mousePos.Y < resizeBarY+10 {
		rl.SetMouseCursor(rl.MouseCursorResizeNS)
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			*isResizing = true
		}
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		*isResizing = false
	}

	if *isResizing {
		newHeight := float32(WindowHeight) - mousePos.Y
		// Example max range for the log box
		maxLogHeight := float32(WindowHeight - 240)
		switch {
		case newHeight < MinLogHeight:
			*logHeight = MinLogHeight
		case newHeight > maxLogHeight:
			*logHeight = maxLogHeight
		default:
			*logHeight = newHeight
		}
	}
}

// drawGUI draws the GUI each frame, including the single log string.
func drawGUI(state *ProcessWindowInfo, font rl.Font, logOffsetY float32, logHeight float32) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	// 1) Basic labels at the top
	labelStartX := float32(20)
	valueStartX := labelStartX + LabelSpacing

	rl.DrawTextEx(font, "Status", rl.NewVector2(labelStartX, 20), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, fmt.Sprintf("%v", state.isConnected), rl.NewVector2(valueStartX, 20), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Task Status", rl.NewVector2(labelStartX, 60), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.statusLabel, rl.NewVector2(valueStartX, 60), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "PCAP File", rl.NewVector2(labelStartX, 100), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.pcapFile, rl.NewVector2(valueStartX, 100), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat File", rl.NewVector2(labelStartX, 140), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatFile, rl.NewVector2(valueStartX, 140), LabelFontSize, 1, rl.Black)

	// 2) Draw the log box
	logBoxY := float32(WindowHeight) - logHeight
	logBox := rl.NewRectangle(0, logBoxY, WindowWidth, logHeight)
	rl.DrawRectangleRec(logBox, rl.Black)

	// 3) Draw the single logs string (split by newline)
	lines := strings.Split(state.logs, "\n")
	y := logBoxY + 10 + logOffsetY
	for _, line := range lines {
		rl.DrawTextEx(font, line, rl.NewVector2(10, y), LogFontSize, 1, rl.White)
		y += LogFontSize + 2
	}

	rl.EndDrawing()
}
