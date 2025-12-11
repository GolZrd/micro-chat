// ==================== TOKEN MANAGER ====================
const TokenManager = {
    ACCESS_TOKEN_TTL: 15 * 60 * 1000,
    refreshTimer: null,

    getAccessToken() {
        return localStorage.getItem('access_token');
    },
    
    setAccessToken(token) {
        this.stopRefreshTimer();
        localStorage.setItem('access_token', token);
        localStorage.setItem('token_set_at', Date.now().toString());
        
        try {
            const claims = this.decodeJWT(token);
            localStorage.setItem('username', claims.name);
            console.log('✅ Token saved, username:', claims.name);
            this.startRefreshTimer();
        } catch (e) {
            console.error('❌ Failed to decode token:', e);
        }
    },

    startRefreshTimer() {
        const refreshInterval = this.ACCESS_TOKEN_TTL - 30000;
        console.log(`⏰ Auto-refresh scheduled in ${refreshInterval/1000} seconds`);
        
        this.refreshTimer = setTimeout(async () => {
            console.log('⏰ Auto-refreshing token...');
            const refreshed = await refreshAccessToken();
            if (!refreshed) {
                console.error('❌ Auto-refresh failed');
                alert('⚠️ Сессия истекла. Пожалуйста, войдите снова.');
                this.clear();
                window.location.href = '/';
            }
        }, refreshInterval);
    },
    
    stopRefreshTimer() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
            this.refreshTimer = null;
        }
    },
    
    decodeJWT(token) {
        const base64Url = token.replace('Bearer ', '').split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(base64));
    },
    
    getUsername() {
        return localStorage.getItem('username');
    },
    
    removeAccessToken() {
        this.stopRefreshTimer();
        localStorage.removeItem('access_token');
        localStorage.removeItem('token_set_at');
    },
    
    getUserId() {
        return localStorage.getItem('user_id');
    },
    
    setUserId(id) {
        localStorage.setItem('user_id', id);
    },
    
    isAuthenticated() {
        return !!this.getAccessToken();
    },

    isTokenExpired() {
        const tokenSetAt = localStorage.getItem('token_set_at');
        if (!tokenSetAt) return true;
        
        const elapsed = Date.now() - parseInt(tokenSetAt);
        return elapsed >= this.ACCESS_TOKEN_TTL;
    },
    
    clear() {
        this.stopRefreshTimer();
        localStorage.removeItem('access_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('username');
        localStorage.removeItem('token_set_at');
    }
};

// ==================== API FUNCTIONS ====================
async function refreshAccessToken() {
    try {
        console.log('🔄 Refreshing access token...');
        
        const response = await fetch('/api/refresh', {
            method: 'POST',
            credentials: 'include'
        });
        
        if (response.ok) {
            const data = await response.json();
            
            if (data.access_token) {
                let accessToken = data.access_token;
                if (!accessToken.startsWith('Bearer ')) {
                    accessToken = 'Bearer ' + accessToken;
                }
                
                TokenManager.setAccessToken(accessToken);
                console.log('✅ Access token refreshed successfully');
                return true;
            } else {
                console.error('❌ No access_token in response');
                return false;
            }
        } else {
            const error = await response.json();
            console.error('❌ Refresh failed:', error.error);

            if (response.status === 401) {
                TokenManager.clear();
                updateAuthStatus();
            }
            return false;
        }
    } catch (error) {
        console.error('❌ Error refreshing token:', error);
        return false;
    }
}

async function apiRequest(url, options = {}) {
    if (TokenManager.isTokenExpired() && TokenManager.isAuthenticated()) {
        console.log('⚠️ Token expired, refreshing before request...');
        const refreshed = await refreshAccessToken();
        if (!refreshed) {
            throw new Error('Session expired');
        }
    }

    const token = TokenManager.getAccessToken();
    
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (token) {
        headers['Authorization'] = token;
    }
    
    let response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include'
    });
    
    if ((response.status === 401 || response.status === 403) && token) {
        console.log('⚠️ Access token expired (status ' + response.status + '), attempting refresh...');
        
        const refreshed = await refreshAccessToken();
        
        if (refreshed) {
            headers['Authorization'] = TokenManager.getAccessToken();
            response = await fetch(url, {
                ...options,
                headers,
                credentials: 'include'
            });
            
            console.log('✅ Request retried with new token');
        } else {
            console.error('❌ Failed to refresh token, logging out...');
            TokenManager.clear();
            updateAuthStatus();
            alert('⚠️ Сессия истекла. Пожалуйста, войдите снова.');
            window.location.href = '/';
            throw new Error('Session expired');
        }
    }
    
    return response;
}

async function checkTokenOnLoad() {
    if (!TokenManager.isAuthenticated()) {
        return;
    }

    if (TokenManager.isTokenExpired()) {
        console.log('⚠️ Token expired on page load, refreshing...');
        const refreshed = await refreshAccessToken();
        
        if (!refreshed) {
            console.log('❌ Could not refresh token on load');
            TokenManager.clear();
            updateAuthStatus();
            return;
        }
    }
    
    TokenManager.startRefreshTimer();
    updateAuthStatus();
}

function updateAuthStatus() {
    const isAuth = TokenManager.isAuthenticated();
    const statusEl = document.getElementById('status');
    const logoutBtn = document.getElementById('logoutBtn');
    const protectedContent = document.querySelectorAll('.protected-content');
    const protectedNav = document.querySelectorAll('.protected-nav');
    const guestOnly = document.querySelectorAll('.guest-only');
    const guestNav = document.querySelectorAll('.guest-nav');
    
    if (isAuth) {
        const username = TokenManager.getUsername();
        const userId = TokenManager.getUserId();
        
        if (statusEl) {
            statusEl.innerHTML = `✅ Вы авторизованы как <strong>${username}</strong> <small>(ID: ${userId})</small>`;
            statusEl.style.color = 'green';
        }
        
        if (logoutBtn) logoutBtn.style.display = 'inline-block';
        
        protectedContent.forEach(el => el.style.display = 'block');
        protectedNav.forEach(el => el.style.display = 'flex');
        guestOnly.forEach(el => el.style.display = 'none');
        guestNav.forEach(el => el.style.display = 'none');

        // ✅ ЗАГРУЖАЕМ КОЛИЧЕСТВО ЧАТОВ ПРИ АВТОРИЗАЦИИ
        loadChatCount();
    } else {
        if (statusEl) {
            statusEl.textContent = '❌ Не авторизован';
            statusEl.style.color = 'red';
        }
        
        if (logoutBtn) logoutBtn.style.display = 'none';
        
        protectedContent.forEach(el => el.style.display = 'none');
        protectedNav.forEach(el => el.style.display = 'none');
        guestOnly.forEach(el => el.style.display = 'block');
        guestNav.forEach(el => el.style.display = 'flex');

        // ✅ СБРАСЫВАЕМ СЧЕТЧИК ПРИ ВЫХОДЕ
        updateChatCount(0);
    }
}

// ==================== FORM HANDLERS ====================
document.addEventListener('DOMContentLoaded', () => {
    // Регистрация
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const data = {
                name: document.getElementById('reg_name').value,
                email: document.getElementById('reg_email').value,
                password: document.getElementById('reg_password').value,
                password_confirm: document.getElementById('reg_password_confirm').value
            };

            try {
                const response = await fetch('/api/register', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                const resultEl = document.getElementById('registerResult');
                
                if (response.ok) {
                    resultEl.innerHTML = `✅ Регистрация успешна! Теперь войдите в систему.`;
                    resultEl.style.background = '#d4edda';
                    resultEl.style.color = '#155724';
                    registerForm.reset();
                } else {
                    resultEl.innerHTML = `❌ Ошибка: ${result.error}`;
                    resultEl.style.background = '#f8d7da';
                    resultEl.style.color = '#721c24';
                }
            } catch (error) {
                document.getElementById('registerResult').innerHTML = `❌ ${error}`;
            }
        });
    }

    // Вход
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const data = {
                email: document.getElementById('login_email').value,
                password: document.getElementById('login_password').value
            };

            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    credentials: 'include',
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                const resultEl = document.getElementById('loginResult');
                
                if (response.ok) {
                    let accessToken = result.access_token;
                    if (!accessToken.startsWith('Bearer ')) {
                        accessToken = 'Bearer ' + accessToken;
                    }
                    
                    TokenManager.setAccessToken(accessToken);
                    TokenManager.setUserId(result.user_id);
                    
                    resultEl.innerHTML = `✅ Добро пожаловать, ${TokenManager.getUsername()}!`;
                    resultEl.style.background = '#d4edda';
                    resultEl.style.color = '#155724';
                    
                    updateAuthStatus();
                    loadChatCount();           // Загружаем количество чатов
                    startChatCountUpdater();   // Запускаем автообновление
                    loginForm.reset();
                } else {
                    resultEl.innerHTML = `❌ Ошибка: ${result.error}`;
                    resultEl.style.background = '#f8d7da';
                    resultEl.style.color = '#721c24';
                }
            } catch (error) {
                document.getElementById('loginResult').innerHTML = `❌ ${error}`;
            }
        });
    }

    // Создание чата
    const createChatForm = document.getElementById('createChatForm');
    if (createChatForm) {
        createChatForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const inputEl = document.getElementById('chat_usernames');
            const usernames = inputEl.value.split(',').map(s => s.trim()).filter(s => s.length > 0);
            const resultEl = document.getElementById('chatResult');

            // Валидация на клиенте
            if (usernames.length === 0) {
                showToast({
                    type: 'warning',
                    title: 'Внимание',
                    message: 'Введите имена пользователей через запятую'
                });
                return;
            }

            try {
                const response = await apiRequest('/api/chat/create', {
                    method: 'POST',
                    body: JSON.stringify({ usernames })
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    // Успех - показываем toast и обновляем UI
                    showToast({
                        type: 'success',
                        title: 'Чат создан!',
                        message: `Чат #${result.chat_id} успешно создан`
                    });

                    // Показываем ссылку на чат
                    if (resultEl) {
                        resultEl.innerHTML = `✅ <a href="/chat?id=${result.chat_id}" style="color: #007bff;">Открыть чат →</a>`;
                        resultEl.style.background = '#d4edda';
                        resultEl.style.color = '#155724';
                    }
                    
                    inputEl.value = '';

                    // Обновляем список чатов и счетчик
                    await loadMyChats();

                    const chatsSection = document.getElementById('chatsSection');
                    if (!chatsSection || !chatsSection.classList.contains('active')) {
                        loadChatCount();
                    }
                } else {
                    // Ошибка - обрабатываем по коду
                    handleCreateChatError(result, resultEl);
                }
            } catch (error) {
                console.error('Network error:', error);
                showToast({
                    type: 'error',
                    title: 'Ошибка сети',
                    message: 'Не удалось подключиться к серверу'
                });
                
                if (resultEl) {
                    resultEl.innerHTML = `❌ Ошибка сети`;
                    resultEl.style.background = '#f8d7da';
                    resultEl.style.color = '#721c24';
                }
            }
        });
    }
});

// Обработка ошибок создания чата
function handleCreateChatError(result, resultEl) {
    console.log('Create chat error:', result);

    switch (result.code) {
        case 'USERS_NOT_FOUND':
            showToast({
                type: 'error',
                title: 'Пользователи не найдены',
                message: 'Следующие пользователи не зарегистрированы:',
                users: result.not_found_users || [],
                duration: 10000
            });
            
            if (resultEl) {
                const usersList = (result.not_found_users || []).join(', ');
                resultEl.innerHTML = `❌ Пользователи не найдены: ${usersList}`;
                resultEl.style.background = '#f8d7da';
                resultEl.style.color = '#721c24';
            }
            break;

        case 'UNAUTHENTICATED':
            showToast({
                type: 'error',
                title: 'Сессия истекла',
                message: 'Пожалуйста, войдите снова'
            });
            
            // Можно добавить редирект на логин
            setTimeout(() => {
                TokenManager.clear();
                updateAuthStatus();
            }, 2000);
            break;

        case 'INVALID_ARGUMENT':
            showToast({
                type: 'warning',
                title: 'Ошибка валидации',
                message: result.error
            });
            
            if (resultEl) {
                resultEl.innerHTML = `❌ ${result.error}`;
                resultEl.style.background = '#fff3cd';
                resultEl.style.color = '#856404';
            }
            break;

        case 'PERMISSION_DENIED':
            showToast({
                type: 'error',
                title: 'Доступ запрещён',
                message: result.error || 'У вас нет прав для создания чата'
            });
            break;

        default:
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: result.error || 'Не удалось создать чат'
            });
            
            if (resultEl) {
                resultEl.innerHTML = `❌ ${result.error || 'Неизвестная ошибка'}`;
                resultEl.style.background = '#f8d7da';
                resultEl.style.color = '#721c24';
            }
    }
}

async function logout() {
    try {
        TokenManager.stopRefreshTimer();

        await fetch('/api/logout', {
            method: 'POST',
            credentials: 'include'
        });
        
        TokenManager.clear();
        updateAuthStatus();
        alert('✅ Вы вышли из системы');
        location.reload();
    } catch (error) {
        alert('❌ Ошибка при выходе: ' + error);
    }
}

async function deleteChat(chatId) {
    if (!confirm(`Вы уверены, что хотите удалить чат #${chatId}?`)) {
        return;
    }

    const chatCard = event.target.closest('.chat-card');
    
    try {
        if (chatCard) {
            chatCard.style.opacity = '0.5';
            chatCard.style.pointerEvents = 'none';
        }

        const response = await apiRequest(`/api/chat/delete/${chatId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            if (chatCard) {
                chatCard.classList.add('deleting');
                setTimeout(() => {
                    // ✅ ОБНОВЛЯЕМ СПИСОК И СЧЕТЧИК
                    loadMyChats();
                }, 300);
            } else {
                loadMyChats();
            }
        } else {
            const error = await response.json();
            alert('❌ Ошибка удаления: ' + (error.error || 'Неизвестная ошибка'));
            
            if (chatCard) {
                chatCard.style.opacity = '1';
                chatCard.style.pointerEvents = 'auto';
            }
        }
    } catch (error) {
        console.error('❌ Delete error:', error);
        alert('❌ Ошибка удаления чата: ' + error.message);
        
        if (chatCard) {
            chatCard.style.opacity = '1';
            chatCard.style.pointerEvents = 'auto';
        }
    }
}

async function loadMyChats() {
    const chatsDiv = document.getElementById('myChats');
    
    // Показываем загрузку только если элемент существует
    if (chatsDiv) {
        chatsDiv.innerHTML = '<p style="color: #666;">⏳ Загрузка...</p>';
    }
    
    try {
        const response = await apiRequest('/api/chat/my');
        const data = await response.json();
        
        console.log('📦 Server response:', data);
        
        if (!response.ok) {
            if (chatsDiv) {
                chatsDiv.innerHTML = `<p style="color: #dc3545;">❌ ${data.error || 'Ошибка загрузки'}</p>`;
            }
            // Обновляем счетчик на 0 при ошибке
            updateChatCount(0);
            return;
        }
        
        let chats = data.chats || [];
        chats = chats.filter(chat => chat && chat.id);
        
        console.log('✅ Filtered chats:', chats);
        
        // ✅ ОБНОВЛЯЕМ СЧЕТЧИК ЧАТОВ
        updateChatCount(chats.length);
        
        // Если элемент для отображения чатов не существует, выходим
        if (!chatsDiv) return;
        
        if (chats.length === 0) {
            chatsDiv.innerHTML = '<p style="color: #666;">📭 У вас пока нет чатов. Создайте первый!</p>';
            return;
        }
        
        let html = '<div class="chats-list">';
        chats.forEach(chat => {
            const chatId = chat.id;
            const users = chat.usernames || [];
            
            let createdDate = 'N/A';
            if (chat.created_at && chat.created_at.seconds) {
                const timestamp = chat.created_at.seconds * 1000;
                const date = new Date(timestamp);
                createdDate = date.toLocaleString('ru-RU', {
                    day: 'numeric',
                    month: 'short',
                    hour: '2-digit',
                    minute: '2-digit'
                });
            }
            
            const usersList = users.join(', ');
            
            html += `
                <div class="chat-card">
                    <div class="chat-card-header">
                        <h3>💬 Чат #${chatId}</h3>
                        <button 
                            onclick="event.stopPropagation(); deleteChat(${chatId})" 
                            class="btn-delete"
                            title="Удалить чат">
                            🗑️
                        </button>
                    </div>
                    <p><strong>👥 Участники:</strong> <span>${usersList}</span></p>
                    <p><strong>📅 Создан:</strong> <span>${createdDate}</span></p>
                    <a href="/chat?id=${chatId}" class="btn-open-chat" onclick="event.stopPropagation();">
                        Открыть чат →
                    </a>
                </div>
            `;
        });
        html += '</div>';
        
        chatsDiv.innerHTML = html;
        
    } catch (error) {
        console.error('❌ Error:', error);
        if (chatsDiv) {
            chatsDiv.innerHTML = `<p style="color: #dc3545;">❌ Ошибка: ${error.message}</p>`;
        }
        // Обновляем счетчик на 0 при ошибке
        updateChatCount(0);
    }
}

// ==================== ОБНОВЛЕНИЕ СЧЕТЧИКА ЧАТОВ ====================
function updateChatCount(count) {
    const chatCountEl = document.getElementById('chatCount');
    if (chatCountEl) {
        chatCountEl.textContent = count;
        
        // Добавляем визуальные эффекты
        if (count > 0) {
            chatCountEl.style.display = 'inline-flex';
            chatCountEl.classList.add('pulse');
            setTimeout(() => {
                chatCountEl.classList.remove('pulse');
            }, 600);
        } else {
            chatCountEl.style.display = 'none';
        }
    }
}

// ==================== ЗАГРУЗКА КОЛИЧЕСТВА ЧАТОВ (без UI) ====================
async function loadChatCount() {
    try {
        const response = await apiRequest('/api/chat/my');
        const data = await response.json();
        
        if (response.ok) {
            let chats = data.chats || [];
            chats = chats.filter(chat => chat && chat.id);
            updateChatCount(chats.length);
        } else {
            updateChatCount(0);
        }
    } catch (error) {
        console.error('❌ Error loading chat count:', error);
        updateChatCount(0);
    }
}

// ==================== АВТООБНОВЛЕНИЕ СЧЕТЧИКА ====================
let chatCountInterval = null;

function startChatCountUpdater() {
    // Обновляем счетчик каждые 30 секунд
    chatCountInterval = setInterval(() => {
        if (TokenManager.isAuthenticated()) {
            loadChatCount();
        }
    }, 30000); // 30 секунд
}

function stopChatCountUpdater() {
    if (chatCountInterval) {
        clearInterval(chatCountInterval);
        chatCountInterval = null;
    }
}

// Останавливаем при выходе
async function logout() {
    try {
        stopChatCountUpdater(); // Останавливаем обновление
        TokenManager.stopRefreshTimer();

        await fetch('/api/logout', {
            method: 'POST',
            credentials: 'include'
        });
        
        TokenManager.clear();
        updateAuthStatus();
        alert('✅ Вы вышли из системы');
        location.reload();
    } catch (error) {
        alert('❌ Ошибка при выходе: ' + error);
    }
}

async function loadUserInfo() {
    const userId = TokenManager.getUserId();
    if (!userId) {
        alert('User ID не найден');
        return;
    }

    const infoDiv = document.getElementById('userInfo');
    if (!infoDiv) return;
    
    infoDiv.innerHTML = '<p style="color: #72767d;">⏳ Загрузка...</p>';

    try {
        const response = await apiRequest(`/api/user/${userId}`);
        const user = await response.json();
        
        if (!response.ok) {
            infoDiv.innerHTML = `<p style="color: #ed4245;">❌ ${user.error}</p>`;
            return;
        }
        
        infoDiv.innerHTML = `
            <div class="user-info-card">
                <p><strong>ID:</strong> <span>${user.id}</span></p>
                <p><strong>Имя:</strong> <span>${user.name}</span></p>
                <p><strong>Email:</strong> <span>${user.email}</span></p>
                <p><strong>Роль:</strong> <span>${user.role || 'Пользователь'}</span></p>
                <p><strong>Создан:</strong> <span>${new Date(user.created_at).toLocaleString('ru-RU')}</span></p>
            </div>
        `;
    } catch (error) {
        infoDiv.innerHTML = `<p style="color: #ed4245;">❌ ${error}</p>`;
    }
}

document.addEventListener('visibilitychange', async () => {
    if (!document.hidden && TokenManager.isAuthenticated()) {
        if (TokenManager.isTokenExpired()) {
            console.log('⚠️ Token expired while away, refreshing...');
            await refreshAccessToken();
        }
    }
});

// ==================== MESSAGE INPUT HANDLER ====================
function initMessageInput() {
    console.log('🎯 Initializing message input...');
    
    const messageInput = document.getElementById('messageInput');
    const messageForm = document.getElementById('messageForm');
    const sendBtn = document.querySelector('.send-btn');
    const charCount = document.getElementById('charCount');
    const charCounter = document.querySelector('.char-counter');
    const emojiBtn = document.querySelector('.emoji-btn');
    const emojiPicker = document.querySelector('.emoji-picker');
    
    console.log('📦 Elements check:', {
        messageInput: !!messageInput,
        messageForm: !!messageForm,
        sendBtn: !!sendBtn,
        emojiBtn: !!emojiBtn,
        emojiPicker: !!emojiPicker
    });
    
    if (!messageInput || !messageForm) {
        console.log('ℹ️ Not a chat page, skipping message input initialization');
        return;
    }
    
    // AUTO-RESIZE
    function autoResize() {
        messageInput.style.height = '22px';
        const newHeight = Math.min(messageInput.scrollHeight, 178);
        messageInput.style.height = newHeight + 'px';
        
        const hasText = messageInput.value.trim().length > 0;
        if (sendBtn) {
            if (hasText) {
                sendBtn.classList.add('active');
                sendBtn.disabled = false;
            } else {
                sendBtn.classList.remove('active');
                sendBtn.disabled = true;
            }
        }
        
        if (charCount) {
            const count = messageInput.value.length;
            charCount.textContent = count;
            
            if (charCounter) {
                charCounter.classList.remove('warning', 'danger');
                if (count > 1800) {
                    charCounter.classList.add('danger');
                } else if (count > 1500) {
                    charCounter.classList.add('warning');
                }
            }
        }
    }
    
    messageInput.addEventListener('input', autoResize);
    messageInput.addEventListener('paste', () => setTimeout(autoResize, 10));
    
    // KEYBOARD
    messageInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            const text = this.value.trim();
            if (text) {
                console.log('⌨️ Enter pressed, triggering submit');
                messageForm.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
            }
        }
    });
    
    if (sendBtn) {
        sendBtn.disabled = true;
    }
    
    console.log('✅ Message input initialized');
    
    // EMOJI PICKER
    if (!emojiBtn || !emojiPicker) {
        console.log('ℹ️ Emoji picker not found');
        return;
    }
    
    console.log('😊 Initializing emoji picker...');
    
    emojiBtn.addEventListener('click', function(e) {
        e.preventDefault();
        e.stopPropagation();
        
        const isActive = emojiPicker.classList.contains('active');
        console.log('🎯 Emoji button clicked, state:', isActive ? 'closing' : 'opening');
        
        emojiPicker.classList.toggle('active');
        this.classList.toggle('active');
    });
    
    document.addEventListener('click', function(e) {
        if (!emojiPicker.contains(e.target) && !emojiBtn.contains(e.target)) {
            if (emojiPicker.classList.contains('active')) {
                console.log('👆 Closing emoji picker');
                emojiPicker.classList.remove('active');
                emojiBtn.classList.remove('active');
            }
        }
    });
    
    emojiPicker.addEventListener('click', function(e) {
        e.stopPropagation();
    });
    
    const emojiGrid = emojiPicker.querySelector('.emoji-grid');
    if (emojiGrid) {
        emojiGrid.addEventListener('click', function(e) {
            if (e.target.tagName === 'SPAN') {
                const emoji = e.target.textContent;
                console.log('😀 Emoji selected:', emoji);
                
                const cursorPos = messageInput.selectionStart || messageInput.value.length;
                const textBefore = messageInput.value.substring(0, cursorPos);
                const textAfter = messageInput.value.substring(cursorPos);
                
                messageInput.value = textBefore + emoji + textAfter;
                
                const newPos = cursorPos + emoji.length;
                messageInput.focus();
                messageInput.setSelectionRange(newPos, newPos);
                
                autoResize();
                console.log('✅ Emoji inserted');
            }
        });
    }
    
    const emojiSearch = emojiPicker.querySelector('.emoji-search');
    if (emojiSearch) {
        emojiSearch.addEventListener('input', function(e) {
            e.stopPropagation();
        });
        
        emojiSearch.addEventListener('click', function(e) {
            e.stopPropagation();
        });
    }
    
    console.log('✅ Emoji picker initialized');
}

// ==================== TOAST УВЕДОМЛЕНИЯ ====================

function showToast({ type = 'error', title, message, users = [], duration = 6000 }) {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️'
    };

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    // Генерируем HTML для списка пользователей
    let usersHtml = '';
    if (users.length > 0) {
        usersHtml = `
            <div class="toast-users">
                ${users.map(u => `<span class="toast-user">@${u}</span>`).join('')}
            </div>
        `;
    }

    toast.innerHTML = `
        <span class="toast-icon">${icons[type] || 'ℹ️'}</span>
        <div class="toast-content">
            <div class="toast-title">${title}</div>
            <div class="toast-message">${message}</div>
            ${usersHtml}
        </div>
        <button class="toast-close" onclick="closeToast(this)">×</button>
    `;

    container.appendChild(toast);

    // Автоматическое закрытие
    if (duration > 0) {
        setTimeout(() => closeToast(toast.querySelector('.toast-close')), duration);
    }
}

function closeToast(btn) {
    const toast = btn.closest('.toast');
    if (toast) {
        toast.classList.add('hiding');
        setTimeout(() => toast.remove(), 300);
    }
}

// ==================== INITIALIZATION ====================
document.addEventListener('DOMContentLoaded', () => {
    console.log('🚀 App initializing...');
    checkTokenOnLoad();
    updateAuthStatus();
    initMessageInput();
    // Запускаем автообновление если авторизован
    if (TokenManager.isAuthenticated()) {
        console.log('👤 Пользователь авторизован, загружаем чаты...');
        setTimeout(() => {
            loadChatCount();
            startChatCountUpdater();
        }, 100);
    }
    console.log('✅ App initialized');
});