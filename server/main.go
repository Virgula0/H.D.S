package main

import (
	be "github.com/Virgula0/progetto-dp/server/backend/cmd"
	fe "github.com/Virgula0/progetto-dp/server/frontend/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Wait for backend to finish, it could have crashed so we can kill FE too eventually
	log.Println("Starting BE routine...")
	be.RunBackend()

	log.Println("Starting FE routine...")
	fe.RunFrontEnd()
	log.Println("All tasks running...")
}
