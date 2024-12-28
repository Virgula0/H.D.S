package restapi

import (
	"github.com/gorilla/mux"

	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/authenticate"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/client"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/handshake"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/logout"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/middlewares"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/raspberrypi"
	"github.com/Virgula0/progetto-dp/server/backend/internal/restapi/register"
)

const RouteIndex = "/v1"
const RouteAuthenticate = "/auth"
const RouteTokenVerifier = "/verify"
const RouteRegister = "/register"
const RouteLogout = "/logout"
const GetClients = "/clients"
const GetDevices = "/devices"
const GetHandshakes = "/handshakes"
const UpdateClientTask = "/assign"

func (h ServiceHandler) InitRoutes(router *mux.Router) {

	authenticateHandler := authenticate.Handler{Usecase: h.Usecase}
	registerHandler := register.Handler{Usecase: h.Usecase}
	authMiddleware := middlewares.TokenAuth{Usecase: h.Usecase}
	logoutHandler := logout.Handler{Usecase: h.Usecase}
	installedClientsHandler := client.Handler{Usecase: h.Usecase}
	installedDevicesHandler := raspberrypi.Handler{Usecase: h.Usecase}
	handshakesHandler := handshake.Handler{Usecase: h.Usecase}

	// Global middleware for loggin requests
	router.Use(middlewares.LogginMiddlware)

	// LOGIN -- NOT AUTHENTICATED --
	loginRouter := router.PathPrefix(RouteIndex).Subrouter()
	loginRouter.
		HandleFunc(RouteAuthenticate, authenticateHandler.LoginHandler).
		Methods("POST")

	// SIGN-UP -- NOT AUTHENTICATED --
	registerRouter := router.PathPrefix(RouteIndex).Subrouter()
	registerRouter.
		HandleFunc(RouteRegister, registerHandler.RegisterHandler).Methods("POST")

	// VERIFY -- AUTHENTICATED --
	verifyRouter := router.PathPrefix(RouteIndex).Subrouter()
	verifyRouter.HandleFunc(RouteTokenVerifier, authenticateHandler.CheckTokenValidity).Methods("GET")
	verifyRouter.Use(authMiddleware.EnsureTokenIsValid)

	// LOGOUT -- AUTHENTICATED --
	logoutRouter := router.PathPrefix(RouteIndex).Subrouter()
	logoutRouter.HandleFunc(RouteLogout, logoutHandler.LogoutUser).Methods("GET")
	logoutRouter.Use(authMiddleware.EnsureTokenIsValid)

	// Get clients installed by user -- AUTHENTICATED --
	installedClientsRouter := router.PathPrefix(RouteIndex).Subrouter()
	installedClientsRouter.HandleFunc(GetClients, installedClientsHandler.ReturnClientsInstalled).Methods("GET")
	installedClientsRouter.Use(authMiddleware.EnsureTokenIsValid)

	// Get raspberry-pi installed by user -- AUTHENTICATED --
	installedDevicesRouter := router.PathPrefix(RouteIndex).Subrouter()
	installedDevicesRouter.HandleFunc(GetDevices, installedDevicesHandler.GetRaspberryPIDevices).Methods("GET")
	installedDevicesRouter.Use(authMiddleware.EnsureTokenIsValid)

	// Get handshake by user -- AUTHENTICATED --
	handshakesRouter := router.PathPrefix(RouteIndex).Subrouter()
	handshakesRouter.HandleFunc(GetHandshakes, handshakesHandler.GetHandshakes).Methods("GET")
	handshakesRouter.Use(authMiddleware.EnsureTokenIsValid)

	handshakesRouter.HandleFunc(UpdateClientTask, handshakesHandler.UpdateClientTask).Methods("POST")
	handshakesRouter.Use(authMiddleware.EnsureTokenIsValid)
}
