package rbac

import (
	"sync"
)

var (
	hierarchyCache   = make(map[string][]string)
	hierarchyCacheMu sync.RWMutex
)

func ResolveRoles(roles []string, hierarchy map[string][]string) []string {
	if len(roles) == 0 {
		return roles
	}

	// Use a map to track unique roles
	resolved := make(map[string]bool)

	// Recursively expand each role
	for _, role := range roles {
		expandRole(role, hierarchy, resolved, make(map[string]bool))
	}

	// Convert map to slice
	result := make([]string, 0, len(resolved))
	for role := range resolved {
		result = append(result, role)
	}

	return result
}

func expandRole(role string, hierarchy map[string][]string, resolved map[string]bool, visited map[string]bool) {
	// If already resolved, skip
	if resolved[role] {
		return
	}

	// Check for cycle
	if visited[role] {
		// Cycle detected, skip to avoid infinite recursion
		return
	}

	// Mark as visited in current path
	visited[role] = true

	// Add this role to resolved set
	resolved[role] = true

	// Expand children (roles this role inherits from)
	if children, ok := hierarchy[role]; ok {
		for _, child := range children {
			expandRole(child, hierarchy, resolved, visited)
		}
	}

	// Remove from visited (backtrack)
	delete(visited, role)
}

func HasRole(roles []string, requiredRole string, hierarchy map[string][]string) bool {
	expanded := ResolveRoles(roles, hierarchy)
	for _, role := range expanded {
		if role == requiredRole {
			return true
		}
	}
	return false
}

func HasAnyRole(roles []string, requiredRoles []string, hierarchy map[string][]string) bool {
	expanded := ResolveRoles(roles, hierarchy)

	// Create a set of expanded roles for faster lookup
	roleSet := make(map[string]bool)
	for _, role := range expanded {
		roleSet[role] = true
	}

	// Check if any required role is in the set
	for _, required := range requiredRoles {
		if roleSet[required] {
			return true
		}
	}

	return false
}

func BuildRoleTree(hierarchy map[string][]string) []*RoleNode {
	// Find root roles (roles that are not children of any other role)
	childRoles := make(map[string]bool)
	for _, children := range hierarchy {
		for _, child := range children {
			childRoles[child] = true
		}
	}

	// Build tree from roots
	var roots []*RoleNode
	for role := range hierarchy {
		if !childRoles[role] {
			// This is a root role
			node := buildRoleNode(role, hierarchy, make(map[string]bool))
			roots = append(roots, node)
		}
	}

	return roots
}

func buildRoleNode(role string, hierarchy map[string][]string, visited map[string]bool) *RoleNode {
	// Prevent cycles
	if visited[role] {
		return &RoleNode{Name: role + " (cycle)", Children: nil}
	}

	visited[role] = true

	node := &RoleNode{
		Name:     role,
		Children: []*RoleNode{},
	}

	// Add children
	if children, ok := hierarchy[role]; ok {
		for _, child := range children {
			childNode := buildRoleNode(child, hierarchy, visited)
			node.Children = append(node.Children, childNode)
		}
	}

	delete(visited, role)

	return node
}

func ClearHierarchyCache() {
	hierarchyCacheMu.Lock()
	hierarchyCache = make(map[string][]string)
	hierarchyCacheMu.Unlock()
}
