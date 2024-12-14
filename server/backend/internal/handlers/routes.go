package handlers

import (
	"github.com/gorilla/mux"

	"github.com/Virgula0/progetto-dp/server/backend/internal/handlers/authenticate"
	"github.com/Virgula0/progetto-dp/server/backend/internal/handlers/logout"
	"github.com/Virgula0/progetto-dp/server/backend/internal/handlers/middlewares"
	"github.com/Virgula0/progetto-dp/server/backend/internal/handlers/register"
)

const RouteIndex = "/v1"
const GroupAdmin = "/admin"
const RouteAuthenticate = "/auth"
const RouteTokenVerifier = "/verify"
const RouteRegister = "/register"
const RouteLogout = "/logout"

func (h ServiceHandler) InitRoutes(router *mux.Router) {

	authenticateHandler := authenticate.Handler{Usecase: h.Usecase}
	registerHandler := register.Handler{Usecase: h.Usecase}
	authMiddleware := middlewares.TokenAuth{Usecase: h.Usecase}
	logoutHandler := logout.Handler{Usecase: h.Usecase}

	// Global middleware for loggin requests
	router.Use(middlewares.LogginMiddlware)

	// LOGIN
	loginRouter := router.PathPrefix(RouteIndex).Subrouter()
	loginRouter.
		HandleFunc(RouteAuthenticate, authenticateHandler.LoginHandler).
		Methods("POST")

	// VERIFY
	verifyRouter := router.PathPrefix(RouteIndex).Subrouter()
	verifyRouter.HandleFunc(RouteTokenVerifier, authenticateHandler.ChekTokenValidity).Methods("GET")
	verifyRouter.Use(authMiddleware.EnsureTokenIsValid)

	// SIGN-UP
	registerRouter := router.PathPrefix(RouteIndex).Subrouter()
	registerRouter.
		HandleFunc(RouteRegister, registerHandler.RegisterHandler).Methods("POST")

	// LOGOUT
	logoutRouter := router.PathPrefix(RouteIndex).Subrouter()
	logoutRouter.HandleFunc(RouteLogout, logoutHandler.LogoutUser).Methods("GET")
	logoutRouter.Use(authMiddleware.EnsureTokenIsValid)

	/*
	   // ADMIN
	   adminRouter := router.PathPrefix(RouteIndex).PathPrefix(GroupAdmin).Subrouter()
	   adminRouter.HandleFunc(UpdatePostLatestRevisionTimestampByID, authMiddleware.IsAdminMiddleware(postHandler.UpdatePostLatestRevisionByID)).Methods("GET")
	*/
}
