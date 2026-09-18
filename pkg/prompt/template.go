package prompt

import (
	"os"
	"path/filepath"
)

// DefaultTemplatePath 查找内置模板：exe 旁 templates、当前目录 star-yi-cli/templates。
func DefaultTemplatePath() string {
	candidates := []string{
		`templates\Star-Yi.zip`,
		`templates\Star-Yi`,
		`star-yi-cli\templates\Star-Yi.zip`,
		`star-yi-cli\templates\Star-Yi`,
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append([]string{
			filepath.Join(dir, "templates", "Star-Yi.zip"),
			filepath.Join(dir, "templates", "Star-Yi"),
		}, candidates...)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}
