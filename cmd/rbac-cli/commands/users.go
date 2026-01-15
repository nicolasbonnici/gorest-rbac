package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nicolasbonnici/gorest-rbac"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage user roles",
	Long: `Manage user role assignments including listing users,
viewing user roles, promoting users to roles, and demoting users from roles.`,
}

// usersListCmd lists all users with their roles
var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users with their roles",
	Long: `Display a list of all users in the system along with their
assigned roles and the last time their roles were updated.`,
	Example: `  # List all users in table format
  rbac-cli users list

  # List all users in JSON format
  rbac-cli users list --output json

  # List all users in YAML format
  rbac-cli users list --output yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		users, err := repo.ListUsers(ctx)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		return outputUsers(users)
	},
}

// usersShowCmd shows roles for a specific user
var usersShowCmd = &cobra.Command{
	Use:   "show <user-id>",
	Short: "Show specific user's roles",
	Long: `Display detailed information about a user's role assignments,
including all assigned roles.`,
	Example: `  # Show user's roles
  rbac-cli users show user-123

  # Show user's roles in JSON format
  rbac-cli users show user-123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		ctx := context.Background()

		roles, err := repo.GetUserRoles(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user roles: %w", err)
		}

		// Create UserRoles struct for output
		userRoles := struct {
			UserID string   `json:"user_id" yaml:"user_id"`
			Roles  []string `json:"roles" yaml:"roles"`
		}{
			UserID: userID,
			Roles:  roles,
		}

		return outputData(userRoles)
	},
}

// usersPromoteCmd assigns a role to a user
var usersPromoteCmd = &cobra.Command{
	Use:   "promote <user-id> <role>",
	Short: "Assign role to user",
	Long: `Promote a user by assigning them a new role. The role must exist
in the system. This operation is audited.`,
	Example: `  # Promote user to admin
  rbac-cli users promote user-123 admin

  # Promote user to moderator
  rbac-cli users promote user-123 moderator`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		role := args[1]
		ctx := context.Background()

		// Get actor (default to "cli-admin" if not set)
		actor, _ := cmd.Flags().GetString("actor")
		if actor == "" {
			actor = "cli-admin"
		}

		if err := repo.AssignRole(ctx, userID, role, actor); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		fmt.Printf("Successfully assigned role '%s' to user '%s'\n", role, userID)
		return nil
	},
}

// usersDemoteCmd removes a role from a user
var usersDemoteCmd = &cobra.Command{
	Use:   "demote <user-id> <role>",
	Short: "Remove role from user",
	Long: `Demote a user by removing a role from their assignments. This operation
is audited.`,
	Example: `  # Demote user from admin
  rbac-cli users demote user-123 admin

  # Remove moderator role from user
  rbac-cli users demote user-123 moderator`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		role := args[1]
		ctx := context.Background()

		// Get actor (default to "cli-admin" if not set)
		actor, _ := cmd.Flags().GetString("actor")
		if actor == "" {
			actor = "cli-admin"
		}

		if err := repo.RemoveRole(ctx, userID, role, actor); err != nil {
			return fmt.Errorf("failed to remove role: %w", err)
		}

		fmt.Printf("Successfully removed role '%s' from user '%s'\n", role, userID)
		return nil
	},
}

func init() {
	// Add subcommands
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersShowCmd)
	usersCmd.AddCommand(usersPromoteCmd)
	usersCmd.AddCommand(usersDemoteCmd)

	// Add flags
	usersPromoteCmd.Flags().String("actor", "", "actor performing the promotion (default: cli-admin)")
	usersDemoteCmd.Flags().String("actor", "", "actor performing the demotion (default: cli-admin)")
}

// outputUsers outputs a list of users based on the format flag
func outputUsers(users []rbac.UserRoles) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(users)

	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(users)

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"User ID", "Roles", "Updated At"})
		table.SetBorder(true)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("|")
		table.SetColumnSeparator("|")
		table.SetRowSeparator("-")

		for _, user := range users {
			roles := strings.Join(user.Roles, ", ")
			if roles == "" {
				roles = "(none)"
			}
			updatedAt := user.UpdatedAt.Format(time.RFC3339)

			table.Append([]string{user.UserID, roles, updatedAt})
		}

		table.Render()
		return nil

	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json, yaml)", outputFormat)
	}
}

// outputData outputs generic data based on the format flag
func outputData(data interface{}) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)

	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(data)

	case "table":
		// For table format, try to display in a simple format
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)

	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json, yaml)", outputFormat)
	}
}
