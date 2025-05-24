/**
 * Main Application - Entry point that coordinates all components
 */
document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const addressInput = document.getElementById('address-input');
    const createAccountBtn = document.getElementById('create-account-btn');
    const currentAccountDisplay = document.getElementById('current-account');
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

    // State
    let currentAccount = null;
    let storageItems = {};
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
    function setCurrentAccount(addr) {
        currentAccount = addr;
        currentAccountDisplay.textContent = addr ? addr : 'No account selected';
    }
    function clearStorageItems() {
        storageItems = {};
        renderStorageList();
        updateTrieBtn.disabled = true;
    }
    function renderStorageList() {
        storageList.innerHTML = '';
        const keys = Object.keys(storageItems);
        if (keys.length === 0) {
            const emptyMsg = document.createElement('div');
            emptyMsg.textContent = 'No storage items. Add some key-value pairs.';
            emptyMsg.className = 'empty-message';
            storageList.appendChild(emptyMsg);
            return;
        }
        for (const key of keys) {
            const value = storageItems[key];
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
                delete storageItems[key];
                renderStorageList();
                if (Object.keys(storageItems).length === 0) updateTrieBtn.disabled = true;
            };
            item.appendChild(keyElem);
            item.appendChild(valueElem);
            item.appendChild(removeBtn);
            storageList.appendChild(item);
        }
    }
    function switchView(view) {
        currentView = view;
        if (view === 'text') {
            textViewBtn.classList.add('active');
            treeViewBtn.classList.remove('active');
            textView.classList.add('active');
            treeView.classList.remove('active');
        } else {
            textViewBtn.classList.remove('active');
            treeViewBtn.classList.add('active');
            textView.classList.remove('active');
            treeView.classList.add('active');
        }
    }
    function updateTrieVisualization(trie) {
        if (!trie) {
            setError('No trie data in response.');
            rootHashElem.textContent = '-';
            trieText.textContent = 'No trie data available.';
            trieDiagram.innerHTML = '';
            return;
        }
        rootHashElem.textContent = trie.rootHash || '-';
        if (trie.textData) {
            trieText.textContent = trie.textData;
        } else {
            trieText.textContent = 'No trie data available.';
        }
        if (trie.trieData) {
            try {
                const trieData = typeof trie.trieData === 'string' ? JSON.parse(trie.trieData) : trie.trieData;
                renderTreeDiagram(trieData);
            } catch (e) {
                trieDiagram.innerHTML = '<div class="error">Error rendering tree diagram</div>';
            }
        } else {
            trieDiagram.innerHTML = '<div class="empty">No tree data available</div>';
        }
    }
    function renderTreeDiagram(data) {
        trieDiagram.innerHTML = '';
        if (!data) {
            trieDiagram.textContent = 'No tree data available';
            return;
        }
        function renderNode(node, depth = 0) {
            const div = document.createElement('div');
            div.style.marginLeft = (depth * 20) + 'px';
            div.textContent = `[${node.type}]` + (node.key ? ` Key:${node.key}` : '') + (node.value ? ` Value:${node.value}` : '') + (node.hash ? ` Hash:${node.hash}` : '');
            trieDiagram.appendChild(div);
            if (node.children) {
                for (const child of node.children) renderNode(child, depth + 1);
            }
        }
        renderNode(data, 0);
    }
    function clearTrieVisualization() {
        rootHashElem.textContent = '-';
        trieText.textContent = 'No trie data available.';
        trieDiagram.innerHTML = '';
    }

    // --- API Helpers ---
    async function apiCall(url, body) {
        setLoading(true);
        setError('');
        try {
            console.log('API CALL', url, body);
            const res = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            const data = await res.json();
            setLoading(false);
            console.log('API RESPONSE', url, data);
            if (!res.ok || data.error) throw new Error(data.error || 'API error');
            return data;
        } catch (e) {
            setLoading(false);
            setError(e.message || 'API error');
            console.error('API ERROR', url, e);
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
        setCurrentAccount(paddedAddr);
        try {
            console.log('Creating/loading account:', paddedAddr);
            const data = await apiCall('/api/account/create', { address: paddedAddr });
            clearStorageItems();
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
        } catch (e) {
            // Show error, but keep the selected account in the UI
            // clearTrieVisualization();
        }
    };

    addStorageBtn.onclick = () => {
        const key = storageKeyInput.value.trim();
        const value = storageValueInput.value.trim();
        if (!/^0x[0-9a-fA-F]+$/.test(key) || !/^0x[0-9a-fA-F]+$/.test(value)) {
            setError('Both key and value must be valid hex strings (0x...)');
            return;
        }
        storageItems[key] = value;
        renderStorageList();
        updateTrieBtn.disabled = false;
        storageKeyInput.value = '';
        storageValueInput.value = '';
    };

    updateTrieBtn.onclick = async () => {
        if (!currentAccount) {
            setError('No account selected.');
            return;
        }
        if (Object.keys(storageItems).length === 0) {
            setError('Please add some storage items.');
            return;
        }
        try {
            console.log('Updating trie for account:', currentAccount, storageItems);
            const data = await apiCall('/api/storage/batch', { address: currentAccount, storage: storageItems });
            if (data && data.trie) {
                updateTrieVisualization(data.trie);
            } else {
                setError('No trie data in response.');
                clearTrieVisualization();
            }
            clearStorageItems();
        } catch (e) {
            // error already shown
        }
    };

    textViewBtn.onclick = () => switchView('text');
    treeViewBtn.onclick = () => switchView('tree');

    // --- Initial State ---
    setCurrentAccount(null);
    clearStorageItems();
    clearTrieVisualization();
    setError('');
    setLoading(false);
});