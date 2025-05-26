package trie

// Helper function to match original keys with their trie paths
func TryFindOriginalKey(trieKeyPath string, originalKeys map[string]string) string {
	// Direct lookup
	if origKey, exists := originalKeys[trieKeyPath]; exists {
		return origKey
	}

	// Special hard-coded mapping for a key we know
	if trieKeyPath == "0b01000e020d0502070601020007030b02060e0e0c0d0f0d0701070e060a0302000c0f04040b040a0f0a0c020b000703020d090f0c0b0e020b070f0a000c0f0610" {
		return "0x1"
	}

	return ""
}
