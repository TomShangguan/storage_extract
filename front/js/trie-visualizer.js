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
        this.treeViewElement = document.getElementById('tree-view'); // This is the outer container for tree view
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
        
        // Update text view - now receives formatted text from backend (without original KV pairs section)
        let textContent = '';
        
        if (data.textData) {
            textContent += data.textData;
        } else {
            textContent += 'No trie structure available.';
        }
        
        this.textViewElement.textContent = textContent;
        
        // Update tree view with enhanced information
        if (data.trieData) {
            try {
                const trieJson = JSON.parse(data.trieData);
                
                // Debug: log the data structure
                console.log('Trie JSON structure:', trieJson);
                console.log('Original KV pairs:', data.originalKVPairs);
                
                // Pass original key-value pairs to tree rendering for enhanced display
                this.renderTreeDiagramBoxed(trieJson, data.originalKVPairs); 
            } catch (e) {
                console.error("Failed to parse trieData JSON:", e);
                const diagramElement = document.getElementById('trie-diagram');
                if (diagramElement) {
                    diagramElement.textContent = 'Error displaying trie: Invalid data format.';
                } else {
                    this.treeViewElement.textContent = 'Error displaying trie: Invalid data format (diagram element missing).';
                }
            }
        } else {
            this.clearTreeDiagram();
        }
    }
    
    /**
     * Pretty print the trie as a human-readable tree (like CLI, not JSON)
     * @param {Object} node - The trie node
     * @param {number} indent - The indentation level
     * @param {string} branchIdx - The branch index
     * @returns {string} - The formatted string
     */
    prettyPrintTrieAscii(node, indent = 0, branchIdx = null) {
        let pad = ' '.repeat(indent);
        let out = '';
        if (node.type === 'branch') {
            out += `${pad}BranchNode (hash: ${node.hash || '-'}):\n`;
            if (node.children) {
                for (let i = 0; i < 16; i++) {
                    const child = node.children.find(c => c && c.key === i.toString(16));
                    if (child) {
                        out += `${pad}  ${i.toString(16)}: \n`;
                        out += this.prettyPrintTrieAscii(child, indent + 4, i);
                    } else {
                        out += `${pad}  ${i.toString(16)}: <nil>\n`;
                    }
                }
            }
        } else if (node.type === 'short') {
            out += `${pad}ShortNode\n`;
            if (node.key) out += `${pad}  Key: ${node.key}\n`;
            if (node.originalKey) out += `${pad}  Original Key: ${node.originalKey}\n`;
            if (node.value) out += `${pad}  Value: ${node.value}\n`;
            if (node.hash) out += `${pad}  Hash: ${node.hash}\n`;
            if (node.children && node.children.length > 0) {
                for (const child of node.children) {
                    out += this.prettyPrintTrieAscii(child, indent + 4);
                }
            }
        } else if (node.type === 'shortNode_value' || node.type === 'value') {
            out += `${pad}Value: ${node.value || ''}\n`;
        }
        return out;
    }
    
    /**
     * Render the tree diagram visually in a hierarchical structure similar to the text view
     * but with visual enhancements
     * @param {Object} rootNodeData - The root node data to render.
     * @param {Array} originalKVPairs - Original key-value pairs before hashing.
     */
    renderTreeDiagramBoxed(rootNodeData, originalKVPairs = []) {
        // Store original key-value pairs for use in rendering
        this.currentOriginalKVPairs = originalKVPairs || [];
        
        // Debug log to verify data is being passed
        if (this.currentOriginalKVPairs.length > 0) {
            console.log('Tree view: Enhanced with', this.currentOriginalKVPairs.length, 'original key-value pairs');
        }
        
        const diagramElement = document.getElementById('trie-diagram');
        if (!diagramElement) {
            console.error("#trie-diagram element not found for rendering.");
            // Fallback to outer container if #trie-diagram is missing, though it shouldn't be.
            this.treeViewElement.textContent = 'Error: Trie diagram container is missing.';
            return;
        }
        diagramElement.innerHTML = ''; // Clear previous content of the specific diagram div

        if (!rootNodeData) {
            diagramElement.textContent = 'No trie data to display.';
            return;
        }

        const treeContainer = document.createElement('div');
        treeContainer.className = 'mpt-tree-container';
        treeContainer.style.fontFamily = 'monospace';
        treeContainer.style.padding = '20px';
        treeContainer.style.overflowX = 'auto';
        
        // Start the recursive rendering from the root node
        this.renderNodeRecursive(rootNodeData, treeContainer, 0, true);
        diagramElement.appendChild(treeContainer);
    }

    /**
     * Recursively renders a node and its children for the tree diagram.
     * @param {Object} node - The current node data.
     * @param {HTMLElement} parentElement - The parent DOM element to append to.
     * @param {number} level - The current depth in the tree.
     * @param {boolean} isLastChildInLevel - Flag if this is the last child in its current level.
     */
    renderNodeRecursive(node, parentElement, level, isLastChildInLevel) {
        if (!node) return;

        const nodeElementWrapper = document.createElement('div');
        nodeElementWrapper.className = 'mpt-node-wrapper';
        nodeElementWrapper.style.paddingLeft = (level * 25) + 'px'; 
        nodeElementWrapper.style.marginTop = level === 0 ? '10px' : '5px';

        // Apply different styling based on node type - match our text visualization
        let nodeTypeForStyle = node.type || 'unknown';
        if (node.type === 'root_short' || node.type === 'root_branch') {
            nodeTypeForStyle = 'root';
        } else {
            nodeTypeForStyle = node.type;
        }
        
        const nodeBox = this.createStyledNode(nodeTypeForStyle, ''); 
        nodeBox.innerHTML = ''; // Clear any default content

        // 1. Title with clear node type identification
        const titleDiv = document.createElement('div');
        titleDiv.className = 'mpt-label';
        
        let nodeTypeDisplay = '';
        if (node.type === 'root_short' || node.type === 'root_branch') {
            nodeTypeDisplay = 'ROOT ';
        }
        
        if (node.type === 'shortNode_value' || (node.isLeaf && (node.type === 'short' || node.type === 'root_short'))) {
            // A short node with a valueNode
            nodeTypeDisplay += 'SHORT NODE';
        } else if (node.type === 'shortNode_extension' || 
                  (node.type === 'short' || node.type === 'root_short')) {
            // A short node with another node
            nodeTypeDisplay += 'SHORT NODE';
        } else if (node.type === 'branch' || node.type === 'root_branch') {
            nodeTypeDisplay += 'BRANCH NODE';
        } else {
            nodeTypeDisplay = node.type.toUpperCase() + ' NODE';
        }
        
        // Add slot info for branch children
        if (node.branchIndex !== undefined && node.branchIndex !== -1) {
            const slotHex = node.branchIndex.toString(16).toUpperCase();
            nodeTypeDisplay += ` (Slot ${slotHex})`;
        }
        
        titleDiv.textContent = nodeTypeDisplay;
        nodeBox.appendChild(titleDiv);

        // 2. Always show key segment (prefix) for short nodes - matches text view
        if (node.key && (node.type === 'short' || node.type === 'root_short' || node.type === 'shortNode_value' || node.isLeaf)) {
            this.addPropertyToNode(nodeBox, 'Key', node.key, '#6a0dad');
        }
        
        // 3. Original key (before hashing) if available - use standard styling and trim leading zeros
        // Note: Backend uses OriginalKey (capitalized) - check both cases for compatibility
        if (node.OriginalKey || node.originalKey) {
            const originalKey = node.OriginalKey || node.originalKey;
            const trimmedOriginalKey = this.trimLeadingZeros(originalKey);
            this.addPropertyToNode(nodeBox, 'Original Key', trimmedOriginalKey, '#2e7d32');
        }
        
        // 4. Value for short nodes with value - enhanced with original value
        if (node.isLeaf || node.type === 'shortNode_value') {
            if (node.value !== undefined && node.value !== null) {
                this.addPropertyToNode(nodeBox, 'Value', node.value, '#1976d2');
                
                // Check for original value first from backend data (preferred method)
                // Note: Backend uses OriginalValue (capitalized) - check both cases for compatibility
                if (node.OriginalValue || node.originalValue) {
                    const originalValue = node.OriginalValue || node.originalValue;
                    const trimmedOriginalValue = this.trimLeadingZeros(originalValue);
                    console.log('Found original value directly from backend:', originalValue);
                    this.addPropertyToNode(nodeBox, 'Original Value', trimmedOriginalValue, '#1565c0');
                }
                // Fallback: try to match using originalKVPairs if backend didn't provide it
                else if (this.currentOriginalKVPairs && this.currentOriginalKVPairs.length > 0) {
                    // Debug logging
                    console.log('Backend originalValue not found, trying manual matching for leaf node:', {
                        nodeKey: node.key,
                        nodeValue: node.value,
                        nodeOriginalKey: node.originalKey,
                        availableKVPairs: this.currentOriginalKVPairs
                    });
                    
                    // Try multiple matching strategies
                    let matchingKV = null;
                    
                    // Strategy 1: Match by original key
                    if (node.originalKey) {
                        matchingKV = this.currentOriginalKVPairs.find(kv => 
                            kv.originalKey === node.originalKey
                        );
                    }
                    
                    // Strategy 2: Match by key hex (with or without 0x prefix)
                    if (!matchingKV && node.key) {
                        const nodeKeyNormalized = node.key.replace(/^0x/, '');
                        matchingKV = this.currentOriginalKVPairs.find(kv => {
                            const kvKeyNormalized = kv.keyHex ? kv.keyHex.replace(/^0x/, '') : '';
                            return kvKeyNormalized === nodeKeyNormalized;
                        });
                    }
                    
                    // Strategy 3: Match by value hex if available
                    if (!matchingKV && node.value) {
                        const nodeValueNormalized = node.value.replace(/^0x/, '');
                        matchingKV = this.currentOriginalKVPairs.find(kv => {
                            const kvValueNormalized = kv.valueHex ? kv.valueHex.replace(/^0x/, '') : '';
                            return kvValueNormalized === nodeValueNormalized;
                        });
                    }
                    
                    console.log('Manual matching result:', matchingKV);
                    
                    if (matchingKV && matchingKV.originalValue) {
                        const trimmedOriginalValue = this.trimLeadingZeros(matchingKV.originalValue);
                        console.log('Found matching KV pair via manual matching, adding original value:', matchingKV.originalValue);
                        this.addPropertyToNode(nodeBox, 'Original Value', trimmedOriginalValue, '#1565c0');
                    }
                }
            }
        }
        
        // 5. Full path - only show if different from just the key or if it's a branch
        if (node.keyPath && (node.type === 'branch' || node.type === 'root_branch' || 
            (node.key && node.keyPath !== node.key))) {
            const pathLabel = (node.type === 'branch' || node.type === 'root_branch') ? 
                'Path Prefix' : 'Full Path';
            this.addPropertyToNode(nodeBox, pathLabel, node.keyPath, '#8e44ad');
        }

        // 6. Hash (only for hash nodes or root node)
        if (node.type === 'hash' || (level === 0 && node.hash)) {
            const hashLabel = level === 0 ? 'Root Hash' : 'Node Hash';
            this.addPropertyToNode(nodeBox, hashLabel, node.hash, '#c0392b');
        }
        
        // 7. For branch nodes, show slot information
        if ((node.type === 'branch' || node.type === 'root_branch') && 
            node.filledSlotCount !== undefined) {
            this.addPropertyToNode(
                nodeBox, 
                'Slots Filled', 
                `${node.filledSlotCount} out of ${node.totalSlotCount || 16}`, 
                '#e67e22'
            );
            
            // Add visual representation of slots
            if (node.slotMap) {
                const slotsContainer = document.createElement('div');
                slotsContainer.className = 'mpt-branch-slots';
                slotsContainer.style.display = 'flex';
                slotsContainer.style.flexWrap = 'wrap';
                slotsContainer.style.gap = '3px';
                slotsContainer.style.marginTop = '8px';
                
                // Create slots grid (4x4 for hex digits)
                for (let i = 0; i < 16; i++) {
                    const slot = i.toString(16);
                    const slotElement = document.createElement('div');
                    slotElement.className = 'mpt-slot';
                    slotElement.textContent = slot.toUpperCase();
                    
                    // Style based on whether the slot is filled
                    const isFilled = node.slotMap[slot] === true;
                    slotElement.className += isFilled ? ' filled' : ' empty';
                    
                    slotElement.style.width = '24px';
                    slotElement.style.height = '24px';
                    slotElement.style.display = 'flex';
                    slotElement.style.alignItems = 'center';
                    slotElement.style.justifyContent = 'center';
                    slotElement.style.borderRadius = '4px';
                    slotElement.style.fontSize = '12px';
                    slotElement.style.fontWeight = 'bold';
                    
                    if (isFilled) {
                        slotElement.style.backgroundColor = '#27ae60';
                        slotElement.style.color = 'white';
                        slotElement.style.border = '1px solid #27ae60';
                    } else {
                        slotElement.style.backgroundColor = '#f8f9fa';
                        slotElement.style.color = '#95a5a6';
                        slotElement.style.border = '1px solid #e9ecef';
                    }
                    
                    slotsContainer.appendChild(slotElement);
                }
                
                nodeBox.appendChild(slotsContainer);
            }
        }
        
        nodeElementWrapper.appendChild(nodeBox);
        parentElement.appendChild(nodeElementWrapper);

        // Render children with better structure
        if (node.children && node.children.length > 0) {
            const childrenContainer = document.createElement('div');
            childrenContainer.className = 'mpt-children-container';
            childrenContainer.style.position = 'relative';
            
            // Add a vertical line to connect children
            if (node.children.length > 1) {
                const connectorLine = document.createElement('div');
                connectorLine.className = 'mpt-connector-line';
                connectorLine.style.position = 'absolute';
                connectorLine.style.left = '12px';
                connectorLine.style.top = '0';
                connectorLine.style.bottom = '0';
                connectorLine.style.width = '2px';
                connectorLine.style.backgroundColor = '#e0e0e0';
                childrenContainer.appendChild(connectorLine);
            }
            
            nodeElementWrapper.appendChild(childrenContainer);

            if (node.type === 'branch' || node.type === 'root_branch') {
                // Sort children by branch index for consistent display
                const sortedChildren = [...node.children].sort((a, b) => 
                    (a.branchIndex ?? 0) - (b.branchIndex ?? 0));
                    
                sortedChildren.forEach((childNode, index) => {
                    this.renderNodeRecursive(
                        childNode, 
                        childrenContainer, 
                        level + 1, 
                        index === node.children.length - 1
                    );
                });
            } else if ((node.type === 'short' || node.type === 'root_short') && !node.isLeaf) {
                // Extension node
                this.renderNodeRecursive(node.children[0], childrenContainer, level + 1, true);
            }
        }
    }
    
    /**
     * Clear the visualization
     */
    clearVisualization() {
        this.rootHashElement.textContent = '-';
        this.textViewElement.textContent = 'No trie data available.';
        this.clearTreeDiagram();
    }

    /**
     * Clears only the tree diagram part of the visualization.
     */
    clearTreeDiagram() {
        const diagramElement = document.getElementById('trie-diagram');
        if (diagramElement) {
            diagramElement.innerHTML = 'No trie data to display.';
        } else {
            // Fallback if #trie-diagram somehow isn't there, clear the parent.
            this.treeViewElement.innerHTML = '<div id="trie-diagram">No trie data to display.</div>';
        }
    }
    
    /**
     * Helper method to create styled node elements
     * @param {string} type - Node type (branch, leaf, etc.)
     * @param {string} title - Node title
     * @param {Object} styles - Additional styles to apply
     * @returns {HTMLElement} The styled node element
     */
    createStyledNode(type, title, styles = {}) {
        const nodeBox = document.createElement('div');
        nodeBox.className = `mpt-node mpt-${type}`;
        nodeBox.style.display = 'inline-block';
        nodeBox.style.borderRadius = '5px';
        nodeBox.style.padding = '8px 12px';
        nodeBox.style.marginBottom = '2px';
        nodeBox.style.maxWidth = '100%';
        nodeBox.style.boxShadow = '0 1px 4px rgba(0,0,0,0.1)';
        
        // Apply type-specific styles
        if (type === 'branch' || type === 'root_branch') {
            nodeBox.style.background = '#e3fcec'; // Solid color instead of gradient
            nodeBox.style.borderLeft = '4px solid #27ae60';
        } else if (type === 'shortNode_value') {
            // ShortNode with valueNode value
            nodeBox.style.background = '#e3f2fd'; // Solid color
            nodeBox.style.borderLeft = '4px solid #2196f3';
        } else if (type === 'shortNode_extension' || type === 'short') {
            // ShortNode with another node as value (extension)
            nodeBox.style.background = '#e8eaf6'; // Solid color
            nodeBox.style.borderLeft = '4px solid #5c6bc0';
        } else if (type === 'root' || type === 'root_short') {
            // Root nodes get special styling
            nodeBox.style.background = '#fff8e1'; // Solid color
            nodeBox.style.borderLeft = '4px solid #ff9800';
        } else {
            nodeBox.style.background = '#f8fafd';
            nodeBox.style.borderLeft = '4px solid #b2bec3';
        }
        
        // Apply additional styles if provided
        for (const [property, value] of Object.entries(styles)) {
            nodeBox.style[property] = value;
        }
        
        // Add title if provided
        if (title) {
            const nodeHeader = document.createElement('div');
            nodeHeader.className = 'mpt-node-header';
            nodeHeader.style.fontWeight = 'bold';
            nodeHeader.style.marginBottom = '4px';
            nodeHeader.textContent = title;
            nodeBox.appendChild(nodeHeader);
        }
        
        return nodeBox;
    }
    
    /**
     * Add a property row to a node box with enhanced display
     * @param {HTMLElement} nodeBox - The container element
     * @param {string} label - Label for the property
     * @param {string} value - Value of the property
     * @param {string} color - Color for the label
     */
    addPropertyToNode(nodeBox, label, value, color = '#1976d2') {
        if (!value && value !== 0) return; // Don't add empty properties
        
        const propertyDiv = document.createElement('div');
        propertyDiv.className = 'mpt-property';
        propertyDiv.style.fontSize = '13px';
        propertyDiv.style.marginTop = '8px';
        propertyDiv.style.wordBreak = 'break-word';
        propertyDiv.style.overflowWrap = 'anywhere';
        propertyDiv.style.padding = '4px 0';
        propertyDiv.style.borderTop = '1px solid rgba(0,0,0,0.05)';
        
        // Add label with color
        const labelSpan = document.createElement('span');
        labelSpan.className = 'mpt-property-label';
        labelSpan.style.color = color;
        labelSpan.style.fontWeight = 'bold';
        labelSpan.style.display = 'block';
        labelSpan.style.marginBottom = '3px';
        labelSpan.textContent = `${label}:`;
        propertyDiv.appendChild(labelSpan);
        
        // For very long values, create a scrollable container
        const valueContainer = document.createElement('div');
        valueContainer.className = 'mpt-property-value-container';
        valueContainer.style.backgroundColor = 'rgba(0,0,0,0.03)';
        valueContainer.style.borderRadius = '4px';
        valueContainer.style.padding = '4px 6px';
        valueContainer.style.maxHeight = '80px';
        valueContainer.style.overflowY = 'auto';
        valueContainer.style.fontFamily = 'Consolas, monospace';
        valueContainer.style.fontSize = '12px';
        valueContainer.style.lineHeight = '1.4';
        
        // Format the value
        const valueSpan = document.createElement('span');
        valueSpan.className = 'mpt-property-value';
        
        // For hash values, format them with 0x prefix
        if (label.toLowerCase().includes('hash') && typeof value === 'string') {
            const formattedValue = value.startsWith('0x') ? value : `0x${value}`;
            valueSpan.textContent = formattedValue;
        }
        // For key values, format them with 0x prefix
        else if ((label === 'Key' || label === 'Original Key' || label.includes('Path')) && 
                 typeof value === 'string' && value.length > 0) {
            const formattedValue = value.startsWith('0x') ? value : `0x${value}`;
            valueSpan.textContent = formattedValue;
        }
        // For values, ensure they have 0x prefix too
        else if (label === 'Value' && typeof value === 'string') {
            const formattedValue = value.startsWith('0x') ? value : `0x${value}`;
            valueSpan.textContent = formattedValue;
        } 
        else {
            valueSpan.textContent = value;
        }
        
        valueContainer.appendChild(valueSpan);
        propertyDiv.appendChild(valueContainer);
        
        nodeBox.appendChild(propertyDiv);
    }
    
    /**
     * Remove leading zeros from hex values, keeping at least one zero
     * @param {string} hexValue - The hex value to trim
     * @returns {string} - The trimmed hex value
     */
    trimLeadingZeros(hexValue) {
        if (!hexValue || typeof hexValue !== 'string') return hexValue;
        
        // Remove 0x prefix if present
        let value = hexValue.startsWith('0x') ? hexValue.slice(2) : hexValue;
        
        // Remove leading zeros but keep at least one character
        value = value.replace(/^0+/, '') || '0';
        
        // Add back 0x prefix
        return '0x' + value;
    }
}