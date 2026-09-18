package template

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// binaryExtensions 视为二进制、不做占位符替换的扩展名。
var binaryExtensions = map[string]bool{
	".jar": true, ".class": true, ".zip": true, ".gz": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
	".pdf": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
}

// IsBinaryPath 根据扩展名判断是否按二进制复制。
func IsBinaryPath(path string) bool {
	return binaryExtensions[strings.ToLower(filepath.Ext(path))]
}

// ProcessFile 将源文件复制到目标，文本文件执行占位符替换。
func ProcessFile(src, dst string, vars Vars) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if IsBinaryPath(src) {
		return copyFileRaw(src, dst)
	}
	return processTextFile(src, dst, vars.Pairs())
}

func copyFileRaw(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// processTextFile 按行流式读取并替换占位符，避免大文件一次性读入内存。
func processTextFile(src, dst string, pairs map[string]string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	const maxLine = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxLine)

	w := bufio.NewWriter(out)
	for scanner.Scan() {
		line := ReplaceLine(scanner.Text(), pairs)
		if _, err := w.WriteString(line); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		// 含 NUL 等不可按行扫描时，回退为整块替换
		return processWholeFile(src, dst, pairs)
	}
	return w.Flush()
}

func processWholeFile(src, dst string, pairs map[string]string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return copyFileRaw(src, dst)
	}
	return os.WriteFile(dst, ReplaceBytes(data, pairs), 0o644)
}
