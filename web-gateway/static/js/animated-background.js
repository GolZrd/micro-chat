// ==================== ANIMATED BACKGROUND ====================
// Динамическое создание анимированных элементов

(function() {
    'use strict';
    
    console.log('🎨 Initializing animated background...');
    
    // Создаем контейнер для плавающих сфер
    function createFloatingOrbs() {
        const container = document.createElement('div');
        container.className = 'floating-orbs';
        
        for (let i = 0; i < 4; i++) {
            const orb = document.createElement('div');
            orb.className = 'orb';
            container.appendChild(orb);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Floating orbs created');
    }
    
    // Создаем анимированные волны
    function createAnimatedWaves() {
        const container = document.createElement('div');
        container.className = 'animated-waves';
        
        for (let i = 0; i < 3; i++) {
            const wave = document.createElement('div');
            wave.className = 'wave';
            container.appendChild(wave);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('🌊 Animated waves created');
    }
    
    // Создаем частицы
    function createParticles() {
        const container = document.createElement('div');
        container.className = 'particles';
        
        // Генерируем 30 частиц
        for (let i = 0; i < 30; i++) {
            const particle = document.createElement('div');
            particle.className = 'particle';
            
            // Случайная позиция
            particle.style.left = Math.random() * 100 + '%';
            particle.style.top = Math.random() * 100 + '%';
            
            // Случайная задержка анимации
            particle.style.animationDelay = Math.random() * 15 + 's';
            
            // Случайная продолжительность
            particle.style.animationDuration = (Math.random() * 10 + 10) + 's';
            
            container.appendChild(particle);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Particles created');
    }
    
    // Создаем светящиеся линии
    function createGlowLines() {
        const container = document.createElement('div');
        container.className = 'glow-lines';
        
        for (let i = 0; i < 3; i++) {
            const line = document.createElement('div');
            line.className = 'glow-line';
            container.appendChild(line);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('💫 Glow lines created');
    }
    
    // Инициализация при загрузке DOM
    document.addEventListener('DOMContentLoaded', function() {
        // Проверяем настройки производительности
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        
        if (!prefersReducedMotion) {
            createFloatingOrbs();
            createAnimatedWaves();
            createParticles();
            createGlowLines();
            console.log('✅ Animated background fully loaded');
        } else {
            console.log('ℹ️ Reduced motion preferred - skipping animations');
        }
    });
    
    // Добавляем интерактивность - курсор создает рябь
    let mouseX = 0;
    let mouseY = 0;
    
    document.addEventListener('mousemove', function(e) {
        mouseX = e.clientX;
        mouseY = e.clientY;
        
        // Создаем эффект ряби при движении мыши
        const ripple = document.createElement('div');
        ripple.style.position = 'fixed';
        ripple.style.width = '100px';
        ripple.style.height = '100px';
        ripple.style.borderRadius = '50%';
        ripple.style.border = '2px solid rgba(88, 101, 242, 0.3)';
        ripple.style.left = mouseX + 'px';
        ripple.style.top = mouseY + 'px';
        ripple.style.transform = 'translate(-50%, -50%)';
        ripple.style.pointerEvents = 'none';
        ripple.style.zIndex = '-1';
        ripple.style.animation = 'rippleEffect 1s ease-out forwards';
        
        document.body.appendChild(ripple);
        
        // Удаляем рябь после анимации
        setTimeout(() => {
            ripple.remove();
        }, 1000);
    });
    
    // Добавляем CSS для эффекта ряби
    const style = document.createElement('style');
    style.textContent = `
        @keyframes rippleEffect {
            0% {
                transform: translate(-50%, -50%) scale(0);
                opacity: 1;
            }
            100% {
                transform: translate(-50%, -50%) scale(3);
                opacity: 0;
            }
        }
    `;
    document.head.appendChild(style);
    
})();