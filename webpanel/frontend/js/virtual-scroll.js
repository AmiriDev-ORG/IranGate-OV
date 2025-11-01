// Virtual Scrolling for Large Datasets (1000+ clients)
// Only renders visible rows for maximum performance

class VirtualScroller {
    constructor(containerId, options = {}) {
        this.container = document.getElementById(containerId);
        this.data = [];
        this.rowHeight = options.rowHeight || 60;
        this.bufferSize = options.bufferSize || 10;
        this.visibleStart = 0;
        this.visibleEnd = 0;
        this.scrollTop = 0;
        this.containerHeight = 0;
        this.rowRenderer = options.rowRenderer || this.defaultRowRenderer;
        
        this.init();
    }

    init() {
        if (!this.container) return;

        // Create virtual scroll container
        this.wrapper = document.createElement('div');
        this.wrapper.className = 'virtual-scroll-wrapper';
        this.wrapper.style.position = 'relative';
        this.wrapper.style.overflow = 'auto';
        this.wrapper.style.height = '600px';

        this.content = document.createElement('div');
        this.content.className = 'virtual-scroll-content';
        this.content.style.position = 'relative';

        this.viewport = document.createElement('div');
        this.viewport.className = 'virtual-scroll-viewport';
        this.viewport.style.position = 'absolute';
        this.viewport.style.top = '0';
        this.viewport.style.left = '0';
        this.viewport.style.right = '0';

        this.content.appendChild(this.viewport);
        this.wrapper.appendChild(this.content);
        this.container.appendChild(this.wrapper);

        // Event listeners
        this.wrapper.addEventListener('scroll', () => this.handleScroll());
        window.addEventListener('resize', () => this.handleResize());
    }

    setData(data) {
        this.data = data;
        this.totalHeight = data.length * this.rowHeight;
        this.content.style.height = this.totalHeight + 'px';
        this.containerHeight = this.wrapper.clientHeight;
        this.render();
    }

    handleScroll() {
        this.scrollTop = this.wrapper.scrollTop;
        this.render();
    }

    handleResize() {
        this.containerHeight = this.wrapper.clientHeight;
        this.render();
    }

    calculateVisibleRange() {
        const startIndex = Math.floor(this.scrollTop / this.rowHeight);
        const endIndex = Math.ceil((this.scrollTop + this.containerHeight) / this.rowHeight);

        this.visibleStart = Math.max(0, startIndex - this.bufferSize);
        this.visibleEnd = Math.min(this.data.length, endIndex + this.bufferSize);
    }

    render() {
        this.calculateVisibleRange();

        const visibleData = this.data.slice(this.visibleStart, this.visibleEnd);
        
        // Calculate offset
        const offsetY = this.visibleStart * this.rowHeight;
        this.viewport.style.transform = `translateY(${offsetY}px)`;

        // Render visible rows
        this.viewport.innerHTML = visibleData.map((item, index) => 
            this.rowRenderer(item, this.visibleStart + index)
        ).join('');
    }

    defaultRowRenderer(item, index) {
        return `
            <div class="virtual-row" style="height: ${this.rowHeight}px; display: flex; align-items: center; padding: 0 1rem; border-bottom: 1px solid rgba(30, 58, 95, 0.3);">
                <span class="text-gray-300">${index + 1}. ${JSON.stringify(item)}</span>
            </div>
        `;
    }

    scrollToIndex(index) {
        const scrollTo = index * this.rowHeight;
        this.wrapper.scrollTop = scrollTo;
    }

    refresh() {
        this.render();
    }

    destroy() {
        this.wrapper.removeEventListener('scroll', this.handleScroll);
        window.removeEventListener('resize', this.handleResize);
        this.container.innerHTML = '';
    }
}

