# Storage Extract Project Progress Log

## 2025/5/18 Progress

### Newly Implemented Components
- **Common Package**
  - Address and Hash types
  - Hex utilities
  - Basic constants
- **Crypto Package**
  - KeccakState interface
  - Basic hashing structure
- **Types Package**
  - StateAccount structure
  - EmptyRootHash constant
- **State Package**
  - Basic StateDB and StateObject structures

### TODOs
1. Fix `Keccak256Hash` implementation 
2. Implement database reading in `getStateObject`
3. Add journal system functionality
4. Implement `GetCommittedState` database reading
5. Fix `SetState` parameter passing

---

## 2025/5/19 Progress

### Newly Implemented Components
- **State Package**
  - Journal system structure
  - Storage change tracking
  - CachingDB interface definition
  - Storage mutation tracking system
  - State update structure
  - Basic database interface
  - Concurrent storage updates in `IntermediateRoot`

### TODOs

1. Implement reading the storage trie from the database in `getTrie`.
2. Add prefetch logic and committed state retrieval from the database in `finalise`.
3. Complete the database interface and implement the mock key-value store for `CachingDB`.
4. Implement the `OpenStorageTrie` and related trie operations.
5. Address all remaining placeholder and unimplemented functions in the state management code.


## 2025/5/20 Progress

### Newly Implemented Components
- **Trie Package**
  - Enhanced documentation for MPT operations
  - Improved insert function logic with detailed examples
  - Better commented code structure for:
    - shortNode handling
    - fullNode operations 
    - nil node cases
  - Clear explanation of MPT path compression mechanism

### TODOs
1. Implement trie delete operations
2. Add support for hash node handling
3. Create high-level user interface for trie operations:
   - Add simple key-value get/set methods
   - Implement trie traversal functions
   - Add trie snapshot functionality
4. Implement storage proof generation:
   - Merkle proof generation
   - Proof verification methods
5. Add trie import/export functionality:
   - JSON serialization support
   - State dump utilities
   - Trie reconstruction from dump