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

// rolesCmd represents the roles command
var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage roles",
	Long: `View and manage roles in the RBAC system, including listing all
available roles and viewing the role hierarchy.`,
}

// rolesListCmd lists all available roles
var rolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available roles",
	Long: `Display a list of all roles defined in the system along with their
descriptions and parent roles (if any).`,
	Example: `  # List all roles in table format
  rbac-cli roles list

  # List all roles in JSON format
  rbac-cli roles list --output json

  # List all roles in YAML format
  rbac-cli roles list --output yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		roles, err := repo.ListRoles(ctx)
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}

		return outputRoles(roles)
	},
}

// rolesHierarchyCmd displays the role hierarchy
var rolesHierarchyCmd = &cobra.Command{
	Use:   "hierarchy",
	Short: "Display role hierarchy tree",
	Long: `Display the hierarchical relationship between roles in a tree format,
showing parent-child relationships and role inheritance.`,
	Example: `  # Display role hierarchy as a tree
  rbac-cli roles hierarchy

  # Display role hierarchy in JSON format
  rbac-cli roles hierarchy --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		hierarchy, err := repo.GetRoleHierarchy(ctx)
		if err != nil {
			return fmt.Errorf("failed to get role hierarchy: %w", err)
		}

		return outputHierarchy(hierarchy)
	},
}

func init() {
	// Add subcommands
	rolesCmd.AddCommand(rolesListCmd)
	rolesCmd.AddCommand(rolesHierarchyCmd)
}

// outputRoles outputs a list of roles based on the format flag
func outputRoles(roles []rbac.Role) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(roles)

	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(roles)

	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Parent", "Created At"})
		table.SetBorder(true)
		table.SetAutoWrapText(true)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("|")
		table.SetColumnSeparator("|")
		table.SetRowSeparator("-")
		table.SetColWidth(40)

		for _, role := range roles {
			parent := role.Parent
			if parent == "" {
				parent = "(root)"
			}
			createdAt := role.CreatedAt.Format(time.RFC3339)

			table.Append([]string{role.Name, role.Description, parent, createdAt})
		}

		table.Render()
		return nil

	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json, yaml)", outputFormat)
	}
}

// outputHierarchy outputs the role hierarchy based on the format flag
func outputHierarchy(hierarchy map[string][]string) error {
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(hierarchy)

	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(hierarchy)

	case "table":
		// For table format, display as a tree
		fmt.Println("Role Hierarchy:")
		fmt.Println(strings.Repeat("=", 50))

		// Find root roles (roles that are not children of any other role)
		allChildren := make(map[string]bool)
		for _, children := range hierarchy {
			for _, child := range children {
				allChildren[child] = true
			}
		}

		// Get all parent roles
		allParents := make(map[string]bool)
		for parent := range hierarchy {
			allParents[parent] = true
		}

		// Find roots (parents that are not children)
		var roots []string
		for parent := range allParents {
			if !allChildren[parent] {
				roots = append(roots, parent)
			}
		}

		// If no hierarchy exists, display a message
		if len(hierarchy) == 0 {
			fmt.Println("(no role hierarchy defined)")
			return nil
		}

		// Print tree for each root
		visited := make(map[string]bool)
		for _, root := range roots {
			printRoleTree(root, hierarchy, "", visited)
		}

		// Print any orphaned hierarchies (roles with children but are themselves children)
		for parent := range hierarchy {
			if !visited[parent] {
				printRoleTree(parent, hierarchy, "", visited)
			}
		}

		return nil

	default:
		return fmt.Errorf("invalid output format: %s (valid options: table, json, yaml)", outputFormat)
	}
}

// printRoleTree recursively prints the role hierarchy as a tree
func printRoleTree(role string, hierarchy map[string][]string, prefix string, visited map[string]bool) {
	if visited[role] {
		return
	}
	visited[role] = true

	fmt.Printf("%s%s\n", prefix, role)

	children := hierarchy[role]
	for i, child := range children {
		isLast := i == len(children)-1
		var newPrefix string
		if isLast {
			fmt.Printf("%s└── ", prefix)
			newPrefix = prefix + "    "
		} else {
			fmt.Printf("%s├── ", prefix)
			newPrefix = prefix + "│   "
		}

		printRoleTree(child, hierarchy, newPrefix, visited)
	}
}
