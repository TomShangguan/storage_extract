# MPT Hashing Process Example

## Initial Structure

Consider a small MPT with two key-value pairs:
```
shortNode("tea")
    └── fullNode
        ├── [1] -> "value1"     // path: "tea1" -> "value1"
        └── [2] -> "value2"     // path: "tea2" -> "value2"
```

## Detailed Hashing Process

### 1. Root Node Processing

```go
hash(shortNode("tea"), force=false)

1. Initial Check:
   - Check cache -> No cached hash
   - Proceed with hashing

2. Create Working Copies:
   collapsed = {
     Key: compact("tea"),
     Val: [to be processed fullNode]
   }
   
   cached = {
     Key: "tea",
     Val: [to be processed fullNode],
     flags.hash: nil
   }
```

### 2. Full Node Processing

```go
hashFullNodeChildren(fullNode)

1. Working States:
   collapsed = {
     Children: [
       0: nilValueNode,
       1: "value1",
       2: "value2",
       3-15: nilValueNode
     ]
   }

   cached = {
     Children: [
       0: nil,
       1: "value1",
       2: "value2",
       3-15: nil
     ],
     flags.hash: nil
   }
```

### 3. Node Size Evaluation

```go
1. fullNode:
   - RLP encoded size >= 32 bytes
   - Action: Create hashNode
   collapsed = hashNode(keccak256(rlp(fullNode)))
   cached.flags.hash = collapsed

2. shortNode:
   - RLP encoded size < 32 bytes
   - Action: Keep original format
   collapsed = original shortNode
   cached.flags.hash = nil
```

### 4. Return Values

```go
return (
    // hashed: What gets passed up
    shortNode {
        Key: compact("tea"),
        Val: hashNode([32 bytes])
    },
    
    // cached: What stays in memory
    shortNode {
        Key: "tea",
        Val: fullNode {
            Children: [nil, "value1", "value2", nil, ...],
            flags.hash: [32 bytes]
        },
        flags.hash: nil
    }
)
```

## Key Design Points

### Storage Optimization
- Nodes < 32 bytes: stored directly in parent
- Nodes >= 32 bytes: stored as hash reference

### Memory Management
- `collapsed`: Optimized for storage/hashing
- `cached`: Optimized for access/updates

### Performance Features
- Hash caching
- Parallel processing for fullNode children
- Memory pooling for hashers

## Common Use Cases

### 1. State Updates
```go
// Adding new value "tea3" -> "value3"
1. Navigate to fullNode
2. Add value at index 3
3. Rehash affected nodes only
```

### 2. State Reads
```go
// Reading "tea1"
1. Use cached structure
2. Follow full paths without database lookups
3. Return "value1" directly
```

### 3. Persistence
```go
// Saving to database
1. Large nodes -> Store hash reference
2. Small nodes -> Store inline in parent
3. Root node -> Always store hash
```