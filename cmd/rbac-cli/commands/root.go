package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/nicolasbonnici/gorest-rbac"
	"github.com/nicolasbonnici/gorest/database"
	_ "github.com/nicolasbonnici/gorest/database/postgres"
	_ "github.com/nicolasbonnici/gorest/database/sqlite"
)

var (
	// Global flags
	cfgFile      string
	dbURL        string
	outputFormat string

	// Version info
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Global database and repository
	db   database.Database
	repo *rbac.Repository
)

// Config represents the CLI configuration file
type Config struct {
	DatabaseURL string `yaml:"database_url"`
	Output      string `yaml:"output"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rbac-cli",
	Short: "GoREST RBAC Plugin CLI",
	Long: `A comprehensive CLI tool for managing RBAC (Role-Based Access Control)
in the GoREST framework.

This tool allows you to:
  - Manage user roles (list, show, promote, demote)
  - View role hierarchies
  - List available roles
  - Configure database connections

Examples:
  # List all users with their roles
  rbac-cli users list

  # Show roles for a specific user
  rbac-cli users show user-123

  # Assign a role to a user
  rbac-cli users promote user-123 admin

  # Remove a role from a user
  rbac-cli users demote user-123 moderator

  # List all available roles
  rbac-cli roles list

  # Display role hierarchy
  rbac-cli roles hierarchy`,
	PersistentPreRunE: initDB,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
		}
		return nil
	},
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .rbac-cli.yaml in current directory)")
	rootCmd.PersistentFlags().StringVar(&dbURL, "db-url", "", "database connection string (overrides config file)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, yaml")

	// Add subcommands
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(configCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		loadConfig(cfgFile)
	} else {
		// Search for config in current directory
		if _, err := os.Stat(".rbac-cli.yaml"); err == nil {
			loadConfig(".rbac-cli.yaml")
		}
	}

	// Environment variable takes precedence
	if envDB := os.Getenv("RBAC_DB_URL"); envDB != "" {
		dbURL = envDB
	}

	if envOutput := os.Getenv("RBAC_OUTPUT"); envOutput != "" {
		outputFormat = envOutput
	}
}

// loadConfig loads configuration from a YAML file
func loadConfig(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return // Silently ignore if config file doesn't exist
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse config file: %v\n", err)
		return
	}

	if dbURL == "" && config.DatabaseURL != "" {
		dbURL = config.DatabaseURL
	}

	if outputFormat == "table" && config.Output != "" {
		outputFormat = config.Output
	}
}

// initDB initializes the database connection
func initDB(cmd *cobra.Command, args []string) error {
	// Skip for commands that don't need DB
	cmdName := cmd.Name()
	if cmdName == "config" || cmdName == "init" || cmdName == "show" || cmdName == "path" ||
		cmdName == "help" || cmdName == "version" || cmdName == "completion" {
		return nil
	}

	if dbURL == "" {
		return fmt.Errorf("database URL is required. Set it via --db-url flag, config file, or RBAC_DB_URL environment variable")
	}

	// Determine driver from URL (auto-detect)
	driver := database.DetectDriver(dbURL)

	// Open database connection
	var err error
	db, err = database.Open(driver, dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize repository
	repo = rbac.NewRepository(db)

	return nil
}

// SetVersion sets version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}
