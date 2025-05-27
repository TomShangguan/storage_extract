# Ethereum Storage Visualizer

## Features

- **Account Management**: Create and manage Ethereum accounts.
- **MPT Visualization**:
    - **Text View**: A hierarchical text representation of the Merkle Patricia Trie (MPT).
    - **Tree View**: An interactive graphical representation of the MPT, clearly showing branch, extension, and leaf nodes, along with their relationships.
- **Multiple Address Support**: Visualize and manage MPTs for multiple Ethereum addresses.

## How to Run

### 1. Local Development Server

This is the default mode and starts a web server to interact with the visualizer.

```bash
go run main.go
```
Alternatively, you can build the executable first and then run it:
```bash
go build -o storage_extract
./storage_extract
```
By default, the server will start on port `8080`. You can specify a different port using the `-port` flag:
```bash
go run main.go -port=8888
# or
./storage_extract -port=8888
```
Once the server is running, open your web browser and navigate to `http://localhost:<port>` (e.g., `http://localhost:8080`).

### 2. GitHub Pages (Demo)

A live demo is available at: [https://tomshangguan.github.io/storage_extract/](https://tomshangguan.github.io/storage_extract/)

> **Note:** The GitHub Pages version may not always reflect the latest features from the main branch. The Merkle Proof functionality is currently under development and not operational in the demo version.

## Using the Web Interface

The web interface provides the following functionalities:

1.  **Create an Account**: Enter an Ethereum address (e.g., `0x...`) to initialize its corresponding state object and storage trie.
2.  **Set Storage**: For a selected account, input key-value pairs. Both keys and values should be provided in hexadecimal format (e.g., key: `0x01`, value: `0x123abc`).
3.  **Update Trie**: After setting or modifying storage key-value pairs, click the "Update Trie" button. This action commits the changes to the in-memory trie and refreshes the MPT visualization (both Text and Tree views).
4.  **View MPT**:
    *   **Text View**: Displays a raw, hierarchical text dump of the current trie structure. This is useful for a quick overview and debugging.
    *   **Tree View**: Renders an interactive, graphical representation of the MPT. Nodes are color-coded or shaped by type (branch, extension, leaf).

## Project Structure

```
/
├── go.mod, go.sum     
├── main.go            
├── README.md          
├── common/            # Utility functions and types (e.g., hex manipulation, custom types)
├── crypto/            # Cryptographic helpers, primarily Keccak256 hashing
├── ethdb/             # Database interface layer (currently uses a mock CachingDB for in-memory storage)
├── front/             # Frontend static files (HTML, CSS, JavaScript)
│   ├── index.html     
│   ├── css/           
│   └── js/            
├── state/             # Core state management, API handlers, and StateDB logic
├── trie/              # Merkle Patricia Trie (MPT) implementation and associated helper functions
│   └── trienode/      # MPT node definitions and specific proof generation/verification logic
└── types/             # Definitions for core Ethereum types (e.g., Address, Hash, StateAccount)
```

### Detailed Descriptions For Important Folders:

-   **`state/`**: This directory is central to managing the Ethereum state.
    -   `api_handlers.go`: Contains HTTP handlers for the backend API. These functions process requests from the frontend for actions like creating accounts, setting storage values, updating tries, and generating/verifying Merkle proofs.
    -   `statedb.go`: Implements the `StateDB` structure, which acts as the primary interface for interacting with the Ethereum state. It manages account objects and their respective storage tries.
    -   `state_object.go`: Defines the `StateObject` type, representing an individual Ethereum account. This includes its nonce, balance (not fully utilized in this visualizer's context), code hash, and the root of its storage trie.
    -   `journal.go`: Implements a journaling system for `StateDB`. This allows for tracking changes made to the state, enabling features like reverting to previous states (though not explicitly exposed in the UI, it's a foundational element for state consistency).
    -   `stateupdate.go`: Manages the process of applying updates to the state, ensuring changes are correctly reflected in the `StateDB` and underlying tries.
    -   `database.go`: Mock implementation for persisting state data.

-   **`trie/`**: **The Heart of Ethereum's Data Structure: Merkle Patricia Trie (MPT) Implementation**

    This directory houses the comprehensive implementation of the Merkle Patricia Trie, a sophisticated data structure crucial for Ethereum's state management, transaction recording, and receipt storage. The MPT allows for efficient and cryptographically verifiable storage and retrieval of key-value pairs.

    -   `trie.go`: This is the core of the MPT. It defines the `Trie` struct and implements fundamental operations like `Update` (for inserting or modifying key-value pairs) and `Hash` (for calculating the trie's root hash). It manages the overall structure and interactions between different node types. The `insert` method within this file handles the intricate logic of adding new data.
    -   `secure_trie.go`: Implements the `StateTrie` struct, which is a specialized version of the MPT. It wraps the basic `Trie` and ensures that all keys are hashed using Keccak256 before being used in the trie.
    -   `node.go`: Defines the fundamental building blocks of the MPT. It introduces the `node` interface and concrete types:
        -   `fullNode`: Represents a branch in the trie with 17 slots (16 for hexadecimal characters '0'-'f', and one for a value if a path terminates at this branch).
        -   `shortNode`: Represents either an extension node (sharing a common path prefix) or a leaf node (storing a value). Its `Key` field stores the path segment, and `Val` points to the next node or holds the actual value.
        -   `hashNode`: A placeholder for a node that has been hashed and whose full data is not currently in memory (typically loaded from a database when needed). It stores the 32-byte hash of the node.
        -   `valueNode`: Represents the actual data (value) stored at the end of a path in the trie.
        This file also includes logic for decoding nodes from their RLP (Recursive Length Prefix) representation (`decodeNode`, `decodeShort`, `decodeFull`).
    -   `node_enc.go`: Complements `node.go` by providing the RLP encoding logic for each node type (`fullNode.encode`, `shortNode.encode`, etc.). RLP is the standard serialization format used throughout Ethereum.
    -   `encoding.go`: Contains utility functions for converting between different key encodings used within the MPT.
    -   `hasher.go`: Manages the hashing of trie nodes. It uses a pool of `hasher` objects (which internally use `crypto.KeccakState`) to efficiently compute Keccak256 hashes of RLP-encoded nodes. Key functions include `hash` (which recursively hashes a node and its children), `shortnodeToHash`, and `fullnodeToHash`. It implements an optimization where nodes smaller than 32 bytes are not hashed but embedded directly in their parent.
    -   `proof.go`: Implements the logic for generating and verifying Merkle proofs.
    -   `print_helper.go`: Provides utility functions for visualizing and debugging the trie structure.
    -   `trie_id.go`: Defines an `ID` struct used to uniquely identify a trie, typically by its owner (e.g., a contract address) and its root hash.
    -   `trienode/` (sub-directory):
        -   `proof.go`: Defines `ProofSet`, a simple in-memory key-value store that implements the `ethdb.KeyValueWriter` and `ethdb.KeyValueReader` interfaces. This is used by the `Prove` and `VerifyProof` functions to temporarily store and retrieve the nodes that form a Merkle proof.

## Setup and Installation

1.  **Prerequisites**:
    *   Go (version 1.23 or later is recommended). You can download it from [golang.org](https://golang.org/dl/).

2.  **Clone the Repository (if applicable)**:
    ```bash
    git clone <repository-url> # Replace <repository-url> with the actual URL
    cd storage_extract
    ```
    If you already have the project files, navigate to the project root directory:
    ```bash
    cd .../storage_extract # Replace ... with your directory path
    ```

3.  **Download Dependencies**:
    This command will analyze your `go.mod` file and download any missing dependencies.
    ```bash
    go mod tidy
    ```
## Future Enhancements (TODOs)

-   **Complete Frontend for Proof Service**:
    -   Enhance the UI for displaying Merkle proofs in a more user-friendly format.
    -   Streamline the workflow for proof generation and verification directly from the interface.
-   **Richer MPT Information Display**:
    -   Provide more contextual details within the MPT visualization (e.g., RLP encoded node data, key before hash for each node).
-   **Persistent Storage**:
    -   Replace the current in-memory mock `CachingDB` with a persistent key-value store to allow state to persist across sessions.
-   **Account Trie Implementation**:
    -   The current version focuses on visualizing the **storage trie** for individual accounts. A key future step is to implement the **account trie** (also known as the state trie), which is managed by the `StateDB` and maps account addresses to their account states (nonce, balance, codeHash, storageRoot).
-   **Block-by-Block State Display**:
    -   In Ethereum, state transitions occur on a block-by-block basis. Future development aims to support viewing and proving account tries and storage tries as they existed at specific historical blocks.

