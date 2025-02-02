package grpcserver

import (
	"github.com/Virgula0/progetto-dp/server/backend/internal/grpcserver/encryption"
	"time"
)

type Option struct {
	Debug           bool
	GrpcURL         string
	GrpcConnTimeout time.Duration

	CACert, CAKey, ServerCert, ServerKey []byte
	ClientConfigStorage                  *encryption.ClientConfigStore
}
