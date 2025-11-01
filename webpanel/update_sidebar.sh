#!/bin/bash
echo "Updating remaining pages with sidebar..."
cd /root/ov/irangate/webpanel/frontend
for page in inbound settings api-docs; do
    echo "Processing ${page}.html..."
    sed -i 's|<script src="js/particles.js"></script>|<link rel="stylesheet" href="css/sidebar.css">\n    <script src="js/particles.js"></script>\n    <script src="js/sidebar.js"></script>|' ${page}.html
    sed -i 's|<nav class="bg-white shadow-lg">|<nav class="bg-white shadow-lg" style="display:none;">|' ${page}.html
    sed -i 's|<body style="background:
    sed -i 's|</body>|    initSidebar('"'"'${page}'"'"');\n    </script>\n    </div>\n</body>|' ${page}.html
    echo "✓ ${page}.html updated"
done
echo ""
echo "All pages updated with sidebar!"
echo "Sidebar features:"
echo "  ✓ Animated slide-in"
echo "  ✓ Left-side positioning"
echo "  ✓ Collapsible on mobile"
echo "  ✓ Smooth transitions"
echo "  ✓ Active page highlighting"