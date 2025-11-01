// IranGate Comprehensive Theme Loader
// This script loads and applies themes across all pages with full visual override

(function() {
    'use strict';

    // Theme switching functionality
    function applySelectedTheme(themeName) {
        // Remove all theme classes from document element (affects all elements)
        document.documentElement.classList.remove('theme-eye-of-sea', 'theme-capitan', 'theme-dragon', 'theme-ocean', 'theme-forest');
        
        // Add selected theme class to document element
        document.documentElement.classList.add(`theme-${themeName}`);
        
        // Also apply to body for compatibility
        document.body.classList.remove('theme-eye-of-sea', 'theme-capitan', 'theme-dragon', 'theme-ocean', 'theme-forest');
        document.body.classList.add(`theme-${themeName}`);
        
        // Update active theme option if on settings page
        const themeOptions = document.querySelectorAll('.theme-option');
        if (themeOptions.length > 0) {
            themeOptions.forEach(option => {
                option.classList.remove('active');
                if (option.dataset.theme === themeName) {
                    option.classList.add('active');
                }
            });
        }
        
        // Save selected theme
        localStorage.setItem('selected_theme', themeName);
        
        // Force re-render of all elements
        document.body.style.display = 'none';
        document.body.offsetHeight; // Trigger reflow
        document.body.style.display = '';
    }

        // Load theme on page load
        function loadTheme() {
            const selectedTheme = localStorage.getItem('selected_theme') || 'eye-of-sea';
            applySelectedTheme(selectedTheme);
        }

        // Theme option click handlers (for settings page)
        function initThemeSwitcher() {
            document.querySelectorAll('.theme-option').forEach(option => {
                option.addEventListener('click', function() {
                    const themeName = this.dataset.theme;
                    applySelectedTheme(themeName);
                    
                    // Show toast if available
                    if (typeof showToast === 'function') {
                        showToast(`Theme changed to ${this.querySelector('.theme-name').textContent}`, 'success');
                    }
                });
            });
        }

        // Cleanup function to prevent memory leaks
        function cleanup() {
            if (observer) {
                observer.disconnect();
            }
        }

        // Cleanup on page unload
        window.addEventListener('beforeunload', cleanup);

    // Override existing CSS with theme variables (OPTIMIZED)
    function applyThemeOverrides() {
        const selectedTheme = localStorage.getItem('selected_theme') || 'eye-of-sea';
        const themeClass = `theme-${selectedTheme}`;
        
        // Apply theme class to document element only (no need to add to all elements)
        document.documentElement.classList.add(themeClass);
        document.body.classList.add(themeClass);
        
        // CSS cascade will handle theme variables automatically
        // No need to manually add class to each element
    }

    // Initialize theme system
    document.addEventListener('DOMContentLoaded', function() {
        loadTheme();
        initThemeSwitcher();
        applyThemeOverrides();
        
        // Reapply theme after a short delay to ensure all elements are loaded
        setTimeout(() => {
            applyThemeOverrides();
        }, 100);
    });

    // Reapply theme when new content is loaded (OPTIMIZED)
    let observer = null;
    try {
        observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.type === 'childList') {
                    const selectedTheme = localStorage.getItem('selected_theme') || 'eye-of-sea';
                    const themeClass = `theme-${selectedTheme}`;
                    
                    mutation.addedNodes.forEach(function(node) {
                        if (node.nodeType === 1) { // Element node
                            // Only add class to the main node, CSS cascade will handle children
                            node.classList.add(themeClass);
                            // No need to iterate through all children - CSS cascade handles it
                        }
                    });
                }
            });
        });

        // Start observing (with debounce to prevent excessive updates)
        let timeoutId;
        if (document.body && observer) {
            observer.observe(document.body, {
                childList: true,
                subtree: true
            });
        }
    } catch (e) {
        console.warn('Failed to initialize MutationObserver:', e);
    }

    // Make functions globally available
    window.applySelectedTheme = applySelectedTheme;
    window.loadTheme = loadTheme;

})();