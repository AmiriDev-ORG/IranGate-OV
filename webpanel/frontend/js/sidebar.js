// Sidebar Component for IranGate Web Panel

function createSidebar(activePage) {
    const username = localStorage.getItem('username') || 'Admin';
    
    return `
        <!-- Sidebar Toggle (Mobile) -->
        <button class="sidebar-toggle" onclick="toggleSidebar()">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
            </svg>
        </button>

        <!-- Sidebar Overlay (Mobile) -->
        <div class="sidebar-overlay" onclick="closeSidebar()"></div>

        <!-- Sidebar -->
        <aside class="sidebar" id="sidebar">
            <!-- Header -->
            <div class="sidebar-header">
                <div class="sidebar-logo">
                    <div class="sidebar-logo-icon">
                        <img src="logo/logo.ico" alt="IranGate Logo" class="w-10 h-10 rounded-lg">
                    </div>
                    <span class="sidebar-logo-text">IranGate</span>
                    <a href="https://t.me/irangate_official" target="_blank" class="sidebar-telegram-link" title="Join Our Telegram Channel">
                        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                            <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.028-1.627 4.476-1.635z"/>
                        </svg>
                    </a>
                </div>
                <div class="sidebar-user">
                    <div class="flex items-center gap-2">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                        </svg>
                        <span class="font-medium">${username}</span>
                    </div>
                </div>
            </div>

            <!-- Menu -->
            <nav class="sidebar-menu">
                <!-- Main Section -->
                <div class="menu-section">
                    <div class="menu-section-title">Main</div>
                    
                    <a href="dashboard.html" class="menu-item ${activePage === 'dashboard' ? 'active' : ''}" title="Dashboard">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
                        </svg>
                        <span>Dashboard</span>
                    </a>

                    <a href="analytics.html" class="menu-item ${activePage === 'analytics' ? 'active' : ''}" title="Analytics">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                        </svg>
                        <span>Analytics</span>
                    </a>
                </div>

                <!-- All Cores Clients Section -->
                <div class="menu-section">
                    <div class="menu-section-title">All Cores Clients</div>
                    
                    <a href="clients.html" class="menu-item ${activePage === 'clients' ? 'active' : ''}" title="OpenVPN Clients">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"/>
                        </svg>
                        <span>OpenVPN Clients</span>
                    </a>

                    <a href="#" class="menu-item coming-soon" title="V2ray" onclick="showComingSoon('V2ray')">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                        </svg>
                        <span>V2ray</span>
                        <span class="badge beta">💫 Coming Soon</span>
                    </a>

                    <a href="#" class="menu-item coming-soon" title="SSH" onclick="showComingSoon('SSH')">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
                        </svg>
                        <span>SSH</span>
                        <span class="badge beta">💫 Coming Soon</span>
                    </a>

                    <a href="ai-core.html" class="menu-item ${activePage === 'ai-core' ? 'active' : ''}" title="AI Core">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                        </svg>
                        <span>AI Core</span>
                    </a>
                </div>

                <!-- Configuration Section -->
                <div class="menu-section">
                    <div class="menu-section-title">Configuration</div>
                    
                    <a href="#" class="menu-item coming-soon" title="Inbound" onclick="showComingSoon('Inbound')">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
                        </svg>
                        <span>Inbound</span>
                        <span class="badge beta">💫 Coming Soon</span>
                    </a>

                    <a href="server-checker.html" class="menu-item ${activePage === 'server-checker' ? 'active' : ''}" title="Server Checker">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                        <span>Server Checker</span>
                    </a>

                    <a href="settings.html" class="menu-item ${activePage === 'settings' ? 'active' : ''}" title="Settings">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                        </svg>
                        <span>Settings</span>
                    </a>
                </div>

                <!-- Advanced Features Section -->
                <div class="menu-section">
                    <div class="menu-section-title">Advanced</div>
                    
                    <a href="#" class="menu-item coming-soon" title="Auto Tunnel" onclick="showComingSoon('Auto Tunnel')">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                        </svg>
                        <span>Auto Tunnel</span>
                        <span class="badge beta">💫 Wait for the surprise</span>
                    </a>
                </div>

                <!-- Other Section -->
                <div class="menu-section">
                    <div class="menu-section-title">Resources</div>
                    
                    <a href="api-docs.html" class="menu-item ${activePage === 'api-docs' ? 'active' : ''}" title="API Docs">
                        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                        </svg>
                        <span>API Docs</span>
                    </a>
                </div>
            </nav>

            <!-- Footer -->
            <div class="sidebar-footer">
                <button onclick="logout()" class="logout-btn">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
                    </svg>
                    <span>Logout</span>
                </button>
            </div>
        </aside>
    `;
}

// Toggle sidebar (mobile)
function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.querySelector('.sidebar-overlay');
    
    sidebar.classList.toggle('open');
    overlay.classList.toggle('active');
}

function closeSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.querySelector('.sidebar-overlay');
    
    sidebar.classList.remove('open');
    overlay.classList.remove('active');
}

// Initialize sidebar
function initSidebar(activePage) {
    const sidebarHTML = createSidebar(activePage);
    document.body.insertAdjacentHTML('afterbegin', sidebarHTML);
    
    // Close sidebar when clicking menu item on mobile
    if (window.innerWidth <= 1024) {
        document.querySelectorAll('.menu-item').forEach(item => {
            item.addEventListener('click', closeSidebar);
        });
    }
}

// Show coming soon modal
function showComingSoon(featureName) {
    const modal = document.createElement('div');
    modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
    modal.innerHTML = `
        <div class="bg-white rounded-lg p-8 max-w-md mx-4">
            <div class="text-center">
                <div class="mb-4">
                    <svg class="w-16 h-16 mx-auto text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                    </svg>
                </div>
                <h3 class="text-xl font-semibold text-gray-800 mb-2">${featureName}</h3>
                <p class="text-gray-600 mb-6 text-2xl">💫 Wait for the surprise!</p>
                <div class="text-left text-sm text-gray-600 mb-6">
                    <p class="mb-2 font-medium">Something amazing is coming:</p>
                    <ul class="list-disc list-inside space-y-1">
                        <li>Automatic tunnel creation based on traffic patterns</li>
                        <li>Intelligent routing rule generation</li>
                        <li>Automated client provisioning</li>
                        <li>Dynamic bandwidth allocation</li>
                    </ul>
                    <p class="mt-3 font-medium text-purple-600">🎁 Stay tuned for the surprise!</p>
                </div>
                <button onclick="this.closest('.fixed').remove()" class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg">
                    Got it
                </button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
}

