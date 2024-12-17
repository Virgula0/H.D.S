package constants

import (
	"os"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
)

var JsonContentType = "application/json"

var JwtSecretKey = []byte(utils.GenerateToken(128))

const TokenConstant = "token"
const Limit = 5

// const for DateTime Format YYYY-MM-DD HH:MM:SS

const DateTimeExample = "2006-01-02 15:04:05"

// Database variables + Config variables
var (
	DBUser     = os.Getenv("DB_USER")
	DBPassword = os.Getenv("DB_PASSWORD")
	DBPort     = os.Getenv("DB_PORT")
	DBHost     = os.Getenv("DB_HOST")
	DBName     = os.Getenv("DB_NAME")

	AllowRegistrations = os.Getenv("ALLOW_REGISTRATIONS")
	DebugEnabled       = os.Getenv("DEBUG")
	WipeTables         = os.Getenv("RESET")

	// GRPC
	GRPC_URL                               = os.Getenv("GRPC_URL")
	GRPC_TIMEOUT, GRPC_TIMEOUT_PARSE_ERROR = time.ParseDuration(os.Getenv("GRPC_TIMEOUT"))
)

var HashCost = 12

// ROLES Constants
type Role string

const RoleString = "role"

// Declare constants of type Role for each role
const (
	ADMIN Role = "ADMIN"
	USER  Role = "USER"
)
