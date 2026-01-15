package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long: `Create and manage the CLI configuration file (.rbac-cli.yaml).
This file stores default settings like database URL and output format.`,
}

// configInitCmd initializes a new config file
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new config file",
	Long: `Create a new .rbac-cli.yaml configuration file in the current directory
with default settings. You can customize the file after creation.`,
	Example: `  # Create config with default settings
  rbac-cli config init

  # Create config with custom database URL
  rbac-cli config init --db-url "postgres://localhost/mydb"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := ".rbac-cli.yaml"

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				return fmt.Errorf("config file already exists. Use --force to overwrite")
			}
		}

		// Get flags
		dbURLFlag, _ := cmd.Flags().GetString("db-url")
		outputFlag, _ := cmd.Flags().GetString("output")

		// Create default config
		config := Config{
			DatabaseURL: dbURLFlag,
			Output:      outputFlag,
		}

		// If no db-url provided, use a placeholder
		if config.DatabaseURL == "" {
			config.DatabaseURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		}

		// If no output provided, use table
		if config.Output == "" {
			config.Output = "table"
		}

		// Marshal to YAML
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// Write to file
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("Configuration file created: %s\n", configPath)
		fmt.Println("\nPlease edit the file to set your database connection string.")
		return nil
	},
}

// configShowCmd shows the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration settings, including values from
the config file, environment variables, and command-line flags.`,
	Example: `  # Show current configuration
  rbac-cli config show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current Configuration:")
		fmt.Println(strings.Repeat("=", 50))

		// Config file location
		configFile := cfgFile
		if configFile == "" {
			configFile = ".rbac-cli.yaml"
		}

		absPath, _ := filepath.Abs(configFile)
		fmt.Printf("Config file:    %s\n", absPath)

		// Check if config file exists
		if _, err := os.Stat(configFile); err != nil {
			fmt.Printf("                (not found)\n")
		} else {
			fmt.Printf("                (loaded)\n")
		}

		fmt.Printf("\nSettings:\n")
		fmt.Printf("Database URL:   %s\n", maskDatabaseURL(dbURL))
		fmt.Printf("Output format:  %s\n", outputFormat)

		fmt.Printf("\nEnvironment Variables:\n")
		if envDB := os.Getenv("RBAC_DB_URL"); envDB != "" {
			fmt.Printf("RBAC_DB_URL:    %s\n", maskDatabaseURL(envDB))
		} else {
			fmt.Printf("RBAC_DB_URL:    (not set)\n")
		}

		if envOutput := os.Getenv("RBAC_OUTPUT"); envOutput != "" {
			fmt.Printf("RBAC_OUTPUT:    %s\n", envOutput)
		} else {
			fmt.Printf("RBAC_OUTPUT:    (not set)\n")
		}

		return nil
	},
}

// configPathCmd shows the path to the config file
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show path to config file",
	Long:  `Display the absolute path to the configuration file.`,
	Example: `  # Show config file path
  rbac-cli config path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := cfgFile
		if configFile == "" {
			configFile = ".rbac-cli.yaml"
		}

		absPath, err := filepath.Abs(configFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		fmt.Println(absPath)
		return nil
	},
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)

	// Add flags for init command
	configInitCmd.Flags().String("db-url", "", "database connection string")
	configInitCmd.Flags().String("output", "table", "default output format (table, json, yaml)")
	configInitCmd.Flags().Bool("force", false, "overwrite existing config file")
}

// maskDatabaseURL masks sensitive parts of the database URL
func maskDatabaseURL(url string) string {
	if url == "" {
		return "(not set)"
	}

	// Mask password in URL
	// Simple masking - just show the protocol and host if possible
	if len(url) > 20 {
		return url[:15] + "...***..."
	}

	return "***"
}
