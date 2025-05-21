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