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
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error creating account:', error);
            throw error;
        }
    }
    
    /**
     * Get account information and trie data
     * @param {string} address - The Ethereum address
     * @returns {Promise} The response promise
     */
    static async getAccount(address) {
        try {
            const response = await fetch('/api/account/get', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error getting account:', error);
            throw error;
        }
    }
    
    /**
     * Update storage with key-value pairs and get updated trie
     * @param {string} address - The Ethereum address
     * @param {Object} storage - The key-value pairs object
     * @returns {Promise} The response promise
     */
    static async updateStorage(address, storage) {
        try {
            const response = await fetch('/api/storage/update', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, storage })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error updating storage:', error);
            throw error;
        }
    }
    
    /**
     * Get a specific storage value
     * @param {string} address - The Ethereum address
     * @param {string} key - The storage key (hex)
     * @returns {Promise} The response promise
     */
    static async getStorageValue(address, key) {
        try {
            const response = await fetch('/api/storage/get', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, key })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error getting storage value:', error);
            throw error;
        }
    }
    
    /**
     * Get a Merkle proof for a key in the trie
     * @param {string} address - The Ethereum address
     * @param {string} key - The storage key (hex)
     * @param {string} root - The root hash for proof verification
     * @returns {Promise} The response promise
     */
    static async getProof(address, key, root) {
        try {
            const response = await fetch('/api/proof', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ address, key, root })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error getting proof:', error);
            throw error;
        }
    }
}