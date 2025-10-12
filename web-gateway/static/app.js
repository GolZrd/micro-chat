// Утилита для работы с токенами
const TokenManager = {
    // Получить access token из localStorage
    getAccessToken() {
        return localStorage.getItem('access_token');
    },
    
    // Сохранить access token
    setAccessToken(token) {
        localStorage.setItem('access_token', token);
        
        // Извлекаем username для отображения
        try {
            const claims = this.decodeJWT(token);
            localStorage.setItem('username', claims.name);
            console.log('✅ Username extracted:', claims.name);
        } catch (e) {
            console.error('❌ Failed to decode token:', e);
        }
    },
    
    // Frontend декодирует JWT для UI
    decodeJWT(token) {
        const base64Url = token.replace('Bearer ', '').split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(base64));
    },
    
    getUsername() {
        return localStorage.getItem('username');
    },
    
    // Удалить access token
    removeAccessToken() {
        localStorage.removeItem('access_token');
    },
    
    // Получить user_id
    getUserId() {
        return localStorage.getItem('user_id');
    },
    
    setUserId(id) {
        localStorage.setItem('user_id', id);
    },
    
    // Проверка наличия токена
    isAuthenticated() {
        return !!this.getAccessToken();
    },
    
    // ✅ ДОБАВЛЕНО: Очистка всех данных
    clear() {
        localStorage.removeItem('access_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('username');
    }
};

// Функция для обновления access token
async function refreshAccessToken() {
    try {
        console.log('🔄 Refreshing access token...');
        
        const response = await fetch('/api/refresh', {
            method: 'POST',
            credentials: 'include' // Важно! Отправляет refresh_token cookie
        });
        
        if (response.ok) {
            const data = await response.json();
            
            if (data.access_token) {
                // Убеждаемся что токен с префиксом Bearer
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
            return false;
        }
    } catch (error) {
        console.error('❌ Error refreshing token:', error);
        return false;
    }
}

// Утилита для API запросов с автоматическим обновлением токена
async function apiRequest(url, options = {}) {
    const token = TokenManager.getAccessToken();
    
    // Добавляем токен в заголовки
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (token) {
        headers['Authorization'] = token; // Токен уже содержит "Bearer "
    }
    
    let response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include' // Важно для cookies!
    });
    
    // Проверяем 401 (Unauthorized) или 403 (PermissionDenied от gRPC)
    if ((response.status === 401 || response.status === 403) && token) {
        console.log('⚠️ Access token expired (status ' + response.status + '), attempting refresh...');
        
        // Пробуем обновить токен
        const refreshed = await refreshAccessToken();
        
        if (refreshed) {
            // Повторяем запрос с новым токеном
            headers['Authorization'] = TokenManager.getAccessToken();
            response = await fetch(url, {
                ...options,
                headers,
                credentials: 'include'
            });
            
            console.log('✅ Request retried with new token');
        } else {
            // Не удалось обновить
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

// Проверка токена при загрузке страницы
async function checkTokenOnLoad() {
    if (!TokenManager.isAuthenticated()) {
        return;
    }

    try {
        // Делаем простой запрос для проверки токена
        const response = await fetch('/api/chat/my', {
            method: 'GET',
            headers: {
                'Authorization': TokenManager.getAccessToken()
            },
            credentials: 'include'
        });

        // Если токен истек, пробуем обновить
        if (response.status === 401 || response.status === 403) {
            console.log('⚠️ Token expired on page load, refreshing...');
            
            const refreshed = await refreshAccessToken();
            
            if (!refreshed) {
                console.log('❌ Could not refresh token, logging out');
                TokenManager.clear();
                updateAuthStatus();
            } else {
                console.log('✅ Token refreshed on page load');
                updateAuthStatus();
            }
        }
    } catch (error) {
        console.error('❌ Error checking token:', error);
    }
}


// ✅ ИЗМЕНЕНО: Показываем username вместо user_id
function updateAuthStatus() {
    const isAuth = TokenManager.isAuthenticated();
    const statusEl = document.getElementById('status');
    const logoutBtn = document.getElementById('logoutBtn');
    const protectedContent = document.querySelectorAll('.protected-content');
    const guestOnly = document.querySelectorAll('.guest-only'); // ✅ НОВОЕ
    
    if (isAuth) {
        const username = TokenManager.getUsername();
        const userId = TokenManager.getUserId();
        
        statusEl.innerHTML = `✅ Вы авторизованы как <strong>${username}</strong> <small>(ID: ${userId})</small>`;
        statusEl.style.color = 'green';
        logoutBtn.style.display = 'inline-block';
        
        // Показываем защищенный контент
        protectedContent.forEach(el => el.style.display = 'block');
        
        // ✅ НОВОЕ: Скрываем формы регистрации и входа
        guestOnly.forEach(el => el.style.display = 'none');
    } else {
        statusEl.textContent = '❌ Не авторизован';
        statusEl.style.color = 'red';
        logoutBtn.style.display = 'none';
        
        // Скрываем защищенный контент
        protectedContent.forEach(el => el.style.display = 'none');
        
        // ✅ НОВОЕ: Показываем формы регистрации и входа
        guestOnly.forEach(el => el.style.display = 'block');
    }
}

// Регистрация
document.getElementById('registerForm').addEventListener('submit', async (e) => {
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
            
            // Очищаем форму
            document.getElementById('registerForm').reset();
        } else {
            resultEl.innerHTML = `❌ Ошибка: ${result.error}`;
            resultEl.style.background = '#f8d7da';
            resultEl.style.color = '#721c24';
        }
    } catch (error) {
        document.getElementById('registerResult').innerHTML = `❌ ${error}`;
    }
});

// Вход
document.getElementById('loginForm').addEventListener('submit', async (e) => {
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
            // ✅ ИЗМЕНЕНО: Убеждаемся что токен с префиксом Bearer
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
            
            // Очищаем форму
            document.getElementById('loginForm').reset();
        } else {
            resultEl.innerHTML = `❌ Ошибка: ${result.error}`;
            resultEl.style.background = '#f8d7da';
            resultEl.style.color = '#721c24';
        }
    } catch (error) {
        document.getElementById('loginResult').innerHTML = `❌ ${error}`;
    }
});

// ✅ ИЗМЕНЕНО: Используем TokenManager.clear()
async function logout() {
    try {
        await fetch('/api/logout', {
            method: 'POST',
            credentials: 'include'
        });
        
        TokenManager.clear();
        updateAuthStatus();
        alert('✅ Вы вышли из системы');
        
        // Перезагружаем страницу для очистки
        location.reload();
    } catch (error) {
        alert('❌ Ошибка при выходе: ' + error);
    }
}

// Создание чата
document.getElementById('createChatForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const usernames = document.getElementById('chat_usernames').value.split(',').map(s => s.trim());

    try {
        const response = await apiRequest('/api/chat/create', {
            method: 'POST',
            body: JSON.stringify({usernames})
        });
        
        const result = await response.json();
        const resultEl = document.getElementById('chatResult');
        
        if (response.ok) {
            resultEl.innerHTML = `✅ Чат создан! <a href="/chat?id=${result.chat_id}" style="color: #007bff;">Открыть →</a>`;
            resultEl.style.background = '#d4edda';
            resultEl.style.color = '#155724';
            
            // Очищаем форму
            document.getElementById('chat_usernames').value = '';
            
            // Обновляем список чатов если он загружен
            loadMyChats();
        } else {
            resultEl.innerHTML = `❌ ${result.error}`;
            resultEl.style.background = '#f8d7da';
            resultEl.style.color = '#721c24';
        }
    } catch (error) {
        document.getElementById('chatResult').innerHTML = `❌ ${error}`;
    }
});

// ✅ НОВОЕ: Функция удаления чата
async function deleteChat(chatId) {
    if (!confirm(`Вы уверены, что хотите удалить чат #${chatId}?`)) {
        return;
    }

    // Находим карточку чата
    const chatCard = event.target.closest('.chat-card');
    
    try {
        // Добавляем анимацию удаления
        if (chatCard) {
            chatCard.style.opacity = '0.5';
            chatCard.style.pointerEvents = 'none';
        }

        const response = await apiRequest(`/api/chat/delete/${chatId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            // Анимация исчезновения
            if (chatCard) {
                chatCard.classList.add('deleting');
                setTimeout(() => {
                    loadMyChats(); // Перезагружаем список после анимации
                }, 300);
            } else {
                loadMyChats();
            }
        } else {
            const error = await response.json();
            alert('❌ Ошибка удаления: ' + (error.error || 'Неизвестная ошибка'));
            
            // Возвращаем карточку в нормальное состояние
            if (chatCard) {
                chatCard.style.opacity = '1';
                chatCard.style.pointerEvents = 'auto';
            }
        }
    } catch (error) {
        console.error('❌ Delete error:', error);
        alert('❌ Ошибка удаления чата: ' + error.message);
        
        // Возвращаем карточку в нормальное состояние
        if (chatCard) {
            chatCard.style.opacity = '1';
            chatCard.style.pointerEvents = 'auto';
        }
    }
}

// Загрузка чатов пользователя с кнопкой удаления
async function loadMyChats() {
    const chatsDiv = document.getElementById('myChats');
    chatsDiv.innerHTML = '<p style="color: #666;">⏳ Загрузка...</p>';
    
    try {
        const response = await apiRequest('/api/chat/my');
        const data = await response.json();
        
        console.log('📦 Server response:', data);
        
        if (!response.ok) {
            chatsDiv.innerHTML = `<p style="color: #dc3545;">❌ ${data.error || 'Ошибка загрузки'}</p>`;
            return;
        }
        
        let chats = data.chats || [];
        chats = chats.filter(chat => chat && chat.id);
        
        console.log('✅ Filtered chats:', chats);
        
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
                            onclick="deleteChat(${chatId})" 
                            class="btn-delete"
                            title="Удалить чат">
                            🗑️
                        </button>
                    </div>
                    <p><strong>👥 Участники:</strong> ${usersList}</p>
                    <p><strong>📅 Создан:</strong> ${createdDate}</p>
                    <a href="/chat?id=${chatId}" class="btn-open-chat">
                       Открыть чат →
                    </a>
                </div>
            `;
        });
        html += '</div>';
        
        chatsDiv.innerHTML = html;
        
    } catch (error) {
        console.error('❌ Error:', error);
        chatsDiv.innerHTML = `<p style="color: #dc3545;">❌ Ошибка: ${error.message}</p>`;
    }
}

// Загрузка информации о пользователе
async function loadUserInfo() {
    const userId = TokenManager.getUserId();
    if (!userId) {
        alert('User ID не найден');
        return;
    }

    const infoDiv = document.getElementById('userInfo');
    infoDiv.innerHTML = '<p style="color: #666;">⏳ Загрузка...</p>';

    try {
        const response = await apiRequest(`/api/user/${userId}`);
        const user = await response.json();
        
        if (!response.ok) {
            infoDiv.innerHTML = `<p style="color: #dc3545;">❌ ${user.error}</p>`;
            return;
        }
        
        infoDiv.innerHTML = `
            <div class="user-info-card">
                <p><strong>🆔 ID:</strong> ${user.id}</p>
                <p><strong>👤 Имя:</strong> ${user.name}</p>
                <p><strong>📧 Email:</strong> ${user.email}</p>
                <p><strong>🎭 Роль:</strong> ${user.role}</p>
                <p><strong>📅 Создан:</strong> ${new Date(user.created_at).toLocaleString('ru-RU')}</p>
            </div>
        `;
    } catch (error) {
        infoDiv.innerHTML = `<p style="color: #dc3545;">❌ ${error}</p>`;
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    updateAuthStatus();
    checkTokenOnLoad(); // Проверяем токен при загрузке
});