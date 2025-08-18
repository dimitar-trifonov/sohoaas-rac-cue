package services

import "strings"

// getInputKeys returns the keys of a map[string]interface{} for logging
func getInputKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, exists := m[key]; exists {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return make(map[string]interface{})
}

func getBool(m map[string]interface{}, key string) bool {
	if val, exists := m[key]; exists {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if val, exists := m[key]; exists {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
		if strSlice, ok := val.([]string); ok {
			return strSlice
		}
	}
	return []string{}
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if val, exists := m[key]; exists {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
		if intVal, ok := val.(int); ok {
			return float64(intVal)
		}
	}
	return 0.0
}

// cleanMarkdownFromJSON removes markdown code block formatting from JSON strings
func cleanMarkdownFromJSON(content string) string {
	// Remove markdown code blocks (```json ... ``` or ``` ... ```)
	content = strings.TrimSpace(content)

	// Remove opening code block markers
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}

	// Remove closing code block markers
	content = strings.TrimSuffix(content, "```")

	// Clean up any remaining whitespace
	content = strings.TrimSpace(content)

	return content
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
