package gui

import (
	"github.com/Virgula0/progetto-dp/client/internal/grpcclient"
	log "github.com/sirupsen/logrus"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InitLoginWindow initializes and displays the login window.
func InitLoginWindow(client *grpcclient.Client) bool {
	initWindow()
	defer rl.CloseWindow()

	// State
	state := loginState{
		username:        "",
		password:        "",
		errorMessage:    "",
		usernameFocused: true,
		passwordFocused: false,
	}

	// UI Rectangles
	ui := loginUI{
		usernameRect:    rl.NewRectangle(100, 90, 200, 30),
		passwordRect:    rl.NewRectangle(100, 140, 200, 30),
		loginButtonRect: rl.NewRectangle(150, 200, 100, 40),
	}

	for !rl.WindowShouldClose() {
		handleInput(&state, &ui)
		if handleLogin(client, &state, &ui) {
			return false
		}
		renderLoginWindow(&state, &ui)
	}

	return rl.WindowShouldClose()
}

// ---------- STRUCTURES ----------

// loginState manages the state of the login window.
type loginState struct {
	username        string
	password        string
	errorMessage    string
	usernameFocused bool
	passwordFocused bool
}

// loginUI holds the UI component positions.
type loginUI struct {
	usernameRect    rl.Rectangle
	passwordRect    rl.Rectangle
	loginButtonRect rl.Rectangle
}

// ---------- INITIALIZATION ----------

func initWindow() {
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(400, 300, "Login Window")
	rl.SetTargetFPS(60)
}

// ---------- INPUT HANDLING ----------

func handleInput(state *loginState, ui *loginUI) {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		mousePoint := rl.GetMousePosition()
		state.usernameFocused = rl.CheckCollisionPointRec(mousePoint, ui.usernameRect)
		state.passwordFocused = rl.CheckCollisionPointRec(mousePoint, ui.passwordRect)
	}

	if state.usernameFocused {
		state.username = HandleTextInput(state.username)
	}
	if state.passwordFocused {
		state.password = HandleTextInput(state.password)
	}
}

// ---------- LOGIN LOGIC ----------

func handleLogin(client *grpcclient.Client, state *loginState, ui *loginUI) bool {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), ui.loginButtonRect) {
		resp, err := client.Authenticate(state.username, state.password)

		if err == nil {
			client.Credentials.Auth.Username = state.username
			client.Credentials.Auth.Password = state.password
			*client.Credentials.JWT = resp.GetDetails()
			return true
		} else {
			log.Errorf("error on Authenticate rpc call: %v", err)
		}

		state.errorMessage = "Invalid username or password"
	}
	return false
}

// ---------- RENDERING ----------

func renderLoginWindow(state *loginState, ui *loginUI) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	// Title
	rl.DrawText("Login", 170, 30, 20, rl.DarkGray)

	// Username Input Box
	drawInputBox("Username:", state.username, ui.usernameRect, 70)

	// Password Input Box
	drawInputBox("Password:", strings.Repeat("*", len(state.password)), ui.passwordRect, 120)

	// Login Button
	rl.DrawRectangleRec(ui.loginButtonRect, rl.SkyBlue)
	rl.DrawText("Login", int32(ui.loginButtonRect.X+20), int32(ui.loginButtonRect.Y+10), 20, rl.Black)

	// Error Message
	if state.errorMessage != "" {
		rl.DrawText(state.errorMessage, 100, 250, 15, rl.Red)
	}

	rl.EndDrawing()
}

func drawInputBox(label, value string, rect rl.Rectangle, labelY int32) {
	rl.DrawRectangleRec(rect, rl.LightGray)
	rl.DrawText(label, int32(rect.X), labelY, 10, rl.DarkGray)
	rl.DrawText(value, int32(rect.X+5), int32(rect.Y+5), 20, rl.Black)
}

// HandleTextInput captures text input for username and password fields.
func HandleTextInput(text string) string {
	key := rl.GetCharPressed()
	for key > 0 {
		if key >= 32 && key <= 125 {
			text += string(key)
		}
		key = rl.GetCharPressed()
	}

	if rl.IsKeyPressed(rl.KeyBackspace) && text != "" {
		text = text[:len(text)-1]
	}
	return text
}
