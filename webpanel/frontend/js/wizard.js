/**
 * OpenVPN Configuration Wizard
 * Advanced multi-step configuration wizard for OpenVPN server and client templates
 */

class OpenVPNWizard {
    constructor() {
        this.currentStep = 1;
        this.totalSteps = 8;
        this.config = {
            // Server Configuration
            name: '',
            deviceType: 'tun',
            deviceNumber: 0,
            protocol: 'tcp',
            port: 443,
            cipher: 'AES-256-GCM',
            
            // Advanced Parameters
            keepalive: true,
            persistKey: true,
            persistTun: true,
            mssFix: true,
            maxClients: false,
            maxClientsCount: null,
            pushDns: false,
            primaryDns: '',
            secondaryDns: '',
            tlsMethod: 'tls-crypt',
            
            // Client Configuration
            clientPrefix: 'yes',
            prefixName: '',
            prefixSeparator: '_',
            resolvRetry: false,
            ignoreUnknownOption: false,
            nobind: false
        };
        
        this.validationErrors = {};
        this.parameterDependencies = new Map();
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.setupParameterDependencies();
        this.updateProgress();
        this.updateStepVisibility();
    }
    
    setupEventListeners() {
        // Modal controls
        document.getElementById('openWizard')?.addEventListener('click', () => this.openWizard());
        document.getElementById('closeWizard')?.addEventListener('click', () => this.closeWizard());
        document.getElementById('prevStep')?.addEventListener('click', () => this.previousStep());
        document.getElementById('nextStep')?.addEventListener('click', () => this.nextStep());
        document.getElementById('finishWizard')?.addEventListener('click', () => this.finishWizard());
        
        // Step 1: Config Name
        document.getElementById('configName')?.addEventListener('input', (e) => {
            this.config.name = e.target.value;
            this.validateConfigName();
            this.updateConfigFileName();
        });
        
        // Step 2: Device Type
        document.querySelectorAll('input[name="deviceType"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.config.deviceType = e.target.value;
                this.updateDeviceOptions();
            });
        });
        
        document.getElementById('deviceNumber')?.addEventListener('change', (e) => {
            this.config.deviceNumber = parseInt(e.target.value);
        });
        
        // Step 3: Protocol
        document.querySelectorAll('input[name="protocol"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.config.protocol = e.target.value;
            });
        });
        
        // Step 4: Port
        document.querySelectorAll('.port-option').forEach(button => {
            button.addEventListener('click', (e) => {
                const port = parseInt(e.target.dataset.port);
                this.selectPort(port);
            });
        });
        
        document.getElementById('customPort')?.addEventListener('input', (e) => {
            const port = parseInt(e.target.value);
            if (port && port >= 1 && port <= 65535) {
                this.selectPort(port);
            }
        });
        
        // Step 5: Cipher
        document.getElementById('cipher')?.addEventListener('change', (e) => {
            this.config.cipher = e.target.value;
        });
        
        // TLS Method
        document.getElementById('tlsMethod')?.addEventListener('change', (e) => {
            this.config.tlsMethod = e.target.value;
        });
        
        // Step 6: Advanced Parameters
        document.getElementById('keepalive')?.addEventListener('change', (e) => {
            this.config.keepalive = e.target.checked;
        });
        
        document.getElementById('persistKey')?.addEventListener('change', (e) => {
            this.config.persistKey = e.target.checked;
        });
        
        document.getElementById('persistTun')?.addEventListener('change', (e) => {
            this.config.persistTun = e.target.checked;
        });
        
        document.getElementById('mssFix')?.addEventListener('change', (e) => {
            this.config.mssFix = e.target.checked;
        });
        
        document.getElementById('nobind')?.addEventListener('change', (e) => {
            this.config.nobind = e.target.checked;
        });
        
        document.getElementById('maxClients')?.addEventListener('change', (e) => {
            this.config.maxClients = e.target.checked;
            this.toggleMaxClientsInput();
        });
        
        // Max clients input validation
        const maxClientsInput = document.querySelector('#maxClientsInput input');
        if (maxClientsInput) {
            maxClientsInput.addEventListener('input', (e) => {
                this.config.maxClientsCount = parseInt(e.target.value);
            });
        }
        
        document.getElementById('pushDns')?.addEventListener('change', (e) => {
            this.config.pushDns = e.target.checked;
            this.toggleDnsSettings();
        });
        
        document.getElementById('primaryDns')?.addEventListener('input', (e) => {
            this.config.primaryDns = e.target.value;
        });
        
        document.getElementById('secondaryDns')?.addEventListener('input', (e) => {
            this.config.secondaryDns = e.target.value;
        });
        
        // Step 7: Client Prefix
        document.querySelectorAll('input[name="clientPrefix"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.config.clientPrefix = e.target.value;
                this.togglePrefixSettings();
            });
        });
        
        document.getElementById('prefixName')?.addEventListener('input', (e) => {
            this.config.prefixName = e.target.value;
            this.updatePrefixPreview();
        });
        
        document.getElementById('prefixSeparator')?.addEventListener('change', (e) => {
            this.config.prefixSeparator = e.target.value;
            this.updatePrefixPreview();
        });
        
        // Step 8: Client Parameters
        document.getElementById('resolvRetry')?.addEventListener('change', (e) => {
            this.config.resolvRetry = e.target.checked;
        });
        
        document.getElementById('ignoreUnknownOption')?.addEventListener('change', (e) => {
            this.config.ignoreUnknownOption = e.target.checked;
        });
        
        document.getElementById('nobind')?.addEventListener('change', (e) => {
            this.config.nobind = e.target.checked;
        });
        
        // Close modal on backdrop click
        document.getElementById('ovpnWizardModal')?.addEventListener('click', (e) => {
            if (e.target.id === 'ovpnWizardModal') {
                this.closeWizard();
            }
        });
    }
    
    setupParameterDependencies() {
        // Define parameter dependencies
        this.parameterDependencies.set('deviceType', ['deviceNumber']);
        this.parameterDependencies.set('protocol', ['port']);
        this.parameterDependencies.set('port', ['cipher']);
        this.parameterDependencies.set('cipher', ['keepalive', 'persistKey', 'persistTun', 'mssFix']);
    }
    
    openWizard() {
        const modal = document.getElementById('ovpnWizardModal');
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        this.resetWizard();
    }
    
    closeWizard() {
        const modal = document.getElementById('ovpnWizardModal');
        modal.classList.add('hidden');
        document.body.style.overflow = '';
    }
    
    resetWizard() {
        this.currentStep = 1;
        this.config = {
            name: '',
            deviceType: 'tun',
            deviceNumber: 0,
            protocol: 'tcp',
            port: 443,
            cipher: 'AES-256-GCM',
            keepalive: true,
            persistKey: true,
            persistTun: true,
            mssFix: true,
            maxClients: false,
            maxClientsCount: null,
            pushDns: false,
            primaryDns: '',
            secondaryDns: '',
            clientPrefix: 'yes',
            prefixName: '',
            prefixSeparator: '_',
            resolvRetry: false,
            ignoreUnknownOption: false,
            nobind: false
        };
        this.validationErrors = {};
        this.updateProgress();
        this.updateStepVisibility();
        this.updateLockedParameters();
    }
    
    previousStep() {
        if (this.currentStep > 1) {
            this.currentStep--;
            this.updateProgress();
            this.updateStepVisibility();
            this.updateButtonStates();
        }
    }
    
    nextStep() {
        if (this.validateCurrentStep()) {
            if (this.currentStep < this.totalSteps) {
                this.currentStep++;
                this.updateProgress();
                this.updateStepVisibility();
                this.updateButtonStates();
                this.updateLockedParameters();
            }
        }
    }
    
    finishWizard() {
        if (this.validateCurrentStep()) {
            this.generateConfigurations();
            this.closeWizard();
        }
    }
    
    validateCurrentStep() {
        let isValid = true;
        this.validationErrors = {};
        
        switch (this.currentStep) {
            case 1:
                isValid = this.validateConfigName();
                break;
            case 2:
                isValid = this.validateDeviceType();
                break;
            case 3:
                isValid = this.validateProtocol();
                break;
            case 4:
                isValid = this.validatePort();
                break;
            case 5:
                isValid = this.validateCipher();
                break;
            case 6:
                isValid = this.validateAdvancedParams();
                break;
            case 7:
                isValid = this.validateClientPrefix();
                break;
            case 8:
                isValid = this.validateClientParams();
                break;
        }
        
        return isValid;
    }
    
    validateConfigName() {
        const name = this.config.name.trim();
        const forbiddenChars = /[^a-zA-Z0-9_-]/;
        const pathTraversalPattern = /\.\.|\/|\\|:|<|>|\||\?|\*/;
        
        if (!name) {
            this.showValidationError('configNameValidation', 'Configuration name is required');
            return false;
        }
        
        if (name.length < 3) {
            this.showValidationError('configNameValidation', 'Configuration name must be at least 3 characters');
            return false;
        }
        
        if (name.length > 50) {
            this.showValidationError('configNameValidation', 'Configuration name must be less than 50 characters');
            return false;
        }
        
        if (pathTraversalPattern.test(name)) {
            this.showValidationError('configNameValidation', 'Configuration name contains forbidden characters that could be used for path traversal attacks');
            return false;
        }
        
        if (forbiddenChars.test(name)) {
            this.showValidationError('configNameValidation', 'Configuration name contains forbidden characters. Only letters, numbers, underscores, and hyphens are allowed');
            return false;
        }
        
        // Check for reserved names
        const reservedNames = ['server', 'client', 'ca', 'cert', 'key', 'dh', 'tls-crypt', 'config', 'template'];
        if (reservedNames.includes(name.toLowerCase())) {
            this.showValidationError('configNameValidation', 'This name is reserved. Please choose a different name');
            return false;
        }
        
        this.showValidationSuccess('configNameValidation', 'Configuration name is valid');
        return true;
    }
    
    validateDeviceType() {
        return true; // Always valid as we have defaults
    }
    
    validateProtocol() {
        return true; // Always valid as we have defaults
    }
    
    validatePort() {
        const port = this.config.port;
        
        if (!port || port < 1 || port > 65535) {
            this.showValidationError('portValidation', 'Port must be between 1 and 65535');
            return false;
        }
        
        this.showValidationSuccess('portValidation', 'Port is valid');
        return true;
    }
    
    validateCipher() {
        return true; // Always valid as we have defaults
    }
    
    validateAdvancedParams() {
        let isValid = true;
        
        if (this.config.maxClients && !this.config.maxClientsCount) {
            const input = document.querySelector('#maxClientsInput input');
            if (input && !input.value) {
                this.showValidationError('maxClientsInput', 'Please enter maximum number of clients');
                isValid = false;
            }
        }
        
        if (this.config.pushDns) {
            if (!this.config.primaryDns) {
                this.showValidationError('primaryDns', 'Primary DNS is required when Push DNS is enabled');
                isValid = false;
            }
            if (!this.config.secondaryDns) {
                this.showValidationError('secondaryDns', 'Secondary DNS is required when Push DNS is enabled');
                isValid = false;
            }
        }
        
        return isValid;
    }
    
    validateClientPrefix() {
        if (this.config.clientPrefix === 'yes' && !this.config.prefixName.trim()) {
            this.showValidationError('prefixName', 'Prefix name is required when using prefix');
            return false;
        }
        return true;
    }
    
    validateClientParams() {
        return true; // Always valid
    }
    
    showValidationError(elementId, message) {
        const element = document.getElementById(elementId);
        if (element) {
            element.textContent = message;
            element.className = 'validation-message error';
        }
    }
    
    showValidationSuccess(elementId, message) {
        const element = document.getElementById(elementId);
        if (element) {
            element.textContent = message;
            element.className = 'validation-message success';
        }
    }
    
    updateProgress() {
        const progressFill = document.getElementById('progressFill');
        const currentStepText = document.getElementById('currentStepText');
        
        if (progressFill) {
            const percentage = (this.currentStep / this.totalSteps) * 100;
            progressFill.style.width = `${percentage}%`;
        }
        
        if (currentStepText) {
            currentStepText.textContent = `Step ${this.currentStep} of ${this.totalSteps}`;
        }
    }
    
    updateStepVisibility() {
        // Hide all steps
        for (let i = 1; i <= this.totalSteps; i++) {
            const step = document.getElementById(`step${i}`);
            if (step) {
                step.classList.remove('active');
            }
        }
        
        // Show current step
        const currentStepElement = document.getElementById(`step${this.currentStep}`);
        if (currentStepElement) {
            currentStepElement.classList.add('active');
        }
    }
    
    updateButtonStates() {
        const prevBtn = document.getElementById('prevStep');
        const nextBtn = document.getElementById('nextStep');
        const finishBtn = document.getElementById('finishWizard');
        
        if (prevBtn) {
            prevBtn.disabled = this.currentStep === 1;
        }
        
        if (nextBtn && finishBtn) {
            if (this.currentStep === this.totalSteps) {
                nextBtn.classList.add('hidden');
                finishBtn.classList.remove('hidden');
            } else {
                nextBtn.classList.remove('hidden');
                finishBtn.classList.add('hidden');
            }
        }
    }
    
    updateConfigFileName() {
        const fileName = this.config.name ? `${this.config.name}.conf` : 'config.conf';
        const element = document.getElementById('configFileName');
        if (element) {
            element.textContent = fileName;
        }
    }
    
    updateDeviceOptions() {
        const deviceNumber = document.getElementById('deviceNumber');
        if (deviceNumber) {
            const options = deviceNumber.querySelectorAll('option');
            options.forEach(option => {
                const value = parseInt(option.value);
                const deviceType = this.config.deviceType;
                option.textContent = `${deviceType}${value}`;
            });
        }
    }
    
    selectPort(port) {
        this.config.port = port;
        
        // Update port option buttons
        document.querySelectorAll('.port-option').forEach(button => {
            button.classList.remove('selected');
            if (parseInt(button.dataset.port) === port) {
                button.classList.add('selected');
            }
        });
        
        // Clear custom port if a predefined port is selected
        if ([443, 80, 1010, 65000].includes(port)) {
            const customPortInput = document.getElementById('customPort');
            if (customPortInput) {
                customPortInput.value = '';
            }
        }
    }
    
    toggleMaxClientsInput() {
        const input = document.getElementById('maxClientsInput');
        if (input) {
            if (this.config.maxClients) {
                input.classList.remove('hidden');
            } else {
                input.classList.add('hidden');
            }
        }
    }
    
    toggleDnsSettings() {
        const settings = document.getElementById('dnsSettings');
        if (settings) {
            if (this.config.pushDns) {
                settings.classList.remove('hidden');
            } else {
                settings.classList.add('hidden');
            }
        }
    }
    
    togglePrefixSettings() {
        const settings = document.getElementById('prefixSettings');
        if (settings) {
            if (this.config.clientPrefix === 'yes') {
                settings.classList.remove('hidden');
            } else {
                settings.classList.add('hidden');
            }
        }
    }
    
    updatePrefixPreview() {
        const preview = document.getElementById('prefixPreview');
        if (preview && this.config.clientPrefix === 'yes') {
            const prefix = this.config.prefixName || 'Server1';
            const separator = this.config.prefixSeparator || '_';
            preview.textContent = `${prefix}${separator}John.ovpn`;
        }
    }
    
    updateLockedParameters() {
        // Update locked parameters display in step 8
        const lockedDeviceType = document.getElementById('lockedDeviceType');
        const lockedProtocol = document.getElementById('lockedProtocol');
        const lockedPort = document.getElementById('lockedPort');
        
        if (lockedDeviceType) {
            lockedDeviceType.textContent = `${this.config.deviceType}${this.config.deviceNumber}`;
        }
        
        if (lockedProtocol) {
            lockedProtocol.textContent = this.config.protocol;
        }
        
        if (lockedPort) {
            lockedPort.textContent = this.config.port;
        }
    }
    
    async generateConfigurations() {
        try {
            // Show loading state
            this.showLoadingState();
            
            // Prepare data for backend API
            const configData = {
                name: this.config.name,
                deviceType: this.config.deviceType,
                deviceNumber: this.config.deviceNumber,
                protocol: this.config.protocol,
                port: this.config.port,
                cipher: this.config.cipher,
                keepalive: this.config.keepalive,
                persistKey: this.config.persistKey,
                persistTun: this.config.persistTun,
                mssFix: this.config.mssFix,
                nobind: this.config.nobind,
                maxClients: this.config.maxClients,
                maxClientsCount: this.config.maxClientsCount,
                pushDns: this.config.pushDns,
                primaryDns: this.config.primaryDns,
                secondaryDns: this.config.secondaryDns,
                prefixName: this.config.prefixName,
                prefixSeparator: this.config.prefixSeparator,
                resolvRetry: this.config.resolvRetry,
                ignoreUnknownOption: this.config.ignoreUnknownOption
            };
            
            // Send to backend API
            const response = await fetch('/api/openvpn/create-configs', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({
                    inbounds: [{
                        id: 'wizard_' + Date.now(),
                        name: configData.name,
                        port: configData.port.toString(),
                        protocol: configData.protocol,
                        clientTemplate: this.generateClientTemplate(),
                        serverTemplate: this.generateServerConfig()
                    }]
                })
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }
            
            const result = await response.json();
            
            if (result.success) {
                this.showSuccessMessage(result);
            } else {
                this.showErrorMessage(result.message || 'Failed to create configuration');
            }
            
        } catch (error) {
            console.error('Error creating configuration:', error);
            this.showErrorMessage('Network error: ' + error.message);
        }
    }
    
    generateServerConfig() {
        let config = `# OpenVPN Server Configuration
# Generated by IranGate Configuration Wizard
# Configuration: ${this.config.name}

# Basic Settings
dev ${this.config.deviceType}${this.config.deviceNumber}
proto ${this.config.protocol}
port ${this.config.port}
cipher ${this.config.cipher}

# Certificate and Key Files
ca /etc/openvpn/ca.crt
cert /etc/openvpn/server.crt
key /etc/openvpn/server.key
dh /etc/openvpn/dh.pem
${this.config.tlsMethod === 'tls-crypt' ? 'tls-crypt /etc/openvpn/tc.key' : 'tls-auth /etc/openvpn/ta.key 0'}

# Network Configuration
server 10.8.0.0 255.255.255.0
ifconfig-pool-persist /var/log/openvpn/ipp.txt

# Client Configuration
client-config-dir /etc/openvpn/ccd
route 192.168.1.0 255.255.255.0

# Advanced Parameters
`;

        if (this.config.keepalive) {
            config += 'keepalive 10 120\n';
        }
        
        if (this.config.persistKey) {
            config += 'persist-key\n';
        }
        
        if (this.config.persistTun) {
            config += 'persist-tun\n';
        }
        
        if (this.config.mssFix) {
            config += 'mssfix 1400\n';
        }
        
        if (this.config.nobind) {
            config += 'nobind\n';
        }
        
        if (this.config.maxClients && this.config.maxClientsCount) {
            config += `max-clients ${this.config.maxClientsCount}\n`;
        }
        
        if (this.config.pushDns) {
            config += `push "dhcp-option DNS ${this.config.primaryDns}"\n`;
            if (this.config.secondaryDns) {
                config += `push "dhcp-option DNS ${this.config.secondaryDns}"\n`;
            }
        }
        
        // Authentication
        config += 'auth SHA512\n';
        
        config += `
# Logging
verb 3
log /var/log/openvpn/openvpn.log
status /var/log/openvpn/status.log 10
`;

        return config;
    }
    
    generateClientTemplate() {
        let template = `# OpenVPN Client Configuration Template
# Generated by IranGate Configuration Wizard
# Configuration: ${this.config.name}

# Client directive
client

# Device and Protocol (locked from server)
dev ${this.config.deviceType}${this.config.deviceNumber}
proto ${this.config.protocol}

# Remote server
remote <SERVER_IP_OR_DOMAIN> ${this.config.port}

# Cipher (locked from server)
cipher ${this.config.cipher}

# Authentication
auth SHA512

# Locked Parameters from Server
`;

        if (this.config.persistKey) {
            template += 'persist-key\n';
        }
        
        if (this.config.persistTun) {
            template += 'persist-tun\n';
        }
        
        if (this.config.mssFix) {
            template += 'mssfix 1400\n';
        }
        
        template += '\n# Client-Only Parameters\n';
        
        if (this.config.resolvRetry) {
            template += 'resolv-retry infinite\n';
        }
        
        if (this.config.ignoreUnknownOption) {
            template += 'ignore-unknown-option block-outside-dns\n';
        }
        
        if (this.config.nobind) {
            template += 'nobind\n';
        }
        
        template += `
# Certificate and Key Files (will be embedded)
# <ca>
# </ca>
# <cert>
# </cert>
# <key>
# </key>
# ${this.config.tlsMethod === 'tls-crypt' ? '<tls-crypt>' : '<tls-auth>'}
# </${this.config.tlsMethod === 'tls-crypt' ? 'tls-crypt' : 'tls-auth'}>
`;

        return template;
    }
    
    showLoadingState() {
        const finishBtn = document.getElementById('finishWizard');
        if (finishBtn) {
            finishBtn.disabled = true;
            finishBtn.innerHTML = '<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>Creating...';
        }
    }
    
    hideLoadingState() {
        const finishBtn = document.getElementById('finishWizard');
        if (finishBtn) {
            finishBtn.disabled = false;
            finishBtn.innerHTML = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>Create Configuration';
        }
    }
    
    showErrorMessage(message) {
        // Hide loading state
        this.hideLoadingState();
        
        // Show error modal
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                <div class="flex items-center mb-4">
                    <svg class="w-12 h-12 text-red-500 mr-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                    <div>
                        <h3 class="text-xl font-semibold text-gray-800">Configuration Failed</h3>
                        <p class="text-red-600 mt-2">${message}</p>
                    </div>
                </div>
                <div class="flex justify-end">
                    <button onclick="this.closest('.fixed').remove()" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">
                        Close
                    </button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    }
    
    async showSuccessMessage(result) {
        try {
            // Hide loading state
            this.hideLoadingState();
            
            // Show success modal
            this.showConfigPreview(result);
            
        } catch (error) {
            console.error('Error showing success message:', error);
            alert('Configuration created successfully!');
        }
    }
    
    generateClientFileName() {
        if (this.config.clientPrefix === 'yes' && this.config.prefixName) {
            return `${this.config.prefixName}${this.config.prefixSeparator}template.ovpn`;
        }
        return 'template.ovpn';
    }
    
    showConfigPreview(result) {
        // Create preview modal
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white rounded-lg p-6 max-w-2xl w-full mx-4">
                <div class="flex justify-between items-center mb-4">
                    <h2 class="text-2xl font-semibold text-gray-800">Configuration Created Successfully!</h2>
                    <button onclick="this.closest('.fixed').remove()" class="text-gray-500 hover:text-gray-700 text-2xl">&times;</button>
                </div>
                <div class="mb-6">
                    <div class="flex items-center mb-4">
                        <svg class="w-12 h-12 text-green-500 mr-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                        </svg>
                        <div>
                            <h3 class="text-lg font-semibold text-gray-800">Configuration: ${this.config.name}</h3>
                            <p class="text-gray-600">OpenVPN configuration has been created and saved</p>
                        </div>
                    </div>
                    <div class="bg-gray-50 rounded-lg p-4">
                        <p class="text-sm text-gray-700 mb-2"><strong>Files Created:</strong></p>
                        <ul class="list-disc list-inside space-y-1">
                            ${(result.data.created_files || []).map(file => `<li class="text-gray-600 font-mono text-sm">/etc/openvpn/${file}</li>`).join('')}
                        </ul>
                        <p class="text-sm text-gray-700 mt-3"><strong>Total Files:</strong> ${result.data.total_files || 0}</p>
                    </div>
                </div>
                <div class="bg-blue-50 rounded-lg p-4 mb-4">
                    <h4 class="font-semibold text-blue-800 mb-2">Next Steps:</h4>
                    <ol class="list-decimal list-inside space-y-1 text-sm text-blue-700">
                        <li>Configure server IP/domain in Settings</li>
                        <li>Start the OpenVPN server service</li>
                        <li>Create client configurations using the template</li>
                    </ol>
                </div>
                <div class="flex justify-end">
                    <button onclick="this.closest('.fixed').remove()" class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded">
                        Close
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    }
}

// Initialize wizard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.ovpnWizard = new OpenVPNWizard();
});
