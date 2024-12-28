package pages

import (
	"github.com/Virgula0/progetto-dp/server/frontend/internal/middlewares"
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

func (h ServiceHandler) InitRoutes(router *mux.Router) {
	loginInstance := login.Page{Usecase: h.Usecase}
	registerInstance := register.Page{Usecase: h.Usecase}
	logoutInstance := logout.Page{Usecase: h.Usecase}
	handshakeInstance := handshakes.Page{Usecase: h.Usecase}
	authenticated := middlewares.TokenAuth{Usecase: h.Usecase}

	router.Use(middlewares.LogginMiddlware)

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
		HandleFunc(Handshake, handshakeInstance.ListHandshakes).
		Methods("GET")
	handshakeRouterTemplate.Use(authenticated.TokenValidation)

}