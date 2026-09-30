// Package names 校验项目名、模块名、包名、应用目录等用户输入。
package names

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
)

var (
	projectNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)
	moduleIDRe    = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	packageSegRe  = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
	groupIDRe     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)
	versionRe     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)
	dirSegRe      = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)
)

// ReservedModules 骨架自带的模块名，业务模块不能重名。
var ReservedModules = map[string]bool{
	"yi-admin":  true,
	"yi-common": true,
	"yi-demo":   true,
}

// reservedTopDirs 应用目录不能占用的顶层目录。
var reservedTopDirs = map[string]bool{
	"yi-admin": true, "yi-common": true, "yi-demo": true,
	".star-yi": true, ".git": true, ".gradle": true, ".idea": true,
	"build": true, "target": true, "src": true, "sql": true, "docs": true,
}

// Java 与 Kotlin 共有的硬关键字，不能作为包段。
var keywords = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true,
	"case": true, "catch": true, "char": true, "class": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extends": true, "final": true, "finally": true, "float": true,
	"for": true, "goto": true, "if": true, "implements": true, "import": true,
	"instanceof": true, "int": true, "interface": true, "long": true, "native": true,
	"new": true, "package": true, "private": true, "protected": true, "public": true,
	"return": true, "short": true, "static": true, "strictfp": true, "super": true,
	"switch": true, "synchronized": true, "this": true, "throw": true, "throws": true,
	"transient": true, "try": true, "void": true, "volatile": true, "while": true,
	"true": true, "false": true, "null": true,
	"as": true, "fun": true, "in": true, "is": true, "object": true, "typealias": true,
	"typeof": true, "val": true, "var": true, "when": true,
}

// ValidateProjectName 项目名同时是文件夹名、Maven artifactId、Gradle rootProject.name。
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("项目名不能为空")
	}
	if !projectNameRe.MatchString(name) {
		return fmt.Errorf("项目名 %q 只能用字母开头，后面跟字母、数字、点、下划线或连字符", name)
	}
	return nil
}

// ValidateModuleID 业务模块名同时是目录名、Maven artifactId、Gradle 项目名。
func ValidateModuleID(id string) error {
	if id == "" {
		return fmt.Errorf("模块名不能为空")
	}
	if !moduleIDRe.MatchString(id) {
		return fmt.Errorf("模块名 %q 只能用小写字母开头，后面跟小写字母、数字或连字符，例如 order-service", id)
	}
	if strings.HasSuffix(id, "-") || strings.Contains(id, "--") {
		return fmt.Errorf("模块名 %q 不能以连字符结尾，也不能有连续的连字符", id)
	}
	if ReservedModules[id] {
		return fmt.Errorf("模块名 %q 是骨架自带模块", id)
	}
	return nil
}

// ValidatePackage 校验包名后缀，允许多段，如 order 或 order.core。
func ValidatePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("包名不能为空")
	}
	for _, seg := range strings.Split(pkg, ".") {
		if !packageSegRe.MatchString(seg) {
			return fmt.Errorf("包名 %q 不合法：每一段只能用小写字母或下划线开头，后面跟小写字母、数字或下划线", pkg)
		}
		if keywords[seg] {
			return fmt.Errorf("包名 %q 里的 %q 是 Java/Kotlin 关键字", pkg, seg)
		}
	}
	return nil
}

// DefaultPackage 由模块名推出包名后缀：order-service → orderservice。
func DefaultPackage(moduleID string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(moduleID) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "app" + s
	}
	if keywords[s] {
		s += "app"
	}
	return s
}

// ClassName 由模块名推出类名前缀：order-service → OrderService。
func ClassName(moduleID string) string {
	var b strings.Builder
	upper := true
	for _, r := range moduleID {
		if r == '-' || r == '_' || r == '.' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "App"
	}
	return b.String()
}

// ValidateGroupID 校验 Maven groupId。
func ValidateGroupID(g string) error {
	if !groupIDRe.MatchString(g) {
		return fmt.Errorf("groupId %q 不合法，应类似 com.star", g)
	}
	return nil
}

// ValidateVersion 校验版本号。
func ValidateVersion(v string) error {
	if !versionRe.MatchString(v) {
		return fmt.Errorf("版本号 %q 不合法，应类似 0.0.1-SNAPSHOT", v)
	}
	return nil
}

// ValidateDisplayName 展示名只要求非空、单行。
func ValidateDisplayName(n string) error {
	if strings.TrimSpace(n) == "" {
		return fmt.Errorf("展示名不能为空")
	}
	if strings.ContainsAny(n, "\r\n") {
		return fmt.Errorf("展示名不能换行")
	}
	return nil
}

// CleanAppsDir 校验并规范应用目录，返回用 / 分隔的相对路径。
func CleanAppsDir(dir string) (string, error) {
	d := strings.TrimSpace(strings.ReplaceAll(dir, `\`, "/"))
	if d == "" {
		return "", fmt.Errorf("应用目录不能为空")
	}
	if strings.HasPrefix(d, "/") || (len(d) > 1 && d[1] == ':') {
		return "", fmt.Errorf("应用目录 %q 必须是相对项目根目录的路径", dir)
	}
	d = path.Clean(d)
	if d == "." {
		return "", fmt.Errorf("应用目录不能是项目根目录本身")
	}
	segs := strings.Split(d, "/")
	for _, s := range segs {
		if s == ".." {
			return "", fmt.Errorf("应用目录 %q 不能跳出项目根目录", dir)
		}
		if !dirSegRe.MatchString(s) {
			return "", fmt.Errorf("应用目录 %q 里的 %q 只能用字母、数字、点、下划线或连字符", dir, s)
		}
	}
	if reservedTopDirs[segs[0]] {
		return "", fmt.Errorf("应用目录不能放在 %q 下", segs[0])
	}
	return d, nil
}
