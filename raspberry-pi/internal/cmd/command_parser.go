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

	cliKey, insecureKey, runKey = "cli", "insecure-login", "run"
)

var cobraCommands = map[string]*cobra.Command{
	cliKey: {
		Use:   cliKey,
		Short: "Daemon CLI management",
	},
	insecureKey: {
		Use:   insecureKey,
		Short: "Authenticate with username and password via args",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !constants.Test {
				return fmt.Errorf("test environment variable is not set")
			}
			if username == "" || password == "" {
				return fmt.Errorf("both --username and --password flags are required")
			}
			return nil
		},
	},
	runKey: {
		Use:   runKey,
		Short: "Authenticate with TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(&username, &password)
		},
	},
}

func runTUI(username, password *string) error {
	model := tui.LoginModel()

	if _, err := tea.NewProgram(model).Run(); err != nil {
		return fmt.Errorf("tui: could not start program -> %v", err)
	}

	*username, *password = model.GetCredentials()
	return nil
}

// setupAuthFlags configures flags for authentication commands.
func setupAuthFlags(cmd *cobra.Command, username, password *string) error {
	cmd.Flags().StringVarP(username, "username", "u", "", "Username for authentication (required)")
	cmd.Flags().StringVarP(password, "password", "p", "", "Password for authentication (required)")

	if err := cmd.MarkFlagRequired("username"); err != nil {
		return err
	}

	return cmd.MarkFlagRequired("password")
}

// AuthCommand returns the parsed AuthRequest structure.
func AuthCommand() (*entities.AuthRequest, error) {
	// Root command setup
	rootCmd := cobraCommands[cliKey]

	// Sub-command for insecure login
	unsecureLogin := cobraCommands[insecureKey]

	// Sub-command for TUI login
	runWithTUI := cobraCommands[runKey]

	// Common flag setup
	if err := setupAuthFlags(unsecureLogin, &username, &password); err != nil {
		return nil, err
	}

	// Add sub-commands to root
	rootCmd.AddCommand(unsecureLogin, runWithTUI)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}

	// Validate credentials
	if username == "" || password == "" {
		return nil, fmt.Errorf("invalid empty credentials")
	}

	return &entities.AuthRequest{
		Username: username,
		Password: password,
	}, nil
}
