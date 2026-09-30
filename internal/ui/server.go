// Package ui 创建器的本地页面。
//
// 只监听 127.0.0.1 的随机端口。页面每隔一段时间发一次心跳；页面关闭、
// 或长时间收不到心跳时，进程自己退出，不会一直占着端口。
package ui

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/create"
	"github.com/star/star-yi-cli/internal/source"
)

//go:embed index.html
var indexHTML string

const (
	idleTimeout  = 150 * time.Second
	startupGrace = 3 * time.Minute
	byeGrace     = 8 * time.Second
)

// Options 页面服务参数。
type Options struct {
	Port        int
	OpenBrowser bool
	Version     string
	Log         func(string)
}

type server struct {
	opts     Options
	token    string
	host     string
	mu       sync.Mutex
	lastSeen atomic.Int64
	busy     atomic.Int32
	quit     chan struct{}
	quitOnce sync.Once
}

// Run 启动页面并阻塞，直到页面关闭或收到 ctx 取消。
func Run(ctx context.Context, opts Options) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
	if err != nil {
		return fmt.Errorf("监听本机端口: %w", err)
	}
	s := &server{opts: opts, token: newToken(), host: ln.Addr().String(), quit: make(chan struct{})}
	s.lastSeen.Store(time.Now().Add(startupGrace - idleTimeout).UnixNano())

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/state", s.api(s.handleState))
	mux.HandleFunc("/api/config", s.api(s.handleSaveConfig))
	mux.HandleFunc("/api/config/reset", s.api(s.handleResetEdition))
	mux.HandleFunc("/api/check", s.api(s.handleCheck))
	mux.HandleFunc("/api/create", s.api(s.handleCreate))
	mux.HandleFunc("/api/pick-dir", s.api(s.handlePickDir))
	mux.HandleFunc("/api/open-dir", s.api(s.handleOpenDir))
	mux.HandleFunc("/api/ping", s.api(func(w http.ResponseWriter, r *http.Request) (any, error) { return "ok", nil }))
	mux.HandleFunc("/api/bye", s.handleBye)
	mux.HandleFunc("/api/quit", s.api(func(w http.ResponseWriter, r *http.Request) (any, error) {
		go func() { time.Sleep(300 * time.Millisecond); s.stop() }()
		return "bye", nil
	}))

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	url := "http://" + s.host + "/"
	s.log("创建器页面：" + url)
	s.log("关闭页面后会自动退出；也可以在这里按 Ctrl+C。")
	if opts.OpenBrowser {
		if err := openURL(url); err != nil {
			s.log("没能自动打开浏览器，请手动打开上面的地址。")
		}
	}

	go s.watchdog()
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case <-ctx.Done():
	case <-s.quit:
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	s.log("创建器已退出。")
	return nil
}

func (s *server) log(msg string) {
	if s.opts.Log != nil {
		s.opts.Log(msg)
	}
}

func (s *server) stop() { s.quitOnce.Do(func() { close(s.quit) }) }

func (s *server) touch() { s.lastSeen.Store(time.Now().UnixNano()) }

func (s *server) watchdog() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-s.quit:
			return
		case <-t.C:
			if s.busy.Load() > 0 {
				s.touch()
				continue
			}
			if time.Since(time.Unix(0, s.lastSeen.Load())) > idleTimeout {
				s.log("页面已关闭。")
				s.stop()
				return
			}
		}
	}
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// sameHost 拒绝别的网站借浏览器访问本机端口。
func (s *server) sameHost(r *http.Request) bool {
	return r.Host == s.host
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !s.sameHost(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.touch()
	page := strings.Replace(indexHTML, "__STAR_YI_TOKEN__", s.token, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(page))
}

func (s *server) handleBye(w http.ResponseWriter, r *http.Request) {
	if !s.sameHost(r) || r.URL.Query().Get("t") != s.token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	s.lastSeen.Store(time.Now().Add(byeGrace - idleTimeout).UnixNano())
	w.WriteHeader(http.StatusNoContent)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) (any, error)

func (s *server) api(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.sameHost(r) || r.Header.Get("X-Star-Yi-Token") != s.token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.touch()
		data, err := fn(w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error(), "data": data})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
	}
}

func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("请求内容不是合法 JSON: %w", err)
	}
	return nil
}

type stateResp struct {
	Version     string            `json:"version"`
	ConfigPath  string            `json:"configPath"`
	Editions    []*config.Edition `json:"editions"`
	Defaults    []*config.Edition `json:"defaults"`
	Cwd         string            `json:"cwd"`
	Home        string            `json:"home"`
	GitOK       bool              `json:"gitOk"`
	CanPickDir  bool              `json:"canPickDir"`
	PathSep     string            `json:"pathSep"`
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) (any, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	p, _ := config.Path()
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	return stateResp{
		Version: s.opts.Version, ConfigPath: p, Editions: cfg.Editions, Defaults: config.Defaults(),
		Cwd: cwd, Home: home, GitOK: source.GitAvailable(), CanPickDir: runtime.GOOS == "windows",
		PathSep: string(filepath.Separator),
	}, nil
}

func (s *server) handleSaveConfig(w http.ResponseWriter, r *http.Request) (any, error) {
	var cfg config.Config
	if err := decode(r, &cfg); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := config.Save(&cfg); err != nil {
		return nil, err
	}
	saved, err := config.Load()
	if err != nil {
		return nil, err
	}
	return saved.Editions, nil
}

func (s *server) handleResetEdition(w http.ResponseWriter, r *http.Request) (any, error) {
	var req struct {
		ID string `json:"id"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	d, ok := config.DefaultEdition(req.ID)
	if !ok {
		return nil, fmt.Errorf("%s 不是内置版本", req.ID)
	}
	return d, nil
}

func (s *server) handleCheck(w http.ResponseWriter, r *http.Request) (any, error) {
	var req struct {
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Repo) == "" {
		return nil, errors.New("仓库地址为空")
	}
	return source.Check(strings.TrimSpace(req.Repo), strings.TrimSpace(req.Branch))
}

type createReq struct {
	create.Options
	EditionID string `json:"edition"`
}

type createResp struct {
	Result *create.Result `json:"result,omitempty"`
	Logs   []string       `json:"logs"`
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) (any, error) {
	var req createReq
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	e, err := cfg.Edition(req.EditionID)
	if err != nil {
		return nil, err
	}
	s.busy.Add(1)
	defer s.busy.Add(-1)

	var mu sync.Mutex
	resp := &createResp{Logs: []string{}}
	o := req.Options
	o.Edition = e
	o.Log = func(line string) {
		mu.Lock()
		resp.Logs = append(resp.Logs, line)
		mu.Unlock()
		s.log(line)
	}
	res, err := create.Run(o)
	resp.Result = res
	if err != nil {
		s.log("创建失败：" + err.Error())
		return resp, err
	}
	return resp, nil
}

func (s *server) handlePickDir(w http.ResponseWriter, r *http.Request) (any, error) {
	var req struct {
		Initial string `json:"initial"`
	}
	_ = decode(r, &req)
	if runtime.GOOS != "windows" {
		return nil, errors.New("只有 Windows 支持选择文件夹对话框，请直接输入路径")
	}
	s.busy.Add(1)
	defer s.busy.Add(-1)
	script := `
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form -Property @{ TopMost = $true; ShowInTaskbar = $false; WindowState = 'Minimized' }
$d = New-Object System.Windows.Forms.FolderBrowserDialog
$d.Description = '选择创建位置（父目录）'
$d.ShowNewFolderButton = $true
if ($env:STAR_YI_INITIAL -and (Test-Path -LiteralPath $env:STAR_YI_INITIAL)) { $d.SelectedPath = $env:STAR_YI_INITIAL }
if ($d.ShowDialog($owner) -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($d.SelectedPath) }
$owner.Dispose()
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.Env = append(os.Environ(), "STAR_YI_INITIAL="+req.Initial)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("打开文件夹对话框失败: %w", err)
	}
	return strings.TrimSpace(strings.TrimPrefix(string(out), "\ufeff")), nil
}

func (s *server) handleOpenDir(w http.ResponseWriter, r *http.Request) (any, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := decode(r, &req); err != nil {
		return nil, err
	}
	info, err := os.Stat(req.Path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("目录不存在：%s", req.Path)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", req.Path)
	case "darwin":
		cmd = exec.Command("open", req.Path)
	default:
		cmd = exec.Command("xdg-open", req.Path)
	}
	_ = cmd.Start()
	go func() { _ = cmd.Wait() }()
	return "ok", nil
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
