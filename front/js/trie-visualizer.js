/**
 * Trie Visualizer - Handles MPT visualization and display
 */
class TrieVisualizer {
    /**
     * Initialize the trie visualizer
     */
    constructor() {
        this.rootHashElement = document.getElementById('root-hash');
        this.textViewElement = document.getElementById('trie-text');
        this.treeViewElement = document.getElementById('trie-diagram');
        this.textViewBtn = document.getElementById('text-view-btn');
        this.treeViewBtn = document.getElementById('tree-view-btn');
        this.textView = document.getElementById('text-view');
        this.treeView = document.getElementById('tree-view');
        
        // Bind view switching events
        this.textViewBtn.addEventListener('click', () => this.switchView('text'));
        this.treeViewBtn.addEventListener('click', () => this.switchView('tree'));
    }
    
    /**
     * Switch between view modes
     * @param {string} viewType - The view type ('text' or 'tree')
     */
    switchView(viewType) {
        if (viewType === 'text') {
            this.textViewBtn.classList.add('active');
            this.treeViewBtn.classList.remove('active');
            this.textView.classList.add('active');
            this.treeView.classList.remove('active');
        } else {
            this.textViewBtn.classList.remove('active');
            this.treeViewBtn.classList.add('active');
            this.textView.classList.remove('active');
            this.treeView.classList.add('active');
        }
    }
    
    /**
     * Update the visualization
     * @param {Object} data - The server response data
     */
    updateVisualization(data) {
        // Update root hash
        this.rootHashElement.textContent = data.rootHash || '-';
        
        // Update text view
        if (data.textData) {
            this.textViewElement.textContent = data.textData;
            this.highlightTextSyntax();
        } else {
            this.textViewElement.textContent = 'No text data available.';
        }
        
        // Update tree view
        if (data.trieData) {
            try {
                const trieData = JSON.parse(data.trieData);
                this.renderTreeDiagram(trieData);
            } catch (error) {
                console.error('Error parsing trie data:', error);
                this.treeViewElement.innerHTML = '<div class="error">Error rendering tree diagram</div>';
            }
        } else {
            this.treeViewElement.innerHTML = '<div class="empty">No tree data available</div>';
        }
    }
    
    /**
     * Highlight text syntax
     */
    highlightTextSyntax() {
        const text = this.textViewElement.textContent;
        let html = text
            .replace(/Short\[([^\]]*)\]/g, '<span class="node-short">Short[$1]</span>')
            .replace(/Branch\[([^\]]*)\]/g, '<span class="node-branch">Branch[$1]</span>')
            .replace(/Hash: ([0-9a-f]+)/gi, 'Hash: <span class="node-hash">$1</span>')
            .replace(/Value: ([0-9a-f]+)/gi, 'Value: <span class="node-value">$1</span>')
            .replace(/Key:([0-9a-f]+)/gi, 'Key:<span class="node-value">$1</span>');
            
        this.textViewElement.innerHTML = html;
    }
    
    /**
     * Render the tree diagram
     * @param {Object} data - The tree data
     */
    renderTreeDiagram(data) {
        // This would typically use D3.js or another library to render the tree
        this.treeViewElement.innerHTML = '<div class="tree-placeholder">Tree diagram will display here</div>';
        
        // Actual implementation would use D3.js or similar to create a visual tree
        // This is just a placeholder
    }
    
    /**
     * Clear the visualization
     */
    clearVisualization() {
        this.rootHashElement.textContent = '-';
        this.textViewElement.textContent = 'No trie data available.';
        this.treeViewElement.innerHTML = '';
    }
}