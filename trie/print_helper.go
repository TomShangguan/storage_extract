package trie

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// TrieNode structure for JSON serialization, for frontend display
type TrieNode struct {
	Type            string          `json:"type"`
	Key             string          `json:"key,omitempty"`
	OriginalKey     string          `json:"originalKey,omitempty"` // Before hashing
	Value           string          `json:"value,omitempty"`
	OriginalValue   string          `json:"originalValue,omitempty"` // Before hashing/encoding
	Hash            string          `json:"hash,omitempty"`
	BranchIndex     int             `json:"branchIndex,omitempty"` // For branch children
	Children        []*TrieNode     `json:"children,omitempty"`
	Depth           int             `json:"depth,omitempty"`
	IsLeaf          bool            `json:"isLeaf,omitempty"`          // Is this a leaf node
	KeyPath         string          `json:"keyPath,omitempty"`         // Full path to this node
	HashedKeyPath   string          `json:"hashedKeyPath,omitempty"`   // Hashed version of the path
	SlotMap         map[string]bool `json:"slotMap,omitempty"`         // Map of all slots in a branch node (filled and empty)
	FilledSlotCount int             `json:"filledSlotCount,omitempty"` // Number of filled slots in a branch
	TotalSlotCount  int             `json:"totalSlotCount,omitempty"`  // Total number of slots in a branch
}

// PrintTrie print the text representation of the current trie structure
func (t *Trie) PrintTrie() {
	fmt.Printf("\n==== Trie Structure ====\n")
	fmt.Printf("Owner: %x\n", t.owner)
	fmt.Printf("Root Hash: %x\n", t.Hash())
	fmt.Printf("Uncommitted Changes: %d\n", t.uncommitted)
	fmt.Println("\nHierarchy:")
	printNode(t.root, "", 0)
	fmt.Println("=====================")
}

// PrintTrieTo writes the trie structure to a strings.Builder instead of stdout
func (t *Trie) PrintTrieTo(w *strings.Builder) {
	fmt.Fprintf(w, "\n==== Trie Structure ====\n")
	fmt.Fprintf(w, "Owner: %x\n", t.owner)
	fmt.Fprintf(w, "Root Hash: %x\n", t.Hash())
	fmt.Fprintf(w, "Uncommitted Changes: %d\n", t.uncommitted)
	fmt.Fprintf(w, "\nHierarchy:\n")
	printNodeTo(t.root, w, "", 0)
	fmt.Fprintf(w, "=====================\n")
}

// PrintTrieToFormatted creates a better formatted text representation for frontend
func (t *Trie) PrintTrieToFormatted(w *strings.Builder) {
	fmt.Fprintf(w, "Hierarchy:\n")
	printNodeFormattedTo(t.root, w, "", 0, true, "", nil)
}

// PrintTrieToFormattedWithKeys creates a better formatted text with key mapping info
func (t *Trie) PrintTrieToFormattedWithKeys(w *strings.Builder, originalKeys map[string]string) {
	fmt.Fprintf(w, "Hierarchy:\n")
	printNodeFormattedTo(t.root, w, "", 0, true, "", originalKeys)
}

// ConvertToJSON convert the trie to JSON format
func (t *Trie) ConvertToJSON() ([]byte, error) {
	rootNode := convertNodeToTrieNode(t.root, 0, -1, nil, nil)
	return json.Marshal(rootNode)
}

// ConvertToJSONWithOriginalKeys convert the trie to JSON format with original keys and values
func (t *Trie) ConvertToJSONWithOriginalKeys(originalKeys map[string]string, originalValues map[string]string) ([]byte, error) {
	rootNode := convertNodeToTrieNode(t.root, 0, -1, originalKeys, originalValues)
	return json.Marshal(rootNode)
}

// convertNodeToTrieNode convert the internal node to a frontend-friendly TrieNode structure
func convertNodeToTrieNode(n node, depth int, branchIndex int, originalKeys map[string]string, originalValues map[string]string) *TrieNode {
	return convertNodeToTrieNodeWithPath(n, depth, branchIndex, originalKeys, originalValues, "")
}

// hexToNibbles converts a hex-encoded byte array to nibble array (each nibble becomes a byte)
func hexToNibbles(hexBytes []byte) []byte {
	nibbles := make([]byte, len(hexBytes)*2)
	for i, b := range hexBytes {
		nibbles[i*2] = b >> 4     // High nibble
		nibbles[i*2+1] = b & 0x0F // Low nibble
	}
	return nibbles
}

// convertNodeToTrieNodeWithPath recursive convert the node to a frontend-friendly TrieNode structure, and track the full path
func convertNodeToTrieNodeWithPath(n node, depth int, branchIndex int, originalKeys map[string]string, originalValues map[string]string, currentPath string) *TrieNode {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *shortNode:
		keyHex := fmt.Sprintf("%x", n.Key)
		fullKeyPath := currentPath + keyHex

		node := &TrieNode{
			Type:        "short",
			Key:         keyHex,
			KeyPath:     fullKeyPath,
			Depth:       depth,
			BranchIndex: branchIndex,
		}

		// Check if we have original key information for this node
		if originalKeys != nil {
			fmt.Printf("DEBUG: Checking key %s against %d original keys\n", keyHex, len(originalKeys))

			// Try different matching strategies
			var originalKey string
			var found bool

			// Strategy 1: Direct key match
			if originalKey, found = originalKeys[keyHex]; found {
				fmt.Printf("DEBUG: Found original key via direct match %s -> %s\n", keyHex, originalKey)
			} else if originalKey, found = originalKeys[fullKeyPath]; found {
				fmt.Printf("DEBUG: Found original key via path match %s -> %s\n", fullKeyPath, originalKey)
			} else {
				// Strategy 2: Try to reverse the nibble encoding process
				// The trie key might be nibble-encoded version of a hash
				// Try to find if any original key's hash matches when nibble-encoded
				for mapKey, mapValue := range originalKeys {
					fmt.Printf("DEBUG: Trying to match %s with map key %s\n", keyHex, mapKey)

					// Check if mapKey could be a hash that when nibble-encoded matches our key
					if len(mapKey) == 64 { // Standard hash length
						// Convert hash to nibbles format
						hashBytes, err := hex.DecodeString(mapKey)
						if err == nil {
							nibbles := hexToNibbles(hashBytes)
							nibblesHex := fmt.Sprintf("%x", nibbles)
							if nibblesHex == keyHex {
								originalKey = mapValue
								found = true
								fmt.Printf("DEBUG: Found match via nibble encoding: %s -> %s\n", keyHex, originalKey)
								break
							}
						}
					}
				}
			}

			if found {
				node.OriginalKey = originalKey
			} else {
				fmt.Printf("DEBUG: No original key found for %s or %s\n", keyHex, fullKeyPath)
			}
		}

		// For root short nodes, make sure we emphasize that it's a root
		if depth == 0 {
			node.Type = "root_short" // Special type for styling in frontend
		}

		// Determine the kind of shortNode based on what's inside Val
		if valueNode, ok := n.Val.(valueNode); ok {
			// This is a shortNode with a valueNode (a terminal node)
			node.Type = "shortNode_value"
			node.Value = fmt.Sprintf("%02x", valueNode)
			node.IsLeaf = true // Keep this for backwards compatibility

			// Try to find original value from the mapping
			if originalValues != nil {
				fmt.Printf("DEBUG: Checking value %s against %d original values\n", node.Value, len(originalValues))

				// Try different value format strategies
				var originalValue string
				var valueFound bool

				// Strategy 1: Direct value match (padded with 0x)
				if originalValue, valueFound = originalValues["0x"+node.Value]; valueFound {
					fmt.Printf("DEBUG: Found original value via 0x-prefixed match %s -> %s\n", node.Value, originalValue)
				} else if originalValue, valueFound = originalValues[node.Value]; valueFound {
					fmt.Printf("DEBUG: Found original value via direct match %s -> %s\n", node.Value, originalValue)
				} else {
					// Strategy 2: Try trimmed version (remove leading zeros)
					trimmedValue := strings.TrimLeft(node.Value, "0")
					if trimmedValue == "" {
						trimmedValue = "0"
					}
					if originalValue, valueFound = originalValues[trimmedValue]; valueFound {
						fmt.Printf("DEBUG: Found original value via trimmed match %s -> %s\n", trimmedValue, originalValue)
					} else if originalValue, valueFound = originalValues["0x"+trimmedValue]; valueFound {
						fmt.Printf("DEBUG: Found original value via 0x-trimmed match %s -> %s\n", trimmedValue, originalValue)
					} else {
						// Strategy 3: Try padded version (add leading zeros to make it 64 chars)
						paddedValue := fmt.Sprintf("%064s", node.Value)
						if originalValue, valueFound = originalValues[paddedValue]; valueFound {
							fmt.Printf("DEBUG: Found original value via padded match %s -> %s\n", paddedValue, originalValue)
						} else if originalValue, valueFound = originalValues["0x"+paddedValue]; valueFound {
							fmt.Printf("DEBUG: Found original value via 0x-padded match %s -> %s\n", paddedValue, originalValue)
						}
					}
				}

				if valueFound {
					node.OriginalValue = originalValue
				} else {
					fmt.Printf("DEBUG: No original value found for %s\n", node.Value)
				}
			}
		} else {
			// This is a shortNode with another node (an extension node)
			node.Type = "shortNode_extension"
			childNode := convertNodeToTrieNodeWithPath(n.Val, depth+1, -1, originalKeys, originalValues, fullKeyPath)
			if childNode != nil {
				node.Children = []*TrieNode{childNode}
			}
		}

		// Only include hash for non-leaf nodes if really needed
		if !node.IsLeaf && n.flags.hash != nil {
			node.Hash = fmt.Sprintf("%x", n.flags.hash)
		}

		return node

	case *fullNode:
		branchNode := &TrieNode{
			Type:            "branch",
			KeyPath:         currentPath,
			Depth:           depth,
			BranchIndex:     branchIndex,
			FilledSlotCount: countFilledSlots(n.Children),
			TotalSlotCount:  16,
			SlotMap:         make(map[string]bool),
		}

		for i, child := range n.Children {
			if child != nil {
				childPath := fmt.Sprintf("%s%x", currentPath, i)
				branchNode.SlotMap[fmt.Sprintf("%x", i)] = true
				branchNode.Children = append(branchNode.Children, convertNodeToTrieNodeWithPath(child, depth+1, i, originalKeys, originalValues, childPath))
			} else {
				branchNode.SlotMap[fmt.Sprintf("%x", i)] = false
			}
		}
		return branchNode

	case hashNode:
		return &TrieNode{
			Type: "hash",
			Hash: fmt.Sprintf("%x", []byte(n)),
		}

	case valueNode:
		return &TrieNode{
			Type:  "value",
			Value: fmt.Sprintf("%x", []byte(n)),
		}

	default:
		return &TrieNode{
			Type: fmt.Sprintf("unknown (%T)", n),
		}
	}
}

// printNode recursive print node and its children with appropriate indentation
func printNode(n node, prefix string, depth int) {
	if n == nil {
		fmt.Printf("%s<nil>\n", strings.Repeat("  ", depth))
		return
	}

	indent := strings.Repeat("  ", depth)

	switch n := n.(type) {
	case *shortNode:
		fmt.Printf("%s%s└─ Short[%s] Key:%x\n", indent, prefix, nodeStateMarker(n.flags), n.Key)
		printNode(n.Val, "", depth+1)

	case *fullNode:
		fmt.Printf("%s%s└─ Branch[%s]\n", indent, prefix, nodeStateMarker(n.flags))
		for i, child := range n.Children {
			if child != nil {
				childPrefix := fmt.Sprintf("[%x] ", i)
				printNode(child, childPrefix, depth+1)
			}
		}

	case hashNode:
		fmt.Printf("%s%s└─ Hash: %x\n", indent, prefix, []byte(n))

	case valueNode:
		if len(n) <= 8 {
			fmt.Printf("%s%s└─ Value: %x\n", indent, prefix, []byte(n))
		} else {
			fmt.Printf("%s%s└─ Value: %x...%x (len=%d)\n",
				indent, prefix, n[:4], n[len(n)-4:], len(n))
		}
	}
}

// printNodeTo is like printNode but writes to a strings.Builder
func printNodeTo(n node, w *strings.Builder, prefix string, depth int) {
	if n == nil {
		fmt.Fprintf(w, "%s<nil>\n", strings.Repeat("  ", depth))
		return
	}

	indent := strings.Repeat("  ", depth)

	switch n := n.(type) {
	case *shortNode:
		fmt.Fprintf(w, "%s%s└─ Short[%s] Key:%x\n", indent, prefix, nodeStateMarker(n.flags), n.Key)
		printNodeTo(n.Val, w, "", depth+1)

	case *fullNode:
		fmt.Fprintf(w, "%s%s└─ Branch[%s]\n", indent, prefix, nodeStateMarker(n.flags))
		for i, child := range n.Children {
			if child != nil {
				childPrefix := fmt.Sprintf("[%x] ", i)
				printNodeTo(child, w, childPrefix, depth+1)
			}
		}

	case hashNode:
		fmt.Fprintf(w, "%s%s└─ Hash: %x\n", indent, prefix, []byte(n))

	case valueNode:
		if len(n) <= 8 {
			fmt.Fprintf(w, "%s%s└─ Value: %x\n", indent, prefix, []byte(n))
		} else {
			fmt.Fprintf(w, "%s%s└─ Value: %x...%x (len=%d)\n",
				indent, prefix, n[:4], n[len(n)-4:], len(n))
		}
	}
}

// printNodeFormattedTo creates better formatted output for frontend display
func printNodeFormattedTo(n node, w *strings.Builder, prefix string, depth int, isLast bool, currentPath string, originalKeys map[string]string) {
	if n == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	connector := "└─"
	if !isLast {
		connector = "├─"
	}
	switch n := n.(type) {
	case *shortNode:
		keyHex := fmt.Sprintf("%x", n.Key)
		fullKeyPath := currentPath + keyHex

		if valueNode, ok := n.Val.(valueNode); ok {
			// This is a shortNode with a valueNode
			fmt.Fprintf(w, "%s%s%s Short Node\n", indent, prefix, connector)
			fmt.Fprintf(w, "%s   Key: %s\n", indent, keyHex)

			// Show the full path if available
			if len(fullKeyPath) > 0 && fullKeyPath != keyHex {
				fmt.Fprintf(w, "%s   Full Path: %s\n", indent, fullKeyPath)
			}

			// Format value as proper hex with 0x prefix
			valueHex := fmt.Sprintf("%x", []byte(valueNode))
			// Trim leading zeros but keep at least one digit
			valueHex = strings.TrimLeft(valueHex, "0")
			if valueHex == "" {
				valueHex = "0"
			}
			// Add 0x prefix
			fmt.Fprintf(w, "%s   Value: 0x%s\n", indent, valueHex)

			// If it's the root node, emphasize that
			if depth == 0 {
				fmt.Fprintf(w, "%s   (Root Node)\n", indent)
			}

		} else {
			// This is a shortNode with another node as value
			fmt.Fprintf(w, "%s%s%s Short Node\n", indent, prefix, connector)
			fmt.Fprintf(w, "%s   Key: %s\n", indent, keyHex)

			// Show the full path if available
			if len(fullKeyPath) > 0 && fullKeyPath != keyHex {
				fmt.Fprintf(w, "%s   Full Path: %s\n", indent, fullKeyPath)
			}

			// If it's the root node, emphasize that
			if depth == 0 {
				fmt.Fprintf(w, "%s   (Root Node)\n", indent)
			}

			printNodeFormattedTo(n.Val, w, "", depth+1, true, fullKeyPath, originalKeys)
		}

	case *fullNode:
		fmt.Fprintf(w, "%s%s%s Branch Node\n", indent, prefix, connector)

		// Display path to this branch if available
		if len(currentPath) > 0 {
			fmt.Fprintf(w, "%s   Path Prefix: %s\n", indent, currentPath)
		}

		// Count actual children
		childCount := 0
		for _, child := range n.Children {
			if child != nil {
				childCount++
			}
		}

		// If this branch has multiple children, indicate that for clarity
		if childCount > 0 {
			fmt.Fprintf(w, "%s   Slots filled: %d/16\n", indent, childCount)
		} else {
			fmt.Fprintf(w, "%s   Empty Branch (no slots filled)\n", indent)
		}

		// Track current child for isLast calculation
		currentChild := 0

		// Print all slots - both populated and nil
		for i, child := range n.Children {
			slotIndex := fmt.Sprintf("%x", i)
			slotPrefix := fmt.Sprintf("[%s] ", slotIndex)

			if child != nil {
				currentChild++
				isLastChild := currentChild == childCount
				childPath := currentPath + slotIndex

				// Add some space to make the indentation consistent
				printNodeFormattedTo(child, w, slotPrefix, depth+1, isLastChild, childPath, originalKeys)
			} else {
				// Print nil branches for all slots to make structure clearer
				nilConnector := "└─"
				if currentChild < childCount {
					nilConnector = "├─"
				}
				fmt.Fprintf(w, "%s%s%s Nil\n", indent, slotPrefix, nilConnector)
			}
		}

	case hashNode:
		hashHex := fmt.Sprintf("%x", []byte(n))
		displayHash := hashHex
		if len(hashHex) > 32 {
			displayHash = hashHex[:16] + "..." + hashHex[len(hashHex)-16:]
		}
		fmt.Fprintf(w, "%s%s%s Hash Node: %s\n", indent, prefix, connector, displayHash)

	case valueNode:
		fmt.Fprintf(w, "%s%s%s Value Node: %x\n", indent, prefix, connector, []byte(n))
	}
}

// countFilledSlots counts the number of non-nil children in a branch node
func countFilledSlots(children [17]node) int {
	count := 0
	for _, child := range children {
		if child != nil {
			count++
		}
	}
	return count
}

// nodeStateMarker
func nodeStateMarker(flag nodeFlag) string {
	if flag.hash != nil {
		if flag.dirty {
			return "H,D" // Hashed and Dirty
		}
		return "H" // Hashed
	}
	if flag.dirty {
		return "D" // Dirty
	}
	return "" // Clean
}

// 为StateTrie实现JSON转换方法
func (t *StateTrie) ConvertToJSON() ([]byte, error) {
	return t.trie.ConvertToJSON()
}

// PrintTrieToFormatted delegates to the underlying trie's method
func (t *StateTrie) PrintTrieToFormatted(w *strings.Builder) {
	t.trie.PrintTrieToFormatted(w)
}

// PrintTrieToFormattedWithKeys delegates to the underlying trie's method
func (t *StateTrie) PrintTrieToFormattedWithKeys(w *strings.Builder, originalKeys map[string]string) {
	t.trie.PrintTrieToFormattedWithKeys(w, originalKeys)
}

// ConvertToJSONWithOriginalKeys delegates to the underlying trie's method
func (t *StateTrie) ConvertToJSONWithOriginalKeys(originalKeys map[string]string, originalValues map[string]string) ([]byte, error) {
	return t.trie.ConvertToJSONWithOriginalKeys(originalKeys, originalValues)
}
