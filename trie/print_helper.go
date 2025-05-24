package trie

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TrieNode 结构用于JSON序列化，便于前端展示
type TrieNode struct {
	Type     string      `json:"type"`
	Key      string      `json:"key,omitempty"`
	Value    string      `json:"value,omitempty"`
	Hash     string      `json:"hash,omitempty"`
	Children []*TrieNode `json:"children,omitempty"`
}

// PrintTrie 打印当前Trie结构的文本表示
func (t *Trie) PrintTrie() {
	fmt.Printf("\n==== Trie Structure ====\n")
	fmt.Printf("Owner: %x\n", t.owner)
	fmt.Printf("Root Hash: %x\n", t.Hash())
	fmt.Printf("Uncommitted Changes: %d\n", t.uncommitted)
	fmt.Println("\nHierarchy:")
	printNode(t.root, "", 0)
	fmt.Println("=====================\n")
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

// ConvertToJSON 将Trie转换为JSON格式
func (t *Trie) ConvertToJSON() ([]byte, error) {
	rootNode := convertNodeToTrieNode(t.root)
	return json.Marshal(rootNode)
}

// convertNodeToTrieNode 将内部节点转换为前端友好的TrieNode结构
func convertNodeToTrieNode(n node) *TrieNode {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *shortNode:
		node := &TrieNode{
			Type: "short",
			Key:  fmt.Sprintf("%x", n.Key),
		}

		childNode := convertNodeToTrieNode(n.Val)
		if childNode != nil {
			node.Children = []*TrieNode{childNode}
		}

		if n.flags.hash != nil {
			node.Hash = fmt.Sprintf("%x", n.flags.hash)
		}

		return node

	case *fullNode:
		node := &TrieNode{
			Type:     "branch",
			Children: make([]*TrieNode, 0, 16),
		}

		for i, child := range n.Children {
			if child == nil {
				continue
			}

			childNode := convertNodeToTrieNode(child)
			if childNode != nil {
				childNode.Key = fmt.Sprintf("%x", i)
				node.Children = append(node.Children, childNode)
			}
		}

		if n.flags.hash != nil {
			node.Hash = fmt.Sprintf("%x", n.flags.hash)
		}

		return node

	case hashNode:
		return &TrieNode{
			Type: "hash",
			Hash: fmt.Sprintf("%x", n),
		}

	case valueNode:
		return &TrieNode{
			Type:  "value",
			Value: fmt.Sprintf("%x", n),
		}
	}

	return nil
}

// printNode 递归打印节点及其子节点，带有适当的缩进
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

// nodeStateMarker 返回表示节点状态的标记
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
