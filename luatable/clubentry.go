package luatable

import "fmt"

// ClubEntryAdapter adapts a ClubEntry-like structure to LuaEntry interface.
// This allows converting club membership data to Lua table format.
type ClubEntryAdapter struct {
	IP     string
	ID     int64
	ClubID int64
}

// GetFields returns a map of field names to values for Lua table generation.
func (e ClubEntryAdapter) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"id":     e.ID,
		"clubid": e.ClubID,
	}
}

// GetKey returns the IP address as the grouping key.
func (e ClubEntryAdapter) GetKey() string {
	return e.IP
}

// ConvertClubEntriesToLuaTable converts a slice of ClubEntry-like structures
// to Lua table format in array structure.
// This is a convenience function for club data.
func ConvertClubEntriesToLuaTable(entries []ClubEntryAdapter) string {
	luaEntries := make([]LuaEntry, len(entries))
	for i := range entries {
		luaEntries[i] = entries[i]
	}
	return ToLuaTable(luaEntries)
}

// ConvertClubEntriesToLuaTableByIP converts a slice of ClubEntry-like structures
// to Lua table format grouped by IP address.
// This is a convenience function for club data.
func ConvertClubEntriesToLuaTableByIP(entries []ClubEntryAdapter) string {
	luaEntries := make([]LuaEntry, len(entries))
	for i := range entries {
		luaEntries[i] = entries[i]
	}
	return ToLuaTableByKey(luaEntries)
}

// ConvertToLuaTable converts club data from file to Lua table format.
// It accepts a parser function that reads and parses the file, returning ClubEntryAdapter entries.
// format can be "array" for array format or "byip" for grouped by IP format.
func ConvertToLuaTable(parser func(string) ([]ClubEntryAdapter, error), filePath string, format string) (string, error) {
	entries, err := parser(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	if format == "byip" {
		return ConvertClubEntriesToLuaTableByIP(entries), nil
	}
	return ConvertClubEntriesToLuaTable(entries), nil
}

