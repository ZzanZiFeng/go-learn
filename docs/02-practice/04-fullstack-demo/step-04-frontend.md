# Step 4: 前端集成

## 目标

创建简单的前端页面演示 API 集成。

## 4.1 HTML 页面

创建 `frontend/index.html`:

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo App - Fullstack Demo</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            margin-bottom: 20px;
            color: #333;
        }
        .auth-section, .todo-section {
            background: white;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
        }
        input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        button {
            padding: 10px 20px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            margin-right: 10px;
        }
        button:hover {
            background: #0056b3;
        }
        button.secondary {
            background: #6c757d;
        }
        button.danger {
            background: #dc3545;
        }
        .todo-list {
            list-style: none;
        }
        .todo-item {
            display: flex;
            align-items: center;
            padding: 15px;
            border-bottom: 1px solid #eee;
        }
        .todo-item:last-child {
            border-bottom: none;
        }
        .todo-item.completed .todo-title {
            text-decoration: line-through;
            color: #888;
        }
        .todo-checkbox {
            width: 20px;
            height: 20px;
            margin-right: 15px;
        }
        .todo-title {
            flex: 1;
        }
        .todo-delete {
            padding: 5px 10px;
            font-size: 12px;
        }
        .user-info {
            background: #e7f3ff;
            padding: 10px;
            border-radius: 4px;
            margin-bottom: 15px;
        }
        .hidden {
            display: none;
        }
        .error {
            color: #dc3545;
            margin-top: 10px;
        }
        .success {
            color: #28a745;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Todo App</h1>

        <!-- 认证区域 -->
        <div id="auth-section" class="auth-section">
            <div id="login-form">
                <h2>登录</h2>
                <div class="form-group">
                    <label>邮箱</label>
                    <input type="email" id="login-email" placeholder="test@example.com">
                </div>
                <div class="form-group">
                    <label>密码</label>
                    <input type="password" id="login-password" placeholder="Password123">
                </div>
                <button onclick="login()">登录</button>
                <button class="secondary" onclick="showRegister()">注册</button>
                <div id="login-error" class="error hidden"></div>
            </div>

            <div id="register-form" class="hidden">
                <h2>注册</h2>
                <div class="form-group">
                    <label>用户名</label>
                    <input type="text" id="register-username" placeholder="testuser">
                </div>
                <div class="form-group">
                    <label>邮箱</label>
                    <input type="email" id="register-email" placeholder="test@example.com">
                </div>
                <div class="form-group">
                    <label>密码</label>
                    <input type="password" id="register-password" placeholder="至少8位，包含大小写字母和数字">
                </div>
                <button onclick="register()">注册</button>
                <button class="secondary" onclick="showLogin()">返回登录</button>
                <div id="register-error" class="error hidden"></div>
            </div>
        </div>

        <!-- 待办区域 -->
        <div id="todo-section" class="todo-section hidden">
            <div class="user-info">
                <span>欢迎，<strong id="username"></strong></span>
                <button class="secondary" style="float:right" onclick="logout()">登出</button>
            </div>

            <h2>我的待办</h2>

            <div class="form-group" style="display:flex;gap:10px">
                <input type="text" id="new-todo" placeholder="添加新待办..." style="flex:1">
                <button onclick="addTodo()">添加</button>
            </div>

            <ul id="todo-list" class="todo-list">
                <!-- 待办列表将在这里渲染 -->
            </ul>
        </div>
    </div>

    <script>
        const API_BASE = '';
        let accessToken = localStorage.getItem('access_token');
        let refreshToken = localStorage.getItem('refresh_token');

        // 初始化
        document.addEventListener('DOMContentLoaded', () => {
            if (accessToken) {
                showTodoSection();
                loadTodos();
            }
        });

        // 显示注册表单
        function showRegister() {
            document.getElementById('login-form').classList.add('hidden');
            document.getElementById('register-form').classList.remove('hidden');
        }

        // 显示登录表单
        function showLogin() {
            document.getElementById('register-form').classList.add('hidden');
            document.getElementById('login-form').classList.remove('hidden');
        }

        // 显示待办区域
        function showTodoSection() {
            document.getElementById('auth-section').classList.add('hidden');
            document.getElementById('todo-section').classList.remove('hidden');
            fetchUserInfo();
        }

        // 显示认证区域
        function showAuthSection() {
            document.getElementById('todo-section').classList.add('hidden');
            document.getElementById('auth-section').classList.remove('hidden');
        }

        // 登录
        async function login() {
            const email = document.getElementById('login-email').value;
            const password = document.getElementById('login-password').value;

            try {
                const response = await fetch(`${API_BASE}/api/auth/login`, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({email, password})
                });

                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || '登录失败');
                }

                const data = await response.json();
                saveTokens(data.access_token, data.refresh_token);
                showTodoSection();
                loadTodos();
            } catch (error) {
                showError('login-error', error.message);
            }
        }

        // 注册
        async function register() {
            const username = document.getElementById('register-username').value;
            const email = document.getElementById('register-email').value;
            const password = document.getElementById('register-password').value;

            try {
                const response = await fetch(`${API_BASE}/api/auth/register`, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({username, email, password})
                });

                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || '注册失败');
                }

                const data = await response.json();
                saveTokens(data.access_token, data.refresh_token);
                showTodoSection();
                loadTodos();
            } catch (error) {
                showError('register-error', error.message);
            }
        }

        // 登出
        function logout() {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            accessToken = null;
            refreshToken = null;
            showAuthSection();
        }

        // 获取用户信息
        async function fetchUserInfo() {
            try {
                const response = await fetchWithAuth(`${API_BASE}/api/users/me`);
                const user = await response.json();
                document.getElementById('username').textContent = user.username;
            } catch (error) {
                console.error('获取用户信息失败:', error);
            }
        }

        // 加载待办列表
        async function loadTodos() {
            try {
                const response = await fetchWithAuth(`${API_BASE}/api/todos`);
                const data = await response.json();
                renderTodos(data.todos || []);
            } catch (error) {
                console.error('加载待办失败:', error);
            }
        }

        // 渲染待办列表
        function renderTodos(todos) {
            const list = document.getElementById('todo-list');
            list.innerHTML = todos.map(todo => `
                <li class="todo-item ${todo.completed ? 'completed' : ''}">
                    <input type="checkbox" class="todo-checkbox"
                        ${todo.completed ? 'checked' : ''}
                        onchange="toggleTodo(${todo.id}, this.checked)">
                    <span class="todo-title">${escapeHtml(todo.title)}</span>
                    <button class="todo-delete danger" onclick="deleteTodo(${todo.id})">删除</button>
                </li>
            `).join('');
        }

        // 添加待办
        async function addTodo() {
            const input = document.getElementById('new-todo');
            const title = input.value.trim();
            if (!title) return;

            try {
                await fetchWithAuth(`${API_BASE}/api/todos`, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({title})
                });
                input.value = '';
                loadTodos();
            } catch (error) {
                console.error('添加待办失败:', error);
            }
        }

        // 切换待办状态
        async function toggleTodo(id, completed) {
            try {
                await fetchWithAuth(`${API_BASE}/api/todos/${id}`, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({completed})
                });
                loadTodos();
            } catch (error) {
                console.error('更新待办失败:', error);
            }
        }

        // 删除待办
        async function deleteTodo(id) {
            try {
                await fetchWithAuth(`${API_BASE}/api/todos/${id}`, {
                    method: 'DELETE'
                });
                loadTodos();
            } catch (error) {
                console.error('删除待办失败:', error);
            }
        }

        // 带认证的 fetch
        async function fetchWithAuth(url, options = {}) {
            options.headers = options.headers || {};
            options.headers['Authorization'] = `Bearer ${accessToken}`;

            let response = await fetch(url, options);

            if (response.status === 401 && refreshToken) {
                // 尝试刷新令牌
                const refreshed = await refreshTokens();
                if (refreshed) {
                    options.headers['Authorization'] = `Bearer ${accessToken}`;
                    response = await fetch(url, options);
                } else {
                    logout();
                }
            }

            return response;
        }

        // 刷新令牌
        async function refreshTokens() {
            try {
                const response = await fetch(`${API_BASE}/api/auth/refresh`, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({refresh_token: refreshToken})
                });

                if (!response.ok) return false;

                const data = await response.json();
                saveTokens(data.access_token, data.refresh_token);
                return true;
            } catch {
                return false;
            }
        }

        // 保存令牌
        function saveTokens(access, refresh) {
            accessToken = access;
            refreshToken = refresh;
            localStorage.setItem('access_token', access);
            localStorage.setItem('refresh_token', refresh);
        }

        // 显示错误
        function showError(elementId, message) {
            const element = document.getElementById(elementId);
            element.textContent = message;
            element.classList.remove('hidden');
            setTimeout(() => element.classList.add('hidden'), 3000);
        }

        // HTML 转义
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // 回车添加待办
        document.getElementById('new-todo')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') addTodo();
        });
    </script>
</body>
</html>
```

## 4.2 功能说明

### 认证功能

- 用户注册（用户名、邮箱、密码）
- 用户登录
- 自动刷新令牌
- 登出

### 待办功能

- 查看待办列表
- 添加新待办
- 标记完成/未完成
- 删除待办

### 令牌管理

- 存储在 localStorage
- 自动附加到请求头
- 401 时自动刷新

## 4.3 API 调用流程

```
┌────────────────────────────────────────────────────────────────┐
│                      前端 API 调用流程                          │
└────────────────────────────────────────────────────────────────┘

 用户操作           前端                    Nginx              后端
    │                │                       │                   │
    │ 1. 登录        │                       │                   │
    │ ──────────────▶│                       │                   │
    │                │ POST /api/auth/login  │                   │
    │                │ ─────────────────────▶│                   │
    │                │                       │ ─────────────────▶│
    │                │                       │◀───────────────── │
    │                │◀───────────────────── │ access + refresh │
    │                │                       │                   │
    │                │ 存储令牌到 localStorage │                   │
    │                │                       │                   │
    │ 2. 获取待办    │                       │                   │
    │ ──────────────▶│                       │                   │
    │                │ GET /api/todos        │                   │
    │                │ Authorization: Bearer │                   │
    │                │ ─────────────────────▶│                   │
    │                │                       │ ─────────────────▶│
    │                │                       │◀───────────────── │
    │                │◀───────────────────── │   待办列表        │
    │                │                       │                   │
    │◀────────────── │ 渲染待办列表          │                   │
    │                │                       │                   │
```

## 项目总结

恭喜！你已完成 Fullstack Demo 项目，掌握了：

| 技能 | 说明 |
|------|------|
| 微服务架构 | 服务拆分、通信 |
| API 网关 | Nginx 反向代理 |
| Docker 部署 | 容器化、编排 |
| 前后端集成 | REST API、JWT |

## 下一步

继续深入学习：

- **[部署轨道](../../03-deployment/)** - 生产环境部署、CI/CD
- **[测试轨道](../../04-testing/)** - 单元测试、集成测试
