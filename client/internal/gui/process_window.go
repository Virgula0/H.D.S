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
	WindowWidth      = 850
	WindowHeight     = 550
	DefaultLogHeight = 300

	LabelSpacing  = 250
	LabelFontSize = 20
	LogFontSize   = 18

	FontPath     = "internal/resources/fonts/Roboto-Black.ttf"
	TopLabelArea = 220
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

var StateUpdateCh = make(chan *StateUpdate, 10)

// RunGUI starts the Raylib window and listens for updates via the channel.
func RunGUI(stateUpdateCh <-chan *StateUpdate) bool {
	initializeWindow()
	defer rl.CloseWindow()

	uiFont := loadUIFont()
	defer rl.UnloadFont(uiFont)

	guiState := newDefaultGUIState()

	// Track scrolling & resizing of the log box
	var (
		logOffsetY   float32 = 0
		logBoxHeight float32 = DefaultLogHeight

		// Whether the user is dragging the log-box boundary
		resizingLogBox bool

		// That means logBoxHeight can be at most WindowHeight - TopLabelArea
		minLogBoxHeight float32 = 50
		maxLogBoxHeight float32 = WindowHeight - TopLabelArea
	)

	// Main application loop: re-draw everything every frame
	for !rl.WindowShouldClose() {
		// Process new state updates
		processStateUpdates(stateUpdateCh, guiState)

		// Handle resizing of the log box
		handleResize(&logBoxHeight, &resizingLogBox, minLogBoxHeight, maxLogBoxHeight)

		// Handle scrolling (mouse wheel)
		handleScrolling(&logOffsetY, logBoxHeight, uiFont, guiState)

		// Draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawGUI(guiState, uiFont, logOffsetY, logBoxHeight)

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
	rl.InitWindow(WindowWidth, WindowHeight, "Process Window - H.D.S")
	rl.SetTargetFPS(60)
}

func loadUIFont() rl.Font {
	// If FontPath is invalid, Raylib will fall back to the default font.
	return rl.LoadFont(FontPath)
}

// handleResize checks mouse input and adjusts logBoxHeight accordingly.
func handleResize(
	logBoxHeight *float32,
	resizingLogBox *bool,
	minHeight float32,
	maxHeight float32,
) {
	// The Y position of the top boundary of the log box
	topOfLogBox := float32(WindowHeight) - *logBoxHeight

	// A small "handle" rectangle to detect mouse hovering for resizing
	// Let's say 6px high, centered around 'topOfLogBox'
	var handleHeight float32 = 6
	resizeHandleRect := rl.NewRectangle(
		0,
		topOfLogBox-handleHeight/2,
		WindowWidth,
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
			if newTop < TopLabelArea {
				newTop = TopLabelArea // Don’t overlap the top labels
			}
			newHeight := float32(WindowHeight) - newTop

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
func handleScrolling(logOffsetY *float32, logBoxHeight float32, font rl.Font, state *ProcessWindowInfo) {
	logBoxY := float32(WindowHeight) - logBoxHeight
	logBoxRect := rl.NewRectangle(0, logBoxY, WindowWidth, logBoxHeight)

	mousePos := rl.GetMousePosition()
	if !rl.CheckCollisionPointRec(mousePos, logBoxRect) {
		return // Only scroll if mouse is inside the log box
	}

	// Scroll speed factor
	const scrollSpeed float32 = 20
	mouseWheelMove := rl.GetMouseWheelMove()
	if mouseWheelMove == 0 {
		return
	}

	// Measure total text height
	state.logsMu.Lock()
	lines := strings.Split(state.renderedLogs, "\n")
	var totalTextHeight float32
	for _, line := range lines {
		size := rl.MeasureTextEx(font, line, LogFontSize, 1)
		totalTextHeight += size.Y
	}
	state.logsMu.Unlock()

	// If total text is smaller than the box, no scrolling needed
	if totalTextHeight < (logBoxHeight - 20) {
		return
	}

	// Adjust offset
	*logOffsetY += mouseWheelMove * scrollSpeed

	const margin = 15
	minOffset := (logBoxHeight - margin) - totalTextHeight // negative
	maxOffset := float32(0)

	// Clamp
	if *logOffsetY > maxOffset {
		*logOffsetY = maxOffset
	}
	if *logOffsetY < minOffset {
		*logOffsetY = minOffset
	}
}

// drawGUI draws the entire GUI every frame.
func drawGUI(state *ProcessWindowInfo, font rl.Font, logOffsetY, logHeight float32) {
	state.logsMu.Lock()
	defer state.logsMu.Unlock()

	// Update renderedLogs with newLogs
	if state.newLogs != "" {
		state.renderedLogs += state.newLogs
		state.newLogs = ""
	}

	// 1) Basic labels at the top
	labelStartX := float32(20)
	valueStartX := labelStartX + LabelSpacing

	rl.DrawTextEx(font, "Status", rl.NewVector2(labelStartX, 20), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.grpcConnected, rl.NewVector2(valueStartX, 20), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Task Status", rl.NewVector2(labelStartX, 60), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.statusLabel, rl.NewVector2(valueStartX, 60), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "PCAP File", rl.NewVector2(labelStartX, 100), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.pcapFile, rl.NewVector2(valueStartX, 100), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat File", rl.NewVector2(labelStartX, 140), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatFile, rl.NewVector2(valueStartX, 140), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat Status", rl.NewVector2(labelStartX, 180), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatStatus, rl.NewVector2(valueStartX, 180), LabelFontSize, 1, rl.Black)

	// 2) Draw the log box rectangle
	logBoxY := float32(WindowHeight) - logHeight
	logBox := rl.NewRectangle(0, logBoxY, WindowWidth, logHeight)
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
		rl.NewVector2(logBox.X+10, logBoxY+10+logOffsetY),
		LogFontSize,
		1,
		rl.White,
	)

	// End scissor mode
	rl.EndScissorMode()
}
