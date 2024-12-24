package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Window Constants
const (
	WindowWidth      = 1280 // Increased width
	WindowHeight     = 900  // Increased height
	UpdateInterval   = 500 * time.Millisecond
	DefaultLogHeight = 400 // Increased default log window height
	LogScrollSpeed   = 20
	MinLogHeight     = 200                                                                                 // Increased minimum log height
	LabelSpacing     = 250                                                                                 // Increased spacing between labels and values
	LabelFontSize    = 20                                                                                  // Adjusted label font size
	LogFontSize      = 18                                                                                  // Improved log font size for readability
	LogLineSpacing   = 26                                                                                  // Adjusted spacing between log lines
	LogScrollSpeedX  = 20                                                                                  // Horizontal scroll speed
	FontPath         = "/home/angelo/Universita/DP/H.D.S/client/internal/resources/fonts/Roboto-Black.ttf" // Custom font path
)

// ProcessWindowInfo holds the state for the GUI.
type ProcessWindowInfo struct {
	sync.Mutex // Ensures thread-safe updates

	IsConnected   bool
	StatusLabel   string
	PCAPFile      string
	HashcatFile   string
	HashcatStatus string
	Logs          string
}

// InitProcessWindow creates and manages the GUI for process information.
func (info *ProcessWindowInfo) InitProcessWindow() bool {
	// Initialize Raylib window
	rl.InitWindow(WindowWidth, WindowHeight, "Process Window")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Load custom font
	font := rl.LoadFont(FontPath)
	defer rl.UnloadFont(font)

	// State for log box scrolling and resizing
	logOffsetY := float32(0)
	logOffsetX := float32(0)
	logHeight := float32(DefaultLogHeight)
	isResizing := false

	// Last label's Y position (ensures resizing respects it)
	lastLabelY := float32(240) + 40 // Adjusted for the larger window

	// Main loop
	lastUpdate := time.Now()
	for !rl.WindowShouldClose() {
		currentTime := time.Now()
		if currentTime.Sub(lastUpdate) >= UpdateInterval {
			info.Lock()
			lastUpdate = currentTime
			info.Unlock()
		}

		// Handle log scrolling only when the mouse is hovering over the log box
		mousePos := rl.GetMousePosition()
		logBoxY := WindowHeight - logHeight
		isHoveringLogBox := mousePos.Y > logBoxY && mousePos.Y < WindowHeight

		if isHoveringLogBox {
			if rl.IsKeyDown(rl.KeyUp) {
				logOffsetY += LogScrollSpeed
			}
			if rl.IsKeyDown(rl.KeyDown) {
				logOffsetY -= LogScrollSpeed
			}
			if rl.IsKeyDown(rl.KeyRight) {
				logOffsetX -= LogScrollSpeedX
			}
			if rl.IsKeyDown(rl.KeyLeft) {
				logOffsetX += LogScrollSpeedX
			}
		}

		// Clamp vertical scrolling
		maxOffsetY := float32(len(strings.Split(info.Logs, "\n"))*LogLineSpacing) - logHeight
		if logOffsetY > 0 {
			logOffsetY = 0
		}
		if logOffsetY < -maxOffsetY {
			logOffsetY = -maxOffsetY
		}

		// Clamp horizontal scrolling
		maxOffsetX := float32(2000) - float32(WindowWidth) // Arbitrary wide content assumption
		if logOffsetX > 0 {
			logOffsetX = 0
		}
		if logOffsetX < -maxOffsetX {
			logOffsetX = -maxOffsetX
		}

		// Handle resizing
		resizeBarY := WindowHeight - logHeight - 5
		isHoveringResizeBar := mousePos.Y > resizeBarY && mousePos.Y < resizeBarY+10

		// Change cursor when hovering over resizing bar
		if isHoveringResizeBar {
			rl.SetMouseCursor(rl.MouseCursorResizeNS)
		} else {
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}

		if rl.IsMouseButtonDown(rl.MouseLeftButton) && isHoveringResizeBar {
			isResizing = true
		}
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			isResizing = false
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}

		if isResizing {
			newHeight := float32(WindowHeight) - mousePos.Y
			maxLogHeight := WindowHeight - lastLabelY
			if newHeight >= MinLogHeight && newHeight <= maxLogHeight {
				logHeight = newHeight
			} else if newHeight < MinLogHeight {
				logHeight = MinLogHeight
			} else if newHeight > maxLogHeight {
				logHeight = maxLogHeight
			}
		}

		// Drawing
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Consistent label-value spacing
		labelStartX := int32(20)
		valueStartX := labelStartX + LabelSpacing

		// Status Label and Dot (Dot comes after status text)
		statusText := fmt.Sprintf("%v", map[bool]string{true: "Connected", false: "Not Connected"}[info.IsConnected])
		rl.DrawTextEx(font, "Status", rl.NewVector2(float32(labelStartX), 20), LabelFontSize, 1, rl.DarkGray)
		rl.DrawTextEx(font, statusText, rl.NewVector2(float32(valueStartX), 20), LabelFontSize, 1, rl.Black)
		rl.DrawCircle(valueStartX+int32(rl.MeasureTextEx(font, statusText, LabelFontSize, 1).X)+15, 30, 10, map[bool]rl.Color{true: rl.Green, false: rl.Yellow}[info.IsConnected])

		// Task Status
		rl.DrawTextEx(font, "Task Status", rl.NewVector2(float32(labelStartX), 60), LabelFontSize, 1, rl.DarkGray)
		rl.DrawTextEx(font, info.StatusLabel, rl.NewVector2(float32(valueStartX), 60), LabelFontSize, 1, rl.Black)

		// PCAP File
		rl.DrawTextEx(font, "PCAP File", rl.NewVector2(float32(labelStartX), 100), LabelFontSize, 1, rl.DarkGray)
		rl.DrawTextEx(font, info.PCAPFile, rl.NewVector2(float32(valueStartX), 100), LabelFontSize, 1, rl.Black)

		// Hashcat file
		rl.DrawTextEx(font, "Hashcat File", rl.NewVector2(float32(labelStartX), 140), LabelFontSize, 1, rl.DarkGray)
		rl.DrawTextEx(font, info.HashcatFile, rl.NewVector2(float32(valueStartX), 140), LabelFontSize, 1, rl.Black)

		// Hashcat Status
		rl.DrawTextEx(font, "Hashcat Status", rl.NewVector2(float32(labelStartX), 180), LabelFontSize, 1, rl.DarkGray)
		rl.DrawTextEx(font, info.HashcatStatus, rl.NewVector2(float32(valueStartX), 180), LabelFontSize, 1, rl.Black)

		// Log Box
		logBox := rl.NewRectangle(0, logBoxY, WindowWidth, logHeight)
		rl.DrawRectangleRec(logBox, rl.Black)
		rl.DrawRectangleLinesEx(logBox, 1, rl.Gray)
		rl.DrawRectangle(0, int32(logBoxY)-5, WindowWidth, 5, rl.Gray)

		// Display logs
		logs := info.Logs

		y := int(logBoxY) + 10 + int(logOffsetY)
		for _, logLine := range strings.Split(logs, "\n") {
			rl.DrawTextEx(font, logLine, rl.NewVector2(10+logOffsetX, float32(y)), LogFontSize, 1, rl.White)
			y += LogLineSpacing
		}

		rl.EndDrawing()
	}

	return rl.WindowShouldClose()
}
