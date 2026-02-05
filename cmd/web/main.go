package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	Port        = "8081"
	StorageRoot = "./storage"
)

// HTML Template for the File Manager - Professional UI
const FileMangerHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexa - مدير الملفات والتخزين</title>
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: 'Cairo', sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container { 
            max-width: 1000px; 
            margin: 0 auto; 
            background: white; 
            padding: 40px;
            border-radius: 20px; 
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        
        .header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 3px solid #667eea;
        }
        
        .header-info h1 {
            color: #2c3e50;
            font-size: 2.5em;
            margin-bottom: 5px;
        }
        
        .server-info {
            display: flex;
            gap: 30px;
            font-size: 0.95em;
        }
        
        .info-item {
            background: #f0f4ff;
            padding: 12px 20px;
            border-radius: 10px;
            border-left: 4px solid #667eea;
        }
        
        .info-label {
            color: #7f8c8d;
            font-size: 0.85em;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .info-value {
            color: #2c3e50;
            font-weight: bold;
            font-size: 1.1em;
            margin-top: 3px;
            font-family: 'Courier New', monospace;
        }
        
        .upload-area { 
            background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
            padding: 40px; 
            border: 2px dashed #667eea; 
            border-radius: 15px; 
            text-align: center; 
            margin-bottom: 40px;
            transition: all 0.3s ease;
        }
        
        .upload-area:hover {
            border-color: #764ba2;
            background: linear-gradient(135deg, #667eea25 0%, #764ba225 100%);
        }
        
        .upload-area h3 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 1.3em;
        }
        
        .file-input-wrapper {
            position: relative;
            overflow: hidden;
            display: inline-block;
        }
        
        .file-input-wrapper input[type=file] {
            position: absolute;
            left: -9999px;
        }
        
        .file-label {
            display: inline-block;
            padding: 12px 30px;
            background: #667eea;
            color: white;
            border-radius: 8px;
            cursor: pointer;
            font-weight: bold;
            transition: background 0.3s ease;
            margin-right: 10px;
        }
        
        .file-label:hover {
            background: #764ba2;
        }
        
        .btn { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; 
            border: none; 
            padding: 12px 35px; 
            border-radius: 8px; 
            cursor: pointer; 
            font-family: 'Cairo', sans-serif;
            font-weight: bold;
            font-size: 1em;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .btn:hover { 
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }
        
        .btn:active {
            transform: translateY(0);
        }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .stat-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 12px;
            text-align: center;
        }
        
        .stat-number {
            font-size: 2em;
            font-weight: bold;
        }
        
        .stat-label {
            font-size: 0.9em;
            opacity: 0.9;
            margin-top: 5px;
        }
        
        table { 
            width: 100%; 
            border-collapse: collapse;
            margin-bottom: 30px;
        }
        
        th { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 18px;
            text-align: right;
            font-weight: bold;
            border-radius: 8px 8px 0 0;
        }
        
        td { 
            padding: 15px 18px; 
            border-bottom: 1px solid #ecf0f1;
        }
        
        tr:hover { 
            background: #f8f9ff;
        }
        
        .file-name {
            display: flex;
            align-items: center;
            gap: 10px;
            color: #2c3e50;
            font-weight: 600;
            word-break: break-word;
        }
        
        .file-icon {
            font-size: 1.5em;
            min-width: 20px;
        }
        
        .file-link {
            color: #667eea;
            text-decoration: none;
            font-weight: bold;
            transition: color 0.3s ease;
        }
        
        .file-link:hover {
            color: #764ba2;
            text-decoration: underline;
        }
        
        .size { 
            color: #7f8c8d; 
            font-size: 0.95em;
            font-family: 'Courier New', monospace;
        }
        
        .actions {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        
        .delete-btn { 
            color: white;
            background: #e74c3c;
            border: none;
            padding: 8px 15px;
            border-radius: 6px;
            cursor: pointer;
            font-weight: bold;
            font-size: 0.9em;
            transition: background 0.3s ease;
        }
        
        .delete-btn:hover {
            background: #c0392b;
        }
        
        .download-btn {
            color: white;
            background: #27ae60;
            border: none;
            padding: 8px 15px;
            border-radius: 6px;
            text-decoration: none;
            cursor: pointer;
            font-weight: bold;
            font-size: 0.9em;
            transition: background 0.3s ease;
            display: inline-block;
        }
        
        .download-btn:hover {
            background: #229954;
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #7f8c8d;
        }
        
        .empty-state-icon {
            font-size: 4em;
            margin-bottom: 20px;
            opacity: 0.5;
        }
        
        .empty-state-text {
            font-size: 1.2em;
            margin-bottom: 10px;
        }
        
        .footer {
            text-align: center;
            padding: 20px;
            color: #95a5a6;
            font-size: 0.9em;
            border-top: 1px solid #ecf0f1;
            margin-top: 30px;
        }
        
        .success-message {
            background: #d4edda;
            color: #155724;
            padding: 15px 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            border-left: 4px solid #28a745;
            display: none;
        }
        
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 15px 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            border-left: 4px solid #f5c6cb;
            display: none;
        }
        
        @media (max-width: 768px) {
            .container { padding: 20px; }
            .header { flex-direction: column; align-items: flex-start; gap: 20px; }
            .server-info { flex-direction: column; gap: 15px; }
            .header-info h1 { font-size: 1.8em; }
            .stats { grid-template-columns: 1fr; }
            .actions { flex-direction: column; }
            .delete-btn, .download-btn { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="header-info">
                <h1>⚡ Nexa Storage</h1>
                <p style="color: #7f8c8d; margin-top: 5px;">نظام التخزين والملفات المتقدم</p>
            </div>
            <div class="server-info">
                <div class="info-item">
                    <div class="info-label">العنوان المحلي</div>
                    <div class="info-value">{{.LocalIP}}:{{.Port}}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">الوقت</div>
                    <div class="info-value">{{.Time}}</div>
                </div>
            </div>
        </div>
        
        <div class="upload-area">
            <h3>📤 رفع ملف جديد</h3>
            <form action="/upload" method="POST" enctype="multipart/form-data">
                <div class="file-input-wrapper">
                    <label class="file-label">اختر الملف</label>
                    <input type="file" name="file" required id="fileInput">
                </div>
                <button type="submit" class="btn">رفع الملف</button>
            </form>
        </div>
        
        {{if .Files}}
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">{{.FileCount}}</div>
                <div class="stat-label">عدد الملفات</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.TotalSize}}</div>
                <div class="stat-label">الحجم الكلي</div>
            </div>
        </div>
        
        <table>
            <thead>
                <tr>
                    <th style="width: 40%;">📄 اسم الملف</th>
                    <th style="width: 15%;">💾 الحجم</th>
                    <th style="width: 25%;">📅 آخر تعديل</th>
                    <th style="width: 20%;">⚙️ الإجراءات</th>
                </tr>
            </thead>
            <tbody>
                {{range .Files}}
                <tr>
                    <td class="file-name">
                        <span class="file-icon">{{.Icon}}</span>
                        <a href="/download?file={{.Name}}" class="file-link" target="_blank">{{.Name}}</a>
                    </td>
                    <td class="size">{{.Size}}</td>
                    <td class="size">{{.ModTime}}</td>
                    <td class="actions">
                        <a href="/download?file={{.Name}}" class="download-btn">تحميل</a>
                        <form action="/delete" method="POST" style="display:inline" onsubmit="return confirm('هل أنت متأكد من حذف هذا الملف؟');">
                            <input type="hidden" name="file" value="{{.Name}}">
                            <button type="submit" class="delete-btn">حذف</button>
                        </form>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{else}}
        <div class="empty-state">
            <div class="empty-state-icon">📂</div>
            <div class="empty-state-text">لا توجد ملفات بعد</div>
            <p style="opacity: 0.7;">ابدأ برفع ملفك الأول الآن</p>
        </div>
        {{end}}
        
        <div class="footer">
            <p>🌐 Nexa Protocol v2.1 | نظام التخزين المحلي المتقدم</p>
            <p style="margin-top: 10px; opacity: 0.7;">يمكن الوصول إليه من أي جهاز على الشبكة المحلية</p>
        </div>
    </div>
    
    <script>
        document.getElementById('fileInput').addEventListener('change', function() {
            if (this.files.length > 0) {
                console.log('Selected file:', this.files[0].name);
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

// Helper function to get the local IP address
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func main() {
	// Ensure storage directory exists
	if _, err := os.Stat(StorageRoot); os.IsNotExist(err) {
		os.Mkdir(StorageRoot, 0755)
	}

	http.HandleFunc("/", enableCORS(dirHandler))
	http.HandleFunc("/upload", enableCORS(uploadHandler))
	http.HandleFunc("/delete", enableCORS(deleteHandler))
	http.HandleFunc("/download", enableCORS(downloadHandler))
	http.HandleFunc("/api/stats", enableCORS(statsHandler))
	http.HandleFunc("/api/list", enableCORS(listHandler))

	localIP := getLocalIP()
	fmt.Println(`
     _____ _ _         __  __                                   
    |  ___(_) | ___   |  \/  | __ _ _ __   __ _  __ _  ___ _ __ 
    | |_  | | |/ _ \  | |\/| |/ _' | '_ \ / _' |/ _' |/ _ \ '__|
    |  _| | | |  __/  | |  | | (_| | | | | (_| | (_| |  __/ |   
    |_|   |_|_|\___|  |_|  |_|\__,_|_| |_|\__,_|\__, |\___|_|   
                                                |___/ v3.0 Ultimate`)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Printf("   [INFO]  Initializing File Storage System...\n")
	fmt.Printf("   [INFO]  Storage Root:      %s\n", StorageRoot)
	fmt.Printf("   [INFO]  Web Interface:     http://%s:%s\n", localIP, Port)
	fmt.Printf("   [INFO]  CORS Policy:       %s\n", "ENABLED (Access-Control-Allow-Origin: *)")
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅  STORAGE SERVER READY")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+Port, nil))
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
				Size:    formatSize(info.Size()),
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
				Size:    formatSize(info.Size()),
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

	localIP := getLocalIP()
	tmpl := template.Must(template.New("fm").Parse(FileMangerHTML))
	tmpl.Execute(w, map[string]interface{}{
		"Files":     fileList,
		"Port":      Port,
		"LocalIP":   localIP,
		"Time":      time.Now().Format("2006-01-02 15:04"),
		"FileCount": len(fileList),
		"TotalSize": formatSize(totalSize),
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

	log.Printf("✅ تم رفع الملف: %s (%s)", filename, formatSize(written))

	// Check if JSON response is expected (for Ajax uploads)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"success","filename":"%s","size":"%s"}`, filename, formatSize(written))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		filename := r.FormValue("file")
		targetPath := filepath.Join(StorageRoot, filepath.Base(filename))

		if err := os.Remove(targetPath); err == nil {
			log.Printf("🗑️  تم حذف الملف: %s", filename)
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

	log.Printf("📥 تم تحميل الملف: %s", filename)
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
	fmt.Fprintf(w, `{"files":%d,"totalSize":%d,"totalSizeFormatted":"%s"}`, fileCount, totalSize, formatSize(totalSize))
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

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
