/**
 * Storage Manager - Handles storage key-value operations
 */
class StorageManager {
    /**
     * Initialize the storage manager
     */
    constructor() {
        this.storageItems = {};
        this.storageList = document.getElementById('storage-list');
        this.keyInput = document.getElementById('storage-key');
        this.valueInput = document.getElementById('storage-value');
        this.addButton = document.getElementById('add-storage-btn');
        this.updateButton = document.getElementById('update-trie-btn');
        
        // Bind events
        this.addButton.addEventListener('click', () => this.addStorageItem());
        
        // Initial state
        this.updateButton.disabled = true;
    }
    
    /**
     * Add a storage item
     */
    addStorageItem() {
        const key = this.keyInput.value.trim();
        const value = this.valueInput.value.trim();
        
        if (!this.isValidHex(key) || !this.isValidHex(value)) {
            alert('Both key and value must be valid hexadecimal strings (starting with 0x)');
            return;
        }
        
        // Add to storage collection
        this.storageItems[key] = value;
        
        // Update UI
        this.renderStorageItems();
        this.updateButton.disabled = false;
        
        // Clear inputs
        this.keyInput.value = '';
        this.valueInput.value = '';
    }
    
    /**
     * Remove a storage item
     * @param {string} key - The key to remove
     */
    removeStorageItem(key) {
        delete this.storageItems[key];
        this.renderStorageItems();
        
        // If no storage items remain, disable the update button
        if (Object.keys(this.storageItems).length === 0) {
            this.updateButton.disabled = true;
        }
    }
    
    /**
     * Render the storage items list
     */
    renderStorageItems() {
        this.storageList.innerHTML = '';
        
        const keys = Object.keys(this.storageItems);
        
        if (keys.length === 0) {
            const emptyMsg = document.createElement('div');
            emptyMsg.textContent = 'No storage items. Add some key-value pairs to create storage.';
            emptyMsg.className = 'empty-message';
            this.storageList.appendChild(emptyMsg);
            return;
        }
        
        for (const key of keys) {
            const value = this.storageItems[key];
            
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
            removeBtn.addEventListener('click', () => this.removeStorageItem(key));
            
            item.appendChild(keyElem);
            item.appendChild(valueElem);
            item.appendChild(removeBtn);
            
            this.storageList.appendChild(item);
        }
    }
    
    /**
     * Get all storage items
     * @returns {Object} The storage items object
     */
    getStorageItems() {
        return this.storageItems;
    }
    
    /**
     * Clear all storage items
     */
    clearStorageItems() {
        this.storageItems = {};
        this.renderStorageItems();
        this.updateButton.disabled = true;
    }
    
    /**
     * Validate a hexadecimal string
     * @param {string} hex - The string to validate
     * @returns {boolean} Whether the string is a valid hex string
     */
    isValidHex(hex) {
        return /^0x[0-9a-f]+$/i.test(hex);
    }
}