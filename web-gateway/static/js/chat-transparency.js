// ==================== CHAT TRANSPARENCY EFFECTS ====================
// Динамическое изменение прозрачности при скролле

(function() {
    'use strict';
    
    console.log('✨ Initializing chat transparency effects...');
    
    document.addEventListener('DOMContentLoaded', function() {
        const messagesContainer = document.querySelector('.messages');
        const chatHeader = document.querySelector('.chat-header');
        const messageInputArea = document.querySelector('.message-input-area');
        
        if (!messagesContainer || !chatHeader || !messageInputArea) {
            console.log('ℹ️ Not on chat page');
            return;
        }
        
        // Проверяем позицию скролла и меняем прозрачность
        messagesContainer.addEventListener('scroll', function() {
            const scrollTop = this.scrollTop;
            const scrollHeight = this.scrollHeight;
            const clientHeight = this.clientHeight;
            const scrollBottom = scrollHeight - clientHeight - scrollTop;
            
            // Затемняем header при скролле вниз
            if (scrollTop > 50) {
                chatHeader.style.background = 'rgba(26, 26, 36, 0.9)';
                chatHeader.style.backdropFilter = 'blur(30px) saturate(200%)';
            } else {
                chatHeader.style.background = 'rgba(26, 26, 36, 0.7)';
                chatHeader.style.backdropFilter = 'blur(20px) saturate(180%)';
            }
            
            // Затемняем input area при скролле вверх
            if (scrollBottom > 50) {
                messageInputArea.style.background = 'rgba(26, 26, 36, 0.9)';
                messageInputArea.style.backdropFilter = 'blur(30px) saturate(200%)';
            } else {
                messageInputArea.style.background = 'rgba(26, 26, 36, 0.65)';
                messageInputArea.style.backdropFilter = 'blur(25px) saturate(180%)';
            }
        });
        
        // Эффект свечения при фокусе на input
        const messageInput = document.getElementById('messageInput');
        if (messageInput) {
            messageInput.addEventListener('focus', function() {
                messageInputArea.style.borderTopColor = 'rgba(88, 101, 242, 0.5)';
                messageInputArea.style.boxShadow = '0 -4px 20px rgba(88, 101, 242, 0.2)';
            });
            
            messageInput.addEventListener('blur', function() {
                messageInputArea.style.borderTopColor = 'rgba(255, 255, 255, 0.1)';
                messageInputArea.style.boxShadow = 'none';
            });
        }
        
        console.log('✅ Chat transparency effects loaded');
    });
    
})();