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

    // Proof functionality
    const proofKeyInput = document.getElementById('proof-key');
    const getProofBtn = document.getElementById('get-proof-btn');
    const proofRootHash = document.getElementById('proof-root-hash');
    const proofValue = document.getElementById('proof-value');

    // State
    let accounts = []; // List of all created/loaded accounts
    let selectedAccount = null; // Currently selected account
    let pendingStorage = {}; // { address: { key: value, ... } }
    let currentView = 'text';

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
            const data = await apiCall('/api/account/get', { address: addr });
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
        } catch (e) {
            // error already shown
            clearTrieVisualization();
        }
    }

    // --- API Helpers ---
    async function apiCall(url, body) {
        setLoading(true);
        setError('');
        try {
            const res = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            const data = await res.json();
            setLoading(false);
            if (!res.ok || data.error) throw new Error(data.error || 'API error');
            return data;
        } catch (e) {
            setLoading(false);
            setError(e.message || 'API error');
            throw e;
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
        if (!accounts.includes(paddedAddr)) {
            accounts.push(paddedAddr);
        }
        if (!pendingStorage[paddedAddr]) {
            pendingStorage[paddedAddr] = {};
        }
        setCurrentAccount(paddedAddr);
        addressInput.value = '';
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
            // First, send the batch to update the pending state
            await apiCall('/api/storage/batch', { address: selectedAccount, storage: items });
            // Debug log
            console.log("Calling /api/trie/update");
            // Then, commit the trie and get the updated trie data
            const data = await apiCall('/api/trie/update', { address: selectedAccount });
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
            // Clear pending storage for this account after commit
            pendingStorage[selectedAccount] = {};
            renderStorageList();
        } catch (e) {
            // error already shown
        }
    };
    textViewBtn.onclick = () => switchView('text');
    treeViewBtn.onclick = () => switchView('tree');

    // --- TrieVisualizer integration ---
    const trieVisualizer = new TrieVisualizer();

    // --- Initial State ---
    renderAccountList();
    setCurrentAccount(null);
    renderStorageList();
    clearTrieVisualization();
    setError('');
    setLoading(false);

    // Proof functionality
    getProofBtn.addEventListener('click', async function() {
        const address = document.getElementById('address-input').value;
        const key = proofKeyInput.value;

        if (!address || !key) {
            showError('Please enter both address and key');
            return;
        }

        try {
            showLoading();
            const result = await ApiClient.getProof(address, key);
            proofRootHash.textContent = result.rootHash;
            proofValue.textContent = result.value || 'Not found';
            hideLoading();
        } catch (error) {
            showError('Failed to get proof: ' + error.message);
            hideLoading();
        }
    });
});