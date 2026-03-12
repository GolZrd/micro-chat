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

    setAvatarUrl: function(url) {
        localStorage.setItem('avatar_url', url || '');
    },

    getAvatarUrl: function() {
        return localStorage.getItem('avatar_url') || '';
    },

    setBio(bio) {
        localStorage.setItem('user_bio', bio || '');
    },

    getBio() {
        return localStorage.getItem('user_bio') || '';
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

                     //Загружаем профиль (аватарку)
                    loadCurrentUserProfile();

                    loadChatCount();           // Загружаем количество чатов
                    startChatCountUpdater();   // Запускаем автообновление
                    NotificationManager.init(); // Инициализируем уведомления
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

// ==================== CHAT SYSTEM ====================

let activeChatId = null;
let activeChatWs = null;
let chatListData = []; // Кэш списка чатов

// ========== Загрузка списка чатов ==========

async function loadChatList() {
    const container = document.getElementById('chatListItems');
    if (!container) return;

    try {
        const response = await apiRequest('/api/chat/my');
        const data = await response.json();

        if (!response.ok) {
            container.innerHTML = '<div class="chat-list-empty"><p>Ошибка загрузки</p></div>';
            updateChatCount(0);
            return;
        }

        let chats = data.chats || [];
        chats = chats.filter(c => c && c.id);
        chatListData = chats;

        updateChatCount(chats.length);
        renderChatList(chats);

    } catch (error) {
        container.innerHTML = '<div class="chat-list-empty"><p>Ошибка</p></div>';
        updateChatCount(0);
    }
}

function renderChatList(chats) {
    const container = document.getElementById('chatListItems');
    if (!container) return;

    if (chats.length === 0) {
        container.innerHTML =
            '<div class="chat-list-empty">' +
            '<i class="fas fa-comments"></i>' +
            '<p>Нет чатов</p>' +
            '</div>';
        return;
    }

    chats.forEach(chat => {
        if (chat.unread_count && chat.unread_count > 0) {
            NotificationManager.unreadCounts[String(chat.id)] = chat.unread_count;
        }
    });

    const myId = String(TokenManager.getUserId());

    let html = '';
    chats.forEach(chat => {
        const chatName = getChatDisplayName(chat);
        const isDirect = chat.is_direct || false;
        const isPublic = chat.is_public || false;
        const isActive = chat.id === activeChatId;
        const members = chat.usernames || [];
        const memberIds = chat.member_ids || [];
        const memberAvatars = chat.member_avatars || {};
        const unreadCount = chat.unread_count || 0;

        // Для direct чата — находим собеседника
        let otherUserId = null;
        let chatAvatarHtml = '';

        if (isDirect && memberIds.length > 0) {
            otherUserId = memberIds.find(function(id) {
                return String(id) !== myId;
            });
            const otherAvatarUrl = otherUserId ? (memberAvatars[String(otherUserId)] || '') : '';
            chatAvatarHtml = avatarHtml(otherAvatarUrl, chatName, 44);
        } else {
            chatAvatarHtml = avatarHtml('', chatName, 44);
        }

        let avatarClass = 'chat-list-item__avatar--direct';
        let typeClass = 'chat-list-item__type--direct';
        let typeText = 'Личный';
        if (!isDirect && isPublic) {
            avatarClass = 'chat-list-item__avatar--public';
            typeClass = 'chat-list-item__type--public';
            typeText = 'Открытый';
        } else if (!isDirect) {
            avatarClass = 'chat-list-item__avatar--group';
            typeClass = 'chat-list-item__type--group';
            typeText = 'Группа';
        }

        let preview = '';
        if (chat.last_message) {
            const sender = chat.last_message_sender || '';
            const text = chat.last_message.length > 40
                ? chat.last_message.substring(0, 40) + '…'
                : chat.last_message;
            preview = sender ? sender + ': ' + text : text;
        } else {
            preview = isDirect ? 'Личный чат' : members.length + ' участников';
        }

        const timeStr = chat.last_message_at
            ? formatChatDate(chat.last_message_at)
            : formatChatDate(chat.created_at);

        const unreadHtml = unreadCount > 0
            ? '<div class="chat-item-unread"><span class="unread-badge">' +
              (unreadCount > 99 ? '99+' : unreadCount) + '</span></div>'
            : '';

        // Сохраняем member_avatars в data-атрибут для использования при открытии чата
        const memberAvatarsJson = JSON.stringify(memberAvatars).replace(/'/g, '&#39;');

        html += '<div class="chat-list-item ' + (isActive ? 'active' : '') +
            (unreadCount > 0 ? ' has-unread' : '') + '"' +
            ' onclick="openChat(' + chat.id + ',\'' + escapeHtml(chatName).replace(/'/g, "\\'") + '\',' + isDirect + ')"' +
            ' data-chat-id="' + chat.id + '"' +
            ' data-chat-name="' + escapeHtml(chatName).toLowerCase() + '"' +
            (otherUserId ? ' data-other-user-id="' + otherUserId + '"' : '') +
            ' data-member-avatars=\'' + memberAvatarsJson + '\'>' +
                '<div class="chat-list-item__avatar ' + avatarClass + '">' +
                    chatAvatarHtml +
                '</div>' +
                '<div class="chat-list-item__body">' +
                    '<div class="chat-list-item__name">' + escapeHtml(chatName) + '</div>' +
                    '<div class="chat-list-item__preview">' + escapeHtml(preview) + '</div>' +
                '</div>' +
                '<div class="chat-list-item__meta">' +
                    '<span class="chat-list-item__time">' + timeStr + '</span>' +
                    '<span class="chat-list-item__type ' + typeClass + '">' + typeText + '</span>' +
                    unreadHtml +
                '</div>' +
            '</div>';
    });

    container.innerHTML = html;
    NotificationManager.updateTotalBadge();
}

// Фильтрация списка чатов
function filterChatList(query) {
    const items = document.querySelectorAll('.chat-list-item');
    const q = query.toLowerCase().trim();

    items.forEach(item => {
        const name = item.dataset.chatName || '';
        item.style.display = name.includes(q) ? 'flex' : 'none';
    });
}

// ========== Открытие чата ==========

function openChat(chatId, chatName, isDirect) {
    closeChatInfo();

    if (activeChatId) {
        NotificationManager.clearActiveChat();
    }

    if (activeChatWs) {
        activeChatWs.close();
        activeChatWs = null;
    }

    activeChatId = chatId;
    NotificationManager.setActiveChat(chatId);

    // Обновляем шапку
    const initials = chatName.substring(0, 2).toUpperCase();
    document.getElementById('chatViewName').textContent = chatName;
    document.getElementById('chatViewStatus').textContent = 'Подключение...';
    document.getElementById('chatViewMessages').innerHTML = '';
    document.getElementById('chatViewInput').value = '';

    // Аватарка в шапке чата
    const avatarContainer = document.getElementById('chatViewAvatar');
    const initialsEl = document.getElementById('chatViewAvatarInitials');

    // Сбрасываем предыдущую аватарку
    if (avatarContainer) {
        const oldImg = avatarContainer.querySelector('img');
        if (oldImg) oldImg.remove();
    }
    if (initialsEl) {
        initialsEl.style.display = '';
        initialsEl.textContent = initials;
    }

    // Для direct чата — ставим аватарку собеседника
    if (isDirect) {
        const chatItem = document.querySelector('.chat-list-item[data-chat-id="' + chatId + '"]');
        if (chatItem) {
            const otherUserId = chatItem.getAttribute('data-other-user-id');
            try {
                const memberAvatars = JSON.parse(chatItem.getAttribute('data-member-avatars') || '{}');
                const url = otherUserId ? (memberAvatars[otherUserId] || '') : '';

                if (url && avatarContainer) {
                    if (initialsEl) initialsEl.style.display = 'none';
                    const img = document.createElement('img');
                    img.src = url;
                    img.style.cssText = 'width:100%;height:100%;object-fit:cover;border-radius:50%;';
                    img.onerror = function() {
                        this.remove();
                        if (initialsEl) initialsEl.style.display = '';
                    };
                    avatarContainer.appendChild(img);
                }
            } catch (e) {}
        }
    }

    document.getElementById('chatViewEmpty').style.display = 'none';
    document.getElementById('chatViewActive').style.display = 'flex';

    const membersPanel = document.getElementById('chatMembersPanel');
    if (membersPanel) membersPanel.style.display = 'none';

    document.querySelectorAll('.chat-list-item').forEach(function(item) {
        item.classList.toggle('active', parseInt(item.dataset.chatId) === chatId);
    });

    connectToChat(chatId);
}

function connectToChat(chatId) {
    let token = TokenManager.getAccessToken();

    if (token && token.startsWith('Bearer ')) {
        token = token.substring(7);
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/chat/${chatId}?token=${encodeURIComponent(token)}`;

    activeChatWs = new WebSocket(wsUrl);

    activeChatWs.onopen = () => {
        document.getElementById('chatViewStatus').textContent = 'Подключено';
    };

    activeChatWs.onmessage = (event) => {
        const msg = JSON.parse(event.data);

        if (msg.type === 'online_users') {
            updateChatOnlineUsers(msg);
        } else {
            displayChatMessage(msg);

            // УВЕДОМЛЕНИЕ: определяем данные и вызываем NotificationManager
            const chatName = document.getElementById('chatViewName').textContent || 'Чат';
            const senderName = msg.from || msg.sender || 'Неизвестный';
            const messageText = msg.text || msg.content || '';

            const currentUsername = TokenManager.getUsername ? TokenManager.getUsername() : null;
            const isMine = currentUsername && senderName === currentUsername;

            if (!isMine) {
                NotificationManager.onNewMessage(
                    chatId,          // ID чата
                    chatName,        // Название чата
                    senderName,      // Имя отправителя
                    messageText,     // Текст сообщения
                    null             // sender_id (нет в WS, фильтруем по username)
                );
            }
        }
    };

    activeChatWs.onclose = () => {
        document.getElementById('chatViewStatus').textContent = 'Отключено';
    };

    activeChatWs.onerror = () => {
        document.getElementById('chatViewStatus').textContent = 'Ошибка';
    };
}

// ========== Сообщения ==========

async function sendActiveMessage() {
    if (pendingFile) {
        sendFileMessage();
        return;
    }

    const input = document.getElementById('chatViewInput');
    const text = input.value.trim();

    if (!text || !activeChatId) return;
    input.value = '';

    try {
        await apiRequest('/api/chat/send', {
            method: 'POST',
            body: JSON.stringify({
                chat_id: activeChatId,
                text: text
            })
        });
    } catch (error) {
        showToast({ type: 'error', title: 'Ошибка', message: 'Не удалось отправить' });
    }
}

// ========== Отображение ==========

function displayChatMessage(msg) {
    const container = document.getElementById('chatViewMessages');
    if (!container) return;
    
    console.log('Display message:', msg); // Для отладки
    
    // Удаляем приветствие если есть
    const welcome = container.querySelector('.chat-welcome');
    if (welcome) welcome.remove();
    
    const currentUsername = TokenManager.getUsername();
    const isOwn = msg.from === currentUsername;
    const isSystem = msg.from === 'system';
    const timeStr = formatMessageTime(msg.sent_at || msg.created_at || msg.timestamp);
    
    // Определяем тип сообщения
    const messageType = msg.message_type || msg.type || 'text';
    
    const msgEl = document.createElement('div');
    msgEl.className = `chat-msg ${isOwn ? 'chat-msg--own' : 'chat-msg--other'}`;
    
    let contentHtml = '';
    
    // Обработка в зависимости от типа
    if (messageType === 'image' || messageType === 'IMAGE') {
        // ИЗОБРАЖЕНИЕ
        let imageUrl = msg.file_url;
        if (imageUrl && imageUrl.startsWith('/')) {
            imageUrl = window.location.origin + imageUrl;
        }
        
        contentHtml = `
            <div class="chat-msg__bubble chat-msg__bubble--image">
                <img src="${escapeHtml(imageUrl)}" 
                     alt="${escapeHtml(msg.file_name || 'Изображение')}"
                     onclick="openImageModal('${escapeHtml(imageUrl)}')"
                     onerror="this.onerror=null; this.style.display='none'; this.parentElement.innerHTML+='<div class=\'chat-msg__text\'>Ошибка загрузки</div>';"
                     style="max-width: 300px; max-height: 200px; border-radius: 8px; cursor: pointer; display: block;">
                ${msg.text && msg.text !== 'Фото' && msg.text !== '📷 Фото' ? 
                    `<div class="chat-msg__text" style="margin-top: 8px;">${escapeHtml(msg.text)}</div>` : ''}
            </div>
        `;
    } 
    else if (messageType === 'file' || messageType === 'FILE') {
        // ФАЙЛ
        const icon = getFileIcon(msg.file_name || '');
        const size = msg.file_size ? formatFileSize(msg.file_size) : '';
        let fileUrl = msg.file_url;
        if (fileUrl && fileUrl.startsWith('/')) {
            fileUrl = window.location.origin + fileUrl;
        }
        
        contentHtml = `
            <div class="chat-msg__bubble chat-msg__bubble--file">
                <a href="${escapeHtml(fileUrl)}" 
                   target="_blank" 
                   download="${escapeHtml(msg.file_name || 'file')}"
                   style="display: flex; align-items: center; gap: 10px; text-decoration: none; color: inherit;">
                    <i class="fas ${icon}" style="font-size: 24px;"></i>
                    <div>
                        <div style="font-weight: bold;">${escapeHtml(msg.file_name || 'Файл')}</div>
                        ${size ? `<div style="font-size: 12px; opacity: 0.7;">${size}</div>` : ''}
                    </div>
                </a>
                ${msg.text && !msg.text.startsWith('📎') && msg.text !== 'Файл' ? 
                    `<div class="chat-msg__text" style="margin-top: 8px;">${escapeHtml(msg.text)}</div>` : ''}
            </div>
        `;
    }
    else if (messageType === 'voice' || messageType === 'VOICE') {
        // ГОЛОСОВОЕ
        contentHtml = createVoicePlayerHtml(msg.file_url, msg.voice_duration || 0, isOwn);
    }
    else {
        // ТЕКСТ
        contentHtml = `<div class="chat-msg__bubble">${escapeHtml(msg.text || '')}</div>`;
    }
    
    // Формируем полное сообщение
    if (isSystem) {
        msgEl.innerHTML = `<div class="chat-msg__system">${escapeHtml(msg.text)}</div>`;
    } else {
        msgEl.innerHTML = `
            ${!isOwn ? `<span class="chat-msg__sender">${escapeHtml(msg.from)}</span>` : ''}
            ${contentHtml}
            <span class="chat-msg__time">${timeStr}</span>
        `;
    }
    
    container.appendChild(msgEl);
    container.scrollTop = container.scrollHeight;
}

function updateChatOnlineUsers(msg) {
    const onlineUsers = msg.online_users || msg.onlineUsers || [];
    const count = onlineUsers.length;

    const statusEl = document.getElementById('chatViewStatus');
    if (statusEl) statusEl.textContent = `${count} в сети`;

    const countEl = document.getElementById('chatViewMemberCount');
    if (countEl) countEl.textContent = count;
}

function toggleChatMembers() {
    const panel = document.getElementById('chatMembersPanel');
    if (!panel) return;
    panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
}

function toggleEmojiPicker() {
    showToast({ type: 'info', title: 'В разработке', message: 'Эмодзи скоро появятся' });
}

function toggleChatSettings() {
    // TODO: настройки чата
    showToast({ type: 'info', title: 'В разработке', message: 'Настройки чата скоро появятся' });
}


function formatMessageTime(sentAt) {
            if (!sentAt) {
                // Fallback если время не пришло
                return new Date().toLocaleTimeString('ru-RU', {
                    hour: '2-digit',
                    minute: '2-digit'
                });
            }

            let date;

            // ISO string: "2025-06-21T12:00:00Z"
            if (typeof sentAt === 'string') {
                date = new Date(sentAt);
            }
            // Proto format: {"seconds": 123456, "nanos": 0}
            else if (sentAt.seconds) {
                date = new Date(sentAt.seconds * 1000);
            }
            else {
                date = new Date(sentAt);
            }

            if (isNaN(date.getTime())) {
                return new Date().toLocaleTimeString('ru-RU', {
                    hour: '2-digit',
                    minute: '2-digit'
                });
            }

            return date.toLocaleTimeString('ru-RU', {
                hour: '2-digit',
                minute: '2-digit'
            });
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

        // Сбрасываем уведомления
        NotificationManager.reset();

         // Очищаем активный чат
        if (activeChatWs) {
            activeChatWs.close();
            activeChatWs = null;
        }
        activeChatId = null;

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

    // Было: <a href="/chat?id=${chatId}">
    // Стало: onclick
    return `
        <div class="chat-card chat-card--direct" onclick="openChat(${chatId}, '${escapeHtml(otherUser)}', true)">
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
                    onclick="event.stopPropagation(); deleteChat(${chatId})" 
                    class="chat-card__delete"
                    title="Удалить чат">
                    <i class="fas fa-trash-alt"></i>
                </button>
                <i class="fas fa-chevron-right chat-card__arrow"></i>
            </div>
        </div>
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
        <div class="chat-card chat-card--group" onclick="openChat(${chatId}, '${escapeHtml(chatName)}', false)">
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
                    onclick="event.stopPropagation(); deleteChat(${chatId})" 
                    class="chat-card__delete"
                    title="Удалить чат">
                    <i class="fas fa-trash-alt"></i>
                </button>
                <i class="fas fa-chevron-right chat-card__arrow"></i>
            </div>
        </div>
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

// ============================================
// ПРОФИЛЬ
// ============================================
async function loadUserInfo() {
    const userId = TokenManager.getUserId();
    if (!userId) return;

    let token = TokenManager.getAccessToken();
    if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

    try {
        const response = await fetch('/api/user/' + userId, {
            headers: { 'Authorization': token }
        });

        if (!response.ok) throw new Error('Failed to load profile');

        const user = await response.json();

        // Загружаем статистику
        const [chatsResp, friendsResp] = await Promise.all([
            fetch('/api/chat/my', { headers: { 'Authorization': token } }).catch(() => null),
            fetch('/api/friends', { headers: { 'Authorization': token } }).catch(() => null),
        ]);

        let chatCount = 0;
        let friendCount = 0;

        if (chatsResp && chatsResp.ok) {
            const chatsData = await chatsResp.json();
            chatCount = chatsData.chats ? chatsData.chats.length : 0;
        }
        if (friendsResp && friendsResp.ok) {
            const friendsData = await friendsResp.json();
            friendCount = friendsData.friends ? friendsData.friends.length : 0;
        }

        renderProfile(user, chatCount, friendCount);

    } catch (err) {
        console.error('Failed to load profile:', err);
        document.getElementById('profilePage').innerHTML =
            '<div class="loading"><p>Ошибка загрузки профиля</p></div>';
    }
}

function renderProfile(user, chatCount, friendCount) {
    const container = document.getElementById('profilePage');
    const initials = user.username ? user.username.substring(0, 2).toUpperCase() : '??';
    const avatarUrl = user.avatar_url || '';
    const bio = user.bio || '';
    const role = user.role || 'user';
    const createdAt = new Date(user.created_at).toLocaleDateString('ru-RU', {
        year: 'numeric', month: 'long', day: 'numeric'
    });

    const roleIcon = role === 'admin' ? 'fa-shield-alt' : 'fa-user';
    const roleLabel = role === 'admin' ? 'Администратор' : 'Пользователь';

    const avatarContent = avatarUrl
        ? '<img src="' + escapeHtml(avatarUrl) + '" alt="avatar">'
        : escapeHtml(initials);

    container.innerHTML =
        '<!-- Шапка -->' +
        '<div class="profile-header">' +
            '<button class="profile-edit-btn" onclick="toggleProfileEdit()" title="Редактировать">' +
                '<i class="fas fa-pen"></i>' +
            '</button>' +
            '<div class="profile-header-content">' +
                '<div class="profile-avatar-container">' +
                    '<div class="profile-avatar" id="profileAvatar">' +
                        avatarContent +
                    '</div>' +
                    '<label class="profile-avatar-upload" title="Сменить аватар">' +
                        '<i class="fas fa-camera"></i>' +
                        '<input type="file" accept="image/jpeg,image/png,image/gif,image/webp" onchange="uploadAvatar(this)">' +
                    '</label>' +
                '</div>' +
                '<div class="profile-header-info">' +
                    '<h1 class="profile-name" id="profileName">' + escapeHtml(user.username) + '</h1>' +
                    '<span class="profile-role">' +
                        '<i class="fas ' + roleIcon + '"></i> ' + roleLabel +
                    '</span>' +
                    '<div class="profile-joined">' +
                        '<i class="fas fa-calendar-alt"></i>' +
                        'На платформе с ' + createdAt +
                    '</div>' +
                '</div>' +
            '</div>' +
        '</div>' +

        '<!-- Статистика -->' +
        '<div class="profile-stats">' +
            '<div class="profile-stat">' +
                '<span class="profile-stat-value">' + chatCount + '</span>' +
                '<span class="profile-stat-label">Чатов</span>' +
            '</div>' +
            '<div class="profile-stat">' +
                '<span class="profile-stat-value">' + friendCount + '</span>' +
                '<span class="profile-stat-label">Друзей</span>' +
            '</div>' +
            '<div class="profile-stat">' +
                '<span class="profile-stat-value">' +
                    (user.role === 'admin' ? '∞' : '—') +
                '</span>' +
                '<span class="profile-stat-label">Уровень</span>' +
            '</div>' +
        '</div>' +

        '<!-- О себе -->' +
        '<div class="profile-section">' +
            '<h3 class="profile-section-title">' +
                '<i class="fas fa-quote-left"></i> О себе' +
            '</h3>' +
            '<div class="profile-bio" id="profileBioView">' +
                (bio
                    ? '<p class="profile-bio-text">' + escapeHtml(bio) + '</p>'
                    : '<p class="profile-bio-empty">Расскажите о себе...</p>'
                ) +
            '</div>' +
        '</div>' +

        '<!-- Информация -->' +
        '<div class="profile-section">' +
            '<h3 class="profile-section-title">' +
                '<i class="fas fa-info-circle"></i> Информация' +
            '</h3>' +

            '<div class="profile-field">' +
                '<div class="profile-field-icon"><i class="fas fa-user"></i></div>' +
                '<div class="profile-field-content">' +
                    '<div class="profile-field-label">Имя пользователя</div>' +
                    '<div class="profile-field-value">' + escapeHtml(user.username) + '</div>' +
                '</div>' +
            '</div>' +

            '<div class="profile-field">' +
                '<div class="profile-field-icon"><i class="fas fa-envelope"></i></div>' +
                '<div class="profile-field-content">' +
                    '<div class="profile-field-label">Email</div>' +
                    '<div class="profile-field-value">' + escapeHtml(user.email) + '</div>' +
                '</div>' +
            '</div>' +

            '<div class="profile-field">' +
                '<div class="profile-field-icon"><i class="fas fa-id-badge"></i></div>' +
                '<div class="profile-field-content">' +
                    '<div class="profile-field-label">ID</div>' +
                    '<div class="profile-field-value">#' + user.id + '</div>' +
                '</div>' +
            '</div>' +
        '</div>' +

        '<!-- Форма редактирования -->' +
        '<div class="profile-section profile-edit-form" id="profileEditForm">' +
            '<h3 class="profile-section-title">' +
                '<i class="fas fa-edit"></i> Редактирование' +
            '</h3>' +

            '<div class="profile-edit-field">' +
                '<label>Имя пользователя</label>' +
                '<input type="text" id="editUsername" value="' + escapeHtml(user.username) + '" maxlength="50">' +
            '</div>' +

            '<div class="profile-edit-field">' +
                '<label>О себе</label>' +
                '<textarea id="editBio" maxlength="500" placeholder="Расскажите о себе...">' +
                    escapeHtml(bio) +
                '</textarea>' +
            '</div>' +

            '<div class="profile-edit-actions">' +
                '<button class="btn-save" onclick="saveProfile()">'+
                    '<i class="fas fa-check"></i> Сохранить' +
                '</button>' +
                '<button class="btn-cancel" onclick="toggleProfileEdit()">' +
                    '<i class="fas fa-times"></i> Отмена' +
                '</button>' +
            '</div>' +
        '</div>';
}

async function loadCurrentUserProfile() {
    var userId = TokenManager.getUserId();
    if (!userId) return;

    var token = TokenManager.getAccessToken();
    if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

    try {
        var response = await fetch('/api/user/' + userId, {
            headers: { 'Authorization': token }
        });
        if (!response.ok) return;

        var user = await response.json();
        TokenManager.setAvatarUrl(user.avatar_url || '');
        applySidebarAvatar(user.avatar_url, user.username);
    } catch (err) {
        console.warn('Failed to load profile:', err);
    }
}

function applySidebarAvatar(avatarUrl, username) {
    var footerAvatar = document.querySelector('.footer-avatar');
    if (!footerAvatar) return;

    var initialsEl = footerAvatar.querySelector('.footer-avatar-initials');
    var oldImg = footerAvatar.querySelector('.footer-avatar-img');

    if (avatarUrl) {
        if (initialsEl) initialsEl.style.display = 'none';
        if (oldImg) {
            oldImg.src = avatarUrl;
        } else {
            var img = document.createElement('img');
            img.className = 'footer-avatar-img';
            img.src = avatarUrl;
            img.onerror = function() {
                this.remove();
                if (initialsEl) initialsEl.style.display = '';
            };
            footerAvatar.appendChild(img);
        }
    } else {
        if (oldImg) oldImg.remove();
        if (initialsEl) {
            initialsEl.style.display = '';
            if (username) initialsEl.textContent = username.substring(0, 2).toUpperCase();
        }
    }
}

// ============================================
// РЕДАКТИРОВАНИЕ
// ============================================

function toggleProfileEdit() {
    const form = document.getElementById('profileEditForm');
    if (form) {
        form.classList.toggle('active');
    }
}

async function saveProfile() {
    const username = document.getElementById('editUsername').value.trim();
    const bio = document.getElementById('editBio').value.trim();

    if (!username) {
        alert('Имя пользователя не может быть пустым');
        return;
    }

    let token = TokenManager.getAccessToken();
    if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

    try {
        const response = await fetch('/api/user/profile', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify({ username, bio })
        });

        if (!response.ok) {
            const data = await response.json();
            alert('Ошибка: ' + (data.error || 'Не удалось сохранить'));
            return;
        }

        // Сохраняем bio
        TokenManager.setBio(bio);

        // Скрываем форму
        toggleProfileEdit();

        // Обновляем страницу профиля
        loadUserInfo();

        // Обновляем sidebar
        var footerUsername = document.getElementById('footerUsername');
        if (footerUsername) footerUsername.textContent = username;

        if (!TokenManager.getAvatarUrl()) {
            var footerInitials = document.getElementById('footerAvatarInitials');
            if (footerInitials) footerInitials.textContent = username.substring(0, 2).toUpperCase();
        }

    } catch (err) {
        alert('Ошибка сохранения: ' + err.message);
    }
}

// ============================================
// ЗАГРУЗКА АВАТАРКИ
// ============================================

async function uploadAvatar(input) {
    const file = input.files[0];
    if (!file) return;

    // Проверки
    if (file.size > 5 * 1024 * 1024) {
        alert('Файл слишком большой. Максимум 5MB');
        input.value = '';
        return;
    }

    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowed.includes(file.type)) {
        alert('Разрешены только JPEG, PNG, GIF, WebP');
        input.value = '';
        return;
    }

    // Показываем прогресс
    const avatar = document.getElementById('profileAvatar');
    const oldContent = avatar.innerHTML;
    avatar.innerHTML = '<div class="avatar-upload-progress"><i class="fas fa-spinner fa-spin"></i></div>';

    let token = TokenManager.getAccessToken();
    if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

    try {
        const formData = new FormData();
        formData.append('avatar', file);

        const response = await fetch('/api/user/avatar', {
            method: 'POST',
            headers: { 'Authorization': token },
            body: formData
        });

        if (!response.ok) {
            const data = await response.json();
            alert('Ошибка: ' + (data.error || 'Не удалось загрузить'));
            avatar.innerHTML = oldContent;
            return;
        }

        const result = await response.json();

        // Обновляем аватар на странице
        avatar.innerHTML = '<img src="' + escapeHtml(result.avatar_url) + '" alt="avatar">';

        // Сохраняем в localStorage
        TokenManager.setAvatarUrl(result.avatar_url);
        applySidebarAvatar(result.avatar_url, TokenManager.getUsername());
        // Обновляем везде
        applyAvatarEverywhere(result.avatar_url, TokenManager.getUsername());

    } catch (err) {
        alert('Ошибка загрузки: ' + err.message);
        avatar.innerHTML = oldContent;
    }

    input.value = '';
}

// ============================================
// AVATAR HELPER
// ============================================
function avatarHtml(avatarUrl, username, size) {
    size = size || 40;
    const initials = username ? username.substring(0, 2).toUpperCase() : '??';

    if (avatarUrl) {
        return '<img src="' + escapeHtml(avatarUrl) + '" alt=""' +
            ' style="width:' + size + 'px;height:' + size + 'px;object-fit:cover;border-radius:50%;display:block;"' +
            ' onerror="this.style.display=\'none\';this.nextElementSibling.style.display=\'flex\'">' +
            '<span style="display:none;align-items:center;justify-content:center;' +
            'width:' + size + 'px;height:' + size + 'px;border-radius:50%;' +
            'background:rgba(108,99,255,0.3);color:#fff;font-weight:600;' +
            'font-size:' + Math.round(size * 0.35) + 'px;">' +
            escapeHtml(initials) + '</span>';
    }

    return '<span style="display:flex;align-items:center;justify-content:center;' +
        'width:' + size + 'px;height:' + size + 'px;border-radius:50%;' +
        'background:rgba(108,99,255,0.3);color:#fff;font-weight:600;' +
        'font-size:' + Math.round(size * 0.35) + 'px;">' +
        escapeHtml(initials) + '</span>';
}

function updateSidebarAvatar(url) {
    const footerAvatar = document.querySelector('.footer-avatar');
    if (footerAvatar && url) {
        const initials = footerAvatar.querySelector('.footer-avatar-initials');
        if (initials) {
            initials.style.display = 'none';
        }
        // Проверяем есть ли уже img
        let img = footerAvatar.querySelector('img');
        if (!img) {
            img = document.createElement('img');
            img.style.width = '100%';
            img.style.height = '100%';
            img.style.objectFit = 'cover';
            img.style.borderRadius = '50%';
            footerAvatar.appendChild(img);
        }
        img.src = url;
    }
}

function applyAvatarEverywhere(avatarUrl, username) {
    const initials = username ? username.substring(0, 2).toUpperCase() : '??';

    // 1. Sidebar footer
    const footerAvatar = document.querySelector('.footer-avatar');
    if (footerAvatar) {
        if (avatarUrl) {
            let img = footerAvatar.querySelector('.footer-avatar-img');
            if (!img) {
                img = document.createElement('img');
                img.className = 'footer-avatar-img';
                footerAvatar.appendChild(img);
            }
            img.src = avatarUrl;

            const initialsEl = footerAvatar.querySelector('.footer-avatar-initials');
            if (initialsEl) initialsEl.style.display = 'none';
        } else {
            const img = footerAvatar.querySelector('.footer-avatar-img');
            if (img) img.remove();

            const initialsEl = footerAvatar.querySelector('.footer-avatar-initials');
            if (initialsEl) {
                initialsEl.style.display = '';
                initialsEl.textContent = initials;
            }
        }
    }
}

// При загрузке страницы, если пользователь уже авторизован:
document.addEventListener('DOMContentLoaded', () => {
    if (TokenManager.isAuthenticated()) {
        loadCurrentUserProfile();
    }
});

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

    // ✅ Аватарка
    const avatarImg = item.querySelector('.friend-avatar-img');
    if (friend.avatar_url && avatarImg) {
        avatarImg.src = friend.avatar_url;
        avatarImg.style.display = 'block';
        avatarImg.onerror = function() {
            this.style.display = 'none';
        };
        // Скрываем инициалы когда есть картинка
        const initialsEl = item.querySelector('.avatar-initials');
        if (initialsEl) initialsEl.style.display = 'none';
    }

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

    container.addEventListener('click', (e) => {
        if (!e.target.closest('.btn-more')) {
            startChatWithFriend(friend.user_id, friend.username);
        }
    });

    const moreBtn = item.querySelector('.btn-more');
    moreBtn.addEventListener('click', (e) => {
        e.stopPropagation();
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

    // ✅ Аватарка
    const avatarImg = item.querySelector('.friend-avatar-img');
    if (request.from_avatar_url && avatarImg) {
        avatarImg.src = request.from_avatar_url;
        avatarImg.style.display = 'block';
        avatarImg.onerror = function() {
            this.style.display = 'none';
        };
        const initialsEl = item.querySelector('.avatar-initials');
        if (initialsEl) initialsEl.style.display = 'none';
    }

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

    // ✅ Аватарка
    const avatarImg = item.querySelector('.search-avatar-img');
    if (user.avatar_url && avatarImg) {
        avatarImg.src = user.avatar_url;
        avatarImg.style.display = 'block';
        avatarImg.onerror = function() {
            this.style.display = 'none';
        };
        const initialsEl = item.querySelector('.avatar-initials');
        if (initialsEl) initialsEl.style.display = 'none';
    }

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

        // Переключаемся на секцию чатов и открываем чат
        showSection('chats');

        // Небольшая задержка чтобы список загрузился
        setTimeout(() => {
            openChat(data.chat_id, username, true);
        }, 300);

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

// ==================== VOICE MESSAGES ====================

let mediaRecorder = null;
let audioChunks = [];
let recordingStartTime = null;
let recordingTimer = null;
let isRecording = false;
let currentAudio = null;
let currentPlayerId = null;

// ========== Запись ==========

async function startVoiceRecording() {
    if (!activeChatId) {
        showToast({ type: 'error', title: 'Ошибка', message: 'Сначала откройте чат' });
        return;
    }

    try {
        const stream = await navigator.mediaDevices.getUserMedia({
            audio: {
                echoCancellation: true,
                noiseSuppression: true,
                sampleRate: 48000
            }
        });

        audioChunks = [];
        recordingStartTime = Date.now();
        isRecording = true;

        mediaRecorder = new MediaRecorder(stream, {
            mimeType: MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
                ? 'audio/webm;codecs=opus'
                : 'audio/webm'
        });

        mediaRecorder.ondataavailable = (event) => {
            if (event.data.size > 0) {
                audioChunks.push(event.data);
            }
        };

        mediaRecorder.onstop = () => {
            stream.getTracks().forEach(track => track.stop());

            if (audioChunks.length > 0) {
                const duration = (Date.now() - recordingStartTime) / 1000;
                const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
                sendVoiceMessage(audioBlob, duration);
            }

            audioChunks = [];
        };

        mediaRecorder.start(100); // чанки каждые 100мс
        showRecordingUI();
        recordingTimer = setInterval(updateRecordingTimer, 100);

    } catch (error) {
        console.error('Microphone error:', error);

        if (error.name === 'NotAllowedError') {
            showToast({
                type: 'error',
                title: 'Нет доступа',
                message: 'Разрешите доступ к микрофону в настройках браузера'
            });
        } else {
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: 'Не удалось получить доступ к микрофону'
            });
        }
    }
}

function stopVoiceRecording() {
    if (mediaRecorder && mediaRecorder.state === 'recording') {
        isRecording = false;
        clearInterval(recordingTimer);
        mediaRecorder.stop();
        hideRecordingUI();
    }
}

function cancelVoiceRecording() {
    if (mediaRecorder && mediaRecorder.state === 'recording') {
        isRecording = false;
        clearInterval(recordingTimer);
        audioChunks = []; // Очищаем чтобы onstop не отправил
        mediaRecorder.stream.getTracks().forEach(track => track.stop());
        mediaRecorder.stop();
        hideRecordingUI();
    }
}

// ========== UI записи ==========

function showRecordingUI() {
    const wrapper = document.getElementById('chatInputWrapper');
    if (!wrapper) return;

    wrapper.classList.add('recording');
    wrapper.innerHTML = `
        <button onclick="cancelVoiceRecording()" class="btn-icon-sm recording-cancel" title="Отмена">
            <i class="fas fa-times"></i>
        </button>
        <div class="recording-indicator">
            <span class="recording-dot"></span>
            <span class="recording-time" id="recordingTime">0:00</span>
        </div>
        <div class="recording-wave">
            <span></span><span></span><span></span><span></span><span></span>
        </div>
        <button onclick="stopVoiceRecording()" class="btn-send recording-send" title="Отправить">
            <i class="fas fa-paper-plane"></i>
        </button>
    `;
}

function hideRecordingUI() {
    const wrapper = document.getElementById('chatInputWrapper');
    if (!wrapper) return;

    wrapper.classList.remove('recording');
    wrapper.innerHTML = `
        <button onclick="toggleEmojiPicker()" class="btn-icon-sm" title="Эмодзи">
            <i class="fas fa-smile"></i>
        </button>
        <input
            type="text"
            id="chatViewInput"
            placeholder="Напишите сообщение..."
            autocomplete="off"
            onkeydown="if(event.key==='Enter') sendActiveMessage()"
        >
        <button onclick="startVoiceRecording()" class="btn-icon-sm btn-voice" title="Голосовое сообщение">
            <i class="fas fa-microphone"></i>
        </button>
        <button onclick="sendActiveMessage()" class="btn-send" title="Отправить">
            <i class="fas fa-paper-plane"></i>
        </button>
    `;

    // Фокус на поле ввода
    const input = document.getElementById('chatViewInput');
    if (input) input.focus();
}

function updateRecordingTimer() {
    const elapsed = (Date.now() - recordingStartTime) / 1000;
    const minutes = Math.floor(elapsed / 60);
    const seconds = Math.floor(elapsed % 60);
    const el = document.getElementById('recordingTime');
    if (el) {
        el.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;
    }
}

// ========== Отправка ==========

async function sendVoiceMessage(audioBlob, duration) {
    if (!activeChatId) return;

    // Минимальная длительность
    if (duration < 0.5) {
        showToast({ type: 'info', title: 'Слишком коротко', message: 'Запишите сообщение длиннее' });
        return;
    }

    const formData = new FormData();
    formData.append('chat_id', activeChatId.toString());
    formData.append('voice', audioBlob, 'voice.webm');
    formData.append('duration', duration.toFixed(1));

    try {
        let token = TokenManager.getAccessToken();

        const response = await fetch('/api/chat/send-voice', {
            method: 'POST',
            headers: {
                'Authorization': token
            },
            body: formData
        });

        if (!response.ok) {
            const data = await response.json();
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: data.error || 'Не удалось отправить'
            });
        }
    } catch (error) {
        console.error('Failed to send voice:', error);
        showToast({
            type: 'error',
            title: 'Ошибка',
            message: 'Не удалось отправить голосовое сообщение'
        });
    }
}

function createVoicePlayerHtml(voiceUrl, duration, isOwn) {
    const id = 'vp_' + Date.now() + '_' + Math.random().toString(36).substr(2, 5);
    const durationStr = formatVoiceDuration(duration);

    return `
        <div class="voice-player ${isOwn ? 'voice-player--own' : 'voice-player--other'}" id="${id}">
            <button class="voice-player__btn" onclick="toggleVoice('${id}', '${voiceUrl}')">
                <i class="fas fa-play" id="${id}_icon"></i>
            </button>
            <div class="voice-player__track">
                <div class="voice-player__progress" id="${id}_bar"></div>
            </div>
            <span class="voice-player__time" id="${id}_time">${durationStr}</span>
        </div>
    `;
}

function formatVoiceDuration(sec) {
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    return `${m}:${s.toString().padStart(2, '0')}`;
}

// ========== Воспроизведение ==========

function toggleVoice(playerId, url) {
    const icon = document.getElementById(playerId + '_icon');
    const bar = document.getElementById(playerId + '_bar');
    const timeEl = document.getElementById(playerId + '_time');

    // Если играет этот — пауза
    if (currentPlayerId === playerId && currentAudio && !currentAudio.paused) {
        currentAudio.pause();
        if (icon) icon.className = 'fas fa-play';
        return;
    }

    // Остановить предыдущий
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;

        if (currentPlayerId) {
            const prevIcon = document.getElementById(currentPlayerId + '_icon');
            const prevBar = document.getElementById(currentPlayerId + '_bar');
            if (prevIcon) prevIcon.className = 'fas fa-play';
            if (prevBar) prevBar.style.width = '0%';
        }
    }

    // Играем новый
    currentAudio = new Audio(url);
    currentPlayerId = playerId;

    if (icon) icon.className = 'fas fa-pause';

    currentAudio.addEventListener('timeupdate', () => {
        if (!currentAudio.duration) return;

        const pct = (currentAudio.currentTime / currentAudio.duration) * 100;
        if (bar) bar.style.width = pct + '%';

        const remaining = currentAudio.duration - currentAudio.currentTime;
        if (timeEl) timeEl.textContent = formatVoiceDuration(remaining);
    });

    currentAudio.addEventListener('ended', () => {
        if (icon) icon.className = 'fas fa-play';
        if (bar) bar.style.width = '0%';
        currentPlayerId = null;

        // Восстанавливаем исходную длительность
        if (timeEl && currentAudio.duration) {
            timeEl.textContent = formatVoiceDuration(currentAudio.duration);
        }
    });

    currentAudio.addEventListener('error', () => {
        if (icon) icon.className = 'fas fa-play';
        showToast({ type: 'error', title: 'Ошибка', message: 'Не удалось воспроизвести' });
        currentPlayerId = null;
    });

    currentAudio.play().catch(err => {
        console.error('Play error:', err);
        if (icon) icon.className = 'fas fa-play';
    });
}

// ==================== CHAT INFO PANEL ====================

let chatInfoData = null; // Кэш данных текущего чата

function toggleChatInfo() {
    const panel = document.getElementById('chatInfoPanel');
    if (!panel) return;

    if (panel.style.display === 'none') {
        openChatInfo();
    } else {
        closeChatInfo();
    }
}

function closeChatInfo() {
    const panel = document.getElementById('chatInfoPanel');
    if (panel) panel.style.display = 'none';

    const overlay = document.getElementById('chatInfoOverlay');
    if (overlay) overlay.style.display = 'none';
}

function openChatInfo() {
    if (!activeChatId) return;

    const panel = document.getElementById('chatInfoPanel');
    if (!panel) return;

    const chat = chatListData.find(c => c.id === activeChatId);
    if (!chat) return;

    chatInfoData = chat;
    renderChatInfo(chat);

    // Создаём overlay
    let overlay = document.getElementById('chatInfoOverlay');
    if (!overlay) {
        overlay = document.createElement('div');
        overlay.id = 'chatInfoOverlay';
        overlay.className = 'chat-info-overlay';
        overlay.onclick = closeChatInfo;

        const chatView = document.getElementById('chatViewActive');
        if (chatView) chatView.appendChild(overlay);
    }
    overlay.style.display = 'block';

    panel.style.display = 'flex';
}

function renderChatInfo(chat) {
    const currentUsername = TokenManager.getUsername();
    const currentUserId = TokenManager.getUserId();
    const isDirect = chat.is_direct || false;
    const isPublic = chat.is_public || false;
    const chatName = getChatDisplayName(chat);
    const initials = chatName.substring(0, 2).toUpperCase();
    const members = chat.usernames || [];
    const memberIds = chat.member_ids || [];
    const memberAvatars = chat.member_avatars || {};
    const isOwner = chat.creator_id === parseInt(currentUserId);
    const myId = String(currentUserId);

    // Аватар чата в панели информации
    const chatInfoAvatarContainer = document.getElementById('chatInfoAvatar');
    const chatInfoInitialsEl = document.getElementById('chatInfoAvatarInitials');

    if (chatInfoInitialsEl) chatInfoInitialsEl.textContent = initials;

    // Убираем старую картинку
    if (chatInfoAvatarContainer) {
        const oldImg = chatInfoAvatarContainer.querySelector('img');
        if (oldImg) oldImg.remove();
        if (chatInfoInitialsEl) chatInfoInitialsEl.style.display = '';
    }

    // Для direct чата — ставим аватарку собеседника
    if (isDirect && chatInfoAvatarContainer) {
        let otherUserId = null;
        for (let i = 0; i < memberIds.length; i++) {
            if (String(memberIds[i]) !== myId) {
                otherUserId = memberIds[i];
                break;
            }
        }
        const avatarUrl = otherUserId ? (memberAvatars[String(otherUserId)] || '') : '';
        if (avatarUrl) {
            if (chatInfoInitialsEl) chatInfoInitialsEl.style.display = 'none';
            const img = document.createElement('img');
            img.src = avatarUrl;
            img.style.cssText = 'width:100%;height:100%;object-fit:cover;border-radius:50%;';
            img.onerror = function() {
                this.remove();
                if (chatInfoInitialsEl) chatInfoInitialsEl.style.display = '';
            };
            chatInfoAvatarContainer.appendChild(img);
        }
    }

    document.getElementById('chatInfoName').textContent = chatName;

    // Тип чата
    const typeEl = document.getElementById('chatInfoType');
    if (isDirect) {
        typeEl.innerHTML = '<i class="fas fa-user"></i><span>Личный чат</span>';
    } else if (isPublic) {
        typeEl.innerHTML = '<i class="fas fa-globe"></i><span>Открытый чат</span>';
    } else {
        typeEl.innerHTML = '<i class="fas fa-lock"></i><span>Закрытый чат</span>';
    }

    // Количество участников
    document.getElementById('chatInfoMemberCount').textContent = members.length;

    // Кнопка добавления
    const addMemberEl = document.getElementById('chatInfoAddMember');
    if (addMemberEl) {
        addMemberEl.style.display = (isOwner && !isDirect) ? 'block' : 'none';
    }

    // Список участников
    const membersContainer = document.getElementById('chatInfoMembers');
    let membersHtml = '';

    members.forEach((username, index) => {
        const isCurrentUser = username === currentUsername;
        const isFirstMember = index === 0;

        // Находим user_id и аватарку для этого участника
        const userId = memberIds[index] || null;
        const memberAvatarUrl = userId ? (memberAvatars[String(userId)] || '') : '';

        // Аватарка участника
        const memberAvatarHtml = avatarHtml(memberAvatarUrl, username, 36);

        // Роль
        let roleHtml = '';
        if (isFirstMember && !isDirect) {
            roleHtml = '<span class="chat-info-member__role">Владелец</span>';
        }

        // "Это вы"
        let youBadge = '';
        if (isCurrentUser) {
            youBadge = '<span class="chat-info-member__you">Вы</span>';
        }

        // Кнопка удаления
        let removeBtn = '';
        if (isOwner && !isCurrentUser && !isDirect) {
            removeBtn =
                '<button class="chat-info-member__remove"' +
                ' onclick="event.stopPropagation(); removeMemberFromChat(\'' + escapeHtml(username) + '\')"' +
                ' title="Удалить">' +
                '<i class="fas fa-times"></i>' +
                '</button>';
        }

        membersHtml +=
            '<div class="chat-info-member">' +
                '<div class="chat-info-member__avatar">' +
                    memberAvatarHtml +
                '</div>' +
                '<div class="chat-info-member__body">' +
                    '<div class="chat-info-member__name">' +
                        escapeHtml(username) + ' ' + youBadge +
                    '</div>' +
                    roleHtml +
                '</div>' +
                removeBtn +
            '</div>';
    });

    membersContainer.innerHTML = membersHtml;

    // Кнопка удаления чата
    const deleteBtn = document.getElementById('chatInfoDeleteBtn');
    if (deleteBtn) {
        if (isOwner || isDirect) {
            deleteBtn.innerHTML = '<i class="fas fa-trash-alt"></i><span>Удалить чат</span>';
        } else {
            deleteBtn.innerHTML = '<i class="fas fa-sign-out-alt"></i><span>Покинуть чат</span>';
        }
        deleteBtn.style.display = 'flex';
    }
}

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        closeChatInfo();
    }
});

// ========== Действия ==========

async function addMemberFromInput() {
    const input = document.getElementById('addMemberInput');
    if (!input) return;

    const username = input.value.trim();
    if (!username) return;

    if (!activeChatId) return;

    try {
        const response = await apiRequest('/api/chat/add-member', {
            method: 'POST',
            body: JSON.stringify({
                chat_id: activeChatId,
                username: username
            })
        });

        if (response.ok) {
            showToast({
                type: 'success',
                title: 'Добавлен',
                message: `${username} добавлен в чат`
            });
            input.value = '';

            // Обновляем список чатов и панель
            await loadChatList();
            const updatedChat = chatListData.find(c => c.id === activeChatId);
            if (updatedChat) {
                renderChatInfo(updatedChat);
            }
        } else {
            const data = await response.json();
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: data.error || 'Не удалось добавить'
            });
        }
    } catch (error) {
        showToast({
            type: 'error',
            title: 'Ошибка',
            message: error.message
        });
    }
}

async function removeMemberFromChat(username) {
    if (!confirm(`Удалить ${username} из чата?`)) return;
    if (!activeChatId) return;

    try {
        const response = await apiRequest('/api/chat/remove-member', {
            method: 'POST',
            body: JSON.stringify({
                chat_id: activeChatId,
                username: username
            })
        });

        if (response.ok) {
            showToast({ type: 'success', title: 'Удалён', message: `${username} удалён из чата` });

            await loadChatList();
            const updatedChat = chatListData.find(c => c.id === activeChatId);
            if (updatedChat) renderChatInfo(updatedChat);
        } else {
            const data = await response.json();
            showToast({ type: 'error', title: 'Ошибка', message: data.error });
        }
    } catch (error) {
        showToast({ type: 'error', title: 'Ошибка', message: error.message });
    }
}

async function leaveOrDeleteChat() {
    if (!activeChatId) return;

    const chat = chatListData.find(c => c.id === activeChatId);
    const currentUserId = TokenManager.getUserId();
    const isOwner = chat && chat.creator_id === parseInt(currentUserId);

    const action = isOwner ? 'удалить' : 'покинуть';
    if (!confirm(`Вы уверены, что хотите ${action} этот чат?`)) return;

    try {
        const response = await apiRequest(`/api/chat/delete/${activeChatId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showToast({
                type: 'success',
                title: 'Готово',
                message: `Чат ${isOwner ? 'удалён' : 'покинут'}`
            });

            // Закрываем чат
            closeChatInfo();
            closeActiveChat();
            loadChatList();
            loadChatCount();
        } else {
            const data = await response.json();
            showToast({
                type: 'error',
                title: 'Ошибка',
                message: data.error
            });
        }
    } catch (error) {
        showToast({
            type: 'error',
            title: 'Ошибка',
            message: error.message
        });
    }
}

// ============================================
// SIDEBAR — СВОРАЧИВАНИЕ / РАЗВОРАЧИВАНИЕ
// ============================================

let sidebarCollapsed = false;
let sidebarAutoCollapsed = false; // флаг: свёрнут автоматически при переходе в чаты

/**
 * Ручное переключение sidebar
 */
function toggleSidebar() {
    const sidebar = document.getElementById('mainSidebar');
    
    if (sidebarCollapsed) {
        expandSidebar();
        sidebarAutoCollapsed = false; // ручное действие сбрасывает авто-флаг
    } else {
        collapseSidebar();
        sidebarAutoCollapsed = false;
    }
}

/**
 * Свернуть sidebar
 */
function collapseSidebar() {
    const sidebar = document.getElementById('mainSidebar');
    sidebar.classList.add('collapsed');
    sidebarCollapsed = true;
    
    // Сохраняем состояние (только ручное)
    if (!sidebarAutoCollapsed) {
        localStorage.setItem('sidebarCollapsed', 'true');
    }
}

/**
 * Развернуть sidebar
 */
function expandSidebar() {
    const sidebar = document.getElementById('mainSidebar');
    sidebar.classList.remove('collapsed');
    sidebarCollapsed = false;
    
    if (!sidebarAutoCollapsed) {
        localStorage.setItem('sidebarCollapsed', 'false');
    }
}

/**
 * Обновлённая showSection — автоматически сворачивает sidebar при переходе в чаты
 */
const originalShowSection = typeof showSection === 'function' ? showSection : null;

// Переключение секций
function showSection(sectionName) {

    // ====== АВТОСВОРАЧИВАНИЕ SIDEBAR ======
    if (sectionName === 'chats' || sectionName === 'activeChat') {
        if (!sidebarCollapsed) {
            collapseSidebar();
            sidebarAutoCollapsed = true;
        }
    } else {
        if (sidebarAutoCollapsed && sidebarCollapsed) {
            expandSidebar();
            sidebarAutoCollapsed = false;
        }
    }
    // ====== КОНЕЦ БЛОКА ======

    // Скрываем все секции
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.remove('active');
    });

    // Убираем active у всех nav items
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });

    // Показываем нужную секцию
    switch(sectionName) {
        case 'home':
            document.getElementById('homeSection').classList.add('active');
            const homeNav = document.querySelector('.nav-item[onclick*="home"]');
            if (homeNav) homeNav.classList.add('active');

            if (TokenManager.isAuthenticated()) {
                const username = TokenManager.getUsername();
                document.getElementById('welcomeTitle').textContent = `Добро пожаловать, ${username}!`;
                document.getElementById('welcomeSubtitle').textContent = 'Рады видеть вас снова';
            } else {
                document.getElementById('welcomeTitle').textContent = 'Добро пожаловать в Micro Chat';
                document.getElementById('welcomeSubtitle').textContent = 'Современный мессенджер для вашего общения';
            }
            break;
        case 'chats':
            document.getElementById('chatsSection').classList.add('active');
            const chatsNav = document.querySelector('.nav-item[onclick*="chats"]');
            if (chatsNav) chatsNav.classList.add('active');
            loadChatList();
            break;
        case 'explore':
            document.getElementById('exploreSection').classList.add('active');
            const exploreNav = document.querySelector('.nav-item[onclick*="explore"]');
            if (exploreNav) exploreNav.classList.add('active');
            loadPublicChats();
            break;
        case 'activeChat':
            document.getElementById('activeChatSection').classList.add('active');
            const chatNavForActive = document.querySelector('.nav-item[onclick*="chats"]');
            if (chatNavForActive) chatNavForActive.classList.add('active');
            break;
        case 'profile':
            document.getElementById('profileSection').classList.add('active');
            const profileNav = document.querySelector('.nav-item[onclick*="profile"]');
            if (profileNav) profileNav.classList.add('active');
            loadUserInfo();
            break;
    }
}

/**
 * Инициализация — восстановление состояния sidebar
 */
function initSidebar() {
    const saved = localStorage.getItem('sidebarCollapsed');
    if (saved === 'true') {
        collapseSidebar();
    }
}

// Вызываем при загрузке
document.addEventListener('DOMContentLoaded', function() {
    initSidebar();
});

// ============================================
// NOTIFICATION MANAGER
// ============================================

const NotificationManager = {

    unreadCounts: {},
    activeChatId: null,
    originalTitle: document.title,
    titleBlinkInterval: null,
    isPageVisible: true,
    notifWs: null,
    reconnectTimer: null,
    initialized: false,

    settings: {
        soundEnabled: true,
        pushEnabled: true,
        toastEnabled: true,
        titleBlinkEnabled: true,
    },

    // ============================================
    // INIT
    // ============================================
    init() {
        if (this.initialized) {
            this.fetchUnreadCounts();
            this.connectWS();
            return;
        }

        this.loadSettings();

        document.addEventListener('visibilitychange', () => {
            this.isPageVisible = !document.hidden;
            if (this.isPageVisible) {
                this.stopTitleBlink();
                this.fetchUnreadCounts();
            }
        });

        this.requestPushPermission();
        this.fetchUnreadCounts();
        this.connectWS();

        this.initialized = true;
        console.log('NotificationManager initialized');
    },

    // ============================================
    // WEBSOCKET
    // ============================================
    connectWS() {
        if (!TokenManager.isAuthenticated()) return;

        // Закрываем старое соединение
        if (this.notifWs) {
            this.notifWs.close();
            this.notifWs = null;
        }

        let token = TokenManager.getAccessToken();
        if (token && token.startsWith('Bearer ')) {
            token = token.substring(7);
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = protocol + '//' + window.location.host +
            '/ws/notifications?token=' + encodeURIComponent(token);

        this.notifWs = new WebSocket(url);

        this.notifWs.onopen = () => {
            console.log('🔔 Notifications connected');
            if (this.reconnectTimer) {
                clearTimeout(this.reconnectTimer);
                this.reconnectTimer = null;
            }
        };

        this.notifWs.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleWsMessage(data);
            } catch (e) {
                console.warn('Parse notification error:', e);
            }
        };

        this.notifWs.onclose = (event) => {
            console.log('🔔 Notifications disconnected');
            this.notifWs = null;

            if (TokenManager.isAuthenticated()) {
                this.reconnectTimer = setTimeout(() => {
                    console.log('🔔 Reconnecting...');
                    this.connectWS();
                }, 3000);
            }
        };

        this.notifWs.onerror = () => {};
    },

    // ============================================
    // ОБРАБОТКА WS СООБЩЕНИЙ
    // ============================================
    handleWsMessage(data) {
        console.log('🔔 WS message received:', data);
        switch (data.type) {
            case 'new_message':
                this.handleNewMessage(data);
                break;
            case 'send-file': // Добавляем обработку ответа на отправку файла
            case 'file_sent':
                this.handleFileMessage(data);
                break;
            default:
                console.log('Unknown notification type:', data.type, data);
        }
},

    // обработчик файловых сообщений
    handleFileMessage(data) {
        console.log('📁 File message received:', data);
        
        // Извлекаем данные из ответа
        const fileData = data.data || data;
        
        // Преобразуем в формат сообщения
        const messageData = {
            type: 'new_message', // Имитируем обычное сообщение
            chat_id: fileData.chat_id || data.chat_id,
            chat_name: fileData.chat_name || data.chat_name,
            sender_name: fileData.from || TokenManager.getUsername(),
            text: fileData.caption || (fileData.is_image ? '📷 Фото' : '📎 Файл'),
            message_type: fileData.is_image ? 'image' : 'file',
            file_url: fileData.file_url,
            file_name: fileData.file_name,
            file_size: fileData.file_size,
            is_image: fileData.is_image,
            sent_at: new Date().toISOString()
        };
        
        // Обрабатываем как новое сообщение
        this.handleNewMessage(messageData);
        
        // Также вызываем колбэк для обновления чата, если он есть
        if (window.ChatView && typeof window.ChatView.displayMessage === 'function') {
            window.ChatView.displayMessage(messageData);
        } else {
            // Ищем функцию displayChatMessage в глобальной области
            const displayFunc = window.displayChatMessage || window.displayMessage;
            if (displayFunc) {
                displayFunc(messageData);
            }
        }
    },

    handleNewMessage(data) {
        console.log('💬 New message handler:', data); // Добавляем логирование
        
        const chatIdStr = String(data.chat_id);
        
        // Проверяем наличие файла в сообщении
        const hasFile = data.file_url || data.message_type === 'image' || data.message_type === 'file';
        
        // Если это файл, возможно текст уже обработан
        let messageText = data.text || '';
        if (hasFile && !messageText) {
            messageText = data.is_image ? '📷 Фото' : '📎 Файл';
        }

        // Если этот чат открыт и вкладка активна — отмечаем прочитанным
        if (String(this.activeChatId) === chatIdStr && this.isPageVisible) {
            this.markAsReadOnServer(data.chat_id);
            
            // Если чат открыт, показываем сообщение сразу
            if (window.ChatView && typeof window.ChatView.addMessage === 'function') {
                window.ChatView.addMessage(data);
            }
            return;
        }

        // Увеличиваем счётчик только если это не наш собственный файл
        // ИЛИ если чат не открыт
        if (String(this.activeChatId) !== chatIdStr) {
            this.unreadCounts[chatIdStr] = (this.unreadCounts[chatIdStr] || 0) + 1;
        }

        // Обновляем UI
        this.updateChatItemBadge(data.chat_id);
        this.updateTotalBadge();

        // Звук (всегда играем для новых сообщений, кроме своих)
        if (this.settings.soundEnabled && data.sender_name !== TokenManager.getUsername()) {
            this.playSound();
        }

        // Toast (для файлов тоже)
        if (this.settings.toastEnabled && data.sender_name !== TokenManager.getUsername()) {
            let displayText = messageText;
            if (hasFile) {
                if (data.is_image) {
                    displayText = '📷 Фото';
                } else if (data.file_name) {
                    displayText = `📎 ${data.file_name}`;
                }
            }
            
            this.showToast(
                data.chat_id,
                data.chat_name || 'Чат',
                data.sender_name || 'Кто-то',
                displayText
            );
        }

        // Push (только если вкладка не активна)
        if (this.settings.pushEnabled && !this.isPageVisible && data.sender_name !== TokenManager.getUsername()) {
            let pushTitle = data.chat_name || 'Чат';
            let pushBody = data.sender_name + ': ';
            
            if (hasFile) {
                if (data.is_image) {
                    pushBody += '📷 Фото';
                } else if (data.file_name) {
                    pushBody += `📎 ${data.file_name}`;
                }
            } else {
                pushBody += messageText;
            }
            
            this.showPushNotification(
                pushTitle,
                pushBody,
                messageText,
                data.chat_id
            );
        }

        // Title blink
        if (this.settings.titleBlinkEnabled && !this.isPageVisible) {
            let blinkText = data.sender_name || 'Кто-то';
            if (hasFile) {
                blinkText += data.is_image ? ': 📷 Фото' : ': 📎 Файл';
            } else {
                blinkText += ': ' + (messageText || 'Новое сообщение');
            }
            this.startTitleBlink(blinkText);
        }
    },
    // ============================================
    // SERVER SYNC
    // ============================================
    async fetchUnreadCounts() {
        if (!TokenManager.isAuthenticated()) return;

        try {
            let token = TokenManager.getAccessToken();
            if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

            const response = await fetch('/api/chat/unread', {
                headers: { 'Authorization': token }
            });
            if (!response.ok) return;

            const data = await response.json();
            this.unreadCounts = {};

            if (data.counts) {
                for (const [chatId, count] of Object.entries(data.counts)) {
                    if (count > 0) {
                        this.unreadCounts[String(chatId)] = count;
                    }
                }
            }

            if (this.activeChatId) {
                delete this.unreadCounts[String(this.activeChatId)];
            }

            this.updateAllChatBadges();
            this.updateTotalBadge();
        } catch (err) {
            console.warn('Failed to fetch unread counts:', err);
        }
    },

    async markAsReadOnServer(chatId) {
        if (!TokenManager.isAuthenticated()) return;
        try {
            let token = TokenManager.getAccessToken();
            if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

            await fetch('/api/chat/read', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': token
                },
                body: JSON.stringify({ chat_id: Number(chatId) })
            });
        } catch (err) {
            console.warn('Failed to mark as read:', err);
        }
    },

    // ============================================
    // READ STATUS
    // ============================================
    markAsRead(chatId) {
        const chatIdStr = String(chatId);
        if (this.unreadCounts[chatIdStr]) {
            delete this.unreadCounts[chatIdStr];
            this.updateChatItemBadge(chatId);
            this.updateTotalBadge();
        }
        this.markAsReadOnServer(chatId);
    },

    setActiveChat(chatId) {
        this.activeChatId = String(chatId);
        this.markAsRead(chatId);
    },

    clearActiveChat() {
        this.activeChatId = null;
    },

    // ============================================
    // BADGES
    // ============================================
    updateChatItemBadge(chatId) {
        const count = this.unreadCounts[String(chatId)] || 0;
        const item = document.querySelector('.chat-list-item[data-chat-id="' + chatId + '"]');
        if (!item) return;

        let container = item.querySelector('.chat-item-unread');
        let badge = container ? container.querySelector('.unread-badge') : null;

        if (count > 0) {
            item.classList.add('has-unread');
            if (!container) {
                container = document.createElement('div');
                container.className = 'chat-item-unread';
                item.appendChild(container);
            }
            if (!badge) {
                badge = document.createElement('span');
                badge.className = 'unread-badge';
                container.appendChild(badge);
            }
            badge.textContent = count > 99 ? '99+' : count;
        } else {
            item.classList.remove('has-unread');
            if (container) container.remove();
        }
    },

    updateAllChatBadges() {
        document.querySelectorAll('.chat-list-item').forEach(item => {
            const id = item.getAttribute('data-chat-id');
            if (id) this.updateChatItemBadge(id);
        });
    },

    updateTotalBadge() {
        const total = Object.values(this.unreadCounts).reduce((s, c) => s + c, 0);
        const badge = document.getElementById('totalUnreadBadge');
        if (badge) {
            badge.textContent = total > 99 ? '99+' : total;
            badge.style.display = total > 0 ? 'flex' : 'none';
        }
        document.title = total > 0
            ? '(' + total + ') ' + this.originalTitle
            : (this.isPageVisible ? this.originalTitle : document.title);
    },

    // ============================================
    // SOUND
    // ============================================
    playSound() {
        try {
            const ctx = new (window.AudioContext || window.webkitAudioContext)();
            const t = ctx.currentTime;

            const o1 = ctx.createOscillator();
            const g1 = ctx.createGain();
            o1.connect(g1); g1.connect(ctx.destination);
            o1.frequency.value = 830; o1.type = 'sine';
            g1.gain.setValueAtTime(0.3, t);
            g1.gain.exponentialRampToValueAtTime(0.01, t + 0.15);
            o1.start(t); o1.stop(t + 0.15);

            const o2 = ctx.createOscillator();
            const g2 = ctx.createGain();
            o2.connect(g2); g2.connect(ctx.destination);
            o2.frequency.value = 1050; o2.type = 'sine';
            g2.gain.setValueAtTime(0.2, t + 0.12);
            g2.gain.exponentialRampToValueAtTime(0.01, t + 0.3);
            o2.start(t + 0.12); o2.stop(t + 0.3);
        } catch (e) {}
    },

    // ============================================
    // TOAST
    // ============================================
    showToast(chatId, chatName, senderName, text) {
        const container = document.getElementById('toastContainer');
        if (!container) return;

        while (container.children.length >= 3) {
            this.removeToast(container.firstChild);
        }

        const initials = senderName ? senderName.substring(0, 2).toUpperCase() : '??';
        const toast = document.createElement('div');
        toast.className = 'toast';
        toast.innerHTML =
            '<div class="toast-avatar">' + this.esc(initials) + '</div>' +
            '<div class="toast-body">' +
                '<div class="toast-header">' +
                    '<div>' +
                        '<span class="toast-sender">' + this.esc(senderName) + '</span>' +
                        ' <span class="toast-chat-name">&middot; ' + this.esc(chatName) + '</span>' +
                    '</div>' +
                    '<span class="toast-time">сейчас</span>' +
                '</div>' +
                '<div class="toast-message">' + this.esc(this.trunc(text, 80)) + '</div>' +
            '</div>' +
            '<button class="toast-close" onclick="NotificationManager.removeToast(this.parentElement)">' +
                '<i class="fas fa-times"></i>' +
            '</button>';

        toast.addEventListener('click', (e) => {
            if (e.target.closest('.toast-close')) return;
            this.removeToast(toast);
            this.openFromNotification(chatId);
        });

        container.appendChild(toast);
        setTimeout(() => this.removeToast(toast), 5000);
    },

    removeToast(t) {
        if (!t || t.classList.contains('toast-hiding')) return;
        t.classList.add('toast-hiding');
        setTimeout(() => { if (t.parentElement) t.parentElement.removeChild(t); }, 300);
    },

    // ============================================
    // PUSH
    // ============================================
    requestPushPermission() {
        if (!('Notification' in window)) return;
        if (Notification.permission === 'default' && TokenManager.isAuthenticated()) {
            this.showPermBanner();
        }
    },

    showPermBanner() {
        if (document.getElementById('notifPermBanner')) return;
        const b = document.createElement('div');
        b.id = 'notifPermBanner';
        b.className = 'notification-permission-banner';
        b.innerHTML =
            '<i class="fas fa-bell" style="font-size:16px;color:#6c63ff;flex-shrink:0;"></i>' +
            '<span>Включить уведомления?</span>' +
            '<button onclick="NotificationManager.askPerm()">Да</button>' +
            '<button onclick="this.parentElement.remove()" style="background:transparent;color:rgba(255,255,255,0.4);padding:5px 8px;">Нет</button>';
        const nav = document.querySelector('.sidebar-nav');
        if (nav) nav.after(b);
    },

    async askPerm() {
        try { await Notification.requestPermission(); } catch (e) {}
        const b = document.getElementById('notifPermBanner');
        if (b) b.remove();
    },

    showPushNotification(chatName, sender, text, chatId) {
        if (!('Notification' in window) || Notification.permission !== 'granted') return;
        try {
            const n = new Notification(sender + ' · ' + chatName, {
                body: this.trunc(text, 100),
                icon: '/static/favicon.ico',
                tag: 'chat-' + chatId,
                renotify: true,
            });
            n.onclick = () => { window.focus(); this.openFromNotification(chatId); n.close(); };
            setTimeout(() => n.close(), 5000);
        } catch (e) {}
    },

    // ============================================
    // TITLE BLINK
    // ============================================
    startTitleBlink(sender, text) {
        if (this.titleBlinkInterval) return;
        const msg = '💬 ' + sender + ': ' + this.trunc(text, 30);
        let show = true;
        this.titleBlinkInterval = setInterval(() => {
            document.title = show ? msg : this.originalTitle;
            show = !show;
        }, 1500);
    },

    stopTitleBlink() {
        if (this.titleBlinkInterval) {
            clearInterval(this.titleBlinkInterval);
            this.titleBlinkInterval = null;
        }
        this.updateTotalBadge();
    },

    // ============================================
    // NAVIGATION
    // ============================================
    openFromNotification(chatId) {
        if (typeof showSection === 'function') showSection('chats');
        setTimeout(() => {
            const item = document.querySelector('.chat-list-item[data-chat-id="' + chatId + '"]');
            if (item) item.click();
        }, 500);
    },

    // ============================================
    // UTILS
    // ============================================
    loadSettings() {
        try {
            const s = localStorage.getItem('notifSettings');
            if (s) this.settings = { ...this.settings, ...JSON.parse(s) };
        } catch (e) {}
    },

    esc(t) {
        if (!t) return '';
        const d = document.createElement('div');
        d.textContent = t;
        return d.innerHTML;
    },

    trunc(t, m) {
        if (!t) return '';
        return t.length > m ? t.substring(0, m) + '…' : t;
    },

    reset() {
        this.unreadCounts = {};
        this.activeChatId = null;
        this.initialized = false;
        this.stopTitleBlink();
        this.updateTotalBadge();
        document.title = this.originalTitle;

        if (this.notifWs) { this.notifWs.close(); this.notifWs = null; }
        if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }

        const b = document.getElementById('notifPermBanner');
        if (b) b.remove();
        const c = document.getElementById('toastContainer');
        if (c) c.innerHTML = '';
    }
};

// ============================================
// ФАЙЛЫ И КАРТИНКИ
// ============================================

let pendingFile = null;

// Выбор файла через кнопку 📎
function handleFileSelect(input) {
    var file = input.files[0];
    if (!file) return;

    if (file.size > 20 * 1024 * 1024) {
        alert('Файл слишком большой. Максимум 20MB');
        input.value = '';
        return;
    }

    pendingFile = file;
    showFilePreview(file);
    input.value = '';
}

// Показать превью файла
function showFilePreview(file) {
    var preview = document.getElementById('filePreview');
    var content = document.getElementById('filePreviewContent');
    if (!preview || !content) return;

    var isImage = file.type.startsWith('image/');
    var sizeText = formatFileSize(file.size);

    if (isImage) {
        var reader = new FileReader();
        reader.onload = function(e) {
            content.innerHTML =
                '<img src="' + e.target.result + '" class="file-preview__image" alt="">' +
                '<div class="file-preview__info">' +
                    '<div class="file-preview__name">' + escapeHtml(file.name) + '</div>' +
                    '<div class="file-preview__size">' + sizeText + '</div>' +
                '</div>';
        };
        reader.readAsDataURL(file);
    } else {
        var icon = getFileIcon(file.name);
        content.innerHTML =
            '<div class="message-file__icon"><i class="fas ' + icon + '"></i></div>' +
            '<div class="file-preview__info">' +
                '<div class="file-preview__name">' + escapeHtml(file.name) + '</div>' +
                '<div class="file-preview__size">' + sizeText + '</div>' +
            '</div>';
    }

    preview.style.display = 'flex';

    // Фокус на поле ввода для добавления подписи
    document.getElementById('chatViewInput').focus();
}

// Отменить загрузку файла
function cancelFileUpload() {
    pendingFile = null;
    var preview = document.getElementById('filePreview');
    if (preview) preview.style.display = 'none';
}

// Отправка файла
async function sendFileMessage() {
    if (!pendingFile || !activeChatId) return;

    var file = pendingFile;
    var textInput = document.getElementById('chatViewInput');
    var text = textInput ? textInput.value.trim() : '';

    // Блокируем повторную отправку
    pendingFile = null;
    var preview = document.getElementById('filePreview');
    if (preview) preview.style.display = 'none';
    if (textInput) textInput.value = '';

    var token = TokenManager.getAccessToken();
    if (!token.startsWith('Bearer ')) token = 'Bearer ' + token;

    var formData = new FormData();
    formData.append('chat_id', activeChatId);
    formData.append('file', file);
    if (text) formData.append('text', text);

    try {
        var response = await fetch('/api/chat/send-file', {
            method: 'POST',
            headers: { 'Authorization': token },
            body: formData
        });

        if (!response.ok) {
            var data = await response.json();
            alert('Ошибка: ' + (data.error || 'Не удалось отправить'));
        }
    } catch (err) {
        alert('Ошибка отправки: ' + err.message);
    }
}

// ============================================
// DRAG & DROP
// ============================================

function initDragDrop() {
    var messagesArea = document.getElementById('chatViewMessages');
    var chatActive = document.getElementById('chatViewActive');
    if (!chatActive) return;

    var dragCounter = 0;

    chatActive.addEventListener('dragenter', function(e) {
        e.preventDefault();
        dragCounter++;
        var overlay = document.getElementById('dragOverlay');
        if (overlay) overlay.style.display = 'flex';
    });

    chatActive.addEventListener('dragleave', function(e) {
        e.preventDefault();
        dragCounter--;
        if (dragCounter <= 0) {
            dragCounter = 0;
            var overlay = document.getElementById('dragOverlay');
            if (overlay) overlay.style.display = 'none';
        }
    });

    chatActive.addEventListener('dragover', function(e) {
        e.preventDefault();
    });

    chatActive.addEventListener('drop', function(e) {
        e.preventDefault();
        dragCounter = 0;
        var overlay = document.getElementById('dragOverlay');
        if (overlay) overlay.style.display = 'none';

        var files = e.dataTransfer.files;
        if (files.length > 0) {
            var file = files[0];
            if (file.size > 20 * 1024 * 1024) {
                alert('Файл слишком большой. Максимум 20MB');
                return;
            }
            pendingFile = file;
            showFilePreview(file);
        }
    });
}

// Инициализация при загрузке
document.addEventListener('DOMContentLoaded', function() {
    initDragDrop();
});

// ============================================
// УТИЛИТЫ
// ============================================

function isImageUrl(url) {
    if (!url) return false;
    var lower = url.toLowerCase();
    return lower.match(/\.(jpg|jpeg|png|gif|webp|svg)/) !== null ||
           lower.includes('/images/');
}

function getFileIcon(filename) {
    if (!filename) return 'fa-file';
    var ext = filename.split('.').pop().toLowerCase();
    var icons = {
        'pdf': 'fa-file-pdf',
        'doc': 'fa-file-word', 'docx': 'fa-file-word',
        'xls': 'fa-file-excel', 'xlsx': 'fa-file-excel',
        'ppt': 'fa-file-powerpoint', 'pptx': 'fa-file-powerpoint',
        'zip': 'fa-file-archive', 'rar': 'fa-file-archive', '7z': 'fa-file-archive',
        'txt': 'fa-file-alt',
        'mp3': 'fa-file-audio', 'wav': 'fa-file-audio',
        'mp4': 'fa-file-video', 'avi': 'fa-file-video', 'mkv': 'fa-file-video',
    };
    return icons[ext] || 'fa-file';
}

function formatFileSize(bytes) {
    if (!bytes) return '';
    if (bytes < 1024) return bytes + ' Б';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' КБ';
    return (bytes / (1024 * 1024)).toFixed(1) + ' МБ';
}

// Просмотр картинки на весь экран
function openImageModal(url) {
    var modal = document.createElement('div');
    modal.className = 'image-modal';
    modal.onclick = function() { modal.remove(); };
    modal.innerHTML =
        '<img src="' + escapeHtml(url) + '" alt="">' +
        '<button class="image-modal__close" onclick="this.parentElement.remove()">' +
            '<i class="fas fa-times"></i>' +
        '</button>';
    document.body.appendChild(modal);
}

// Закрытие по Escape
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
        var modal = document.querySelector('.image-modal');
        if (modal) modal.remove();
    }
});

// Вставка из буфера обмена (Ctrl+V)
document.addEventListener('paste', function(e) {
    if (!activeChatId) return;

    var chatInput = document.getElementById('chatViewInput');
    if (document.activeElement !== chatInput) return;

    var items = e.clipboardData.items;
    for (var i = 0; i < items.length; i++) {
        if (items[i].type.startsWith('image/')) {
            e.preventDefault();
            var file = items[i].getAsFile();
            if (file) {
                pendingFile = file;
                showFilePreview(file);
            }
            break;
        }
    }
});

// ==================== INITIALIZATION ====================
document.addEventListener('DOMContentLoaded', () => {
    console.log('🚀 App initializing...');
    checkTokenOnLoad();
    updateAuthStatus();
    NotificationManager.init();
    initMessageInput();
    // Запускаем автообновление если авторизован
    if (TokenManager.isAuthenticated()) {
        loadCurrentUserProfile();
        console.log('👤 Пользователь авторизован, загружаем чаты...');
        setTimeout(() => {
            loadChatCount();
            startChatCountUpdater();
        }, 100);
    }
    console.log('✅ App initialized');
});