/**
 * Main Application - Entry point that coordinates all components
 */
document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const addressInput = document.getElementById('address-input');
    const createAccountBtn = document.getElementById('create-account-btn');
    const accountListElem = document.getElementById('account-list');
    const storageKeyInput = document.getElementById('storage-key');
    const storageValueInput = document.getElementById('storage-value');
    const addStorageBtn = document.getElementById('add-storage-btn');
    const storageList = document.getElementById('storage-list');
    const updateTrieBtn = document.getElementById('update-trie-btn');
    const rootHashElem = document.getElementById('root-hash');
    const textViewBtn = document.getElementById('text-view-btn');
    const treeViewBtn = document.getElementById('tree-view-btn');
    const textView = document.getElementById('text-view');
    const treeView = document.getElementById('tree-view');
    const trieText = document.getElementById('trie-text');
    const trieDiagram = document.getElementById('trie-diagram');
    const errorMessage = document.getElementById('error-message');
    const loadingMessage = document.getElementById('loading-message');

    // Storage value retrieval functionality
    const getValueKeyInput = document.getElementById('get-value-key');
    const getValueBtn = document.getElementById('get-value-btn');
    const currentValue = document.getElementById('current-value');
    const originalValue = document.getElementById('original-value');
    const proofGenerated = document.getElementById('proof-generated');

    // Proof functionality
    const proofKeyInput = document.getElementById('proof-key');
    const proofRootInput = document.getElementById('proof-root');
    const getProofBtn = document.getElementById('get-proof-btn');
    const useCurrentRootBtn = document.getElementById('use-current-root-btn');
    const proofRootHash = document.getElementById('proof-root-hash');
    const proofValue = document.getElementById('proof-value');

    // State
    let accounts = []; // List of all created/loaded accounts
    let selectedAccount = null; // Currently selected account
    let pendingStorage = {}; // { address: { key: value, ... } }
    let currentView = 'text';

    // Initialize TrieVisualizer early
    const trieVisualizer = new TrieVisualizer();

    // --- UI Helpers ---
    function setLoading(loading) {
        loadingMessage.style.display = loading ? '' : 'none';
    }
    function setError(msg) {
        if (msg) {
            errorMessage.textContent = msg;
            errorMessage.style.display = '';
        } else {
            errorMessage.textContent = '';
            errorMessage.style.display = 'none';
        }
    }
    function renderAccountList() {
        accountListElem.innerHTML = '';
        if (accounts.length === 0) {
            const li = document.createElement('li');
            li.textContent = 'No accounts yet.';
            li.className = 'empty-message';
            accountListElem.appendChild(li);
            return;
        }
        accounts.forEach(addr => {
            const li = document.createElement('li');
            li.textContent = addr;
            li.className = 'account-item' + (addr === selectedAccount ? ' selected' : '');
            li.title = 'Click to select this account';
            li.onclick = () => selectAccount(addr);
            accountListElem.appendChild(li);
        });
    }
    function setCurrentAccount(addr) {
        selectedAccount = addr;
        renderAccountList();
        renderStorageList();
        clearTrieVisualization();
        if (addr) {
            fetchAndShowTrie(addr);
        }
    }
    function renderStorageList() {
        storageList.innerHTML = '';
        if (!selectedAccount) {
            storageList.innerHTML = '<div class="empty-message">Select an account first.</div>';
            updateTrieBtn.disabled = true;
            return;
        }
        const items = pendingStorage[selectedAccount] || {};
        const keys = Object.keys(items);
        if (keys.length === 0) {
            storageList.innerHTML = '<div class="empty-message">No pending storage. Add key-value pairs.</div>';
            updateTrieBtn.disabled = true;
            return;
        }
        keys.forEach(key => {
            const value = items[key];
            const item = document.createElement('div');
            item.className = 'storage-item';
            const keyElem = document.createElement('div');
            keyElem.className = 'storage-key';
            keyElem.textContent = key;
            const valueElem = document.createElement('div');
            valueElem.className = 'storage-value';
            valueElem.textContent = value;
            const removeBtn = document.createElement('button');
            removeBtn.textContent = 'Remove';
            removeBtn.onclick = () => {
                delete pendingStorage[selectedAccount][key];
                renderStorageList();
            };
            item.appendChild(keyElem);
            item.appendChild(valueElem);
            item.appendChild(removeBtn);
            storageList.appendChild(item);
        });
        updateTrieBtn.disabled = false;
    }
    
    function switchView(view) {
        trieVisualizer.switchView(view);
    }
    function formatTrieTextHierarchy(textData) {
        // Try to pretty-print the hierarchy from textData (if it's JSON or a string tree)
        if (!textData) return 'No trie data available.';
        // If it's JSON, pretty print; if it's a string, try to preserve indentation
        try {
            const obj = typeof textData === 'string' ? JSON.parse(textData) : textData;
            return JSON.stringify(obj, null, 2);
        } catch {
            // fallback: preserve whitespace/indentation
            return textData.replace(/\n/g, '\n');
        }
    }
    function updateTrieVisualization(trie) {
        trieVisualizer.updateVisualization(trie);
    }
    function renderTreeDiagram(data) {
        trieDiagram.innerHTML = '';
        if (!data) {
            trieDiagram.textContent = 'No tree data available';
            return;
        }
        // --- New MPT visual layout ---
        function renderNode(node, isRoot = false) {
            if (node.type === 'branch') {
                // Branch node: horizontal row of 16 slots (0-f)
                const branchBox = document.createElement('div');
                branchBox.className = 'mpt-node mpt-branch';
                branchBox.style.display = 'flex';
                branchBox.style.flexDirection = 'column';
                branchBox.style.alignItems = 'center';
                branchBox.style.margin = '24px auto';
                // Root label
                if (isRoot) {
                    const rootLabel = document.createElement('div');
                    rootLabel.className = 'mpt-label';
                    rootLabel.innerHTML = `<b>root</b>`;
                    branchBox.appendChild(rootLabel);
                } else {
                    const label = document.createElement('div');
                    label.className = 'mpt-label';
                    label.innerHTML = `<b>BranchNode</b>`;
                    branchBox.appendChild(label);
                }
                // Hash
                if (node.hash) {
                    const hashDiv = document.createElement('div');
                    hashDiv.className = 'mpt-hash';
                    hashDiv.innerHTML = `<span class='mpt-label-hash'>Hash:</span> ${node.hash}`;
                    branchBox.appendChild(hashDiv);
                }
                // 0-f row
                const row = document.createElement('div');
                row.className = 'mpt-branch-row';
                row.style.display = 'flex';
                row.style.justifyContent = 'center';
                row.style.margin = '12px 0 0 0';
                for (let i = 0; i < 16; i++) {
                    const slot = document.createElement('div');
                    slot.className = 'mpt-branch-slot';
                    slot.textContent = i.toString(16);
                    slot.style.position = 'relative';
                    slot.style.width = '28px';
                    slot.style.height = '28px';
                    slot.style.display = 'flex';
                    slot.style.alignItems = 'center';
                    slot.style.justifyContent = 'center';
                    slot.style.margin = '0 2px';
                    slot.style.border = '1.5px solid #b2bec3';
                    slot.style.background = '#f8fafd';
                    slot.style.fontFamily = 'monospace';
                    slot.style.fontWeight = 'bold';
                    // If child exists, render connector and child node
                    if (node.children && node.children[i]) {
                        slot.style.background = '#e3fcec';
                        // Connector
                        const connector = document.createElement('div');
                        connector.className = 'mpt-branch-connector';
                        connector.style.position = 'absolute';
                        connector.style.left = '50%';
                        connector.style.top = '100%';
                        connector.style.width = '2px';
                        connector.style.height = '18px';
                        connector.style.background = '#27ae60';
                        connector.style.transform = 'translateX(-50%)';
                        slot.appendChild(connector);
                    }
                    row.appendChild(slot);
                }
                branchBox.appendChild(row);
                // Children row (below)
                const childrenRow = document.createElement('div');
                childrenRow.className = 'mpt-branch-children-row';
                childrenRow.style.display = 'flex';
                childrenRow.style.justifyContent = 'center';
                childrenRow.style.marginTop = '18px';
                for (let i = 0; i < 16; i++) {
                    if (node.children && node.children[i]) {
                        const childBox = renderNode(node.children[i], false);
                        childBox.style.margin = '0 2px';
                        childrenRow.appendChild(childBox);
                    } else {
                        // Empty slot for alignment
                        const empty = document.createElement('div');
                        empty.style.width = '28px';
                        empty.style.height = '1px';
                        empty.style.margin = '0 2px';
                        childrenRow.appendChild(empty);
                    }
                }
                branchBox.appendChild(childrenRow);
                return branchBox;
            } else if (node.type === 'leaf' || node.type === 'short') {
                // Leaf/short node: colored box with key/value/hash
                const leafBox = document.createElement('div');
                leafBox.className = 'mpt-node mpt-' + node.type;
                leafBox.style.display = 'flex';
                leafBox.style.flexDirection = 'column';
                leafBox.style.alignItems = 'center';
                leafBox.style.margin = '24px auto';
                const label = document.createElement('div');
                label.className = 'mpt-label';
                label.innerHTML = `<b>${node.type === 'leaf' ? 'LeafNode' : 'ShortNode'}</b>`;
                leafBox.appendChild(label);
                if (node.key) {
                    const keyDiv = document.createElement('div');
                    keyDiv.className = 'mpt-key';
                    keyDiv.innerHTML = `<span class='mpt-label-key'>Key:</span> ${node.key}`;
                    leafBox.appendChild(keyDiv);
                }
                if (node.value) {
                    const valueDiv = document.createElement('div');
                    valueDiv.className = 'mpt-value';
                    valueDiv.innerHTML = `<span class='mpt-label-value'>Value:</span> ${node.value}`;
                    leafBox.appendChild(valueDiv);
                }
                if (node.hash) {
                    const hashDiv = document.createElement('div');
                    hashDiv.className = 'mpt-hash';
                    hashDiv.innerHTML = `<span class='mpt-label-hash'>Hash:</span> ${node.hash}`;
                    leafBox.appendChild(hashDiv);
                }
                return leafBox;
            } else {
                // Unknown node type fallback
                const box = document.createElement('div');
                box.className = 'mpt-node';
                box.textContent = node.type || 'Unknown node';
                return box;
            }
        }
        // Center the root node
        const rootWrapper = document.createElement('div');
        rootWrapper.style.display = 'flex';
        rootWrapper.style.justifyContent = 'center';
        rootWrapper.appendChild(renderNode(data, true));
        trieDiagram.appendChild(rootWrapper);
    }
    function clearTrieVisualization() {
        trieVisualizer.updateVisualization({});
    }
    async function fetchAndShowTrie(addr) {
        try {
            setLoading(true);
            setError('');
            const data = await ApiClient.getAccount(addr);
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
                if (data.trie.rootHash) {
                    rootHashElem.textContent = data.trie.rootHash;
                }
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
            setLoading(false);
        } catch (e) {
            setError(e.message || 'Failed to fetch trie data');
            clearTrieVisualization();
            setLoading(false);
        }
    }

    // --- Event Handlers ---
    createAccountBtn.onclick = async () => {
        const addr = addressInput.value.trim();
        if (!/^0x[0-9a-fA-F]{1,40}$/.test(addr)) {
            setError('Please enter a valid Ethereum address (0x...)');
            return;
        }
        const paddedAddr = '0x' + addr.slice(2).padStart(40, '0');
        
        try {
            setLoading(true);
            setError('');
            await ApiClient.createAccount(paddedAddr);
            
            if (!accounts.includes(paddedAddr)) {
                accounts.push(paddedAddr);
            }
            if (!pendingStorage[paddedAddr]) {
                pendingStorage[paddedAddr] = {};
            }
            setCurrentAccount(paddedAddr);
            addressInput.value = '';
            setLoading(false);
        } catch (e) {
            setError(e.message || 'Failed to create account');
            setLoading(false);
        }
    };
    
    function selectAccount(addr) {
        setCurrentAccount(addr);
    }
    
    addStorageBtn.onclick = () => {
        if (!selectedAccount) {
            setError('Select an account first.');
            return;
        }
        const key = storageKeyInput.value.trim();
        const value = storageValueInput.value.trim();
        if (!/^0x[0-9a-fA-F]+$/.test(key) || !/^0x[0-9a-fA-F]+$/.test(value)) {
            setError('Both key and value must be valid hex strings (0x...)');
            return;
        }
        if (!pendingStorage[selectedAccount]) {
            pendingStorage[selectedAccount] = {};
        }
        pendingStorage[selectedAccount][key] = value;
        renderStorageList();
        storageKeyInput.value = '';
        storageValueInput.value = '';
        setError('');
    };
    
    updateTrieBtn.onclick = async () => {
        if (!selectedAccount) {
            setError('No account selected.');
            return;
        }
        const items = pendingStorage[selectedAccount] || {};
        if (Object.keys(items).length === 0) {
            setError('Please add some storage items.');
            return;
        }
        try {
            setLoading(true);
            setError('');
            // Call the consolidated storage update endpoint
            const data = await ApiClient.updateStorage(selectedAccount, items);
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
                if (data.trie.rootHash) {
                    rootHashElem.textContent = data.trie.rootHash;
                }
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
            // Clear pending storage for this account after successful update
            pendingStorage[selectedAccount] = {};
            renderStorageList();
            setLoading(false);
        } catch (e) {
            setError(e.message || 'Failed to update storage');
            setLoading(false);
        }
    };
    
    textViewBtn.onclick = () => switchView('text');
    treeViewBtn.onclick = () => switchView('tree');

    // Storage value retrieval functionality
    getValueBtn.addEventListener('click', async function() {
        if (!selectedAccount) {
            setError('Please select an account first');
            return;
        }
        const key = getValueKeyInput.value.trim();

        if (!key) {
            setError('Please enter a storage key');
            return;
        }

        if (!/^0x[0-9a-fA-F]+$/.test(key)) {
            setError('Key must be a valid hex string (0x...)');
            return;
        }

        try {
            setLoading(true);
            setError('');
            const result = await ApiClient.getStorageValue(selectedAccount, key);
            
            // Format the current value with 0x prefix if not zero
            const currentVal = result.value && result.value !== '0' ? '0x' + result.value : '0x0';
            currentValue.textContent = currentVal;
            
            // Show if original value matches (from the originalMatch boolean)
            originalValue.textContent = result.originalMatch ? 'Matches' : 'No match/Not set';
            
            // Proof is generated automatically during storage updates
            proofGenerated.textContent = 'Available';
            
            setLoading(false);
        } catch (error) {
            setError('Failed to get storage value: ' + error.message);
            currentValue.textContent = '-';
            originalValue.textContent = '-';
            proofGenerated.textContent = '-';
            setLoading(false);
        }
    });

    // Proof functionality
    getProofBtn.addEventListener('click', async function() {
        if (!selectedAccount) {
            setError('Please select an account first');
            return;
        }
        const key = proofKeyInput.value.trim();
        const rootHash = proofRootInput.value.trim();

        if (!key) {
            setError('Please enter a storage key');
            return;
        }

        if (!rootHash) {
            setError('Please enter a root hash');
            return;
        }

        if (!/^0x[0-9a-fA-F]+$/.test(key)) {
            setError('Key must be a valid hex string (0x...)');
            return;
        }

        if (!/^0x[0-9a-fA-F]{64}$/.test(rootHash)) {
            setError('Root hash must be a valid 64-character hex string (0x...)');
            return;
        }

        try {
            setLoading(true);
            setError('');
            
            // Use the user-provided root hash for proof verification
            const result = await ApiClient.getProof(selectedAccount, key, rootHash);
            
            // The backend returns just the value from proof verification
            const proofVal = result.value && result.value !== '' ? '0x' + result.value : 'Not found';
            proofValue.textContent = proofVal;
            
            // Display the root hash that was used for verification
            proofRootHash.textContent = rootHash;
            
            setLoading(false);
        } catch (error) {
            setError('Failed to get proof: ' + error.message);
            proofRootHash.textContent = '-';
            proofValue.textContent = '-';
            setLoading(false);
        }
    });

    // Use Current Root button functionality
    useCurrentRootBtn.addEventListener('click', async function() {
        if (!selectedAccount) {
            setError('Please select an account first');
            return;
        }

        try {
            setLoading(true);
            setError('');
            
            // Get the current account data to obtain the root hash
            const accountData = await ApiClient.getAccount(selectedAccount);
            
            if (accountData && accountData.trie && accountData.trie.rootHash) {
                // Fill the root input field with the current root hash
                proofRootInput.value = accountData.trie.rootHash;
                setError(''); // Clear any previous errors
            } else {
                setError('No root hash available. Please update storage first.');
            }
            
            setLoading(false);
        } catch (error) {
            setError('Failed to get current root hash: ' + error.message);
            setLoading(false);
        }
    });

    // --- Initial State ---
    renderAccountList();
    setCurrentAccount(null);
    renderStorageList();
    clearTrieVisualization();
    setError('');
    setLoading(false);
});