package constants

import (
	"fmt"
	"os"
)

const TimeOut = 10

var BackendBaseURL = fmt.Sprintf("http://%s:%s/v1/", os.Getenv("BACKEND_HOST"), os.Getenv("BACKEND_PORT"))

const SessionTokenName = "session_token"

type CustomType string

var AuthToken = CustomType("token")

const JSONContentType = "application/json"
const HTMLContentType = "text/html;charset=UTF-8"
const FileToCrackString = "FILE_TO_CRACK"

const MaxUploadSize = 10 << 28 // 2,68435456 GB

// Views

const (
	LoginView     = "login.html"
	RegisterView  = "register.html"
	HandshakeView = "handshake.html"
	ClientView    = "clients.html"
	DeviceView    = "raspberrypi.html"
	WelcomeView   = "welcome.html"
)

// Endpoints FE
const (
	RouteIndex       = "/"
	Login            = "/login"
	HandshakePage    = "/handshakes"
	ClientPage       = "/clients"
	RaspberryPIPage  = "/raspberrypi"
	Register         = "/register"
	Logout           = "/logout"
	SubmitTask       = "/submit-task"
	DeleteClient     = "/delete-client"
	DeleteRaspberry  = "/delete-raspberrypi"
	DeleteHandshake  = "/delete-handshake"
	CreateHandshake  = "/create-handshake"
	UpdateEncryption = "/update-encryption"
	UpdatePassword   = "/update-password"
)

// Endpoints BE
const (
	BackendVerifyEndpoint    = "verify"
	BackendAuthEndpoint      = "auth"
	BackendLogoutEndpoint    = "logout"
	BackendRegisterEndpoint  = "register"
	BackendGetHandshakes     = "handshakes"
	BackendGetClients        = "clients"
	BackendGetRaspberryPi    = "devices"
	BackendUpdateClientTask  = "assign"
	BackendDeleteClient      = "delete/client"
	BackendHandshake         = "manage/handshake"
	BackendDeleteRaspberryPI = "delete/raspberrypi"
	UpdateClientEncryption   = "encryption-status"
	UpdateUserPassword       = "user/password"
)
