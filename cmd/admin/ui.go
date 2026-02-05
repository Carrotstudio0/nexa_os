package main

const LoginHTML = `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexa Protocol | Login</title>
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@300;400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #6C5DD3;
            --bg: #1f2029;
            --card: #2a2b36;
            --text: #ffffff;
        }
        body {
            font-family: 'Cairo', sans-serif;
            background-color: var(--bg);
            color: var(--text);
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            overflow: hidden;
            background-image: 
                radial-gradient(at 0% 0%, hsla(253,16%,7%,1) 0, transparent 50%), 
                radial-gradient(at 50% 0%, hsla(225,39%,30%,1) 0, transparent 50%), 
                radial-gradient(at 100% 0%, hsla(339,49%,30%,1) 0, transparent 50%);
        }
        .login-card {
            background: rgba(42, 43, 54, 0.7);
            backdrop-filter: blur(20px);
            padding: 2.5rem;
            border-radius: 20px;
            width: 100%;
            max-width: 400px;
            box-shadow: 0 20px 50px rgba(0,0,0,0.3);
            border: 1px solid rgba(255,255,255,0.1);
            animation: float 6s ease-in-out infinite;
        }
        h1 { margin-bottom: 2rem; text-align: center; color: var(--primary); font-weight: 700; letter-spacing: 2px; }
        input {
            width: 100%;
            padding: 15px;
            margin-bottom: 1rem;
            border-radius: 12px;
            border: 1px solid rgba(255,255,255,0.1);
            background: rgba(0,0,0,0.2);
            color: white;
            font-family: 'Cairo', sans-serif;
            box-sizing: border-box;
            transition: all 0.3s;
        }
        input:focus { outline: none; border-color: var(--primary); box-shadow: 0 0 15px rgba(108, 93, 211, 0.3); }
        button {
            width: 100%;
            padding: 15px;
            border-radius: 12px;
            border: none;
            background: var(--primary);
            color: white;
            font-weight: bold;
            cursor: pointer;
            font-size: 1.1rem;
            font-family: 'Cairo', sans-serif;
            transition: transform 0.2s;
        }
        button:hover { transform: scale(1.02); box-shadow: 0 10px 20px rgba(108,93,211,0.4); }
        .error {
            background: rgba(231, 76, 60, 0.2);
            color: #ff6b6b;
            padding: 10px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
        }
        @keyframes float {
            0% { transform: translateY(0px); }
            50% { transform: translateY(-10px); }
            100% { transform: translateY(0px); }
        }
        .demo-info {
            margin-top: 20px;
            text-align: center;
            font-size: 0.8rem;
            opacity: 0.5;
        }
    </style>
</head>
<body>
    <div class="login-card">
        <h1>NEXA ID</h1>
        {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
        <form method="POST">
            <input type="text" name="user" placeholder="اسم المستخدم" required autocomplete="off">
            <input type="password" name="pass" placeholder="كلمة المرور" required>
            <button type="submit">تسجيل الدخول</button>
        </form>
        <div class="demo-info">
            Default: admin / admin123
        </div>
    </div>
</body>
</html>`

const LayoutHTML = `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexa Dashboard</title>
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@300;400;700&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary: #6C5DD3;
            --secondary: #A0D7E7;
            --bg: #1f2029;
            --card: #2a2b36;
            --text: #ffffff;
            --text-mute: #808191;
            --success: #66e2d5;
        }
        * { box-sizing: border-box; }
        body {
            font-family: 'Cairo', sans-serif;
            background-color: var(--bg);
            color: var(--text);
            margin: 0;
            display: flex;
            height: 100vh;
            overflow: hidden;
        }
        
        /* Sidebar */
        .sidebar {
            width: 280px;
            background: var(--card);
            border-left: 1px solid rgba(255,255,255,0.05);
            display: flex;
            flex-direction: column;
            padding: 20px;
        }
        .logo { 
            font-size: 1.8rem; font-weight: 700; color: var(--text); margin-bottom: 40px; 
            display: flex; align-items: center; gap: 10px;
        }
        .logo i { color: var(--primary); }
        .menu-item {
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 12px;
            cursor: pointer;
            color: var(--text-mute);
            transition: all 0.3s;
            display: flex;
            align-items: center;
            gap: 15px;
        }
        .menu-item:hover, .menu-item.active {
            background: var(--primary);
            color: white;
        }
        .menu-item i { font-size: 1.2rem; }
        .user-profile {
            margin-top: auto;
            border-top: 1px solid rgba(255,255,255,0.1);
            padding-top: 20px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .avatar {
            width: 40px; height: 40px; background: linear-gradient(45deg, var(--primary), #a29bfe);
            border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: bold;
        }

        /* Main Content */
        .main {
            flex: 1;
            padding: 40px;
            overflow-y: auto;
            background-image: radial-gradient(at 10% 10%, rgba(108, 93, 211, 0.1) 0, transparent 50%);
        }
        
        /* Tabs */
        .tab-content { display: none; animation: fadeIn 0.4s; }
        .tab-content.active { display: block; }

        /* Dashboard Cards */
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .card {
            background: var(--card);
            padding: 25px;
            border-radius: 20px;
            border: 1px solid rgba(255,255,255,0.05);
        }
        .card h3 { color: var(--text-mute); margin: 0 0 10px 0; font-size: 0.9rem; }
        .card .value { font-size: 2rem; font-weight: 700; }

        /* Terminal */
        .terminal {
            background: #111;
            border-radius: 12px;
            padding: 20px;
            font-family: 'Courier New', monospace;
            min-height: 400px;
            border: 1px solid #333;
            display: flex;
            flex-direction: column;
        }
        .output { flex: 1; overflow-y: auto; margin-bottom: 20px; color: #a29bfe; font-size: 0.9rem; white-space: pre-wrap; }
        .input-line { display: flex; gap: 10px; align-items: center; }
        .input-line input {
            background: transparent; border: none; color: white; width: 100%; font-family: inherit; font-size: 1rem;
        }
        .input-line input:focus { outline: none; }

        /* Table */
        table { width: 100%; border-collapse: collapse; background: var(--card); border-radius: 15px; overflow: hidden; }
        th, td { padding: 15px; text-align: right; border-bottom: 1px solid rgba(255,255,255,0.05); }
        th { background: rgba(0,0,0,0.2); color: var(--text-mute); }
        tr:hover { background: rgba(255,255,255,0.02); }

        .btn { padding: 8px 15px; border-radius: 8px; border: none; cursor: pointer; background: var(--primary); color: white; }
        .btn-danger { background: #e74c3c; }

        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="logo"><i class="fas fa-network-wired"></i> NEXA PROTOCOL</div>
        <div class="menu-item active" onclick="showTab('dashboard', this)"><i class="fas fa-home"></i> نظرة عامة</div>
        <div class="menu-item" onclick="showTab('terminal', this)"><i class="fas fa-terminal"></i> الطرفية</div>
        <div class="menu-item" onclick="showTab('network', this)"><i class="fas fa-globe"></i> الشبكة</div>
        {{if eq .Role "admin"}}
        <div class="menu-item" onclick="showTab('users', this)"><i class="fas fa-users"></i> المستخدمين</div>
        {{end}}
        
        <div class="user-profile">
            <div class="avatar">{{printf "%.1s" .Username}}</div>
            <div>
                <div style="font-weight:bold">{{.Username}}</div>
                <div style="font-size:0.8rem; color:var(--text-mute)">{{.Role}}</div>
            </div>
            <a href="/logout" style="margin-right:auto; color:#ff6b6b;"><i class="fas fa-sign-out-alt"></i></a>
        </div>
    </div>

    <div class="main">
        <!-- Dashboard -->
        <div id="dashboard" class="tab-content active">
            <h1 style="margin-bottom:30px">لوحة القيادة</h1>
            <div class="grid">
                <div class="card">
                    <h3>حالة النظام</h3>
                    <div class="value" style="color:var(--success)">{{.SystemStatus}}</div>
                </div>
                <div class="card">
                    <h3>المستخدمين النشطين</h3>
                    <div class="value">{{.UserCount}}</div>
                </div>
                <div class="card">
                    <h3>إجمالي العمليات</h3>
                    <div class="value">{{.LogCount}}</div>
                </div>
            </div>

            <!-- Ledger Preview (Mock) -->
            <div class="card">
                <h3>آخر الكتل (Blockchain Ledger)</h3>
                <table>
                    <tr><th>Hash</th><th>Timestamp</th><th>Validator</th></tr>
                    <tr><td>0x8F3...A21</td><td>Just now</td><td>Node-01</td></tr>
                    <tr><td>0x1B4...99C</td><td>2 mins ago</td><td>Node-02</td></tr>
                    <tr><td>0xE55...F12</td><td>5 mins ago</td><td>Node-01</td></tr>
                </table>
            </div>
        </div>

        <!-- Terminal -->
        <div id="terminal" class="tab-content">
            <h1>محطة الأوامر</h1>
            <div class="terminal">
                <div class="output" id="termOutput">
                    Welcome to Nexa Interactive Terminal v2.0
                    Connected to {{.Username}}@localhost...
                </div>
                <div class="input-line">
                    <span style="color:var(--success)">➜</span>
                    <input type="text" id="termInput" placeholder="أدخل الأمر هنا (مثلاً: PING, LIST, HELP)..." autocomplete="off">
                </div>
            </div>
            <div style="margin-top:10px; display:flex; gap:10px; align-items:center;">
                <label style="cursor:pointer; display:flex; gap:5px; align-items:center;">
                    <input type="checkbox" id="useDNS"> استخدام DNS (.nexa)
                </label>
            </div>
        </div>

        <!-- Network Map (Mock) -->
        <div id="network" class="tab-content">
            <h1>خريطة الشبكة</h1>
            <div style="height:400px; background:rgba(0,0,0,0.2); border-radius:20px; display:flex; align-items:center; justify-content:center; color:var(--text-mute);">
                <div style="text-align:center">
                    <i class="fas fa-project-diagram" style="font-size:4rem; margin-bottom:20px; color:var(--primary);"></i>
                    <p>جاري البحث عن العقد المتصلة...</p>
                    <p style="font-size:0.8rem">تم اكتشاف عقدة واحدة (Localhost)</p>
                </div>
            </div>
        </div>

        <!-- Users (Admin only) -->
        <div id="users" class="tab-content">
            <h1>إدارة المستخدمين</h1>
            <div class="card">
                 <form action="/admin/users" method="POST" style="display:flex; gap:10px; margin-bottom:20px;">
                    <input type="hidden" name="action" value="add">
                    <input name="username" placeholder="اسم المستخدم" style="padding:10px; border-radius:8px; border:1px solid #444; background:#222; color:white;">
                    <input name="password" placeholder="كلمة المرور" type="password" style="padding:10px; border-radius:8px; border:1px solid #444; background:#222; color:white;">
                    <select name="role" style="padding:10px; border-radius:8px; border:1px solid #444; background:#222; color:white;">
                        <option value="user">User</option>
                        <option value="admin">Admin</option>
                    </select>
                    <button class="btn">إضافة</button>
                </form>
                <table>
                    <tr><th>User</th><th>Role</th><th>Action</th></tr>
                    <!-- Data populated from server loop would go here, but for this demo using static placeholders or would need range -->
                    <tr><td>admin</td><td><span style="padding:3px 8px; background:rgba(108,93,211,0.2); color:var(--primary); border-radius:4px">Admin</span></td><td>-</td></tr>
                    <tr><td>user1</td><td>User</td><td>-</td></tr>
                </table>
            </div>
        </div>
    </div>

    <script>
        function showTab(id, el) {
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            document.getElementById(id).classList.add('active');
            document.querySelectorAll('.menu-item').forEach(m => m.classList.remove('active'));
            el.classList.add('active');
        }

        const termInput = document.getElementById('termInput');
        termInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                const cmd = this.value;
                if(!cmd) return;
                
                const output = document.getElementById('termOutput');
                output.innerHTML += '\n<span style="color:var(--text-mute)">➜ ' + cmd + '</span>';
                this.value = '';
                
                const useDNS = document.getElementById('useDNS').checked;

                fetch('/api/command', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: 'cmd=' + encodeURIComponent(cmd) + '&dns=' + useDNS
                })
                .then(r => r.json())
                .then(data => {
                    const color = data.status === 'ERROR' ? '#ff6b6b' : '#66e2d5';
                    output.innerHTML += '\n<span style="color:'+color+'">' + data.result + '</span>';
                    output.scrollTop = output.scrollHeight;
                });
            }
        });
    </script>
</body>
</html>
`
