/**
 * Account Manager - Handles Ethereum account operations
 */
class AccountManager {
    /**
     * Initialize the account manager
     */
    constructor() {
        this.currentAccount = null;
        this.addressInput = document.getElementById('address-input');
        this.createButton = document.getElementById('create-account-btn');
        this.accountDisplay = document.getElementById('current-account');
        
        this.createButton.addEventListener('click', () => this.createAccount());
    }
    
    /**
     * Create or load an Ethereum account
     * @returns {Promise<boolean>} Success indicator
     */
    async createAccount() {
        let address = this.addressInput.value.trim();
        
        // Format the address - if it's a partial address, pad it with zeros
        if (address.startsWith('0x')) {
            const hexPart = address.slice(2);
            if (hexPart.length > 0 && hexPart.length < 40 && /^[0-9a-f]+$/i.test(hexPart)) {
                address = '0x' + hexPart.padStart(40, '0');
                this.addressInput.value = address;
            }
        }
        
        if (!this.isValidAddress(address)) {
            alert('Please enter a valid Ethereum address (0x followed by 40 hex characters)');
            return false;
        }
        
        try {
            await ApiClient.createAccount(address);
            this.currentAccount = address;
            this.updateAccountDisplay();
            return true;
        } catch (error) {
            console.error('Error creating account:', error);
            alert('Failed to create account: ' + error.message);
            return false;
        }
    }
    
    /**
     * Update the account display in the UI
     */
    updateAccountDisplay() {
        if (this.currentAccount) {
            this.accountDisplay.textContent = this.currentAccount;
            this.accountDisplay.classList.add('highlight');
            setTimeout(() => {
                this.accountDisplay.classList.remove('highlight');
            }, 2000);
        } else {
            this.accountDisplay.textContent = 'No account selected';
        }
    }
    
    /**
     * Get the current account address
     * @returns {string|null} The current account address
     */
    getCurrentAccount() {
        return this.currentAccount;
    }
    
    /**
     * Validate an Ethereum address
     * @param {string} address - The address to validate
     * @returns {boolean} Whether the address is valid
     */
    isValidAddress(address) {
        return /^0x[0-9a-fA-F]{40}$/.test(address);
    }
}