package model

import (
	"encoding/json"
)

// CleanToolSchemaForGemini removes fields that Gemini doesn't support from tool parameters
func CleanToolSchemaForGemini(tools []Tool) []Tool {
	if len(tools) == 0 {
		return tools
	}

	cleaned := make([]Tool, len(tools))
	for i, tool := range tools {
		cleaned[i] = tool
		if tool.Function.Parameters != nil && len(tool.Function.Parameters) > 0 {
			var schema map[string]interface{}
			if err := json.Unmarshal(tool.Function.Parameters, &schema); err == nil {
				cleanedSchema := cleanSchema(schema)
				if cleanedJSON, err := json.Marshal(cleanedSchema); err == nil {
					cleaned[i].Function.Parameters = cleanedJSON
				}
			}
		}
	}
	return cleaned
}

// cleanSchema recursively removes unsupported fields from JSON schema
func cleanSchema(schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		return nil
	}

	// Create a copy to avoid modifying the original
	result := make(map[string]interface{})
	for k, v := range schema {
		result[k] = v
	}

	// Remove unsupported fields
	delete(result, "$schema")
	delete(result, "additionalProperties")
	delete(result, "$defs")
	delete(result, "definitions")

	// Recursively clean nested schemas
	if props, ok := result["properties"].(map[string]interface{}); ok {
		cleanedProps := make(map[string]interface{})
		for key, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				cleanedProps[key] = cleanSchema(propMap)
			} else {
				cleanedProps[key] = prop
			}
		}
		result["properties"] = cleanedProps
	}

	// Clean items schema for arrays
	if items, ok := result["items"].(map[string]interface{}); ok {
		result["items"] = cleanSchema(items)
	}

	// Clean anyOf/oneOf/allOf
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if arr, ok := result[key].([]interface{}); ok {
			cleaned := make([]interface{}, len(arr))
			for i, item := range arr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					cleaned[i] = cleanSchema(itemMap)
				} else {
					cleaned[i] = item
				}
			}
			result[key] = cleaned
		}
	}

	return result
}

// CleanToolSchemaJSON cleans tool schema from raw JSON bytes
func CleanToolSchemaJSON(toolsJSON []byte) ([]byte, error) {
	var tools []Tool
	if err := json.Unmarshal(toolsJSON, &tools); err != nil {
		return toolsJSON, err
	}

	cleaned := CleanToolSchemaForGemini(tools)
	return json.Marshal(cleaned)
}

