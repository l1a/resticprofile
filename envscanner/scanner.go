package envscanner

import (
	"fmt"
	"os"
	"regexp"
	"slices"

	"gopkg.in/yaml.v3"
)

// envVarRegex defines the pattern for finding environment variable references.
// It looks for the string ".Env." followed by a valid variable name (letters, numbers, underscore)
// and captures the variable name itself.
var envVarRegex = regexp.MustCompile(`\.Env\.([A-Za-z_][A-Za-z0-9_]*)`)

// ScanForEnvVariables reads a YAML configuration file, locates the section for a
// specific profile or group, and recursively scans it for all references to
// environment variables in the format ".Env.VAR_NAME".
// It returns a unique, sorted list of the variable names it finds (e.g., "HOME", "RESTIC_PASSWORD").
func ScanForEnvVariables(configFile string, profileOrGroupName string) ([]string, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", configFile, err)
	}

	// The root node is a document node, its content is the main mapping node.
	if len(root.Content) == 0 {
		return nil, nil // Empty file
	}
	mainNode := root.Content[0]

	// Find the node for the specific profile or group.
	targetNode := findNodeByKey(mainNode, "profiles", profileOrGroupName)
	if targetNode == nil {
		targetNode = findNodeByKey(mainNode, "groups", profileOrGroupName)
	}

	if targetNode == nil {
		// Profile or group not found, which is not an error, just means no variables to scan.
		return nil, nil
	}

	// Scan the target node and all its children for environment variable references.
	foundVars := make(map[string]struct{})
	scanNode(targetNode, foundVars)

	// Convert map keys to a slice
	var result []string
	for v := range foundVars {
		result = append(result, v)
	}
	slices.Sort(result) // Sort for deterministic output

	return result, nil
}

// findNodeByKey traverses a YAML mapping node to find a specific nested value.
// It first looks for a `topLevelKey` (like "profiles" or "groups") and then, within that
// section, searches for the `nestedKey` (the name of the profile or group).
// It returns the YAML node corresponding to the nested key, or nil if not found.
func findNodeByKey(root *yaml.Node, topLevelKey, nestedKey string) *yaml.Node {
	if root.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(root.Content); i += 2 {
		keyNode := root.Content[i]
		if keyNode.Value == topLevelKey {
			valueNode := root.Content[i+1]
			if valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					nestedKeyNode := valueNode.Content[j]
					if nestedKeyNode.Value == nestedKey {
						return valueNode.Content[j+1]
					}
				}
			}
			return nil // Found top-level key but it's not a map, or nested key not found
		}
	}
	return nil
}

// scanNode recursively traverses a yaml.Node and all of its children.
// If it finds a scalar string node, it scans the string value for any .Env.*
// patterns. All unique variable names found are added to the `foundVars` map.
func scanNode(node *yaml.Node, foundVars map[string]struct{}) {
	if node == nil {
		return
	}

	// If the node is a string, scan it.
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		matches := envVarRegex.FindAllStringSubmatch(node.Value, -1)
		for _, match := range matches {
			if len(match) > 1 {
				foundVars[match[1]] = struct{}{}
			}
		}
	}

	// Recurse into child nodes.
	for _, child := range node.Content {
		scanNode(child, foundVars)
	}
}
