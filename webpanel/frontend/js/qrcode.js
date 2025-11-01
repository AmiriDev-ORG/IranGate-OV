// QR Code generation and management

// Generate QR code for client
async function generateClientQRCode(clientName, size = 256) {
    try {
        const params = new URLSearchParams({
            size: size,
            level: 'M' // Medium error correction
        });

        const response = await fetch(`/api/clients/${clientName}/qrcode?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        const result = await response.json();

        if (result.success) {
            return result.data;
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('Error generating QR code:', error);
        throw error;
    }
}

// Show QR code modal
function showQRCodeModal(clientName, qrCodeData) {
    const modal = document.getElementById('qrcodeModal');
    if (!modal) {
        createQRCodeModal();
    }

    // Update modal content
    const modalTitle = document.getElementById('qrcodeModalTitle');
    const qrImage = document.getElementById('qrcodeImage');
    const downloadBtn = document.getElementById('downloadQRCodeBtn');
    const copyBtn = document.getElementById('copyQRCodeBtn');

    modalTitle.textContent = `QR Code - ${clientName}`;
    qrImage.src = qrCodeData.data_url;
    qrImage.alt = `QR Code for ${clientName}`;

    // Update download button
    downloadBtn.onclick = () => downloadQRCode(clientName, qrCodeData.size);

    // Update copy button
    copyBtn.onclick = () => copyQRCodeToClipboard(qrCodeData.data_url);

    // Show modal
    modal.classList.remove('hidden');
}

// Create QR code modal if it doesn't exist
function createQRCodeModal() {
    const modalHTML = `
        <div id="qrcodeModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 hidden z-50">
            <div class="flex items-center justify-center min-h-screen p-4">
                <div class="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
                    <div class="flex justify-between items-center mb-4">
                        <h3 id="qrcodeModalTitle" class="text-lg font-semibold text-gray-800">QR Code</h3>
                        <button id="closeQRCodeModal" class="text-gray-400 hover:text-gray-600">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                            </svg>
                        </button>
                    </div>
                    
                    <div class="text-center mb-6">
                        <img id="qrcodeImage" src="" alt="QR Code" class="mx-auto max-w-full h-auto rounded-lg">
                    </div>
                    
                    <div class="flex justify-center space-x-4">
                        <button id="downloadQRCodeBtn" 
                                class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg">
                            Download
                        </button>
                        <button id="copyQRCodeBtn" 
                                class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg">
                            Copy Image
                        </button>
                    </div>
                    
                    <div class="mt-4 p-3 bg-gray-50 rounded-lg">
                        <p class="text-xs text-gray-600 text-center">
                            Scan this QR code with your VPN client to automatically configure the connection.
                        </p>
                    </div>
                </div>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // Add event listener for close button
    document.getElementById('closeQRCodeModal').addEventListener('click', hideQRCodeModal);
    
    // Close modal when clicking outside
    document.getElementById('qrcodeModal').addEventListener('click', function(e) {
        if (e.target === this) {
            hideQRCodeModal();
        }
    });
}

// Hide QR code modal
function hideQRCodeModal() {
    const modal = document.getElementById('qrcodeModal');
    if (modal) {
        modal.classList.add('hidden');
    }
}

// Download QR code
async function downloadQRCode(clientName, size = 512) {
    try {
        const params = new URLSearchParams({
            size: size,
            level: 'H' // High error correction for downloads
        });

        const response = await fetch(`/api/clients/${clientName}/qrcode-download?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        if (response.ok) {
            const blob = await response.blob();
            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = `${clientName}_qrcode.png`;
            link.style.display = 'none';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);
            
            showSuccess('QR code downloaded successfully');
        } else {
            throw new Error('Failed to download QR code');
        }
    } catch (error) {
        console.error('Error downloading QR code:', error);
        showError('Failed to download QR code');
    }
}

// Copy QR code to clipboard
async function copyQRCodeToClipboard(dataUrl) {
    try {
        // Convert data URL to blob
        const response = await fetch(dataUrl);
        const blob = await response.blob();
        
        // Copy to clipboard
        await navigator.clipboard.write([
            new ClipboardItem({
                'image/png': blob
            })
        ]);
        
        showSuccess('QR code copied to clipboard');
    } catch (error) {
        console.error('Error copying QR code:', error);
        showError('Failed to copy QR code to clipboard');
    }
}

// Generate QR code for client (main function)
async function generateQRCodeForClient(clientName, size = 256) {
    try {
        showLoading('Generating QR code...');
        
        const qrCodeData = await generateClientQRCode(clientName, size);
        showQRCodeModal(clientName, qrCodeData);
        
        hideLoading();
    } catch (error) {
        hideLoading();
        showError('Failed to generate QR code: ' + error.message);
    }
}

// Batch generate QR codes
async function batchGenerateQRCodes(clientNames, size = 256) {
    try {
        showLoading('Generating QR codes...');
        
        const params = new URLSearchParams();
        clientNames.forEach(name => {
            params.append('clients', name);
        });
        params.append('size', size);
        params.append('level', 'M');

        const response = await fetch(`/api/qrcode/batch-generate?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        const result = await response.json();

        if (result.success) {
            hideLoading();
            showBatchQRCodeModal(result.data);
        } else {
            throw new Error(result.message);
        }
    } catch (error) {
        hideLoading();
        console.error('Error batch generating QR codes:', error);
        showError('Failed to generate QR codes: ' + error.message);
    }
}

// Show batch QR code modal
function showBatchQRCodeModal(batchData) {
    const modal = document.getElementById('batchQRCodeModal');
    if (!modal) {
        createBatchQRCodeModal();
    }

    const container = document.getElementById('batchQRCodeContainer');
    container.innerHTML = '';

    Object.entries(batchData.results).forEach(([clientName, qrData]) => {
        const qrCard = document.createElement('div');
        qrCard.className = 'bg-white border border-gray-200 rounded-lg p-4 text-center';
        
        qrCard.innerHTML = `
            <h4 class="font-semibold text-gray-800 mb-2">${clientName}</h4>
            <img src="${qrData.data_url}" alt="QR Code for ${clientName}" class="mx-auto mb-3 max-w-full h-auto">
            <div class="flex justify-center space-x-2">
                <button onclick="downloadQRCode('${clientName}', ${qrData.size})" 
                        class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">
                    Download
                </button>
                <button onclick="copyQRCodeToClipboard('${qrData.data_url}')" 
                        class="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm">
                    Copy
                </button>
            </div>
        `;
        
        container.appendChild(qrCard);
    });

    // Show modal
    modal.classList.remove('hidden');
}

// Create batch QR code modal
function createBatchQRCodeModal() {
    const modalHTML = `
        <div id="batchQRCodeModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 hidden z-50">
            <div class="flex items-center justify-center min-h-screen p-4">
                <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full p-6">
                    <div class="flex justify-between items-center mb-6">
                        <h3 class="text-lg font-semibold text-gray-800">Batch QR Codes</h3>
                        <button id="closeBatchQRCodeModal" class="text-gray-400 hover:text-gray-600">
                            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                            </svg>
                        </button>
                    </div>
                    
                    <div id="batchQRCodeContainer" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-96 overflow-y-auto">
                        <!-- QR codes will be loaded here -->
                    </div>
                </div>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // Add event listener for close button
    document.getElementById('closeBatchQRCodeModal').addEventListener('click', hideBatchQRCodeModal);
    
    // Close modal when clicking outside
    document.getElementById('batchQRCodeModal').addEventListener('click', function(e) {
        if (e.target === this) {
            hideBatchQRCodeModal();
        }
    });
}

// Hide batch QR code modal
function hideBatchQRCodeModal() {
    const modal = document.getElementById('batchQRCodeModal');
    if (modal) {
        modal.classList.add('hidden');
    }
}

// Utility functions
function getToken() {
    return localStorage.getItem('authToken');
}

function showLoading(message) {
    // Create or update loading indicator
    let loading = document.getElementById('loadingIndicator');
    if (!loading) {
        loading = document.createElement('div');
        loading.id = 'loadingIndicator';
        loading.className = 'fixed top-4 left-1/2 transform -translate-x-1/2 bg-blue-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
        document.body.appendChild(loading);
    }
    loading.textContent = message;
}

function hideLoading() {
    const loading = document.getElementById('loadingIndicator');
    if (loading) {
        loading.remove();
    }
}

function showSuccess(message) {
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

function showError(message) {
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 5000);
}
