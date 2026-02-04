// Package luatable provides functionality to convert tabular data
// to Lua table format. It supports multiple output formats including
// array format and grouped format (e.g., by key field).
//
// Example usage:
//
//	entries := []ClubEntry{...}
//	luaTable := luatable.ToLuaTable(entries)
//	fmt.Println(luaTable)
package luatable

import (
	"fmt"
	"sort"
	"strings"
)

// LuaEntry represents a data entry that can be converted to Lua table format.
// It provides methods to get field values for Lua table generation.
type LuaEntry interface {
	// GetFields returns a map of field names to values for Lua table generation.
	// Keys are field names (e.g., "id", "clubid"), values are the actual values.
	GetFields() map[string]interface{}

	// GetKey returns the key value for grouping (e.g., IP address).
	// This is used when grouping entries by a key field.
	GetKey() string
}

// ToLuaTable converts a slice of entries to Lua table format in array structure.
// Format: {[1]={id=xxx, clubid=xxx}, [2]={id=xxx, clubid=xxx}, ...}
func ToLuaTable(entries []LuaEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	var sb strings.Builder
	sb.WriteString("{\n")

	for i, entry := range entries {
		if i > 0 {
			sb.WriteString(",\n")
		}
		fields := entry.GetFields()
		sb.WriteString(fmt.Sprintf("  [%d]={", i+1))
		sb.WriteString(formatFields(fields))
		sb.WriteString("}")
	}

	sb.WriteString("\n}")
	return sb.String()
}

// ToLuaTableByKey converts a slice of entries to Lua table format grouped by key.
// Format: {["key"]={{id=xxx, clubid=xxx}, ...}, ...}
func ToLuaTableByKey(entries []LuaEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	// Group entries by key
	groups := make(map[string][]LuaEntry)
	for _, entry := range entries {
		key := entry.GetKey()
		groups[key] = append(groups[key], entry)
	}

	var sb strings.Builder
	sb.WriteString("{\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	firstKey := true
	for _, key := range keys {
		if !firstKey {
			sb.WriteString(",\n")
		}
		firstKey = false

		groupEntries := groups[key]
		sb.WriteString(fmt.Sprintf("  [\"%s\"]={\n", escapeLuaString(key)))
		for i, entry := range groupEntries {
			if i > 0 {
				sb.WriteString(",\n")
			}
			fields := entry.GetFields()
			sb.WriteString("    {")
			sb.WriteString(formatFields(fields))
			sb.WriteString("}")
		}
		sb.WriteString("\n  }")
	}

	sb.WriteString("\n}")
	return sb.String()
}

// formatFields formats a map of fields into Lua table field syntax.
// Example: id=123, clubid=456
func formatFields(fields map[string]interface{}) string {
	var parts []string
	for name, value := range fields {
		parts = append(parts, formatField(name, value))
	}
	// Sort field names for consistent output
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// formatField formats a single field name-value pair for Lua table.
func formatField(name string, value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%s=\"%s\"", name, escapeLuaString(v))
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%s=%d", name, v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%s=%d", name, v)
	case float32, float64:
		return fmt.Sprintf("%s=%g", name, v)
	case bool:
		if v {
			return fmt.Sprintf("%s=true", name)
		}
		return fmt.Sprintf("%s=false", name)
	default:
		// Fallback to string representation
		return fmt.Sprintf("%s=\"%v\"", name, escapeLuaString(fmt.Sprintf("%v", v)))
	}
}

// escapeLuaString escapes special characters in Lua strings.
// Currently handles basic escaping for quotes and backslashes.
func escapeLuaString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

