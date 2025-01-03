package gui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Virgula0/progetto-dp/client/internal/constants"
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	rl "github.com/gen2brain/raylib-go/raylib"
	log "github.com/sirupsen/logrus"
)

const (
	WindowWidth      = 850
	WindowHeight     = 550
	DefaultLogHeight = 350

	LabelSpacing  = 250
	LabelFontSize = 20
	LogFontSize   = 18

	FontPath = "/home/angelo/Universita/DP/H.D.S/client/internal/resources/fonts/Roboto-Black.ttf"
)

// ------------------------------------------------------------------------------------
// STATE MANAGEMENT
// ------------------------------------------------------------------------------------

// StateUpdate describes how to update the GUI state in one shot.
type StateUpdate struct {
	StatusLabel   string
	IsConnected   bool
	PCAPFile      string
	HashcatFile   string
	HashcatStatus string
	LogContent    string // New log content to append
}

// ProcessWindowInfo stores the current GUI state.
type ProcessWindowInfo struct {
	isConnected   bool
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

	state.isConnected = update.IsConnected
	state.statusLabel = update.StatusLabel
	state.pcapFile = update.PCAPFile
	state.hashcatFile = update.HashcatFile
	state.hashcatStatus = update.HashcatStatus

	// Append only the new log content
	state.newLogs += update.LogContent
}

// newDefaultGUIState initializes the default GUI state.
func newDefaultGUIState() *ProcessWindowInfo {
	return &ProcessWindowInfo{
		isConnected:      false,
		statusLabel:      "Initializing...",
		pcapFile:         "N/A",
		hashcatFile:      "N/A",
		hashcatStatus:    "Idle",
		renderedLogs:     "",
		newLogs:          "",
		lastReadPosition: 0,
	}
}

// ------------------------------------------------------------------------------------
// LOGGER GOROUTINE
// ------------------------------------------------------------------------------------

// GuiLogger reads logs incrementally and sends updates to the GUI.
func GuiLogger(otherInfos map[string]string, ctx context.Context, stateUpdateCh chan<- *StateUpdate) {
	var lastReadPosition int

	for {
		if ctx.Err() != nil {
			log.Error("Log routine terminated")
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
				StatusLabel:   otherInfos[constants.HashcatStatus],
				IsConnected:   true,
				PCAPFile:      otherInfos[constants.PCAPFile],
				HashcatFile:   otherInfos[constants.HashcatFile],
				HashcatStatus: constants.CrackStatus,
				LogContent:    newContent,
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// ------------------------------------------------------------------------------------
// MAIN GUI LOOP
// ------------------------------------------------------------------------------------

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
	)

	// Main application loop: re-draw everything every frame
	for !rl.WindowShouldClose() {
		// Process new state updates
		processStateUpdates(stateUpdateCh, guiState)

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

// ------------------------------------------------------------------------------------
// RAYLIB UI LOGIC
// ------------------------------------------------------------------------------------

func initializeWindow() {
	rl.InitWindow(WindowWidth, WindowHeight, "Process Window - Incremental Logs")
	rl.SetTargetFPS(120)
}

func loadUIFont() rl.Font {
	// If FontPath is invalid, Raylib will fall back to the default font.
	return rl.LoadFont(FontPath)
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
	rl.DrawTextEx(font, fmt.Sprintf("%v", state.isConnected), rl.NewVector2(valueStartX, 20), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Task Status", rl.NewVector2(labelStartX, 60), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.statusLabel, rl.NewVector2(valueStartX, 60), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "PCAP File", rl.NewVector2(labelStartX, 100), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.pcapFile, rl.NewVector2(valueStartX, 100), LabelFontSize, 1, rl.Black)

	rl.DrawTextEx(font, "Hashcat File", rl.NewVector2(labelStartX, 140), LabelFontSize, 1, rl.DarkGray)
	rl.DrawTextEx(font, state.hashcatFile, rl.NewVector2(valueStartX, 140), LabelFontSize, 1, rl.Black)

	// 2) Draw Logs
	logBoxY := float32(WindowHeight) - logHeight
	logBox := rl.NewRectangle(0, logBoxY, WindowWidth, logHeight)
	rl.DrawRectangleRec(logBox, rl.Black)

	// Render logs
	rl.DrawTextEx(font, state.renderedLogs, rl.NewVector2(10, logBoxY+10+logOffsetY), LogFontSize, 1, rl.White)
}
