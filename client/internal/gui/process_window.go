package gui

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	rl "github.com/gen2brain/raylib-go/raylib"
	log "github.com/sirupsen/logrus"
)

const (
	windowWidth      = 850
	windowHeight     = 550
	defaultLogHeight = 300

	labelSpacing  = 250
	labelFontSize = 20
	logFontSize   = 18

	fontPath     = "internal/resources/fonts/Roboto-Black.ttf"
	jetBrainPath = "internal/resources/fonts/JetBrainsMono-Regular.ttf"

	topLabelArea = 220

	// Button dimensions
	buttonWidth  float32 = 120
	buttonHeight float32 = 30
	buttonMargin float32 = 10
)

// Track scrolling & resizing of the log box
var (
	logOffsetX   float32 = 0
	logOffsetY   float32 = 0
	logBoxHeight float32 = defaultLogHeight

	resizingLogBox bool

	minLogBoxHeight float32 = 50
	maxLogBoxHeight float32 = windowHeight - topLabelArea

	showModal = false
)

// ------------------------------------------------------------------------------------
// STATE MANAGEMENT
// ------------------------------------------------------------------------------------

// StateUpdate describes how to update the GUI state in one shot.
type StateUpdate struct {
	StatusLabel   string
	GRPCConnected string
	PCAPFile      string
	HashcatFile   string
	HashcatStatus string
	LogContent    string // New log content to append
}

// ProcessWindowInfo stores the current GUI state.
type ProcessWindowInfo struct {
	grpcConnected string
	statusLabel   string
	pcapFile      string
	hashcatFile   string
	hashcatStatus string

	// Logs
	renderedLogs string // Already rendered logs
	newLogs      string // New logs to append next frame

	logsMu           sync.Mutex
	lastReadPosition int // Tracks last read position in grpcclient logs
}

// applyStateUpdate applies the new state to the GUI.
func applyStateUpdate(state *ProcessWindowInfo, update *StateUpdate) {
	state.logsMu.Lock()
	defer state.logsMu.Unlock()

	// Only apply if there's a difference
	if update.GRPCConnected != "" {
		state.grpcConnected = update.GRPCConnected
	}
	if update.StatusLabel != "" && update.StatusLabel != state.statusLabel {
		state.statusLabel = update.StatusLabel
	}
	if update.PCAPFile != "" && update.PCAPFile != state.pcapFile {
		state.pcapFile = update.PCAPFile
	}
	if update.HashcatFile != "" && update.HashcatFile != state.hashcatFile {
		state.hashcatFile = update.HashcatFile
	}
	if update.HashcatStatus != "" && update.HashcatStatus != state.hashcatStatus {
		state.hashcatStatus = update.HashcatStatus
	}

	// If there's new log content, append it
	if update.LogContent != "" {
		state.newLogs += update.LogContent
	}
}

// newDefaultGUIState initializes the default GUI state.
func newDefaultGUIState() *ProcessWindowInfo {
	return &ProcessWindowInfo{
		grpcConnected:    "Initialized...",
		statusLabel:      "Initializing...",
		pcapFile:         "N/A",
		hashcatFile:      "N/A",
		hashcatStatus:    "Idle",
		renderedLogs:     "",
		newLogs:          "",
		lastReadPosition: 0,
	}
}

// GuiLogger reads logs incrementally and sends updates to the GUI.
func GuiLogger(ctx context.Context, stateUpdateCh chan<- *StateUpdate) {
	var lastReadPosition int
	for {
		if ctx.Err() != nil {
			log.Warn("Log routine terminated")
			return
		}

		// Read current logs
		currentLogs := grpcclient.ReadLogs()

		// Extract only new content since the last read position
		if len(currentLogs) > lastReadPosition {
			newContent := currentLogs[lastReadPosition:]
			lastReadPosition = len(currentLogs)

			// Send state update with new logs
			stateUpdateCh <- &StateUpdate{
				LogContent: newContent,
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

var StateUpdateCh = make(chan *StateUpdate, 1)

// RunGUI starts the Raylib window and listens for updates via the channel.
func RunGUI(stateUpdateCh <-chan *StateUpdate) bool {
	initializeWindow()
	defer rl.CloseWindow()

	uiFont := loadUIFont()
	defer rl.UnloadFont(uiFont)
	fontJetBrains := rl.LoadFont(jetBrainPath)
	defer rl.UnloadFont(fontJetBrains)

	guiState := newDefaultGUIState()

	// Main application loop: re-draw everything every frame
	for !rl.WindowShouldClose() {
		// Process new state updates
		processStateUpdates(stateUpdateCh, guiState)

		// Handle resizing of the log box
		handleResize(&logBoxHeight, &resizingLogBox, minLogBoxHeight, maxLogBoxHeight)

		// Handle scrolling (mouse wheel)
		handleScrolling(&logOffsetY, &logOffsetX, logBoxHeight, uiFont, guiState)

		// Draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawGUI(guiState, uiFont, logOffsetX, logOffsetY, logBoxHeight)

		// Draw the "Copy Logs" button
		buttonX := windowWidth - buttonWidth - buttonMargin
		buttonY := float32(windowHeight) - logBoxHeight - buttonHeight - buttonMargin
		buttonRect := rl.NewRectangle(buttonX, buttonY, buttonWidth, buttonHeight)

		rl.DrawRectangleRec(buttonRect, rl.SkyBlue)
		rl.DrawRectangleLinesEx(buttonRect, 2, rl.RayWhite)
		rl.DrawText("Copy Logs", int32(buttonX+10), int32(buttonY+7), 20, rl.Black)

		// Check if the button is clicked
		mousePos := rl.GetMousePosition()
		if rl.CheckCollisionPointRec(mousePos, buttonRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			guiState.logsMu.Lock()
			rl.SetClipboardText(guiState.renderedLogs)
			guiState.logsMu.Unlock()
			log.Warn("Logs copied to clipboard")
			showModal = true
		}

		if showModal {
			if DrawMessageWindow(fontJetBrains, OkIcon, "Logs copied", "OK") {
				showModal = false
			}
		}

		rl.EndDrawing()
	}

	return rl.WindowShouldClose()
}

// processStateUpdates reads from the StateUpdate channel without blocking.
func processStateUpdates(ch <-chan *StateUpdate, state *ProcessWindowInfo) {
	select {
	case update := <-ch:
		applyStateUpdate(state, update)
	default:
		// No updates pending
	}
}

func initializeWindow() {
	runtime.LockOSThread() // FIXES GRAPHICS CLIENT RENDERING PROBLEMS
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(windowWidth, windowHeight, "Process Window - H.D.S")
	rl.SetTargetFPS(60)
}

func loadUIFont() rl.Font {
	// If fontPath is invalid, Raylib will fall back to the default font.
	return rl.LoadFont(fontPath)
}

// handleResize checks mouse input and adjusts logBoxHeight accordingly.
func handleResize(
	logBoxHeight *float32,
	resizingLogBox *bool,
	minHeight float32,
	maxHeight float32,
) {
	// The Y position of the top boundary of the log box
	topOfLogBox := float32(windowHeight) - *logBoxHeight

	// A small "handle" rectangle to detect mouse hovering for resizing
	// Let's say 6px high, centered around 'topOfLogBox'
	var handleHeight float32 = 6
	resizeHandleRect := rl.NewRectangle(
		0,
		topOfLogBox-handleHeight/2,
		windowWidth,
		handleHeight,
	)

	mousePos := rl.GetMousePosition()
	mouseInHandle := rl.CheckCollisionPointRec(mousePos, resizeHandleRect)

	// If the user is hovering over the handle (and not yet dragging), change cursor
	if mouseInHandle && !*resizingLogBox {
		rl.SetMouseCursor(rl.MouseCursorResizeNS)
	} else if !*resizingLogBox {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	// Start resizing if the user left-clicks on the handle
	if mouseInHandle && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		*resizingLogBox = true
	}

	// If resizing, update the height as the mouse moves
	if *resizingLogBox {
		rl.SetMouseCursor(rl.MouseCursorResizeNS) // Force the NS-resize cursor
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			newTop := mousePos.Y
			if newTop < topLabelArea {
				newTop = topLabelArea // Don’t overlap the top labels
			}
			newHeight := float32(windowHeight) - newTop

			// Clamp the new height
			if newHeight < minHeight {
				newHeight = minHeight
			} else if newHeight > maxHeight {
				newHeight = maxHeight
			}
			*logBoxHeight = newHeight
		} else {
			// If the mouse was released, stop resizing
			*resizingLogBox = false
		}
	}
}

// handleScrolling changes logOffsetY based on mouse wheel movement.
// Measure the total text height to know how far can scroll.
func handleScrolling(
	logOffsetY *float32,
	logOffsetX *float32,
	logBoxHeight float32,
	font rl.Font,
	state *ProcessWindowInfo,
) {
	logBoxY := float32(windowHeight) - logBoxHeight
	logBoxRect := rl.NewRectangle(0, logBoxY, windowWidth, logBoxHeight)

	mousePos := rl.GetMousePosition()
	if !rl.CheckCollisionPointRec(mousePos, logBoxRect) {
		return // Only scroll if the mouse is inside the log box
	}

	const scrollSpeed float32 = 20
	mouseWheelMove := rl.GetMouseWheelMove()

	if mouseWheelMove != 0 {
		if rl.IsKeyDown(rl.KeyLeftShift) {
			handleHorizontalScrolling(logOffsetX, scrollSpeed, mouseWheelMove, font, state)
		} else {
			handleVerticalScrolling(logOffsetY, scrollSpeed, mouseWheelMove, logBoxHeight, font, state)
		}
	}
}

func handleVerticalScrolling(
	logOffsetY *float32,
	scrollSpeed float32,
	mouseWheelMove float32,
	logBoxHeight float32,
	font rl.Font,
	state *ProcessWindowInfo,
) {
	state.logsMu.Lock()
	lines := strings.Split(state.renderedLogs, "\n")
	var totalTextHeight float32
	for _, line := range lines {
		size := rl.MeasureTextEx(font, line, logFontSize, 1)
		totalTextHeight += size.Y
	}
	state.logsMu.Unlock()

	if totalTextHeight < (logBoxHeight - 20) {
		return // No vertical scrolling needed
	}

	*logOffsetY += mouseWheelMove * scrollSpeed

	const margin = 15
	minOffset := float32(windowHeight) - logBoxHeight - totalTextHeight + margin
	var maxOffset float32 = 0.0

	if *logOffsetY > maxOffset {
		*logOffsetY = maxOffset
	}
	if *logOffsetY < minOffset {
		*logOffsetY = minOffset
	}
}

func handleHorizontalScrolling(
	logOffsetX *float32,
	scrollSpeed float32,
	mouseWheelMove float32,
	font rl.Font,
	state *ProcessWindowInfo,
) {
	state.logsMu.Lock()
	var maxLineWidth float32
	lines := strings.Split(state.renderedLogs, "\n")
	for _, line := range lines {
		size := rl.MeasureTextEx(font, line, logFontSize, 1)
		if size.X > maxLineWidth {
			maxLineWidth = size.X
		}
	}
	state.logsMu.Unlock()

	if maxLineWidth < windowWidth {
		return // No horizontal scrolling needed
	}

	*logOffsetX -= mouseWheelMove * scrollSpeed // Invert for natural scrolling

	const margin = 15
	minOffsetX := float32(windowWidth) - maxLineWidth - margin
	var maxOffsetX float32 = 0.0

	if *logOffsetX > maxOffsetX {
		*logOffsetX = maxOffsetX
	}
	if *logOffsetX < minOffsetX {
		*logOffsetX = minOffsetX
	}
}

// drawGUI draws the entire GUI every frame.
func drawGUI(state *ProcessWindowInfo, font rl.Font, logOffsetX, logOffsetY, logHeight float32) {
	state.logsMu.Lock()
	defer state.logsMu.Unlock()

	// Update renderedLogs with newLogs
	if state.newLogs != "" {
		state.renderedLogs += state.newLogs
		state.newLogs = ""
	}

	// 1) Basic labels at the top
	labelStartX := float32(20)
	valueStartX := labelStartX + labelSpacing

	rl.DrawTextEx(font, "Status", rl.NewVector2(labelStartX, 20), labelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.grpcConnected, rl.NewVector2(valueStartX, 20), labelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Task Status", rl.NewVector2(labelStartX, 60), labelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.statusLabel, rl.NewVector2(valueStartX, 60), labelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "PCAP File", rl.NewVector2(labelStartX, 100), labelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.pcapFile, rl.NewVector2(valueStartX, 100), labelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat File", rl.NewVector2(labelStartX, 140), labelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatFile, rl.NewVector2(valueStartX, 140), labelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat Status", rl.NewVector2(labelStartX, 180), labelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatStatus, rl.NewVector2(valueStartX, 180), labelFontSize, 1, rl.Black)

	// 2) Draw the log box rectangle
	logBoxY := float32(windowHeight) - logHeight
	logBox := rl.NewRectangle(0, logBoxY, windowWidth, logHeight)
	rl.DrawRectangleRec(logBox, rl.Black)

	rl.BeginScissorMode(
		int32(logBox.X),
		int32(logBox.Y),
		int32(logBox.Width),
		int32(logBox.Height),
	)

	// Now draw the logs inside the clipped region
	rl.DrawTextEx(
		font,
		state.renderedLogs,
		rl.NewVector2(logBox.X+10+logOffsetX, logBoxY+10+logOffsetY),
		logFontSize,
		1,
		rl.White,
	)

	// End scissor mode
	rl.EndScissorMode()
}
