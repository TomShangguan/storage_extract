/**
 * API Client - Handles all communication with the backend API
 */
class ApiClient {
    /**
     * Create a new account
     * @param {string} address - The Ethereum address
     * @returns {Promise} The response promise
     */
    static async createAccount(address) {
        try {
            const response = await fetch('/api/account/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error creating account:', error);
            throw error;
        }
    }
    
    /**
     * Set a storage key-value pair
     * @param {string} address - The Ethereum address
     * @param {string} key - The storage key (hex)
     * @param {string} value - The storage value (hex)
     * @returns {Promise} The response promise
     */
    static async setStorage(address, key, value) {
        try {
            const response = await fetch('/api/storage/set', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, key, value })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error setting storage:', error);
            throw error;
        }
    }
    
    /**
     * Set multiple storage key-value pairs
     * @param {string} address - The Ethereum address
     * @param {Object} storage - The key-value pairs object
     * @returns {Promise} The response promise
     */
    static async setBatchStorage(address, storage) {
        try {
            const response = await fetch('/api/storage/batch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, storage })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error setting batch storage:', error);
            throw error;
        }
    }
    
    /**
     * Update the trie and get visualization data
     * @param {string} address - The Ethereum address
     * @returns {Promise} The response promise
     */
    static async updateTrie(address) {
        try {
            const response = await fetch('/api/trie/update', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error updating trie:', error);
            throw error;
        }
    }
    
    /**
     * Get a Merkle proof for a key in the trie
     * @param {string} address - The Ethereum address
     * @param {string} key - The storage key (hex)
     * @returns {Promise} The response promise
     */
    static async getProof(address, key) {
        try {
            const response = await fetch('/api/proof', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, key })
            });
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Error getting proof:', error);
            throw error;
        }
    }
}