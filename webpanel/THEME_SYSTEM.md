# IranGate Comprehensive Theme System

## Overview
The IranGate webpanel now includes a **COMPLETE** theme system that changes **EVERY** visual element including:
- Sidebar colors and styling
- All buttons and their hover effects  
- All borders and containers
- All text colors and typography
- All background colors and gradients
- All animations and effects
- All form elements and inputs
- All tables and data displays
- All status indicators and progress bars
- All scrollbars and UI elements

**This is a FULL visual transformation, not just background changes!**

## Available Themes

### 1. Eye of Sea (Default)
- **Colors**: Navy blue, dark blue, cyan accents
- **Style**: Professional, modern
- **Best for**: Business environments, professional use

### 2. Capitan
- **Colors**: Black, white, gray
- **Style**: Minimalist, modern
- **Best for**: Clean, professional look

### 3. Dragon
- **Colors**: Red, orange, black
- **Style**: Bold, energetic
- **Best for**: High-energy environments, gaming setups

### 4. Ocean
- **Colors**: Blue, teal, cyan
- **Style**: Calming, aquatic
- **Best for**: Relaxed environments, nature lovers

### 5. Forest
- **Colors**: Green, earth tones
- **Style**: Natural, organic
- **Best for**: Eco-friendly setups, nature themes

## How to Use

### Switching Themes
1. Go to **Settings** page
2. Scroll to **Webpanel Theme** section
3. Click on any theme preview to switch
4. Theme changes are applied immediately
5. Your selection is saved automatically

### Online Themes
- Click **"🌐 Get Online Themes"** link
- Download additional themes from the community repository
- Follow installation instructions for custom themes

### Custom Themes
1. Use the **Custom Theme** section
2. Add your own CSS in the **Custom CSS** field
3. Add JavaScript functionality in the **Custom JavaScript** field
4. Click **Save Custom Theme**

## Technical Details

### File Structure
```
/opt/irangate/webpanel/frontend/
├── css/
│   ├── themes.css          # Theme definitions
│   ├── dark-theme.css      # Default theme
│   └── ...
├── js/
│   ├── theme-loader.js     # Theme switching logic
│   └── ...
└── settings.html           # Theme configuration page
```

### CSS Variables
Each theme defines CSS custom properties:
- `--primary-bg`: Main background color
- `--secondary-bg`: Secondary background
- `--accent-color`: Primary accent color
- `--text-primary`: Main text color
- `--text-secondary`: Secondary text color
- And more...

### JavaScript API
```javascript
// Apply a theme programmatically
applySelectedTheme('dragon');

// Load saved theme
loadTheme();

// Get current theme
const currentTheme = localStorage.getItem('selected_theme');
```

## Development

### Adding New Themes
1. Add theme CSS variables to `themes.css`
2. Add theme preview styles
3. Update theme switcher HTML in `settings.html`
4. Add theme option to JavaScript

### Theme Structure
```css
.theme-your-theme {
    --primary-bg: #your-color;
    --secondary-bg: #your-color;
    --accent-color: #your-color;
    /* ... more variables */
}
```

## Browser Support
- Modern browsers with CSS custom properties support
- Chrome 49+, Firefox 31+, Safari 9.1+
- IE 11+ (with polyfill)

## Troubleshooting

### Theme Not Loading
1. Check browser console for errors
2. Clear browser cache
3. Verify CSS files are accessible

### Custom CSS Not Working
1. Check CSS syntax
2. Use `!important` for overrides
3. Verify selectors are correct

### Theme Not Persisting
1. Check localStorage is enabled
2. Clear browser data and try again
3. Check for JavaScript errors

## Support
For theme-related issues or requests:
- Check the GitHub repository
- Create an issue with theme details
- Join the community Discord
