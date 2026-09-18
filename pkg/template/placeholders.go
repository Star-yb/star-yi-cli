package template

import "strings"

// Vars 模板占位符替换变量。
type Vars struct {
	ProjectArtifactID string
	ProjectName       string
	GroupID           string
	Version           string
	AppsModulePrefix  string
	DefaultAPIPrefix  string
	SQLSubDir         string
}

// Pairs 返回占位符 → 值的映射（含花括号键名）。
func (v Vars) Pairs() map[string]string {
	return map[string]string{
		"{projectArtifactId}": v.ProjectArtifactID,
		"{projectName}":       v.ProjectName,
		"{groupId}":           v.GroupID,
		"{version}":           v.Version,
		"{appsModulePrefix}":  v.AppsModulePrefix,
		"{defaultApiPrefix}":  v.DefaultAPIPrefix,
		"{sqlSubDir}":         v.SQLSubDir,
	}
}

// ReplaceLine 对单行文本执行占位符替换。
func ReplaceLine(line string, pairs map[string]string) string {
	for k, val := range pairs {
		line = strings.ReplaceAll(line, k, val)
	}
	return line
}

// ReplaceBytes 对整块文本执行占位符替换。
func ReplaceBytes(data []byte, pairs map[string]string) []byte {
	s := string(data)
	for k, val := range pairs {
		s = strings.ReplaceAll(s, k, val)
	}
	return []byte(s)
}
