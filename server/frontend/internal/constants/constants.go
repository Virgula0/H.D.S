package constants

import (
	"fmt"
	"os"
)

const TimeOut = 10

var BackendBaseURL = fmt.Sprintf("http://%s:%s/v1/", os.Getenv("BACKEND_HOST"), os.Getenv("BACKEND_PORT"))

const SessionTokenName = "session_token"
const AuthToken = "token"

const JSONContentType = "application/json"
const HTMLContentType = "text/html"

// Views

const (
	LoginView     = "login.html"
	RegisterView  = "register.html"
	HandshakeView = "handshake.html"
	ClientView    = "clients.html"
	DeviceView    = "raspberrypi.html"
)

// Endpoints FE
const (
	RouteIndex      = "/"
	Login           = "/login"
	HandshakePage   = "/handshakes"
	ClientPage      = "/clients"
	RaspberryPIPage = "/raspberrypi"
	Register        = "/register"
	Logout          = "/logout"
)

// Endpoints BE
const (
	BackendVerifyEndpoint   = "verify"
	BackendAuthEndpoint     = "auth"
	BackendLogoutEndpoint   = "logout"
	BackendRegisterEndpoint = "register"
	BackendGetHandshakes    = "handshakes"
	BackendGetClients       = "clients"
	BackendGetRaspberryPi   = "devices"
)
