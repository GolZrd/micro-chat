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
            localStorage.setItem('username', claims.username);
            console.log('✅ Token saved, username:', claims.username);
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
    const appContainer = document.getElementById('appContainer');
    const protectedContent = document.querySelectorAll('.protected-content');
    const protectedNav = document.querySelectorAll('.protected-nav');
    const guestOnly = document.querySelectorAll('.guest-only');
    const guestNav = document.querySelectorAll('.guest-nav');

    // Footer sidebar элементы
    const footerUser = document.getElementById('sidebarFooterUser');
    const footerUsername = document.getElementById('footerUsername');
    const footerInitials = document.getElementById('footerAvatarInitials');
    const footerIndicator = document.getElementById('footerOnlineIndicator');
    const footerStatus = document.getElementById('footerStatus');

    if (isAuth) {
        const username = TokenManager.getUsername();
        const userId = TokenManager.getUserId();

        // Классы для layout
        document.body.classList.add('authenticated');
        if (appContainer) appContainer.classList.remove('guest-mode');

        // Статус авторизации
        if (statusEl) {
            statusEl.innerHTML = `✅ Вы авторизованы как <strong>${username}</strong> <small>(ID: ${userId})</small>`;
            statusEl.style.color = 'green';
        }

        // Показываем/скрываем элементы
        protectedContent.forEach(el => el.style.removeProperty('display'));
        protectedNav.forEach(el => el.style.removeProperty('display'));
        guestOnly.forEach(el => el.style.display = 'none');
        guestNav.forEach(el => el.style.display = 'none');

        // Footer профиль
        if (footerUser) footerUser.style.display = 'flex';
        if (footerUsername) footerUsername.textContent = username;
        if (footerInitials) footerInitials.textContent = username.substring(0, 2).toUpperCase();
        if (footerIndicator) {
            footerIndicator.classList.remove('offline');
            footerIndicator.classList.add('online');
        }
        if (footerStatus) {
            footerStatus.textContent = 'В сети';
            footerStatus.classList.add('online');
        }

        // Загружаем данные
        loadChatCount();
        loadFriendsWithPresence();
        loadFriendRequests();
        startPresence();

    } else {
        // Классы для layout
        document.body.classList.remove('authenticated');
        if (appContainer) appContainer.classList.add('guest-mode');

        // Статус
        if (statusEl) {
            statusEl.textContent = '❌ Не авторизован';
            statusEl.style.color = 'red';
        }

        // Скрываем элементы
        protectedContent.forEach(el => el.style.display = 'none');
        protectedNav.forEach(el => el.style.display = 'none');
        guestOnly.forEach(el => el.style.removeProperty('display'));
        guestNav.forEach(el => el.style.removeProperty('display'));

        // Footer профиль
        if (footerUser) footerUser.style.display = 'none';
        if (footerIndicator) {
            footerIndicator.classList.remove('online');
            footerIndicator.classList.add('offline');
        }
        if (footerStatus) {
            footerStatus.textContent = 'Не в сети';
            footerStatus.classList.remove('online');
        }

        // Сбрасываем
        updateChatCount(0);
        stopPresence();
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
                username: document.getElementById('reg_username').value,
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
                    await loadFriendsWithPresence();
                    startPresence();
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
            
            const name = document.getElementById('chat_name').value.trim();
            const usernamesStr = document.getElementById('chat_usernames').value.trim();
            const chatType = document.querySelector('input[name="chat_type"]:checked')?.value || 'private';

            if (!usernamesStr) {
                showToast({ type: 'error', title: 'Ошибка', message: 'Укажите участников' });
                return;
            }

            const usernames = usernamesStr.split(',').map(u => u.trim()).filter(u => u);

            try {
                const response = await apiRequest('/api/chat/create', {
                    method: 'POST',
                    body: JSON.stringify({
                        name: name,
                        usernames: usernames,
                        is_public: chatType === 'public'
                    })
                });

                const data = await response.json();

                if (!response.ok) {
                    showToast({ type: 'error', title: 'Ошибка', message: data.error });
                    return;
                }

                showToast({ type: 'success', title: 'Чат создан', message: `ID: ${data.chat_id}` });
                closeCreateChatModal();
                loadMyChats();
                loadChatCount();

            } catch (error) {
                showToast({ type: 'error', title: 'Ошибка', message: error.message });
            }
              
        });
    }
});

// ==================== ОТКРЫТЫЕ ЧАТЫ ====================

let searchPublicTimeout = null;

function searchPublicChats(query) {
    clearTimeout(searchPublicTimeout);
    searchPublicTimeout = setTimeout(() => loadPublicChats(query), 300);
}

async function loadPublicChats(search = '') {
    const container = document.getElementById('publicChats');
    if (!container) return;

    container.innerHTML = '<div class="loading-chats">Поиск...</div>';

    try {
        const url = search
            ? `/api/chat/public?search=${encodeURIComponent(search)}`
            : '/api/chat/public';

        const response = await apiRequest(url);
        const data = await response.json();

        if (!response.ok) {
            container.innerHTML = `<div class="no-chats"><p>${data.error}</p></div>`;
            return;
        }

        const chats = data.chats || [];

        if (chats.length === 0) {
            container.innerHTML = '<div class="no-chats"><p>Открытых чатов не найдено</p></div>';
            return;
        }

        let html = '<div class="chats-grid">';
        chats.forEach(chat => {
            html += `
                <div class="chat-card chat-card--public">
                    <div class="chat-card__public-icon">
                        <i class="fas fa-globe"></i>
                    </div>
                    <div class="chat-card__body">
                        <div class="chat-card__name">${escapeHtml(chat.name)}</div>
                        <div class="chat-card__members">
                            <i class="fas fa-users"></i>
                            <span>${chat.member_count} участников</span>
                            <span>· создал ${escapeHtml(chat.creator_name)}</span>
                        </div>
                        <div class="chat-card__meta">${formatChatDate(chat.created_at)}</div>
                    </div>
                    <button onclick="joinPublicChat(${chat.id}, '${escapeHtml(chat.name)}')" class="btn-join">
                        <i class="fas fa-sign-in-alt"></i>
                        Войти
                    </button>
                </div>
            `;
        });
        html += '</div>';

        container.innerHTML = html;

    } catch (error) {
        container.innerHTML = `<div class="no-chats"><p>Ошибка: ${error.message}</p></div>`;
    }
}

async function joinPublicChat(chatId, chatName) {
    try {
        const response = await apiRequest('/api/chat/join', {
            method: 'POST',
            body: JSON.stringify({ chat_id: chatId })
        });

        const data = await response.json();

        if (!response.ok) {
            showToast({ type: 'error', title: 'Ошибка', message: data.error });
            return;
        }

        showToast({ type: 'success', title: 'Вы присоединились!', message: chatName });
        loadPublicChats();
        loadChatCount();

    } catch (error) {
        showToast({ type: 'error', title: 'Ошибка', message: error.message });
    }
}

// ==================== УПРАВЛЕНИЕ УЧАСТНИКАМИ ====================

async function addMemberToChat(chatId) {
    const username = prompt('Введите имя пользователя:');
    if (!username) return;

    try {
        const response = await apiRequest('/api/chat/add-member', {
            method: 'POST',
            body: JSON.stringify({
                chat_id: chatId,
                username: username.trim()
            })
        });

        if (response.ok) {
            showToast({ type: 'success', title: 'Добавлен', message: `${username} добавлен в чат` });
        } else {
            const data = await response.json();
            showToast({ type: 'error', title: 'Ошибка', message: data.error });
        }
    } catch (error) {
        showToast({ type: 'error', title: 'Ошибка', message: error.message });
    }
}

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

    if (chatsDiv) {
        chatsDiv.innerHTML = '<div class="loading-chats">Загрузка чатов...</div>';
    }

    try {
        const response = await apiRequest('/api/chat/my');
        const data = await response.json();

        if (!response.ok) {
            if (chatsDiv) {
                chatsDiv.innerHTML = `<div class="no-chats"><p>${data.error || 'Ошибка загрузки'}</p></div>`;
            }
            updateChatCount(0);
            return;
        }

        let chats = data.chats || [];
        chats = chats.filter(chat => chat && chat.id);

        updateChatCount(chats.length);

        if (!chatsDiv) return;

        if (chats.length === 0) {
            chatsDiv.innerHTML = '<div class="no-chats"><p>У вас пока нет чатов. Создайте первый!</p></div>';
            return;
        }

        // Разделяем на личные и групповые
        const directChats = chats.filter(c => c.is_direct);
        const groupChats = chats.filter(c => !c.is_direct);

        let html = '';

        // Личные чаты
        if (directChats.length > 0) {
            html += `
                <div class="chats-section">
                    <div class="chats-section-title">
                        <i class="fas fa-user"></i>
                        <span>Личные сообщения</span>
                        <span class="chats-section-count">${directChats.length}</span>
                    </div>
                    <div class="chats-grid">
            `;

            directChats.forEach(chat => {
                html += renderDirectChatCard(chat);
            });

            html += '</div></div>';
        }

        // Групповые чаты
        if (groupChats.length > 0) {
            html += `
                <div class="chats-section">
                    <div class="chats-section-title">
                        <i class="fas fa-users"></i>
                        <span>Групповые чаты</span>
                        <span class="chats-section-count">${groupChats.length}</span>
                    </div>
                    <div class="chats-grid">
            `;

            groupChats.forEach(chat => {
                html += renderGroupChatCard(chat);
            });

            html += '</div></div>';
        }

        chatsDiv.innerHTML = html;

    } catch (error) {
        console.error('❌ Error:', error);
        if (chatsDiv) {
            chatsDiv.innerHTML = `<div class="no-chats"><p>Ошибка: ${error.message}</p></div>`;
        }
        updateChatCount(0);
    }
}

function renderDirectChatCard(chat) {
    const chatId = chat.id;
    const otherUser = getChatDisplayName(chat);
    const initials = otherUser.substring(0, 2).toUpperCase();
    const createdDate = formatChatDate(chat.created_at);

    return `
        <a href="/chat?id=${chatId}" class="chat-card chat-card--direct">
            <div class="chat-card__avatar">
                <span class="chat-card__initials">${escapeHtml(initials)}</span>
                <span class="chat-card__online-dot"></span>
            </div>
            <div class="chat-card__body">
                <div class="chat-card__name">${escapeHtml(otherUser)}</div>
                <div class="chat-card__meta">Личный чат · ${createdDate}</div>
            </div>
            <div class="chat-card__actions">
                <button 
                    onclick="event.preventDefault(); event.stopPropagation(); deleteChat(${chatId})" 
                    class="chat-card__delete"
                    title="Удалить чат">
                    <i class="fas fa-trash-alt"></i>
                </button>
                <i class="fas fa-chevron-right chat-card__arrow"></i>
            </div>
        </a>
    `;
}

function renderGroupChatCard(chat) {
    const chatId = chat.id;
    const users = chat.usernames || [];
    const chatName = chat.name || `Чат #${chatId}`;
    const createdDate = formatChatDate(chat.created_at);
    const memberCount = users.length;

    // Показываем первые 3 аватара
    const avatarUsers = users.slice(0, 3);
    const extraCount = users.length - 3;

    let avatarsHtml = '<div class="chat-card__avatars-stack">';
    avatarUsers.forEach((user, i) => {
        const userInitials = user.substring(0, 2).toUpperCase();
        avatarsHtml += `
            <div class="chat-card__stacked-avatar" style="z-index: ${3 - i}">
                ${escapeHtml(userInitials)}
            </div>
        `;
    });
    if (extraCount > 0) {
        avatarsHtml += `
            <div class="chat-card__stacked-avatar chat-card__stacked-extra" style="z-index: 0">
                +${extraCount}
            </div>
        `;
    }
    avatarsHtml += '</div>';

    return `
        <a href="/chat?id=${chatId}" class="chat-card chat-card--group">
            ${avatarsHtml}
            <div class="chat-card__body">
                <div class="chat-card__name">${escapeHtml(chatName)}</div>
                <div class="chat-card__members">
                    <i class="fas fa-users"></i>
                    <span>${memberCount} участников</span>
                </div>
                <div class="chat-card__meta">${createdDate}</div>
            </div>
            <div class="chat-card__actions">
                <button 
                    onclick="event.preventDefault(); event.stopPropagation(); deleteChat(${chatId})" 
                    class="chat-card__delete"
                    title="Удалить чат">
                    <i class="fas fa-trash-alt"></i>
                </button>
                <i class="fas fa-chevron-right chat-card__arrow"></i>
            </div>
        </a>
    `;
}

// Получить отображаемое имя чата
function getChatDisplayName(chat) {
    if (!chat.is_direct) {
        return chat.name || `Чат #${chat.id}`;
    }

    const currentUsername = TokenManager.getUsername();
    const users = chat.usernames || [];
    const otherUser = users.find(u => u !== currentUsername);

    return otherUser || chat.name || `Чат #${chat.id}`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Универсальный парсер даты (поддерживает оба формата)
function formatChatDate(createdAt) {
    if (!createdAt) return 'N/A';

    let date;

    // Формат proto: {"seconds": 123456, "nanos": 0}
    if (createdAt.seconds) {
        date = new Date(createdAt.seconds * 1000);
    }
    // Формат ISO string: "2025-06-21T12:00:00Z"
    else if (typeof createdAt === 'string') {
        date = new Date(createdAt);
    }
    else {
        return 'N/A';
    }

    if (isNaN(date.getTime())) return 'N/A';

    return date.toLocaleString('ru-RU', {
        day: 'numeric',
        month: 'short',
        hour: '2-digit',
        minute: '2-digit'
    });
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
    // Обновляем счетчик каждые 5 минут
    chatCountInterval = setInterval(() => {
        if (TokenManager.isAuthenticated()) {
            loadChatCount();
        }
    }, 300000); // 300 секунд
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
                <p><strong>Имя:</strong> <span>${user.username}</span></p>
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

// ==================== FRIENDS ====================

let friendsList = [];
let friendRequests = [];
let searchTimeout = null;
let currentDropdownFriend = null; // Хранит данные друга для dropdown

// Инициализация dropdown
function initFriendDropdown() {
    const dropdown = document.getElementById('friendDropdown');
    if (!dropdown) return;

    // Обработчики для пунктов меню
    dropdown.querySelectorAll('.dropdown-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.stopPropagation();
            const action = item.dataset.action;
            
            if (currentDropdownFriend) {
                switch (action) {
                    case 'chat':
                        startChatWithFriend(currentDropdownFriend.user_id, currentDropdownFriend.username);
                        break;
                    case 'remove':
                        removeFriend(currentDropdownFriend.user_id, currentDropdownFriend.username);
                        break;
                }
            }
            
            closeFriendDropdown();
        });
    });

    // Закрытие при клике вне
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.friend-dropdown') && !e.target.closest('.btn-more')) {
            closeFriendDropdown();
        }
    });

    // Закрытие при скролле
    document.querySelector('.right-sidebar .sidebar-content')?.addEventListener('scroll', () => {
        closeFriendDropdown();
    });

    // Закрытие при нажатии Escape
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeFriendDropdown();
        }
    });
}

function openFriendDropdown(button, friend) {
    const dropdown = document.getElementById('friendDropdown');
    if (!dropdown) return;

    currentDropdownFriend = friend;

    // Убираем active с других кнопок
    document.querySelectorAll('.btn-more.active').forEach(btn => {
        btn.classList.remove('active');
    });

    // Добавляем active к текущей кнопке
    button.classList.add('active');

    // Позиционируем dropdown
    const rect = button.getBoundingClientRect();
    const dropdownHeight = 120; // Примерная высота
    
    // Проверяем, помещается ли снизу
    const spaceBelow = window.innerHeight - rect.bottom;
    const showAbove = spaceBelow < dropdownHeight;

    dropdown.style.left = `${rect.left - 150 + rect.width}px`; // Выравниваем по правому краю кнопки
    
    if (showAbove) {
        dropdown.style.top = `${rect.top - dropdownHeight - 5}px`;
    } else {
        dropdown.style.top = `${rect.bottom + 5}px`;
    }

    dropdown.classList.add('open');
}

function closeFriendDropdown() {
    const dropdown = document.getElementById('friendDropdown');
    if (dropdown) {
        dropdown.classList.remove('open');
    }
    
    document.querySelectorAll('.btn-more.active').forEach(btn => {
        btn.classList.remove('active');
    });
    
    currentDropdownFriend = null;
}

// Загрузка друзей
async function loadFriends() {
    try {
        const response = await apiRequest('/api/friends');
        const data = await response.json();

        if (!response.ok) {
            console.error('Failed to load friends:', data.error);
            return;
        }

        friendsList = data.friends || [];
        renderFriends();
    } catch (error) {
        console.error('Error loading friends:', error);
    }
}

function renderFriends() {
    const container = document.getElementById('friendsList');
    const totalCount = document.getElementById('totalFriendsCount');

    if (!container) return;

    // Обновляем счётчик
    if (totalCount) {
        totalCount.textContent = friendsList.length;
    }

    if (friendsList.length === 0) {
        container.innerHTML = '<p class="empty-text">Список друзей пуст</p>';
        return;
    }

    // Сортируем: онлайн сначала, потом по имени
    const sorted = [...friendsList].sort((a, b) => {
        if (a.is_online && !b.is_online) return -1;
        if (!a.is_online && b.is_online) return 1;
        return a.username.localeCompare(b.username);
    });

    container.innerHTML = '';

    sorted.forEach(friend => {
        const item = createFriendItem(friend);
        container.appendChild(item);
    });
}

function createFriendItem(friend) {
    const template = document.getElementById('friendItemTemplate');
    const item = template.content.cloneNode(true);
    const container = item.querySelector('.friend-item');

    const initials = friend.username.substring(0, 2).toUpperCase();

    item.querySelector('.avatar-initials').textContent = initials;
    item.querySelector('.friend-name').textContent = friend.username;

    const indicator = item.querySelector('.online-indicator');
    const statusEl = item.querySelector('.friend-status');

    if (friend.is_online) {
        indicator.classList.remove('offline');
        indicator.classList.add('online');
        statusEl.textContent = 'В сети';
        statusEl.classList.add('online');
        statusEl.classList.remove('offline');
    } else {
        indicator.classList.remove('online');
        indicator.classList.add('offline');
        statusEl.textContent = formatLastSeen(friend.last_seen_at);
        statusEl.classList.remove('online');
        statusEl.classList.add('offline');
    }

    // Клик по элементу — открыть чат
    container.addEventListener('click', (e) => {
        if (!e.target.closest('.btn-more')) {
            startChatWithFriend(friend.user_id, friend.username);
        }
    });

    // Кнопка "ещё" (3 точки)
    const moreBtn = item.querySelector('.btn-more');
    moreBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        
        // Если уже открыто для этого друга — закрываем
        if (moreBtn.classList.contains('active')) {
            closeFriendDropdown();
        } else {
            openFriendDropdown(moreBtn, friend);
        }
    });

    return item;
}

// Загрузка заявок
async function loadFriendRequests() {
    try {
        const response = await apiRequest('/api/friends/requests');
        const data = await response.json();

        if (!response.ok) {
            console.error('Failed to load requests:', data.error);
            return;
        }

        friendRequests = data.requests || [];
        renderFriendRequests();
    } catch (error) {
        console.error('Error loading friend requests:', error);
    }
}

function renderFriendRequests() {
    const section = document.getElementById('friendRequestsSection');
    const container = document.getElementById('friendRequests');
    const badge = document.getElementById('requestsBadge');
    const navBadge = document.getElementById('friendRequestsCount');

    // Обновляем бейджи
    if (badge) badge.textContent = friendRequests.length;
    if (navBadge) {
        navBadge.textContent = friendRequests.length;
        navBadge.style.display = friendRequests.length > 0 ? 'inline-flex' : 'none';
    }

    // Показываем/скрываем секцию
    if (section) {
        section.style.display = friendRequests.length > 0 ? 'block' : 'none';
    }

    if (!container) return;

    container.innerHTML = '';

    friendRequests.forEach(request => {
        const item = createRequestItem(request);
        container.appendChild(item);
    });
}

function createRequestItem(request) {
    const template = document.getElementById('friendRequestItemTemplate');
    const item = template.content.cloneNode(true);

    const initials = request.from_username.substring(0, 2).toUpperCase();
    const date = new Date(request.created_at).toLocaleDateString('ru-RU');

    item.querySelector('.avatar-initials').textContent = initials;
    item.querySelector('.request-name').textContent = request.from_username;
    item.querySelector('.request-date').textContent = date;

    item.querySelector('.btn-accept').addEventListener('click', (e) => {
        e.stopPropagation();
        acceptFriendRequest(request.id);
    });

    item.querySelector('.btn-reject').addEventListener('click', (e) => {
        e.stopPropagation();
        rejectFriendRequest(request.id);
    });

    return item;
}

// Поиск пользователей
async function searchUsers(query) {
    const resultsDiv = document.getElementById('searchResults');
    if (!resultsDiv) return;

    if (query.length < 2) {
        resultsDiv.innerHTML = '';
        return;
    }

    resultsDiv.innerHTML = '<p class="empty-text">Поиск...</p>';

    try {
        const response = await apiRequest(`/api/users/search?q=${encodeURIComponent(query)}&limit=10`);
        const data = await response.json();

        if (!response.ok) {
            resultsDiv.innerHTML = `<p class="empty-text">${data.error}</p>`;
            return;
        }

        const users = data.users || [];

        if (users.length === 0) {
            resultsDiv.innerHTML = '<p class="empty-text">Никого не найдено</p>';
            return;
        }

        resultsDiv.innerHTML = '';
        users.forEach(user => {
            resultsDiv.appendChild(createSearchResultItem(user));
        });
    } catch (error) {
        console.error('Search error:', error);
        resultsDiv.innerHTML = '<p class="empty-text">Ошибка поиска</p>';
    }
}

function createSearchResultItem(user) {
    const template = document.getElementById('searchResultItemTemplate');
    const item = template.content.cloneNode(true);

    const initials = user.username.substring(0, 2).toUpperCase();

    item.querySelector('.avatar-initials').textContent = initials;
    item.querySelector('.result-name').textContent = user.username;

    const statusText = item.querySelector('.result-status');
    const addBtn = item.querySelector('.btn-add-friend');

    switch (user.friendship_status) {
        case 'friends':
            statusText.textContent = 'В друзьях';
            addBtn.innerHTML = '<i class="fas fa-check"></i>';
            addBtn.disabled = true;
            break;
        case 'pending_sent':
            statusText.textContent = 'Заявка отправлена';
            addBtn.innerHTML = '<i class="fas fa-clock"></i>';
            addBtn.disabled = true;
            break;
        case 'pending_received':
            statusText.textContent = 'Принять заявку';
            addBtn.innerHTML = '<i class="fas fa-user-check"></i>';
            break;
        default:
            statusText.textContent = '';
            addBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                sendFriendRequest(user.id, user.username);
            });
    }

    return item;
}

// Действия
async function sendFriendRequest(userId, username) {
    try {
        const response = await apiRequest('/api/friends/request', {
            method: 'POST',
            body: JSON.stringify({ user_id: userId })
        });

        if (response.ok) {
            showToast({
                type: 'success',
                title: 'Заявка отправлена',
                message: `Заявка отправлена ${username}`
            });
            const input = document.getElementById('friendSearchInput');
            if (input && input.value) searchUsers(input.value);
        } else {
            const data = await response.json();
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: data.error
            });
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

async function acceptFriendRequest(requestId) {
    try {
        const response = await apiRequest(`/api/friends/accept/${requestId}`, {
            method: 'POST'
        });

        if (response.ok) {
            showToast({
                type: 'success',
                title: 'Принято',
                message: 'Теперь вы друзья!'
            });
            loadFriendRequests();
            loadFriends();
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

async function rejectFriendRequest(requestId) {
    try {
        await apiRequest(`/api/friends/reject/${requestId}`, {
            method: 'POST'
        });
        loadFriendRequests();
    } catch (error) {
        console.error('Error:', error);
    }
}

async function removeFriend(friendId, friendName) {
    if (!confirm(`Удалить ${friendName} из друзей?`)) return;

    try {
        const response = await apiRequest(`/api/friends/${friendId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showToast({
                type: 'success',
                title: 'Удалено',
                message: `${friendName} удалён из друзей`
            });
            loadFriends();
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

async function startChatWithFriend(userId, username) {
    try {
        showToast({
            type: 'info',
            title: 'Открываем чат',
            message: `Подключение к чату с ${username}...`
        });

        const response = await apiRequest('/api/chat/direct', {
            method: 'POST',
            body: JSON.stringify({
                user_id: userId,
                username: username
            })
        });

        const data = await response.json();

        if (!response.ok) {
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: data.error || 'Не удалось открыть чат'
            });
            return;
        }

        if (data.created) {
            showToast({
                type: 'success',
                title: 'Чат создан',
                message: `Новый чат с ${username}`
            });
            // Обновляем счётчик чатов
            loadChatCount();
        }

        // Переходим в чат
        window.location.href = `/chat?id=${data.chat_id}`;

    } catch (error) {
        console.error('Error starting chat:', error);
        showToast({
            type: 'error',
            title: 'Ошибка',
            message: 'Не удалось открыть чат'
        });
    }
}

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    // Инициализируем dropdown
    initFriendDropdown();

    // Поиск с debounce
    const searchInput = document.getElementById('friendSearchInput');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                searchUsers(e.target.value.trim());
            }, 300);
        });
    }

    // Загружаем данные если авторизован
    if (typeof TokenManager !== 'undefined' && TokenManager.isAuthenticated()) {
        loadFriendRequests();
        loadFriendsWithPresence()
        startPresence();
    }
});

// Загружаем друзей и СРАЗУ обновляем их статусы
async function loadFriendsWithPresence() {
    await loadFriends();
    await updateFriendsPresence();
}

// ==================== PRESENCE SYSTEM ====================

let heartbeatInterval = null;
let presenceInterval = null;

// ========== Запуск/Остановка ==========

function startPresence() {
    console.log('🟢 Starting presence system');

    // Сразу отправляем heartbeat
    sendHeartbeat();

    // Heartbeat каждые 30 секунд
    heartbeatInterval = setInterval(sendHeartbeat, 30000);

    // Обновляем статусы друзей каждые 15 секунд
    presenceInterval = setInterval(updateFriendsPresence, 15000);

    // При закрытии вкладки/браузера
    window.addEventListener('beforeunload', stopPresence);

    // Оптимизация: реже heartbeat когда вкладка скрыта
    document.addEventListener('visibilitychange', handleVisibilityChange);
}

function stopPresence() {
    console.log('🔴 Stopping presence system');

    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
    }
    if (presenceInterval) {
        clearInterval(presenceInterval);
        presenceInterval = null;
    }
}

function handleVisibilityChange() {
    if (document.hidden) {
        // Вкладка скрыта — heartbeat реже (60 сек)
        clearInterval(heartbeatInterval);
        heartbeatInterval = setInterval(sendHeartbeat, 60000);

        // Статусы друзей реже (30 сек)
        clearInterval(presenceInterval);
        presenceInterval = setInterval(updateFriendsPresence, 30000);
    } else {
        // Вкладка активна — возвращаем нормальную частоту
        clearInterval(heartbeatInterval);
        sendHeartbeat(); // Сразу отправить
        heartbeatInterval = setInterval(sendHeartbeat, 30000);

        clearInterval(presenceInterval);
        updateFriendsPresence(); // Сразу обновить
        presenceInterval = setInterval(updateFriendsPresence, 15000);
    }
}

// ========== API Вызовы ==========

async function sendHeartbeat() {
    try {
        await apiRequest('/api/presence/heartbeat', {
            method: 'POST',
            body: JSON.stringify({})
        });
    } catch (error) {
        // Тихо проглатываем — heartbeat не критичен
        console.debug('Heartbeat failed:', error);
    }
}

async function updateFriendsPresence() {
    // Не обновляем если нет друзей
    if (!friendsList || friendsList.length === 0) return;

    // Собираем user_id друзей
    const userIds = friendsList
        .map(f => f.user_id)
        .filter(id => id && id > 0);

    if (userIds.length === 0) return;

    try {
        const response = await apiRequest('/api/presence/friends', {
            method: 'POST',
            body: JSON.stringify({ user_ids: userIds })
        });

        if (!response.ok) return;

        const data = await response.json();
        const presences = data.presences || [];

        console.log('📡 Presence response:', presences);

        // Создаём карту user_id → presence
        const presenceMap = {};
        presences.forEach(p => {
            presenceMap[p.user_id] = p;
        });

        // Обновляем данные друзей
        let changed = false;
        friendsList.forEach(friend => {
            const p = presenceMap[friend.user_id];
            const wasOnline = friend.is_online;

            if (p) {
                friend.is_online = p.is_online;
                friend.last_seen_at = p.last_seen_at || null;
            } else {
                friend.is_online = false;
            }

            if (wasOnline !== friend.is_online) {
                changed = true;
            }
        });

        console.log('👥 Updated friends:', friendsList.map(f => 
            `${f.username}: ${f.is_online ? 'online' : 'offline'}`
        ));

        // Перерисовываем только если что-то изменилось
        if (changed) {
            console.log('👥 Friends presence updated');
            renderFriends();
        }

    } catch (error) {
        console.debug('Failed to update presence:', error);
    }
}

// ========== Форматирование "последний раз в сети" ==========

function formatLastSeen(lastSeenAt) {
    if (!lastSeenAt) return 'Не в сети';

    const date = new Date(lastSeenAt);
    if (isNaN(date.getTime())) return 'Не в сети';

    const now = new Date();
    const diffMs = now - date;
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffSec < 30) return 'Только что';
    if (diffMin < 1) return 'Меньше минуты назад';
    if (diffMin === 1) return 'Минуту назад';
    if (diffMin < 5) return `${diffMin} минуты назад`;
    if (diffMin < 60) return `${diffMin} мин назад`;
    if (diffHours === 1) return 'Час назад';
    if (diffHours < 24) return `${diffHours} ч назад`;
    if (diffDays === 1) return 'Вчера';
    if (diffDays < 7) return `${diffDays} дн назад`;

    return date.toLocaleDateString('ru-RU', {
        day: 'numeric',
        month: 'short'
    });
}

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