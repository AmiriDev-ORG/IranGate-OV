// Tech Bubbles - Animated Particle Network Background
// Minimal, elegant particle system with connections

class ParticleNetwork {
    constructor() {
        this.canvas = null;
        this.ctx = null;
        this.particles = [];
        this.particleCount = 60;
        this.maxDistance = 150;
        this.mousePosition = { x: null, y: null };
        this.colors = {
            particles: 'rgba(100, 255, 218, 0.8)',    // Cyan - brighter
            lines: 'rgba(100, 255, 218, 0.25)',       // Cyan transparent - more visible
            mouseGlow: 'rgba(0, 212, 255, 0.4)'       // Blue glow - stronger
        };
    }

    init() {
        // Create canvas
        this.canvas = document.createElement('canvas');
        this.canvas.id = 'particleCanvas';
        document.body.insertBefore(this.canvas, document.body.firstChild);
        
        this.ctx = this.canvas.getContext('2d');
        this.resize();
        this.createParticles();
        
        // Event listeners
        window.addEventListener('resize', () => this.resize());
        window.addEventListener('mousemove', (e) => this.updateMousePosition(e));
        window.addEventListener('mouseout', () => this.resetMousePosition());
        
        // Start animation
        this.animate();
    }

    resize() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }

    createParticles() {
        this.particles = [];
        for (let i = 0; i < this.particleCount; i++) {
            this.particles.push({
                x: Math.random() * this.canvas.width,
                y: Math.random() * this.canvas.height,
                vx: (Math.random() - 0.5) * 0.5,
                vy: (Math.random() - 0.5) * 0.5,
                radius: Math.random() * 2 + 2
            });
        }
    }

    updateMousePosition(e) {
        this.mousePosition.x = e.clientX;
        this.mousePosition.y = e.clientY;
    }

    resetMousePosition() {
        this.mousePosition.x = null;
        this.mousePosition.y = null;
    }

    drawParticles() {
        this.particles.forEach(particle => {
            // Update position
            particle.x += particle.vx;
            particle.y += particle.vy;

            // Bounce off edges
            if (particle.x < 0 || particle.x > this.canvas.width) {
                particle.vx = -particle.vx;
            }
            if (particle.y < 0 || particle.y > this.canvas.height) {
                particle.vy = -particle.vy;
            }

            // Mouse interaction
            if (this.mousePosition.x !== null) {
                const dx = this.mousePosition.x - particle.x;
                const dy = this.mousePosition.y - particle.y;
                const distance = Math.sqrt(dx * dx + dy * dy);
                
                if (distance < 100) {
                    const force = (100 - distance) / 100;
                    particle.x -= dx * force * 0.03;
                    particle.y -= dy * force * 0.03;
                }
            }

            // Draw particle with glow
            this.ctx.beginPath();
            this.ctx.arc(particle.x, particle.y, particle.radius, 0, Math.PI * 2);
            
            // Gradient glow
            const gradient = this.ctx.createRadialGradient(
                particle.x, particle.y, 0,
                particle.x, particle.y, particle.radius * 5
            );
            gradient.addColorStop(0, this.colors.particles);
            gradient.addColorStop(0.5, 'rgba(100, 255, 218, 0.3)');
            gradient.addColorStop(1, 'rgba(100, 255, 218, 0)');
            
            this.ctx.fillStyle = gradient;
            this.ctx.fill();
        });
    }

    drawConnections() {
        for (let i = 0; i < this.particles.length; i++) {
            for (let j = i + 1; j < this.particles.length; j++) {
                const dx = this.particles[i].x - this.particles[j].x;
                const dy = this.particles[i].y - this.particles[j].y;
                const distance = Math.sqrt(dx * dx + dy * dy);

                if (distance < this.maxDistance) {
                    const opacity = (1 - distance / this.maxDistance) * 0.5;
                    
                    this.ctx.beginPath();
                    this.ctx.moveTo(this.particles[i].x, this.particles[i].y);
                    this.ctx.lineTo(this.particles[j].x, this.particles[j].y);
                    this.ctx.strokeStyle = `rgba(100, 255, 218, ${opacity * 0.3})`;
                    this.ctx.lineWidth = 0.5;
                    this.ctx.stroke();
                }
            }
        }
    }

    drawMouseGlow() {
        if (this.mousePosition.x !== null) {
            const gradient = this.ctx.createRadialGradient(
                this.mousePosition.x, this.mousePosition.y, 0,
                this.mousePosition.x, this.mousePosition.y, 150
            );
            gradient.addColorStop(0, this.colors.mouseGlow);
            gradient.addColorStop(1, 'rgba(0, 212, 255, 0)');
            
            this.ctx.fillStyle = gradient;
            this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        }
    }

    animate() {
        // Clear canvas
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        // Draw elements
        this.drawConnections();
        this.drawParticles();
        this.drawMouseGlow();
        
        // Continue animation
        requestAnimationFrame(() => this.animate());
    }
}

// Initialize particle network when DOM is loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        const network = new ParticleNetwork();
        network.init();
    });
} else {
    const network = new ParticleNetwork();
    network.init();
}

