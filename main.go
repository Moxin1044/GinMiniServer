package main

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gin-gonic/gin"
)

const version = "1.0.0"

// ============ TUI 样式 ============

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
)

// ============ TUI 模型 ============

type tickMsg time.Time

type model struct {
	host      string
	port      int
	dir       string
	password  string
	logDir    string
	startTime time.Time
	spinner   int
	quitting  bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func initialModel(host string, port int, dir, password, logDir string) model {
	return model{
		host:      host,
		port:      port,
		dir:       dir,
		password:  password,
		logDir:    logDir,
		startTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.spinner = (m.spinner + 1) % len(spinnerFrames)
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "\n  " + statusStyle.Render("●") + " 服务已停止，再见！\n\n"
	}

	logo := logoStyle.Render(`
  ______      ____                    
 / ___(_)__  / __/__ _____  _____ ____
/ (_ / / _ \_\ \/ -_) __/ |/ / -_) __/
\___/_/_//_/___/\__/_/  |___/\__/_/   
                                      
`)

	uptime := time.Since(m.startTime).Round(time.Second)
	now := time.Now().Format("2006-01-02 15:04:05")
	displayHost := m.host
	if m.host == "0.0.0.0" {
		displayHost = "localhost"
	}

	authStatus := valueStyle.Render("无")
	if m.password != "" {
		authStatus = warnStyle.Render("已启用")
	}

	logStatus := valueStyle.Render("关闭")
	if m.logDir != "" {
		logStatus = warnStyle.Render(m.logDir)
	}

	infoItems := []struct {
		label string
		value string
	}{
		{"版本", fmt.Sprintf("v%s", version)},
		{"平台", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
		{"时间", now},
		{"运行", uptime.String()},
		{"监听", fmt.Sprintf("http://%s:%d", displayHost, m.port)},
		{"目录", m.dir},
		{"认证", authStatus},
		{"日志", logStatus},
	}

	maxLabel := 0
	for _, item := range infoItems {
		if len(item.label) > maxLabel {
			maxLabel = len(item.label)
		}
	}

	var infoLines string
	for i, item := range infoItems {
		pad := strings.Repeat(" ", maxLabel-len(item.label))
		line := "  " + labelStyle.Render("◆ ") + labelStyle.Render(item.label) + pad + "  " + item.value
		if i < len(infoItems)-1 {
			line += "\n"
		}
		infoLines += line
	}

	divider := dividerStyle.Render("  " + strings.Repeat("─", 46))
	spinner := statusStyle.Render(spinnerFrames[m.spinner])
	status := fmt.Sprintf("  %s %s", spinner, statusStyle.Render("服务运行中"))
	hint := hintStyle.Render("  按 Ctrl+C 或 q 停止服务")

	content := logo + "\n" + titleStyle.Render("  GinMiniServer — 轻量级静态文件服务器") + "\n\n" + infoLines + "\n" + divider + "\n" + status + "\n" + hint
	box := borderStyle.Render(content)
	return "\n" + box + "\n"
}

// ============ 网页文件浏览器 ============

const fileListTpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Path}} — GinMiniServer</title>
<style>
  :root {
    --bg: #0d1117;
    --surface: #161b22;
    --surface-hover: #1c2330;
    --border: #30363d;
    --text: #e6edf3;
    --text-dim: #8b949e;
    --accent: #58a6ff;
    --accent-2: #bc8cff;
    --green: #3fb950;
    --yellow: #d29922;
    --red: #f85149;
    --radius: 10px;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    padding: 24px;
  }
  .container { max-width: 960px; margin: 0 auto; }

  /* 顶部标题栏 */
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    flex-wrap: wrap;
    gap: 12px;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .brand-icon {
    width: 36px; height: 36px;
    background: linear-gradient(135deg, var(--accent), var(--accent-2));
    border-radius: 8px;
    display: flex; align-items: center; justify-content: center;
    font-size: 18px; font-weight: 800; color: #fff;
  }
  .brand-text { font-size: 18px; font-weight: 700; }
  .brand-text span { color: var(--accent); }

  /* 搜索框 */
  .search-wrap { position: relative; flex: 0 0 auto; }
  .search {
    width: 240px;
    padding: 8px 14px 8px 36px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text);
    font-size: 14px;
    outline: none;
    transition: border-color .2s;
  }
  .search:focus { border-color: var(--accent); }
  .search-icon {
    position: absolute; left: 12px; top: 50%;
    transform: translateY(-50%);
    color: var(--text-dim); font-size: 14px;
  }

  /* 面包屑 */
  .breadcrumb {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 16px;
    font-size: 14px;
    color: var(--text-dim);
    flex-wrap: wrap;
  }
  .breadcrumb a {
    color: var(--accent);
    text-decoration: none;
    padding: 4px 8px;
    border-radius: 6px;
    transition: background .2s;
  }
  .breadcrumb a:hover { background: var(--surface); }
  .breadcrumb .sep { color: var(--border); }
  .breadcrumb .current { color: var(--text); font-weight: 600; padding: 4px 8px; }

  /* 文件列表 */
  .file-list {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
  }
  .file-row {
    display: grid;
    grid-template-columns: 36px 1fr 120px 160px;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    color: var(--text);
    transition: background .15s;
    gap: 12px;
  }
  .file-row:last-child { border-bottom: none; }
  .file-row:hover { background: var(--surface-hover); }
  .file-row.header-row {
    font-size: 12px;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: .5px;
    cursor: default;
    background: rgba(255,255,255,.02);
  }
  .file-row.header-row:hover { background: rgba(255,255,255,.02); }
  .file-icon { font-size: 18px; text-align: center; }
  .file-name { font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .file-name a { color: var(--text); text-decoration: none; }
  .file-row:hover .file-name a { color: var(--accent); }
  .file-size { font-size: 13px; color: var(--text-dim); text-align: right; }
  .file-time { font-size: 13px; color: var(--text-dim); text-align: right; }

  /* 空目录 */
  .empty {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-dim);
  }
  .empty-icon { font-size: 48px; margin-bottom: 12px; }

  /* 页脚 */
  .footer {
    margin-top: 24px;
    text-align: center;
    font-size: 12px;
    color: var(--text-dim);
  }
  .footer a { color: var(--accent); text-decoration: none; }

  /* 响应式 */
  @media (max-width: 640px) {
    .file-row { grid-template-columns: 32px 1fr 80px; }
    .file-time { display: none; }
    .search { width: 160px; }
  }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <div class="brand">
      <div class="brand-icon">G</div>
      <div class="brand-text">Gin<span>MiniServer</span></div>
    </div>
    <div class="search-wrap">
      <span class="search-icon">🔍</span>
      <input type="text" class="search" id="search" placeholder="搜索文件..." oninput="filterFiles()">
    </div>
  </div>

  <div class="breadcrumb">
    <a href="/">🏠 根目录</a>
    {{range .Breadcrumbs}}
    <span class="sep">/</span>
    <a href="{{.URL}}">{{.Name}}</a>
    {{end}}
    {{if .CurrentName}}<span class="sep">/</span><span class="current">{{.CurrentName}}</span>{{end}}
  </div>

  <div class="file-list" id="fileList">
    <div class="file-row header-row">
      <div class="file-icon"></div>
      <div class="file-name">名称</div>
      <div class="file-size">大小</div>
      <div class="file-time">修改时间</div>
    </div>
    {{if .ParentURL}}
    <a class="file-row" href="{{.ParentURL}}" data-name="..">
      <div class="file-icon">📁</div>
      <div class="file-name">..</div>
      <div class="file-size">—</div>
      <div class="file-time">—</div>
    </a>
    {{end}}
    {{range .Entries}}
    <a class="file-row" href="{{.URL}}" data-name="{{.Name}}">
      <div class="file-icon">{{.Icon}}</div>
      <div class="file-name">{{.Name}}</div>
      <div class="file-size">{{.Size}}</div>
      <div class="file-time">{{.ModTime}}</div>
    </a>
    {{else}}
    {{if not .ParentURL}}
    <div class="empty">
      <div class="empty-icon">📂</div>
      <div>此目录为空</div>
    </div>
    {{end}}
    {{end}}
  </div>

  <div class="footer">
    GinMiniServer v{{.Version}} · <a href="https://github.com/Moxin1044/GinMiniServer">Powered by Moxin1044</a>
  </div>
</div>
<script>
function filterFiles() {
  const q = document.getElementById('search').value.toLowerCase();
  document.querySelectorAll('#fileList .file-row:not(.header-row)').forEach(function(row) {
    const name = (row.dataset.name || '').toLowerCase();
    row.style.display = name.indexOf(q) !== -1 ? '' : 'none';
  });
}
</script>
</body>
</html>`

type breadcrumbItem struct {
	Name string
	URL  string
}

type fileEntry struct {
	Name    string
	URL     string
	Icon    string
	Size    string
	ModTime string
}

type pageData struct {
	Path         string
	Breadcrumbs  []breadcrumbItem
	CurrentName  string
	ParentURL    string
	Entries      []fileEntry
	Version      string
}

var tpl = template.Must(template.New("files").Parse(fileListTpl))

// fileIcon 根据扩展名返回文件图标
func fileIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "🐹"
	case ".js", ".jsx", ".ts", ".tsx":
		return "📜"
	case ".py":
		return "🐍"
	case ".html", ".htm":
		return "🌐"
	case ".css", ".scss":
		return "🎨"
	case ".json", ".xml", ".yaml", ".yml", ".toml":
		return "⚙️"
	case ".md", ".txt", ".rst":
		return "📄"
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".ico":
		return "🖼️"
	case ".mp3", ".wav", ".flac", ".ogg":
		return "🎵"
	case ".mp4", ".avi", ".mkv", ".mov", ".webm":
		return "🎬"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "📦"
	case ".pdf":
		return "📕"
	case ".doc", ".docx":
		return "📘"
	case ".xls", ".xlsx":
		return "📗"
	case ".ppt", ".pptx":
		return "📙"
	case ".exe", ".msi", ".dmg", ".app":
		return "⚙️"
	case ".sh", ".bash", ".bat", ".ps1":
		return "🔧"
	case ".c", ".cpp", ".h", ".hpp", ".rs", ".java", ".rb":
		return "💻"
	default:
		return "📄"
	}
}

// formatSize 格式化文件大小
func formatSize(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// serveFiles 自定义文件服务 handler
func serveFiles(root string) gin.HandlerFunc {
	rootAbs, _ := filepath.Abs(root)

	return func(c *gin.Context) {
		relPath := c.Param("filepath")
		if relPath == "" {
			relPath = "/"
		}
		// 清理路径
		relPath = strings.TrimPrefix(relPath, "/")

		fullPath := filepath.Join(rootAbs, relPath)

		info, err := os.Stat(fullPath)
		if err != nil {
			c.Status(http.StatusNotFound)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(404, "<h1>404 Not Found</h1><p>文件或目录不存在</p>")
			return
		}

		// 如果是文件，直接服务文件
		if !info.IsDir() {
			http.ServeFile(c.Writer, c.Request, fullPath)
			return
		}

		// 目录：检查是否有 index.html
		indexPath := filepath.Join(fullPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(c.Writer, c.Request, indexPath)
			return
		}

		// 渲染目录列表
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			c.String(http.StatusInternalServerError, "读取目录失败")
			return
		}

		// 分离目录和文件，分别排序
		var dirs, files []os.DirEntry
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, e)
			} else {
				files = append(files, e)
			}
		}
		sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

		// 构建面包屑
		var breadcrumbs []breadcrumbItem
		var currentName string
		parts := strings.Split(strings.TrimSuffix(relPath, "/"), "/")
		if relPath != "" && relPath != "." {
			for i := 0; i < len(parts); i++ {
				if parts[i] == "" {
					continue
				}
				url := "/" + strings.Join(parts[:i+1], "/") + "/"
				if i == len(parts)-1 {
					currentName = parts[i]
				} else {
					breadcrumbs = append(breadcrumbs, breadcrumbItem{Name: parts[i], URL: url})
				}
			}
		}

		// 父目录链接
		var parentURL string
		if relPath != "" && relPath != "." {
			parentParts := parts[:len(parts)-1]
			if len(parentParts) == 0 || (len(parentParts) == 1 && parentParts[0] == "") {
				parentURL = "/"
			} else {
				parentURL = "/" + strings.Join(parentParts, "/") + "/"
			}
		}

		// 构建文件条目
		var fileEntries []fileEntry
		for _, e := range append(dirs, files...) {
			info, _ := e.Info()
			name := e.Name()
			url := "/" + relPath
			if relPath != "" && !strings.HasSuffix(url, "/") {
				url += "/"
			} else if relPath == "" {
				url = "/"
			}
			if e.IsDir() {
				url += name + "/"
			} else {
				url += name
			}

			size := "—"
			if !e.IsDir() && info != nil {
				size = formatSize(info.Size())
			}
			modTime := "—"
			if info != nil {
				modTime = info.ModTime().Format("2006-01-02 15:04")
			}

			fileEntries = append(fileEntries, fileEntry{
				Name:    name,
				URL:     url,
				Icon:    fileIcon(name, e.IsDir()),
				Size:    size,
				ModTime: modTime,
			})
		}

		data := pageData{
			Path:        "/" + relPath,
			Breadcrumbs: breadcrumbs,
			CurrentName: currentName,
			ParentURL:   parentURL,
			Entries:     fileEntries,
			Version:     version,
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		tpl.Execute(c.Writer, data)
	}
}

// authMiddleware Basic Auth 中间件
func authMiddleware(password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if password == "" {
			c.Next()
			return
		}
		user, pass, ok := c.Request.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte("admin")) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			c.Header("WWW-Authenticate", `Basic realm="GinMiniServer"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

// logger 日志记录器（线程安全写入文件）
type fileLogger struct {
	mu sync.Mutex
	f  *os.File
}

func newFileLogger(dir string) (*fileLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	logPath := filepath.Join(dir, time.Now().Format("2006-01-02")+".log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	return &fileLogger{f: f}, nil
}

func (l *fileLogger) write(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.f.WriteString(line)
}

func (l *fileLogger) close() {
	if l.f != nil {
		l.f.Close()
	}
}

// logMiddleware 日志中间件，将请求记录到文件
func logMiddleware(logger *fileLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		size := c.Writer.Size()

		line := fmt.Sprintf("%s | %3d | %13s | %15s | %-7s | %10d | %s\n",
			start.Format("2006-01-02 15:04:05"),
			status,
			latency,
			clientIP,
			method,
			size,
			path,
		)
		logger.write(line)
	}
}

// ============ 主函数 ============

func main() {
	port := flag.Int("port", 8080, "HTTP 服务端口")
	dir := flag.String("dir", ".", "要服务的文件目录")
	host := flag.String("host", "0.0.0.0", "绑定的主机地址")
	password := flag.String("password", "", "访问密码（留空则无认证）")
	logDir := flag.String("log", "", "日志输出目录（留空则不记录日志）")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	if *showVersion {
		fmt.Printf("GinMiniServer v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	absDir := *dir
	if absDir == "." {
		if wd, err := os.Getwd(); err == nil {
			absDir = wd
		}
	}

	// 解析日志目录
	absLogDir := ""
	var logger *fileLogger
	if *logDir != "" {
		absLogDir, _ = filepath.Abs(*logDir)
		var err error
		logger, err = newFileLogger(absLogDir)
		if err != nil {
			fmt.Printf("  [错误] %v\n", err)
			os.Exit(1)
		}
		defer logger.close()
	}

	// 设置 Gin Release 模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 日志中间件
	if logger != nil {
		r.Use(logMiddleware(logger))
	}

	// 认证中间件
	r.Use(authMiddleware(*password))

	// 自定义文件服务
	r.GET("/*filepath", serveFiles(*dir))

	// 后台启动 HTTP 服务
	addr := fmt.Sprintf("%s:%d", *host, *port)
	go func() {
		if err := r.Run(addr); err != nil {
			fmt.Printf("\n  [错误] 服务启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	// 启动 TUI
	p := tea.NewProgram(initialModel(*host, *port, absDir, *password, absLogDir), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI 启动失败: %v\n", err)
		os.Exit(1)
	}
}
