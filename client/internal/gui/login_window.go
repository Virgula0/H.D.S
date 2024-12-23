package gui

import (
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// GUI Constants
const (
	screenWidth  = 400
	screenHeight = 300
)

// InitLoginWindow initializes and displays the login window.
func InitLoginWindow(client *grpcclient.Client) bool {
	// Initialize Raylib
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(screenWidth, screenHeight, "Login Window")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Input State
	username := ""
	password := ""
	errorMessage := ""
	usernameFocused := true
	passwordFocused := false

	// Rectangle positions
	usernameRect := rl.NewRectangle(100, 90, 200, 30)
	passwordRect := rl.NewRectangle(100, 140, 200, 30)
	loginButtonRect := rl.NewRectangle(150, 200, 100, 40)

	for !rl.WindowShouldClose() {
		// Input Handling
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePoint := rl.GetMousePosition()
			usernameFocused = rl.CheckCollisionPointRec(mousePoint, usernameRect)
			passwordFocused = rl.CheckCollisionPointRec(mousePoint, passwordRect)
		}

		// Text Input
		if usernameFocused {
			username = HandleTextInput(username)
		}
		if passwordFocused {
			password = HandleTextInput(password)
		}

		// Check for Login Button Press
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), loginButtonRect) {
			if resp, err := client.Authenticate(username, password); err == nil {
				client.Credentials.Auth.Username = username
				client.Credentials.Auth.Password = password
				*client.Credentials.JWT = resp.Details
				return false
			} else {
				errorMessage = "Invalid username or password"
			}
		}

		// Drawing
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Title
		rl.DrawText("Login", 170, 30, 20, rl.DarkGray)

		// Username Input Box
		rl.DrawRectangleRec(usernameRect, rl.LightGray)
		rl.DrawText("Username:", 100, 70, 10, rl.DarkGray)
		rl.DrawText(username, int32(usernameRect.X+5), int32(usernameRect.Y+5), 20, rl.Black)

		// Password Input Box
		rl.DrawRectangleRec(passwordRect, rl.LightGray)
		rl.DrawText("Password:", 100, 120, 10, rl.DarkGray)
		rl.DrawText(strings.Repeat("*", len(password)), int32(passwordRect.X+5), int32(passwordRect.Y+5), 20, rl.Black)

		// Login Button
		rl.DrawRectangleRec(loginButtonRect, rl.SkyBlue)
		rl.DrawText("Login", int32(loginButtonRect.X+20), int32(loginButtonRect.Y+10), 20, rl.Black)

		// Error Message
		if errorMessage != "" {
			rl.DrawText(errorMessage, 100, 250, 15, rl.Red)
		}

		rl.EndDrawing()
	}

	return rl.WindowShouldClose()
}

// HandleTextInput captures text input for username and password fields.
func HandleTextInput(text string) string {
	key := rl.GetCharPressed()
	for key > 0 {
		if key >= 32 && key <= 125 {
			text += string(rune(key))
		}
		key = rl.GetCharPressed()
	}

	if rl.IsKeyPressed(rl.KeyBackspace) && len(text) > 0 {
		text = text[:len(text)-1]
	}
	return text
}
