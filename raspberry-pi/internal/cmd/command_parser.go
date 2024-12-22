package cmd

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var (
	username string
	password string
)

// AuthCommand returns the parsed AuthRequest structure.
func AuthCommand() (*entities.AuthRequest, error) {
	var authCmd = &cobra.Command{
		Use:   "run",
		Short: "Authenticate with username and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure both flags are provided
			if username == "" || password == "" {
				return fmt.Errorf("both --username and --password flags are required")
			}

			// Return AuthRequest struct
			return nil
		},
	}

	// Define flags
	authCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication (required)")
	authCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication (required)")

	err := authCmd.MarkFlagRequired("username")
	if err != nil {
		return nil, err
	}

	err = authCmd.MarkFlagRequired("password")
	if err != nil {
		return nil, err
	}

	// Execute the command
	if err := authCmd.Execute(); err != nil {
		return nil, err
	}

	return &entities.AuthRequest{
		Username: username,
		Password: password,
	}, nil
}
