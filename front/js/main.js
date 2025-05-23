/**
 * Main Application - Entry point that coordinates all components
 */
document.addEventListener('DOMContentLoaded', function() {
    // Initialize components
    const storageManager = new StorageManager();
    const trieVisualizer = new TrieVisualizer();
    const accountManager = new AccountManager();
    
    // Get DOM elements
    const updateTrieBtn = document.getElementById('update-trie-btn');
    
    // Bind events
    updateTrieBtn.addEventListener('click', updateTrie);
    
    /**
     * Update the trie
     */
    async function updateTrie() {
        const currentAddress = accountManager.getCurrentAccount();
        if (!currentAddress) {
            alert('Please create an account first');
            return;
        }
        
        const storageItems = storageManager.getStorageItems();
        
        if (Object.keys(storageItems).length === 0) {
            alert('Please add some storage items');
            return;
        }
        
        try {
            // First, set batch storage
            await ApiClient.setBatchStorage(currentAddress, storageItems);
            
            // Then update the trie
            const data = await ApiClient.updateTrie(currentAddress);
            
            // Update visualization
            trieVisualizer.updateVisualization(data);
            
            // Clear storage items
            storageManager.clearStorageItems();
            
        } catch (error) {
            console.error('Error updating trie:', error);
            alert('Failed to update trie: ' + error.message);
        }
    }
});