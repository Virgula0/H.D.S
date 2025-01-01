package pages

import (
	"github.com/Virgula0/progetto-dp/server/frontend/internal/middlewares"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/clients"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/welcome"
	"github.com/gorilla/mux"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/login"

	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/handshakes"
	logout "github.com/Virgula0/progetto-dp/server/frontend/internal/pages/logout"
	"github.com/Virgula0/progetto-dp/server/frontend/internal/pages/register"
)

const RouteIndex = constants.RouteIndex
const Login = constants.Login
const Register = constants.Register
const Logout = constants.Logout
const Handshake = constants.HandshakePage
const Clients = constants.ClientPage
const Devices = constants.RaspberryPIPage
const HandshakeSubmission = constants.SubmitTask
const DeleteRaspberryPI = constants.DeleteRaspberry
const DeleteClient = constants.DeleteClient
const DeleteHandshake = constants.DeleteHandshake

// InitRoutes
//
// Initializes routes for the FE
//
//nolint:funlen // this function does not have logic, init routes can have a huge length
func (h ServiceHandler) InitRoutes(router *mux.Router) {
	loginInstance := login.Page{Usecase: h.Usecase}
	registerInstance := register.Page{Usecase: h.Usecase}
	logoutInstance := logout.Page{Usecase: h.Usecase}
	handshakeInstance := handshakes.Page{Usecase: h.Usecase}
	clientsInstance := clients.Page{Usecase: h.Usecase}
	devicesInstance := raspberrypi.Page{Usecase: h.Usecase}
	welcomeInstance := welcome.Page{Usecase: h.Usecase}
	authenticated := middlewares.TokenAuth{Usecase: h.Usecase}

	router.Use(middlewares.LoggingMiddleware)

	// LOGIN
	loginRouter := router.PathPrefix(RouteIndex).Subrouter()
	loginRouter.
		HandleFunc(Login, loginInstance.PerformLogin).
		Methods("POST")

	loginRouterTemplate := router.PathPrefix(RouteIndex).Subrouter()
	loginRouterTemplate.
		HandleFunc(Login, loginInstance.LoginTemplate).
		Methods("GET")
	loginRouterTemplate.Use(authenticated.CheckCookieExistence)

	// LOGOUT
	logoutRouter := router.PathPrefix(RouteIndex).Subrouter()
	logoutRouter.
		HandleFunc(Logout, logoutInstance.Logout).
		Methods("GET")
	logoutRouter.Use(authenticated.TokenValidation)

	// REGISTER
	registerRouterTemplate := router.PathPrefix(RouteIndex).Subrouter()
	registerRouterTemplate.
		HandleFunc(Register, registerInstance.Register).
		Methods("GET")
	registerRouterTemplate.Use(authenticated.CheckCookieExistence)

	registerRouter := router.PathPrefix(RouteIndex).Subrouter()
	registerRouter.
		HandleFunc(Register, registerInstance.PerformRegistration).
		Methods("POST")

	// HANDSHAKES
	handshakeRouterTemplate := router.PathPrefix(RouteIndex).Subrouter()
	handshakeRouterTemplate.
		HandleFunc(Handshake, handshakeInstance.TemplateHandshake).
		Methods("GET")
	handshakeRouterTemplate.Use(authenticated.TokenValidation)

	handshakeRouter := router.PathPrefix(RouteIndex).Subrouter()
	handshakeRouter.
		HandleFunc(HandshakeSubmission, handshakeInstance.UpdateTask).
		Methods("POST")
	handshakeRouter.Use(authenticated.TokenValidation)

	handshakeRouter.
		HandleFunc(DeleteHandshake, handshakeInstance.DeleteHandshake).
		Methods("POST")
	handshakeRouter.Use(authenticated.TokenValidation)

	// Clients
	clientsRouterTemplate := router.PathPrefix(RouteIndex).Subrouter()
	clientsRouterTemplate.
		HandleFunc(Clients, clientsInstance.ListClients).
		Methods("GET")
	clientsRouterTemplate.Use(authenticated.TokenValidation)

	clientsRouterTemplate.
		HandleFunc(DeleteClient, clientsInstance.DeleteClient).
		Methods("POST")
	clientsRouterTemplate.Use(authenticated.TokenValidation)

	// RaspberryPI
	devicesRouterTemplate := router.PathPrefix(RouteIndex).Subrouter()
	devicesRouterTemplate.
		HandleFunc(Devices, devicesInstance.ListRaspberryPI).
		Methods("GET")
	devicesRouterTemplate.Use(authenticated.TokenValidation)

	devicesRouterTemplate.
		HandleFunc(DeleteRaspberryPI, devicesInstance.DeleteRaspberryPI).
		Methods("POST")
	devicesRouterTemplate.Use(authenticated.TokenValidation)

	// Welcome page
	welcomeTemplate := router
	welcomeTemplate.
		HandleFunc(RouteIndex, welcomeInstance.WelcomeTemplate).
		Methods("GET")
}
