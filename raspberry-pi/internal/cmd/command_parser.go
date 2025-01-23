package cmd

import (
	"fmt"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/constants"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/entities"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var (
	username string
	password string
)

func runTUI(username, password *string) error {
	model := tui.LoginModel()

	if _, err := tea.NewProgram(model).Run(); err != nil {
		return fmt.Errorf("tui: could not start program -> %v", err)
	}

	*username, *password = model.GetCredentials()
	return nil
}

// AuthCommand returns the parsed AuthRequest structure.
func AuthCommand() (*entities.AuthRequest, error) {

	rootCmd := &cobra.Command{
		Use:   "",
		Short: "Daemon CLI management",
	}

	unsecureLogin := &cobra.Command{
		Use:   "insecure-login",
		Short: "Authenticate with username and password via args",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure both flags are provided

			if !constants.Test {
				return fmt.Errorf("test environment variable is not set")
			}

			if username == "" || password == "" {
				return fmt.Errorf("both --username and --password flags are required")
			}

			// Return AuthRequest struct
			return nil
		},
	}

	tt := &cobra.Command{
		Use:   "run",
		Short: "Authenticate with tui",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(&username, &password)
		},
	}

	unsecureLogin.Flags().StringVarP(&username, "username", "u", "", "Username for authentication (required)")
	unsecureLogin.Flags().StringVarP(&password, "password", "p", "", "Password for authentication (required)")

	err := unsecureLogin.MarkFlagRequired("username")
	if err != nil {
		return nil, err
	}

	err = unsecureLogin.MarkFlagRequired("password")
	if err != nil {
		return nil, err
	}

	rootCmd.AddCommand(unsecureLogin, tt)

	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("invalid empty credentials")
	}

	return &entities.AuthRequest{
		Username: username,
		Password: password,
	}, nil

}
