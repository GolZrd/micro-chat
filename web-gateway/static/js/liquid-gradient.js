// ==================== LIQUID GRADIENT BACKGROUND ====================

(function() {
    'use strict';
    
    console.log('🌊 Initializing Liquid Gradient background...');
    
    // Создаем блобы
    function createBlobs() {
        const container = document.createElement('div');
        container.className = 'liquid-blobs';
        
        for (let i = 0; i < 3; i++) {
            const blob = document.createElement('div');
            blob.className = 'blob';
            container.appendChild(blob);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Liquid blobs created');
    }
    
    // Создаем плавающие точки
    function createFloatingDots() {
        const container = document.createElement('div');
        container.className = 'floating-dots';
        
        for (let i = 0; i < 40; i++) {
            const dot = document.createElement('div');
            dot.className = 'dot';
            
            dot.style.left = Math.random() * 100 + '%';
            dot.style.bottom = '0';
            dot.style.animationDelay = Math.random() * 15 + 's';
            dot.style.animationDuration = (Math.random() * 10 + 10) + 's';
            
            container.appendChild(dot);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Floating dots created');
    }
    
    // Создаем анимированную волну с SVG
    function createWaveLine() {
        const container = document.createElement('div');
        container.className = 'wave-line';
        
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('viewBox', '0 0 1200 200');
        svg.setAttribute('preserveAspectRatio', 'none');
        
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        
        let offset = 0;
        
        function animateWave() {
            offset += 0.5;
            const d = `M0,100 Q${150 + Math.sin(offset * 0.01) * 50},${80 + Math.sin(offset * 0.02) * 20} 300,100 T600,100 T900,100 T1200,100 L1200,200 L0,200 Z`;
            path.setAttribute('d', d);
            path.setAttribute('fill', 'url(#waveGradient)');
            path.setAttribute('opacity', '0.3');
            
            requestAnimationFrame(animateWave);
        }
        
        // Создаем градиент
        const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
        const gradient = document.createElementNS('http://www.w3.org/2000/svg', 'linearGradient');
        gradient.setAttribute('id', 'waveGradient');
        gradient.setAttribute('x1', '0%');
        gradient.setAttribute('y1', '0%');
        gradient.setAttribute('x2', '100%');
        gradient.setAttribute('y2', '0%');
        
        const stop1 = document.createElementNS('http://www.w3.org/2000/svg', 'stop');
        stop1.setAttribute('offset', '0%');
        stop1.setAttribute('style', 'stop-color:rgba(79, 172, 254, 0.8);stop-opacity:1');
        
        const stop2 = document.createElementNS('http://www.w3.org/2000/svg', 'stop');
        stop2.setAttribute('offset', '50%');
        stop2.setAttribute('style', 'stop-color:rgba(102, 126, 234, 0.8);stop-opacity:1');
        
        const stop3 = document.createElementNS('http://www.w3.org/2000/svg', 'stop');
        stop3.setAttribute('offset', '100%');
        stop3.setAttribute('style', 'stop-color:rgba(118, 75, 162, 0.8);stop-opacity:1');
        
        gradient.appendChild(stop1);
        gradient.appendChild(stop2);
        gradient.appendChild(stop3);
        defs.appendChild(gradient);
        svg.appendChild(defs);
        svg.appendChild(path);
        container.appendChild(svg);
        
        document.body.insertBefore(container, document.body.firstChild);
        animateWave();
        
        console.log('🌊 Wave line created');
    }
    
    // Инициализация
    document.addEventListener('DOMContentLoaded', function() {
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        
        if (!prefersReducedMotion) {
            createBlobs();
            createFloatingDots();
            createWaveLine();
            console.log('✅ Liquid Gradient background loaded');
        }
    });
    
})();