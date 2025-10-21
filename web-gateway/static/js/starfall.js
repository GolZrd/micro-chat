// ==================== STARFALL BACKGROUND ====================
// Космический фон с падающими звездами

(function() {
    'use strict';
    
    console.log('🌟 Initializing Starfall background...');
    
    // Создаем звездное поле
    function createStarfield() {
        const container = document.createElement('div');
        container.className = 'starfield';
        
        // Генерируем звезды разных размеров
        for (let i = 0; i < 200; i++) {
            const star = document.createElement('div');
            star.className = 'star';
            
            // Случайный размер
            const size = Math.random();
            if (size < 0.6) {
                star.classList.add('small');
            } else if (size < 0.9) {
                star.classList.add('medium');
            } else {
                star.classList.add('large');
            }
            
            // Случайная позиция
            star.style.left = Math.random() * 100 + '%';
            star.style.top = Math.random() * 100 + '%';
            
            // Случайная задержка мерцания
            star.style.animationDelay = Math.random() * 3 + 's';
            star.style.animationDuration = (Math.random() * 2 + 2) + 's';
            
            container.appendChild(star);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Starfield created');
    }
    
    // Создаем падающие звезды
    function createShootingStars() {
        const container = document.createElement('div');
        container.className = 'shooting-stars';
        
        // Генерируем падающие звезды с интервалом
        setInterval(() => {
            const star = document.createElement('div');
            star.className = 'shooting-star';
            
            // Случайный цвет следа
            const colorType = Math.random();
            if (colorType < 0.33) {
                star.classList.add('purple');
            } else if (colorType < 0.66) {
                star.classList.add('yellow');
            }
            
            // Случайная начальная позиция (только верхняя и правая часть)
            star.style.left = (Math.random() * 50 + 50) + '%';
            star.style.top = (Math.random() * 30) + '%';
            
            // Случайная задержка
            star.style.animationDelay = (Math.random() * 2) + 's';
            
            container.appendChild(star);
            
            // Удаляем звезду после анимации
            setTimeout(() => {
                star.remove();
            }, 3000);
        }, 800); // Новая звезда каждые 800ms
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('💫 Shooting stars created');
    }
    
    // Создаем кометы (редкие большие падающие звезды)
    function createComets() {
        setInterval(() => {
            const comet = document.createElement('div');
            comet.className = 'comet';
            
            // Случайная начальная позиция
            comet.style.left = (Math.random() * 30 + 70) + '%';
            comet.style.top = (Math.random() * 20) + '%';
            
            const shootingStars = document.querySelector('.shooting-stars');
            if (shootingStars) {
                shootingStars.appendChild(comet);
                
                // Удаляем комету после анимации
                setTimeout(() => {
                    comet.remove();
                }, 5000);
            }
        }, 5000); // Комета каждые 5 секунд
        
        console.log('☄️ Comets initialized');
    }
    
    // Создаем созвездия
    function createConstellations() {
        const container = document.createElement('div');
        container.className = 'constellation';
        
        // Создаем несколько созвездий
        const constellations = [
            // Созвездие 1
            [
                { x: 10, y: 20 },
                { x: 15, y: 25 },
                { x: 20, y: 22 },
                { x: 18, y: 28 }
            ],
            // Созвездие 2
            [
                { x: 70, y: 30 },
                { x: 75, y: 35 },
                { x: 78, y: 32 },
                { x: 82, y: 38 }
            ],
            // Созвездие 3
            [
                { x: 40, y: 60 },
                { x: 45, y: 65 },
                { x: 50, y: 62 }
            ]
        ];
        
        constellations.forEach((points, idx) => {
            for (let i = 0; i < points.length - 1; i++) {
                const line = document.createElement('div');
                line.className = 'constellation-line';
                
                const dx = points[i + 1].x - points[i].x;
                const dy = points[i + 1].y - points[i].y;
                const length = Math.sqrt(dx * dx + dy * dy);
                const angle = Math.atan2(dy, dx) * (180 / Math.PI);
                
                line.style.width = length + '%';
                line.style.left = points[i].x + '%';
                line.style.top = points[i].y + '%';
                line.style.transform = `rotate(${angle}deg)`;
                line.style.animationDelay = (idx * 1.5) + 's';
                
                container.appendChild(line);
            }
            
            // Добавляем звезды в узлах созвездия
            points.forEach(point => {
                const star = document.createElement('div');
                star.className = 'star large';
                star.style.left = point.x + '%';
                star.style.top = point.y + '%';
                star.style.animationDelay = (idx * 0.5) + 's';
                container.appendChild(star);
            });
        });
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('🌌 Constellations created');
    }
    
    // Создаем космическую пыль
    function createSpaceDust() {
        const container = document.createElement('div');
        container.className = 'space-dust';
        
        for (let i = 0; i < 100; i++) {
            const particle = document.createElement('div');
            particle.className = 'dust-particle';
            
            particle.style.left = Math.random() * 100 + '%';
            particle.style.bottom = '0';
            particle.style.animationDelay = Math.random() * 20 + 's';
            particle.style.animationDuration = (Math.random() * 15 + 15) + 's';
            
            container.appendChild(particle);
        }
        
        document.body.insertBefore(container, document.body.firstChild);
        console.log('✨ Space dust created');
    }
    
    // Инициализация
    document.addEventListener('DOMContentLoaded', function() {
        const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        
        if (!prefersReducedMotion) {
            createStarfield();
            createShootingStars();
            createComets();
            createConstellations();
            createSpaceDust();
            console.log('✅ Starfall background fully loaded');
        } else {
            // Только статичные звезды
            createStarfield();
            console.log('ℹ️ Reduced motion - only static stars');
        }
    });
    
})();