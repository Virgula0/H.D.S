package constants

import (
	"os"
	"strings"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/utils"
)

const UserIDKey = "userID"

var JSONContentType = "application/json"
var JwtSecretKey = []byte(utils.GenerateToken(128))

type MyTokenKey string

var TokenConstant = MyTokenKey("token")

const Limit = 5

// const for DateTime Format YYYY-MM-DD HH:MM:SS

const DateTimeExample = "2006-01-02 15:04:05"

// Database variables + Config variables
var (
	ServerHost = os.Getenv("BACKEND_HOST")
	ServerPort = os.Getenv("BACKEND_PORT")

	DBUser     = os.Getenv("DB_USER")
	DBPassword = os.Getenv("DB_PASSWORD")
	DBPort     = os.Getenv("DB_PORT")
	DBHost     = os.Getenv("DB_HOST")
	DBName     = os.Getenv("DB_NAME")

	AllowRegistrations = os.Getenv("ALLOW_REGISTRATIONS")
	DebugEnabled       = strings.ToLower(os.Getenv("DEBUG")) == "true"
	WipeTables         = os.Getenv("RESET")

	// GRPC
	GrpcURL                            = os.Getenv("GRPC_URL")
	GrpcTimeout, GrpcTimeoutParseError = time.ParseDuration(os.Getenv("GRPC_TIMEOUT"))

	// TCP
	TCPAddress = os.Getenv("TCP_ADDRESS")
	TCPPort    = os.Getenv("TCP_PORT")
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

// Statuses for handshake assignments
const (
	NothingStatus = "nothing"
	PendingStatus = "pending"
	WorkingStatus = "working"
)
