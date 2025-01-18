package gui

import (
	"github.com/gen2brain/raylib-go/raylib"
)

const (
	messageWidth  = 400
	messageHeight = 150
)

const (
	AlertIcon    rune = '⚠' // Warning symbol
	OkIcon       rune = '✓' // Checkmark symbol
	InfoIcon     rune = 'i' // Information symbol
	QuestionIcon rune = '?' // Question mark symbol
	ErrorIcon    rune = 'X' // Error symbol
)

// DrawMessageWindow displays a modal window with an icon, a title, and a button to dismiss it.
func DrawMessageWindow(font rl.Font, icon rune, title, buttonText string) bool {
	// Define modal dimensions and positions
	modalWidth := float32(messageWidth)
	modalHeight := float32(messageHeight)
	modalX := (float32(rl.GetScreenWidth()) - modalWidth) / 2
	modalY := (float32(rl.GetScreenHeight())-modalHeight)/2 - 100
	modalRect := rl.NewRectangle(modalX, modalY, modalWidth, modalHeight)

	// Define button dimensions and positions
	buttonWidth := float32(100)
	buttonHeight := float32(30)
	buttonX := modalX + (modalWidth-buttonWidth)/2
	buttonY := modalY + modalHeight - buttonHeight - 20
	buttonRect := rl.NewRectangle(buttonX, buttonY, buttonWidth, buttonHeight)

	// Draw modal background
	rl.DrawRectangleRec(modalRect, rl.Black)
	rl.DrawRectangleLinesEx(modalRect, 2, rl.LightGray)

	// Draw icon
	iconSize := float32(32)
	iconX := modalX + 20
	iconY := modalY + 20
	rl.DrawTextEx(font, string(icon), rl.Vector2{X: iconX, Y: iconY}, iconSize, 1, rl.White)

	// Draw title
	textX := iconX + iconSize + 10
	textY := modalY + 20
	rl.DrawText(title, int32(textX), int32(textY), 20, rl.White)

	// Draw button
	rl.DrawRectangleRec(buttonRect, rl.SkyBlue)
	rl.DrawRectangleLinesEx(buttonRect, 2, rl.DarkBlue)
	buttonTextX := buttonX + (buttonWidth-float32(len(buttonText))*10)/2
	buttonTextY := buttonY + (buttonHeight-20)/2
	rl.DrawText(buttonText, int32(buttonTextX), int32(buttonTextY), 20, rl.White)

	// Check button click
	mousePos := rl.GetMousePosition()
	if rl.CheckCollisionPointRec(mousePos, buttonRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Button clicked, close the modal
		return true
	}

	return false
}
