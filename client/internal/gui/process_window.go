package gui

import (
	"context"
	"fmt"
	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	rl "github.com/gen2brain/raylib-go/raylib"
	log "github.com/sirupsen/logrus"
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

var StateUpdateCh = make(chan StateUpdate, 1)

// GuiLogger periodically sends GUI updates until the context is canceled.
func GuiLogger(otherInfos map[string]string, ctx context.Context) {
	for {
		StateUpdateCh <- StateUpdate{
			StatusLabel:   otherInfos[constants.HashcatStatus],
			IsConnected:   true,
			PCAPFile:      otherInfos[constants.PCAPFile],
			HashcatFile:   otherInfos[constants.HashcatFile],
			HashcatStatus: constants.CrackStatus,
			LogContent:    grpcclient.Logs.String(),
		}
		if ctx.Err() != nil {
			log.Warn("Log goroutine killed")
			return
		}
	}
}

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
func applyStateUpdate(state *ProcessWindowInfo, update *StateUpdate) {
	state.isConnected = update.IsConnected
	state.statusLabel = update.StatusLabel
	state.pcapFile = update.PCAPFile
	state.hashcatFile = update.HashcatFile
	state.hashcatStatus = update.HashcatStatus
	// Replace the entire logs string
	state.logs = update.LogContent
}

// RunGUI starts the Raylib window and listens for updates via the channel.
// RunGUI starts the main Raylib-based GUI loop.
func RunGUI(stateUpdateCh <-chan StateUpdate) bool {
	initializeWindow()
	defer rl.CloseWindow()

	uiFont := loadUIFont()
	defer rl.UnloadFont(uiFont)

	guiState := newDefaultGUIState()

	// Track scrolling & resizing of the log box
	var (
		logOffsetY   float32 = 0
		logBoxHeight float32 = DefaultLogHeight
		isResizing           = false
	)

	// Main application loop
	for !rl.WindowShouldClose() {
		// Handle incoming state updates (non-blocking)
		processStateUpdates(stateUpdateCh, guiState)

		// Handle user input and update GUI state
		handleGUIInput(&logOffsetY, &logBoxHeight, &isResizing, guiState)

		// Draw all GUI elements
		drawGUI(guiState, uiFont, logOffsetY, logBoxHeight)
	}

	return rl.WindowShouldClose()
}

// initializeWindow sets up the Raylib window and framerate.
func initializeWindow() {
	rl.InitWindow(WindowWidth, WindowHeight, "Process Window - Single Log String")
	rl.SetTargetFPS(60)
}

// loadUIFont loads the custom font if available; otherwise falls back to Raylib’s default font.
func loadUIFont() rl.Font {
	return rl.LoadFont(FontPath)
}

// newDefaultGUIState returns a ProcessWindowInfo struct with initial default values.
func newDefaultGUIState() *ProcessWindowInfo {
	return &ProcessWindowInfo{
		isConnected:   false,
		statusLabel:   "Initializing...",
		pcapFile:      "N/A",
		hashcatFile:   "N/A",
		hashcatStatus: "Idle",
		logs:          "", // initially empty
	}
}

// processStateUpdates reads from the StateUpdate channel without blocking.
// If there is an update, it applies it to the current state.
func processStateUpdates(ch <-chan StateUpdate, state *ProcessWindowInfo) {
	select {
	case update := <-ch:
		applyStateUpdate(state, &update)
	default:
		// No updates pending
	}
}

// handleGUIInput aggregates all the user input handling needed to update log scrolling/resizing.
func handleGUIInput(
	logOffsetY *float32,
	logBoxHeight *float32,
	isResizing *bool,
	state *ProcessWindowInfo,
) {
	mousePos := rl.GetMousePosition()

	handleLogScrolling(mousePos, logOffsetY, *logBoxHeight, state)
	handleLogBoxResizing(mousePos, logBoxHeight, isResizing)
}

// handleLogScrolling manages vertical scrolling within the log box area.
func handleLogScrolling(mousePos rl.Vector2, logOffsetY *float32, logBoxHeight float32, state *ProcessWindowInfo) {
	// The top Y coordinate of the log box
	logBoxTopY := float32(WindowHeight) - logBoxHeight

	// Allow scrolling if the mouse is over the log box region
	if mousePos.Y > logBoxTopY && mousePos.Y < float32(WindowHeight) {
		if rl.IsKeyDown(rl.KeyUp) {
			*logOffsetY += LogScrollSpeed
		}
		if rl.IsKeyDown(rl.KeyDown) {
			*logOffsetY -= LogScrollSpeed
		}
	}

	// Estimate total text height to clamp scroll offset
	lines := strings.Split(state.logs, "\n")
	totalTextHeight := float32(len(lines)) * (LogFontSize + 2)
	maxOffsetY := totalTextHeight - logBoxHeight

	if *logOffsetY > 0 {
		*logOffsetY = 0
	}
	if *logOffsetY < -maxOffsetY {
		*logOffsetY = -maxOffsetY
	}
}

// handleLogBoxResizing manages mouse-based resizing of the log box height.
func handleLogBoxResizing(mousePos rl.Vector2, logBoxHeight *float32, isResizing *bool) {
	// Grab the Y coordinate for the resizing "grip"
	resizeBarY := float32(WindowHeight) - *logBoxHeight - 5

	// If mouse is over the resize bar, change cursor icon
	if mousePos.Y > resizeBarY && mousePos.Y < resizeBarY+10 {
		rl.SetMouseCursor(rl.MouseCursorResizeNS)
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			*isResizing = true
		}
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	// If mouse is released, stop resizing
	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		*isResizing = false
	}

	// Update log box height while resizing
	if *isResizing {
		newHeight := float32(WindowHeight) - mousePos.Y
		maxLogHeight := float32(WindowHeight - 240) // Example constraint

		switch {
		case newHeight < MinLogHeight:
			*logBoxHeight = MinLogHeight
		case newHeight > maxLogHeight:
			*logBoxHeight = maxLogHeight
		default:
			*logBoxHeight = newHeight
		}
	}
}

// drawGUI draws the GUI each frame, including the single log string.
func drawGUI(state *ProcessWindowInfo, font rl.Font, logOffsetY, logHeight float32) {
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
