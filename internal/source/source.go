// Package source 把骨架仓库下载到本地目录。
//
// 远程地址优先用 git 浅克隆；git 不可用或克隆失败时，GitHub 仓库改下载压缩包。
// 仓库地址也可以是本机路径：是 git 仓库时克隆已提交的内容，否则直接复制目录。
package source

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Request 一次下载。
type Request struct {
	Repo   string
	Branch string
	// Dest 目标目录，必须不存在。
	Dest string
	Log  func(string)
}

// skipDirs 复制本地目录时跳过的构建产物和工具目录。
var skipDirs = map[string]bool{
	".git": true, ".gradle": true, ".idea": true, ".vscode": true, ".kotlin": true,
	"build": true, "target": true, "out": true, "node_modules": true,
}

var githubRe = regexp.MustCompile(`^(?:https?://github\.com/|git@github\.com:|ssh://git@github\.com/)([^/]+)/([^/]+?)(?:\.git)?/?$`)

func (r *Request) log(format string, a ...any) {
	if r.Log != nil {
		r.Log(fmt.Sprintf(format, a...))
	}
}

// LocalDir 仓库地址是本机目录时返回它的绝对路径。
func LocalDir(repo string) (string, bool) {
	p := strings.TrimPrefix(repo, "file://")
	if strings.Contains(p, "://") || strings.HasPrefix(p, "git@") {
		return "", false
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return abs, true
	}
	return "", false
}

// GitAvailable 本机是否装了 git。
func GitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Fetch 下载仓库内容到 Dest，并去掉 .git。
func Fetch(r Request) error {
	if _, err := os.Stat(r.Dest); err == nil {
		return fmt.Errorf("下载目录 %s 已存在", r.Dest)
	}
	if local, ok := LocalDir(r.Repo); ok {
		if isGitRepo(local) && GitAvailable() {
			r.log("从本机仓库克隆 %s%s", local, branchNote(r.Branch))
			if err := gitClone(r, local, false); err != nil {
				return err
			}
		} else {
			r.log("复制本机目录 %s", local)
			if err := CopyTree(local, r.Dest, true); err != nil {
				return err
			}
		}
		return RemoveAll(filepath.Join(r.Dest, ".git"))
	}

	var gitErr error
	if GitAvailable() {
		r.log("git 克隆 %s%s", r.Repo, branchNote(r.Branch))
		gitErr = gitClone(r, r.Repo, true)
		if gitErr == nil {
			return RemoveAll(filepath.Join(r.Dest, ".git"))
		}
		_ = RemoveAll(r.Dest)
		r.log("git 克隆失败：%v", gitErr)
	} else {
		gitErr = errors.New("本机没有 git")
		r.log("本机没有找到 git")
	}

	m := githubRe.FindStringSubmatch(r.Repo)
	if m == nil {
		return fmt.Errorf("下载 %s 失败：%w", r.Repo, gitErr)
	}
	r.log("改为下载 GitHub 压缩包")
	if err := downloadGitHubZip(r, m[1], m[2]); err != nil {
		_ = RemoveAll(r.Dest)
		return fmt.Errorf("下载 %s 失败：git 克隆 %v；压缩包 %w", r.Repo, gitErr, err)
	}
	return nil
}

func branchNote(b string) string {
	if b == "" {
		return "（默认分支）"
	}
	return "（分支 " + b + "）"
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func gitCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=never")
	return cmd
}

func gitClone(r Request, from string, shallow bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	args := []string{"-c", "advice.detachedHead=false", "clone", "--quiet"}
	if shallow {
		args = append(args, "--depth", "1", "--single-branch")
	}
	if r.Branch != "" {
		args = append(args, "--branch", r.Branch)
	}
	args = append(args, "--", from, r.Dest)
	out, err := gitCommand(ctx, args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return errors.New(msg)
	}
	return nil
}

// Check 检查仓库地址能否访问，返回一句说明。
func Check(repo, branch string) (string, error) {
	if local, ok := LocalDir(repo); ok {
		if !isGitRepo(local) {
			return "本机目录，创建时直接复制（不是 git 仓库）", nil
		}
		if branch == "" {
			return "本机 git 仓库，创建时克隆默认分支", nil
		}
		if !GitAvailable() {
			return "本机 git 仓库，但没有找到 git，创建时会直接复制目录", nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := gitCommand(ctx, "-C", local, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run(); err != nil {
			return "", fmt.Errorf("本机仓库里没有分支 %s", branch)
		}
		return "本机 git 仓库，分支 " + branch + " 存在", nil
	}
	if !GitAvailable() {
		if githubRe.MatchString(repo) {
			return "本机没有 git，创建时会下载 GitHub 压缩包", nil
		}
		return "", errors.New("本机没有 git，非 GitHub 地址无法下载")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"ls-remote", "--heads", "--", repo}
	if branch != "" {
		args = append(args, branch)
	}
	out, err := gitCommand(ctx, args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if ctx.Err() != nil {
			msg = "30 秒内没有响应"
		}
		if githubRe.MatchString(repo) {
			return "", fmt.Errorf("git 访问失败：%s。创建时会再尝试下载 GitHub 压缩包", msg)
		}
		return "", fmt.Errorf("git 访问失败：%s", msg)
	}
	if branch != "" && strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("仓库可以访问，但没有分支 %s", branch)
	}
	if branch == "" {
		return "仓库可以访问，将使用默认分支", nil
	}
	return "仓库可以访问，分支 " + branch + " 存在", nil
}

func downloadGitHubZip(r Request, owner, repo string) error {
	ref := r.Branch
	if ref == "" {
		ref = "HEAD"
	} else {
		ref = "refs/heads/" + ref
	}
	url := fmt.Sprintf("https://codeload.github.com/%s/%s/zip/%s", owner, repo, ref)
	r.log("下载 %s", url)

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "star-yi-src-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return extractZipStripRoot(tmp.Name(), r.Dest)
}

// extractZipStripRoot 解压并去掉 GitHub 压缩包里的顶层目录。
func extractZipStripRoot(zipPath, dest string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destAbs, 0o755); err != nil {
		return err
	}
	for _, f := range zr.File {
		name := strings.ReplaceAll(f.Name, `\`, "/")
		i := strings.Index(name, "/")
		if i < 0 {
			continue
		}
		rel := name[i+1:]
		if rel == "" {
			continue
		}
		target := filepath.Join(destAbs, filepath.FromSlash(rel))
		if r, err := filepath.Rel(destAbs, target); err != nil || strings.HasPrefix(r, "..") {
			return fmt.Errorf("压缩包里有越界路径 %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	mode := os.FileMode(0o644)
	if f.Mode()&0o111 != 0 {
		mode = 0o755
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// CopyTree 复制目录。skipBuild 为 true 时跳过 .git、build、target 等目录。
func CopyTree(src, dst string, skipBuild bool) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if d.IsDir() && rel != "." && skipBuild && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(p, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if mode&0o200 == 0 {
		mode |= 0o200
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// RemoveAll 删除目录。git 在 Windows 上把对象文件设为只读，先去掉只读再删。
func RemoveAll(p string) error {
	if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	_ = filepath.WalkDir(p, func(q string, d fs.DirEntry, err error) error {
		if err == nil {
			if d.IsDir() {
				_ = os.Chmod(q, 0o755)
			} else {
				_ = os.Chmod(q, 0o644)
			}
		}
		return nil
	})
	return os.RemoveAll(p)
}
