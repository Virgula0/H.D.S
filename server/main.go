package main

import (
	"fmt"
	"sync"

	"github.com/Virgula0/progetto-dp/server/backend/cmd"
)

func main() {
	// Create a WaitGroup
	var wg sync.WaitGroup

	// Add one to the WaitGroup for the goroutine
	wg.Add(1)

	// Start the RunBackend function in a goroutine
	go func() {
		defer wg.Done() // Mark this goroutine as done when it completes
		cmd.RunBackend()
	}()

	// Wait for backend to finish, it could have crashed so we can kill FE too eventually
	fmt.Println("Starting backend routine...")
	wg.Wait()
	fmt.Println("All tasks completed")
}
