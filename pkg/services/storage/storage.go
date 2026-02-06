package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/analytics"
	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
)

const (
	StorageRoot = "./storage"
)

// ShareLinkStore
var (
	shareLinks = make(map[string]string)
	shareMutex sync.RWMutex

	// Telemetry
	uploadBytes   int64
	downloadBytes int64
	netManager    *network.NetworkManager
	govManager    *governance.GovernanceManager
	metricsMutex  sync.Mutex
)

func reportMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		metricsMutex.Lock()
		upSpeed := uploadBytes
		downSpeed := downloadBytes
		uploadBytes = 0
		downloadBytes = 0
		metricsMutex.Unlock()

		if netManager != nil {
			netManager.UpdateDeviceMetrics("svc-storage", network.DeviceMetrics{
				RequestsPerSec: 0, // Not tracking req/s yet for storage specifically
				LastActivity:   time.Now().Unix(),
				Custom: map[string]interface{}{
					"upload_speed_bps":   upSpeed,
					"download_speed_bps": downSpeed,
				},
			})
			netManager.UpdateServiceMetrics("storage", map[string]interface{}{
				"upload_speed":   utils.FormatSize(upSpeed) + "/s",
				"download_speed": utils.FormatSize(downSpeed) + "/s",
			})
		}
	}
}

// HTML Template (New Professional UI - No backticks in JS for Go compatibility)
const FileMangerHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA | Ultimate Cloud</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
    <style>
        :root {
            --primary: #6366f1;
            --secondary: #ec4899;
            --accent: #06b6d4;
            --success: #10b981;
            --bg: #020617;
            --sidebar-bg: rgba(15, 23, 42, 0.8);
            --content-bg: rgba(30, 41, 59, 0.4);
            --glass: rgba(255, 255, 255, 0.05);
            --border: rgba(255, 255, 255, 0.1);
            --text: #f8fafc;
            --text-muted: #94a3b8;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; outline: none; }
        
        body { 
            font-family: 'Outfit', 'Cairo', sans-serif; 
            background: var(--bg);
            background-image: 
                radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.15) 0, transparent 50%),
                radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.15) 0, transparent 50%);
            color: var(--text);
            height: 100vh;
            overflow: hidden;
            display: flex;
        }
        
        /* Sidebar */
        .sidebar {
            width: 280px;
            background: var(--sidebar-bg);
            backdrop-filter: blur(20px);
            border-left: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 30px 20px;
            z-index: 10;
        }

        .logo {
            font-size: 2rem;
            font-weight: 900;
            margin-bottom: 40px;
            background: linear-gradient(to right, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .nav-item {
            padding: 14px 20px;
            margin-bottom: 8px;
            border-radius: 12px;
            cursor: pointer;
            transition: all 0.3s;
            color: var(--text-muted);
            display: flex;
            align-items: center;
            gap: 12px;
            font-weight: 600;
        }

        .nav-item:hover, .nav-item.active {
            background: rgba(99, 102, 241, 0.1);
            color: var(--text);
            border-right: 3px solid var(--primary);
        }

        .nav-item i { width: 20px; }

        /* Main Content */
        .main-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            background: var(--content-bg);
            margin: 20px;
            margin-right: 0; /* RTL Fix */
            border-radius: 30px 0 0 30px;
            border: 1px solid var(--border);
            border-right: none;
            overflow: hidden;
            position: relative;
        }

        /* Top Bar */
        .top-bar {
            padding: 24px 40px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-bottom: 1px solid var(--border);
            background: rgba(15, 23, 42, 0.6);
        }

        .search-box {
            position: relative;
            width: 400px;
        }

        .search-box input {
            width: 100%;
            background: var(--glass);
            border: 1px solid var(--border);
            padding: 12px 50px 12px 20px;
            border-radius: 12px;
            color: white;
            font-family: inherit;
        }

        .search-box i {
            position: absolute;
            right: 20px;
            top: 50%;
            transform: translateY(-50%);
            color: var(--text-muted);
        }

        .actions {
            display: flex;
            gap: 15px;
        }

        .btn {
            padding: 10px 24px;
            border-radius: 12px;
            border: none;
            font-weight: 700;
            cursor: pointer;
            transition: all 0.3s;
            display: flex;
            align-items: center;
            gap: 8px;
            font-family: inherit;
        }

        .btn-primary {
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            color: white;
            box-shadow: 0 4px 15px rgba(99, 102, 241, 0.3);
        }
        
        .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 8px 25px rgba(99, 102, 241, 0.4); }

        .btn-glass {
            background: var(--glass);
            color: var(--text);
            border: 1px solid var(--border);
        }

        .btn-glass:hover { background: rgba(255,255,255,0.1); }

        /* File Area */
        .file-area {
            flex: 1;
            padding: 30px 40px;
            overflow-y: auto;
        }

        .breadcrumbs {
            margin-bottom: 20px;
            color: var(--text-muted);
            font-size: 0.9rem;
            display: flex;
            gap: 8px;
        }

        .breadcrumbs span { cursor: pointer; color: var(--accent); }
        .breadcrumbs span:hover { text-decoration: underline; }

        .grid-view {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
            gap: 20px;
        }

        .file-card {
            background: var(--glass);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 20px;
            text-align: center;
            transition: all 0.3s;
            cursor: pointer;
            position: relative;
        }

        .file-card:hover {
            background: rgba(255,255,255,0.08);
            transform: translateY(-5px);
            border-color: var(--primary);
        }

        .file-icon {
            font-size: 3.5rem;
            margin-bottom: 15px;
            display: block;
        }

        .file-name {
            font-size: 0.95rem;
            font-weight: 600;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 5px;
        }

        .file-meta {
            font-size: 0.75rem;
            color: var(--text-muted);
        }

        .context-menu {
            position: absolute;
            top: 10px;
            left: 10px;
            opacity: 0;
            transition: 0.2s;
        }

        .file-card:hover .context-menu { opacity: 1; }

        .modal {
            display: none;
            position: fixed;
            top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.7);
            backdrop-filter: blur(5px);
            z-index: 1000;
            justify-content: center;
            align-items: center;
        }

        .modal-content {
            background: #0f172a;
            padding: 40px;
            border-radius: 24px;
            border: 1px solid var(--border);
            width: 450px;
            text-align: center;
            box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5);
            animation: popIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
        }

        @keyframes popIn {
            from { transform: scale(0.9); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }

        #qr-code { padding: 20px; background: white; border-radius: 12px; margin: 20px auto; width: fit-content; }
        
        .drag-overlay {
            position: absolute;
            top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(99, 102, 241, 0.9);
            z-index: 50;
            display: flex;
            justify-content: center;
            align-items: center;
            flex-direction: column;
            opacity: 0;
            pointer-events: none;
            transition: 0.3s;
        }
        
        .drag-active .drag-overlay { opacity: 1; pointer-events: all; }

        @media (max-width: 768px) {
            body { flex-direction: column; }
            .sidebar { width: 100%; flex-direction: row; overflow-x: auto; padding: 10px; border-left: none; border-bottom: 1px solid var(--border); }
            .nav-item { margin: 0 5px; flex-shrink: 0; }
            .main-content { margin: 0; border-radius: 0; border: none; }
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="logo"><i class="fas fa-cube"></i> NEXA CLOUD</div>
        <div class="nav-item active" onclick="loadFiles('')"><i class="fas fa-home"></i> الرئيسيه</div>
        <div class="nav-item" onclick="loadFiles('incoming')"><i class="fas fa-inbox"></i> الوارد (Incoming)</div>
        <div class="nav-item" onclick="loadFiles('shared')"><i class="fas fa-share-alt"></i> مشترك (Shared)</div>
        <div class="nav-item" onclick="loadFiles('vault')"><i class="fas fa-lock"></i> الخزنة (Vault)</div>
        <div class="nav-item" onclick="loadFiles('backup')"><i class="fas fa-sync"></i> النسخ الاحتياطي</div>
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid var(--border); display: none;" id="cat-filters">
            <div style="color: var(--text-muted); font-size: 0.8rem; margin-bottom: 10px; font-weight:700;">التصنيفات</div>
            <div class="nav-item" onclick="filterType('image')"><i class="fas fa-image"></i> صور</div>
            <div class="nav-item" onclick="filterType('video')"><i class="fas fa-video"></i> فيديو</div>
            <div class="nav-item" onclick="filterType('doc')"><i class="fas fa-file-alt"></i> مستندات</div>
        </div>
    </div>

    <div class="main-content" id="dropZone">
        <div class="drag-overlay">
            <i class="fas fa-cloud-upload-alt" style="font-size: 5rem; color: white; margin-bottom: 20px;"></i>
            <h2 style="color: white;">أفلت الملفات للرفع الفوري</h2>
        </div>
        <div class="top-bar">
            <div class="search-box">
                <input type="text" id="searchInput" placeholder="بحث في الملفات..." onkeyup="searchFiles()">
                <i class="fas fa-search"></i>
            </div>
            <div class="actions">
                 <button class="btn btn-glass" onclick="createFolder()"><i class="fas fa-folder-plus"></i></button>
                 <button class="btn btn-primary" onclick="document.getElementById('fileInput').click()">
                    <i class="fas fa-cloud-upload"></i> رفع ملف
                 </button>
                 <input type="file" id="fileInput" hidden multiple onchange="handleFileSelect(this.files)">
            </div>
        </div>
        <div class="file-area">
            <div class="breadcrumbs" id="breadcrumbs"><span onclick="loadFiles('')">الرئيسية</span></div>
            <div class="grid-view" id="fileList"></div>
        </div>
    </div>

    <div id="shareModal" class="modal">
        <div class="modal-content">
            <h2 style="margin-bottom: 20px;">مشاركة الملف</h2>
            <div id="qr-code"></div>
            <p style="color: var(--text-muted); margin: 15px 0;">انسخ الرابط أو امسح الكود</p>
            <input type="text" id="shareLink" readonly style="width: 100%; padding: 10px; border-radius: 8px; border: 1px solid var(--border); background: var(--glass); color: white; text-align: center;">
            <div style="margin-top: 20px; display: flex; gap: 10px; justify-content: center;">
                <button class="btn btn-primary" onclick="copyLink()">نسخ الرابط</button>
                <button class="btn btn-glass" onclick="closeModal()">إغلاق</button>
            </div>
        </div>
    </div>

    <script>
        let currentPath = ''; let allFiles = [];

        document.addEventListener('DOMContentLoaded', () => {
            loadFiles(''); setupDragDrop();
            if(window.innerWidth > 768) document.getElementById('cat-filters').style.display = 'block';
        });

        function loadFiles(path) {
            currentPath = path; updateBreadcrumbs(path);
            if (path.includes('vault')) {
                const pin = prompt("🔐 الخزنة مشفرة. رمز الحماية:");
                if (pin !== '1234') { alert("رمز خاطئ!"); loadFiles(''); return; }
            }
            // FIXED: Relative paths to support proxying
            fetch('api/list?dir=' + encodeURIComponent(path))
                .then(res => res.json())
                .then(data => { allFiles = data || []; renderFiles(allFiles); })
                .catch(err => { console.error(err); document.getElementById('fileList').innerHTML = '<div style="color:#ef4444; padding:40px; text-align:center;">خطأ في الاتصال بالسيرفر</div>'; });
        }

        function renderFiles(files) {
            const container = document.getElementById('fileList'); container.innerHTML = '';
            if (!files || files.length === 0) {
                container.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 50px; color: var(--text-muted);">المجلد فارغ</div>';
                return;
            }
            files.forEach(file => {
                const div = document.createElement('div'); div.className = 'file-card';
                div.onclick = () => { if (file.IsDir) loadFiles(currentPath ? currentPath + '/' + file.Name : file.Name); };
                const icon = file.IsDir ? '📁' : getIcon(file.Name);
                let actions = '';
                if (!file.IsDir) {
                    actions = '<button class="btn btn-sm btn-glass" onclick="openShare(\'' + file.Name + '\')" style="padding: 5px 10px;"><i class="fas fa-share-alt"></i></button>';
                    actions += '<a href="download?file=' + encodeURIComponent(currentPath ? currentPath + '/' + file.Name : file.Name) + '" class="btn btn-sm btn-glass" style="padding: 5px 10px; text-decoration:none;"><i class="fas fa-download"></i></a>';
                }
                actions += '<button class="btn btn-sm btn-glass" onclick="deleteFile(\'' + file.Name + '\')" style="padding: 5px 10px; color: #ef4444;"><i class="fas fa-trash"></i></button>';

                div.innerHTML = '<div class="file-icon">' + icon + '</div>' +
                                '<div class="file-name" title="' + file.Name + '">' + file.Name + '</div>' +
                                '<div class="file-meta">' + file.Size + '</div>' +
                                '<div class="context-menu" onclick="event.stopPropagation()">' + actions + '</div>';
                container.appendChild(div);
            });
        }

        function getIcon(name) {
            const ext = name.split('.').pop().toLowerCase();
            const map = { 'pdf':'📕', 'doc':'📄', 'txt':'📝', 'jpg':'🖼️', 'png':'🖼️', 'mp4':'🎬', 'mp3':'🎵', 'zip':'📦', 'exe':'⚙️' };
            return map[ext] || '📄';
        }

        function filterType(type) {
            let filtered = [];
            if (type === 'image') filtered = allFiles.filter(f => /\.(jpg|jpeg|png|gif)$/i.test(f.Name));
            if (type === 'video') filtered = allFiles.filter(f => /\.(mp4|mov|avi)$/i.test(f.Name));
            if (type === 'doc') filtered = allFiles.filter(f => /\.(pdf|doc|docx|txt)$/i.test(f.Name));
            renderFiles(filtered);
        }

        function searchFiles() {
            const q = document.getElementById('searchInput').value.toLowerCase();
            renderFiles(allFiles.filter(f => f.Name.toLowerCase().includes(q)));
        }

        function updateBreadcrumbs(path) {
            const bc = document.getElementById('breadcrumbs');
            if (!path) { bc.innerHTML = '<span onclick="loadFiles(\'\')">الرئيسية</span>'; return; }
            const parts = path.split('/');
            let html = '<span onclick="loadFiles(\'\')">الرئيسية</span>', acc = '';
            parts.forEach(p => { acc += (acc ? '/' : '') + p; html += ' / <span onclick="loadFiles(\'' + acc + '\')">' + p + '</span>'; });
            bc.innerHTML = html;
        }

        function openShare(filename) {
            const filePath = currentPath ? currentPath + '/' + filename : filename;
            fetch('api/share?file=' + encodeURIComponent(filePath))
                .then(res => res.json())
                .then(data => {
                    document.getElementById('shareLink').value = data.link;
                    document.getElementById('qr-code').innerHTML = '';
                    new QRCode(document.getElementById('qr-code'), { text: data.link, width: 128, height: 128 });
                    document.getElementById('shareModal').style.display = 'flex';
                });
        }
        function closeModal() { document.getElementById('shareModal').style.display = 'none'; }
        function copyLink() { document.getElementById("shareLink").select(); document.execCommand("copy"); alert("تم النسخ"); }
        function deleteFile(filename) {
            if(!confirm('حذف النهائي؟')) return;
            const filePath = currentPath ? currentPath + '/' + filename : filename;
            fetch('delete?file=' + encodeURIComponent(filePath), { method: 'POST' }).then(() => loadFiles(currentPath));
        }
        function createFolder() {
            const name = prompt("اسم المجلد:");
            if (name) {
                const path = currentPath ? currentPath + '/' + name : name;
                fetch('api/mkdir?dir=' + encodeURIComponent(path)).then(() => loadFiles(currentPath));
            }
        }
        function setupDragDrop() {
            const dz = document.getElementById('dropZone');
            document.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('drag-active'); });
            document.addEventListener('dragleave', e => { if (!e.relatedTarget || !e.relatedTarget.closest('.drag-overlay')) dz.classList.remove('drag-active'); });
            document.addEventListener('drop', e => { e.preventDefault(); dz.classList.remove('drag-active'); if(e.dataTransfer.files.length) handleFileSelect(e.dataTransfer.files); });
        }
        function handleFileSelect(files) {
            const fd = new FormData();
            for (let i=0; i<files.length; i++) fd.append('file', files[i]);
            if (currentPath) fd.append('dir', currentPath);
            fetch('upload', { method: 'POST', body: fd }).then(res => { if(res.ok) loadFiles(currentPath); else alert('خطأ في الرفع'); });
        }
    </script>
</body>
</html>
`

type FileInfo struct {
	Name   string
	Size   string
	Time   string
	IsDir  bool
	IsLink bool
}

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	netManager = nm
	govManager = gm
	go reportMetrics()

	// Ensure professional storage structure
	os.MkdirAll(StorageRoot, 0755)
	os.MkdirAll(filepath.Join(StorageRoot, "public"), 0755)
	os.MkdirAll(filepath.Join(StorageRoot, "locked"), 0755)
	subDirs := []string{"incoming", "shared", "vault", "backup", "temp"}
	for _, sub := range subDirs {
		path := filepath.Join(StorageRoot, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
	}

	// Start Auto-Backup Routine (Every 5 minutes)
	go startAutoBackup()

	mux := http.NewServeMux()
	mux.HandleFunc("/", enableCORS(webHandler))
	mux.HandleFunc("/upload", enableCORS(uploadHandler))
	mux.HandleFunc("/delete", enableCORS(deleteHandler))
	mux.HandleFunc("/download", enableCORS(downloadHandler))
	mux.HandleFunc("/api/list", enableCORS(listAPIHandler))
	mux.HandleFunc("/api/stats", enableCORS(statsHandler))
	mux.HandleFunc("/api/share", enableCORS(shareAPIHandler))
	mux.HandleFunc("/api/mkdir", enableCORS(mkdirAPIHandler))
	mux.HandleFunc("/s/", enableCORS(handleSharedLink))

	cfg := config.Get()
	portStr := fmt.Sprintf("%d", cfg.Services.Storage.Port)
	localIP := utils.GetLocalIP()
	utils.LogInfo("Storage", fmt.Sprintf("Root: %s", StorageRoot))
	utils.LogInfo("Storage", fmt.Sprintf("UI:   http://%s:%s", localIP, portStr))
	utils.SaveEndpoint("storage", fmt.Sprintf("http://%s:%s", localIP, portStr))

	server := &http.Server{
		Addr:    "0.0.0.0:" + portStr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogError("Storage", "Failed to start server", err)
	}
}

func startAutoBackup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		source := filepath.Join(StorageRoot, "incoming")
		dest := filepath.Join(StorageRoot, "backup")
		files, _ := os.ReadDir(source)
		for _, f := range files {
			if !f.IsDir() {
				srcFile, err := os.Open(filepath.Join(source, f.Name()))
				if err != nil {
					continue
				}
				dstFile, err := os.Create(filepath.Join(dest, f.Name()))
				if err != nil {
					srcFile.Close()
					continue
				}
				io.Copy(dstFile, srcFile)
				srcFile.Close()
				dstFile.Close()
			}
		}
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		next(w, r)
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.New("fm").Parse(FileMangerHTML))
	tmpl.Execute(w, nil)
}

func listAPIHandler(w http.ResponseWriter, r *http.Request) {
	subDir := r.URL.Query().Get("dir")
	if strings.Contains(subDir, "..") {
		subDir = ""
	}
	readPath := filepath.Join(StorageRoot, subDir)
	files, err := os.ReadDir(readPath)
	if err != nil {
		if subDir == "" {
			os.MkdirAll(StorageRoot, 0755)
			files, _ = os.ReadDir(readPath)
		} else {
			http.Error(w, "Directory not found", 404)
			return
		}
	}
	var fileList []FileInfo
	for _, f := range files {
		info, _ := f.Info()
		fileList = append(fileList, FileInfo{
			Name:  f.Name(),
			Size:  utils.FormatSize(info.Size()),
			Time:  info.ModTime().Format("02/01 15:04"),
			IsDir: f.IsDir(),
		})
	}
	sort.Slice(fileList, func(i, j int) bool {
		if fileList[i].IsDir != fileList[j].IsDir {
			return fileList[i].IsDir
		}
		return fileList[i].Time > fileList[j].Time
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileList)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	files, _ := os.ReadDir(StorageRoot)
	totalSize := int64(0)
	fileCount := 0
	for _, f := range files {
		if !f.IsDir() {
			info, _ := f.Info()
			totalSize += info.Size()
			fileCount++
		}
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"files":%d,"totalSize":%d,"totalSizeFormatted":"%s"}`, fileCount, totalSize, utils.FormatSize(totalSize))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(500 << 20) // 500MB
	files := r.MultipartForm.File["file"]
	targetDir := r.FormValue("dir")
	if strings.Contains(targetDir, "..") {
		targetDir = ""
	}
	saveDir := filepath.Join(StorageRoot, targetDir)
	os.MkdirAll(saveDir, 0755)

	for _, header := range files {
		// Governance Check: File Size
		if govManager != nil {
			policy := govManager.PolicyEngine.GetPolicy()
			if header.Size > int64(policy.MaxUploadSizeMB)*1024*1024 {
				govManager.ReportEvent("Security", governance.LevelAction,
					fmt.Sprintf("Blocked large upload: %s", header.Filename),
					fmt.Sprintf("Size %d bytes exceeds policy %d MB", header.Size, policy.MaxUploadSizeMB),
					"Upload Rejected")
				http.Error(w, "File exceeds system policy", http.StatusForbidden)
				return
			}
		}

		file, _ := header.Open()

		// Safe filename with check
		filename := filepath.Base(header.Filename)
		targetPath := filepath.Join(saveDir, filename)
		if _, err := os.Stat(targetPath); err == nil {
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)
			filename = fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext)
			targetPath = filepath.Join(saveDir, filename)
		}

		dst, _ := os.Create(targetPath)
		written, _ := io.Copy(dst, file)

		utils.LogSuccess("Storage", fmt.Sprintf("Uploaded: %s (%s)", filename, utils.FormatSize(written)))

		// Track in analytics
		sessionID := "unknown"
		if cookie, err := r.Cookie("session_id"); err == nil {
			sessionID = cookie.Value
		}
		analytics.GetManager().TrackFile(sessionID, analytics.FileActivity{
			Action:   "upload",
			FileName: filename,
			Path:     targetPath,
			FileSize: written,
			Status:   "success",
		})

		metricsMutex.Lock()
		uploadBytes += written
		metricsMutex.Unlock()
		dst.Close()
		file.Close()
	}
	w.WriteHeader(200)
}

func mkdirAPIHandler(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if strings.Contains(dir, "..") || dir == "" {
		return
	}
	os.MkdirAll(filepath.Join(StorageRoot, dir), 0755)
	w.WriteHeader(200)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if strings.Contains(file, "..") {
		return
	}
	os.RemoveAll(filepath.Join(StorageRoot, file))
	utils.LogInfo("Storage", "Deleted: "+file)

	// Track in analytics
	sessionID := "unknown"
	if cookie, err := r.Cookie("session_id"); err == nil {
		sessionID = cookie.Value
	}
	analytics.GetManager().TrackFile(sessionID, analytics.FileActivity{
		Action:   "delete",
		FileName: filepath.Base(file),
		Path:     file,
		Status:   "success",
	})

	w.WriteHeader(200)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if strings.Contains(file, "..") {
		return
	}
	path := filepath.Join(StorageRoot, file)
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(file))

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	defer f.Close()

	// Track download size
	n, _ := io.Copy(w, f)

	// Track in analytics
	sessionID := "unknown"
	if cookie, err := r.Cookie("session_id"); err == nil {
		sessionID = cookie.Value
	}
	analytics.GetManager().TrackFile(sessionID, analytics.FileActivity{
		Action:   "download",
		FileName: filepath.Base(file),
		Path:     file,
		FileSize: n,
		Status:   "success",
	})

	metricsMutex.Lock()
	downloadBytes += n
	metricsMutex.Unlock()
}

func shareAPIHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		return
	}
	hasher := md5.New()
	hasher.Write([]byte(file + time.Now().String()))
	token := hex.EncodeToString(hasher.Sum(nil))[:8]
	shareMutex.Lock()
	shareLinks[token] = file
	shareMutex.Unlock()
	localIP := utils.GetLocalIP()
	cfg := config.Get()
	link := fmt.Sprintf("http://%s:%d/s/%s", localIP, cfg.Services.Storage.Port, token)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"link":"%s"}`, link)
}

func handleSharedLink(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/s/")
	shareMutex.RLock()
	file, exists := shareLinks[token]
	shareMutex.RUnlock()
	if !exists {
		http.Error(w, "Link expired", 404)
		return
	}
	path := filepath.Join(StorageRoot, file)
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(file))
	http.ServeFile(w, r, path)
}
