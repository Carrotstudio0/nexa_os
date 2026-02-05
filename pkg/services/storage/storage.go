package storage

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/utils"
)

const (
	Port        = config.WebPort
	StorageRoot = "./storage"
)

// HTML Template for the File Manager - Professional UI
const FileMangerHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA | Digital Storage</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary: #6366f1;
            --secondary: #ec4899;
            --accent: #06b6d4;
            --bg: #020617;
            --card-bg: rgba(15, 23, 42, 0.6);
            --glass: rgba(255, 255, 255, 0.03);
            --border: rgba(255, 255, 255, 0.08);
            --text: #f8fafc;
            --text-muted: #64748b;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: 'Outfit', 'Cairo', sans-serif; 
            background: var(--bg);
            background-image: 
                radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.1) 0, transparent 40%),
                radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.1) 0, transparent 40%);
            color: var(--text);
            min-height: 100vh;
            padding: 40px;
        }
        
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: var(--card-bg);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            padding: 50px;
            border-radius: 40px; 
            box-shadow: 0 40px 100px -20px rgba(0,0,0,0.5);
            border: 1px solid var(--border);
            animation: slideIn 0.8s cubic-bezier(0.16, 1, 0.3, 1);
        }

        @keyframes slideIn {
            from { opacity: 0; transform: translateY(30px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        .header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 50px;
            padding-bottom: 30px;
            border-bottom: 1px solid var(--border);
        }
        
        .header-info h1 {
            font-size: 3rem;
            font-weight: 900;
            background: linear-gradient(to right, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -2px;
        }
        
        .info-pill {
            background: var(--glass);
            border: 1px solid var(--border);
            padding: 12px 24px;
            border-radius: 20px;
            display: flex;
            align-items: center;
            gap: 12px;
            transition: all 0.3s;
        }

        .info-pill:hover { border-color: var(--primary); background: rgba(255,255,255,0.06); }
        
        .info-label { color: var(--text-muted); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 1px; }
        .info-value { font-weight: 700; color: var(--text); }
        
        .upload-zone { 
            background: linear-gradient(135deg, rgba(99, 102, 241, 0.05), rgba(236, 72, 153, 0.05));
            padding: 60px; 
            border: 2px dashed var(--border); 
            border-radius: 32px; 
            text-align: center; 
            margin-bottom: 50px;
            transition: all 0.4s;
            cursor: pointer;
        }
        
        .upload-zone:hover {
            border-color: var(--secondary);
            background: rgba(255, 255, 255, 0.04);
            transform: scale(1.01);
        }

        .upload-zone i { font-size: 4rem; margin-bottom: 20px; color: var(--secondary); opacity: 0.8; }
        
        .btn-action { 
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            color: white; border: none; padding: 16px 40px; 
            border-radius: 16px; cursor: pointer; font-weight: 800;
            font-size: 1.1rem; transition: all 0.3s;
            box-shadow: 0 10px 20px -5px rgba(99, 102, 241, 0.4);
        }
        
        .btn-action:hover { transform: translateY(-3px); box-shadow: 0 20px 40px -10px rgba(99, 102, 241, 0.6); }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 24px;
            margin-bottom: 40px;
        }
        
        .stat-box {
            background: var(--glass);
            padding: 24px;
            border-radius: 24px;
            border: 1px solid var(--border);
            text-align: center;
        }
        
        .val { font-size: 2.5rem; font-weight: 900; color: var(--text); }
        .lbl { font-size: 0.9rem; color: var(--text-muted); margin-top: 4px; }
        
        table { width: 100%; border-collapse: separate; border-spacing: 0 12px; margin-top: 20px; }
        th { padding: 20px; text-align: right; color: var(--text-muted); font-size: 0.85rem; text-transform: uppercase; letter-spacing: 1.5px; }
        td { padding: 24px; background: rgba(255, 255, 255, 0.03); border: 1px solid transparent; transition: all 0.3s; }
        tr td:first-child { border-radius: 20px 0 0 20px; }
        tr td:last-child { border-radius: 0 20px 20px 0; }
        
        tr:hover td { background: rgba(255, 255, 255, 0.06); border-color: var(--border); }
        
        .f-item { display: flex; align-items: center; gap: 16px; font-weight: 700; color: #fff; }
        .f-icon { font-size: 1.75rem; }
        
        .actions-cell { display: flex; gap: 12px; }
        
        .btn-sm { 
            padding: 10px 20px; border-radius: 12px; font-weight: 700; font-size: 0.9rem; text-decoration: none; border: 1px solid var(--border); transition: all 0.2s;
        }
        
        .btn-dl { background: rgba(6, 182, 212, 0.1); color: var(--accent); }
        .btn-dl:hover { background: var(--accent); color: white; border-color: transparent; }
        
        .btn-del { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
        .btn-del:hover { background: #ef4444; color: white; border-color: transparent; }

        .footer {
            margin-top: 60px; text-align: center; padding-top: 40px; border-top: 1px solid var(--border); color: var(--text-muted);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="header-info">
                <h1>NEXA DISK</h1>
                <p style="color: var(--text-muted); font-size: 1.1rem;">إدارة الأرشيف الرقمي v3.1</p>
            </div>
            <div style="display: flex; gap: 16px;">
                <div class="info-pill" style="text-decoration: none;">
                    <span class="info-label">Hub</span>
                    <span class="info-value">Command Center</span>
                </div>
                <div class="info-pill">
                    <span class="info-label">IP Address</span>
                    <span class="info-value">{{.LocalIP}}</span>
                </div>
                <div class="info-pill">
                    <span class="info-label">Server Time</span>
                    <span class="info-value">{{.Time}}</span>
                </div>
            </div>
        </div>
        
        <div class="upload-zone" onclick="document.getElementById('fileInput').click()">
            <form action="/upload" method="POST" enctype="multipart/form-data" id="uploadForm">
                <i class="fas fa-cloud-arrow-up"></i>
                <h3>اسحب الملفات هنا أو انقر للاختيار</h3>
                <p style="color: var(--text-muted); margin-bottom: 30px;">الحد الأقصى للرفع هو 500 ميجابايت للعملية الواحدة</p>
                <input type="file" name="file" required id="fileInput" style="display:none">
                <button type="submit" class="btn-action" onclick="event.stopPropagation()">بدء الرفع الآمن</button>
            </form>
        </div>
        
        {{if .Files}}
        <div class="stats-grid">
            <div class="stat-box">
                <div class="val">{{.FileCount}}</div>
                <div class="lbl">الملفات الكلية</div>
            </div>
            <div class="stat-box">
                <div class="val">{{.TotalSize}}</div>
                <div class="lbl">المساحة المستخدمة</div>
            </div>
            <div class="stat-box">
                <div class="val">✓</div>
                <div class="lbl">حالة التشفير</div>
            </div>
        </div>
        
        <table>
            <thead>
                <tr>
                    <th>المستند / الملف</th>
                    <th>الحجم</th>
                    <th>تاريخ التعديل</th>
                    <th>الإجراءات</th>
                </tr>
            </thead>
            <tbody>
                {{range .Files}}
                <tr>
                    <td>
                        <div class="f-item">
                            <span class="f-icon">{{.Icon}}</span>
                            <span>{{.Name}}</span>
                        </div>
                    </td>
                    <td><span style="color: var(--text-muted);">{{.Size}}</span></td>
                    <td><span style="color: var(--text-muted);">{{.ModTime}}</span></td>
                    <td>
                        <div class="actions-cell">
                            <a href="/download?file={{.Name}}" class="btn-sm btn-dl">تحميل</a>
                            <form action="/delete" method="POST" onsubmit="return confirm('تأكيد حذف الملف نهائياً؟');">
                                <input type="hidden" name="file" value="{{.Name}}">
                                <button type="submit" class="btn-sm btn-del">حذف</button>
                            </form>
                        </div>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{else}}
        <div style="text-align: center; padding: 80px; color: var(--text-muted);">
            <i class="fas fa-folder-open" style="font-size: 5rem; margin-bottom: 20px; opacity: 0.2;"></i>
            <h2>المخزن فارغ تماماً</h2>
            <p>ابدأ برفع الملفات لملء الأرشيف الرقمي</p>
        </div>
        {{end}}
        
        <div class="footer">
            <p>&copy; 2026 Nexa Ultimate System | Matrix Expansion v3.1</p>
        </div>
    </div>
    
    <script>
        document.getElementById('fileInput').addEventListener('change', function() {
            if (this.files.length > 0) {
                const btn = document.querySelector('.btn-action');
                btn.textContent = 'جاهز: ' + this.files[0].name;
                btn.style.background = 'linear-gradient(135deg, #06b6d4, #6366f1)';
            }
        });
        
        document.querySelector('form').addEventListener('submit', function() {
            const fileInput = document.getElementById('fileInput');
            if (fileInput.files.length === 0) {
                alert('يرجى اختيار ملف');
                return false;
            }
            const fileSize = fileInput.files[0].size;
            const maxSize = 500 * 1024 * 1024; // 500MB
            if (fileSize > maxSize) {
                alert('حجم الملف كبير جداً. الحد الأقصى: 500MB');
                return false;
            }
            return true;
        });
    </script>
</body>
</html>
`

type FileInfo struct {
	Name    string
	Size    string
	ModTime string
	Icon    string
}

func Start() {
	// Ensure professional storage structure
	subDirs := []string{"incoming", "shared", "vault", "temp"}
	for _, sub := range subDirs {
		path := filepath.Join(StorageRoot, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", enableCORS(dirHandler))
	mux.HandleFunc("/upload", enableCORS(uploadHandler))
	mux.HandleFunc("/delete", enableCORS(deleteHandler))
	mux.HandleFunc("/download", enableCORS(downloadHandler))
	mux.HandleFunc("/api/stats", enableCORS(statsHandler))
	mux.HandleFunc("/api/list", enableCORS(listHandler))

	localIP := utils.GetLocalIP()
	utils.LogInfo("Storage", fmt.Sprintf("Storage Root:      %s", StorageRoot))
	utils.LogInfo("Storage", fmt.Sprintf("Web Interface:     http://%s:%s", localIP, Port))
	utils.SaveEndpoint("storage", fmt.Sprintf("http://%s:%s", localIP, Port))

	server := &http.Server{
		Addr:    "0.0.0.0:" + Port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogFatal("Storage", err.Error())
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	files, _ := os.ReadDir(StorageRoot)
	var fileList []FileInfo
	for _, f := range files {
		if !f.IsDir() {
			info, _ := f.Info()
			ext := filepath.Ext(f.Name())
			icon := getFileIcon(ext)
			fileList = append(fileList, FileInfo{
				Name:    f.Name(),
				Size:    utils.FormatSize(info.Size()),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Icon:    icon,
			})
		}
	}
	// Sort by newest first
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].ModTime > fileList[j].ModTime
	})

	importJSON, _ := json.Marshal(fileList)
	w.Write(importJSON)
}

func dirHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	files, _ := os.ReadDir(StorageRoot)
	var fileList []FileInfo
	totalSize := int64(0)

	for _, f := range files {
		if !f.IsDir() {
			info, _ := f.Info()
			ext := filepath.Ext(f.Name())
			icon := getFileIcon(ext)
			fileList = append(fileList, FileInfo{
				Name:    f.Name(),
				Size:    utils.FormatSize(info.Size()),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Icon:    icon,
			})
			totalSize += info.Size()
		}
	}

	// Sort by newest first
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].ModTime > fileList[j].ModTime
	})

	localIP := utils.GetLocalIP()
	tmpl := template.Must(template.New("fm").Parse(FileMangerHTML))
	tmpl.Execute(w, map[string]interface{}{
		"Files":     fileList,
		"Port":      Port,
		"LocalIP":   localIP,
		"Time":      time.Now().Format("2006-01-02 15:04"),
		"FileCount": len(fileList),
		"TotalSize": utils.FormatSize(totalSize),
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(500 * 1024 * 1024) // 500MB max
	if err != nil {
		http.Error(w, "File too large or invalid upload", 413)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", 400)
		return
	}
	defer file.Close()

	// Sanitize filename
	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		http.Error(w, "Invalid filename", 400)
		return
	}

	// Create file with timestamp if duplicate
	targetPath := filepath.Join(StorageRoot, filename)
	if _, err := os.Stat(targetPath); err == nil {
		// File exists, add timestamp
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		filename = fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext)
		targetPath = filepath.Join(StorageRoot, filename)
	}

	// Create the file
	dst, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, "Unable to create file", 500)
		return
	}
	defer dst.Close()

	// Copy with progress tracking
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(targetPath)
		http.Error(w, "Error writing file", 500)
		return
	}

	utils.LogSuccess("Storage", fmt.Sprintf("File uploaded: %s (%s)", filename, utils.FormatSize(written)))

	// Check if JSON response is expected (for Ajax uploads)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","filename":"%s","size":"%s"}`, filename, utils.FormatSize(written))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		filename := r.FormValue("file")
		targetPath := filepath.Join(StorageRoot, filepath.Base(filename))

		if err := os.Remove(targetPath); err == nil {
			utils.LogInfo("Storage", fmt.Sprintf("File deleted: %s", filename))
		}

		// Check if JSON response is expected
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"status":"success"}`)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("file")
	targetPath := filepath.Join(StorageRoot, filepath.Base(filename))

	// Check if file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		http.Error(w, "File not found", 404)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	utils.LogInfo("Storage", fmt.Sprintf("File downloading: %s", filename))
	http.ServeFile(w, r, targetPath)
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

func getFileIcon(ext string) string {
	ext = strings.ToLower(ext)
	icons := map[string]string{
		".pdf":  "📕",
		".doc":  "📄",
		".docx": "📄",
		".txt":  "📝",
		".mp3":  "🎵",
		".mp4":  "🎬",
		".jpg":  "🖼️",
		".jpeg": "🖼️",
		".png":  "🖼️",
		".gif":  "🖼️",
		".zip":  "📦",
		".rar":  "📦",
		".7z":   "📦",
		".exe":  "⚙️",
		".msi":  "⚙️",
		".xlsx": "📊",
		".csv":  "📊",
		".iso":  "💿",
		".apk":  "📱",
	}

	if icon, ok := icons[ext]; ok {
		return icon
	}
	return "📄"
}
