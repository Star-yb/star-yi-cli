package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/meta"
)

// Gradle 对应 Star-Yi Arc：settings.gradle.kts 登记 include 和目录，yi-admin 加 project 依赖。
type Gradle struct{}

func (Gradle) Name() string      { return "gradle" }
func (Gradle) BuildFile() string { return "build.gradle.kts" }

var (
	rootNameRe    = regexp.MustCompile(`(?m)^[ \t]*rootProject\.name[ \t]*=[ \t]*"([^"]*)"`)
	descriptionRe = regexp.MustCompile(`(?m)^description[ \t]*=[ \t]*"([^"]*)"`)
	groupRe       = regexp.MustCompile(`(?m)^[ \t]*group[ \t]*=[ \t]*"([^"]*)"`)
	versionRe     = regexp.MustCompile(`(?m)^[ \t]*version[ \t]*=[ \t]*"([^"]*)"`)
	projectDirRe  = regexp.MustCompile(`(?m)^[ \t]*project\([ \t]*":([^"]+)"[ \t]*\)\.projectDir[ \t]*=[ \t]*file\([ \t]*"([^"]*)"[ \t]*\)`)
	depsOpenRe    = regexp.MustCompile(`^[ \t]*dependencies[ \t]*\{`)
)

type textFile struct {
	path string
	text string
	crlf bool
}

func readText(p string) (*textFile, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	s := string(data)
	crlf := strings.Contains(s, "\r\n")
	return &textFile{path: p, text: strings.ReplaceAll(s, "\r\n", "\n"), crlf: crlf}, nil
}

func (f *textFile) save() error {
	s := f.text
	if f.crlf {
		s = strings.ReplaceAll(s, "\n", "\r\n")
	}
	return os.WriteFile(f.path, []byte(s), 0o644)
}

// replaceFirst 只改第一处匹配的引号内的值。
func replaceFirst(re *regexp.Regexp, text, value string) (string, bool) {
	loc := re.FindStringSubmatchIndex(text)
	if loc == nil {
		return text, false
	}
	return text[:loc[2]] + value + text[loc[3]:], true
}

func firstValue(re *regexp.Regexp, text string) string {
	if m := re.FindStringSubmatch(text); m != nil {
		return m[1]
	}
	return ""
}

// kotlinString 转义放进 Kotlin 字符串字面量的值。
func kotlinString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`)
	return r.Replace(s)
}

func (Gradle) Validate(root string, p *meta.Project) error {
	for _, rel := range []string{
		"settings.gradle.kts",
		"build.gradle.kts",
		filepath.Join(p.AdminModule, "build.gradle.kts"),
		filepath.Join(p.CommonModule, "build.gradle.kts"),
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			return fmt.Errorf("找不到 %s，这不是 Gradle 版的 Star-Yi 骨架", rel)
		}
	}
	return nil
}

func (Gradle) Info(root string) (Info, error) {
	settings, err := readText(filepath.Join(root, "settings.gradle.kts"))
	if err != nil {
		return Info{}, err
	}
	bf, err := readText(filepath.Join(root, "build.gradle.kts"))
	if err != nil {
		return Info{}, err
	}
	info := Info{
		Name:        firstValue(rootNameRe, settings.text),
		DisplayName: firstValue(descriptionRe, bf.text),
		GroupID:     firstValue(groupRe, bf.text),
		Version:     firstValue(versionRe, bf.text),
	}
	if info.Name == "" {
		info.Name = filepath.Base(root)
	}
	if info.DisplayName == "" {
		info.DisplayName = info.Name
	}
	return info, nil
}

func (Gradle) Rename(root string, p *meta.Project, spec RenameSpec) error {
	settings, err := readText(filepath.Join(root, "settings.gradle.kts"))
	if err != nil {
		return err
	}
	if t, ok := replaceFirst(rootNameRe, settings.text, kotlinString(spec.Name)); ok {
		settings.text = t
	} else {
		settings.text = fmt.Sprintf("rootProject.name = \"%s\"\n", kotlinString(spec.Name)) + settings.text
	}

	bf, err := readText(filepath.Join(root, "build.gradle.kts"))
	if err != nil {
		return err
	}
	bf.text, _ = replaceFirst(descriptionRe, bf.text, kotlinString(spec.DisplayName))
	if spec.GroupID != "" {
		if t, ok := replaceFirst(groupRe, bf.text, kotlinString(spec.GroupID)); ok {
			bf.text = t
		} else {
			return fmt.Errorf("根 build.gradle.kts 里没有 group = \"...\"，无法改 groupId")
		}
	}
	if spec.Version != "" {
		if t, ok := replaceFirst(versionRe, bf.text, kotlinString(spec.Version)); ok {
			bf.text = t
		} else {
			return fmt.Errorf("根 build.gradle.kts 里没有 version = \"...\"，无法改版本")
		}
	}
	if err := settings.save(); err != nil {
		return err
	}
	return bf.save()
}

func (Gradle) WriteModuleBuild(root string, p *meta.Project, spec ModuleSpec, info Info) error {
	dir := filepath.Join(root, filepath.FromSlash(path.Join(p.AppsDir, spec.ID)))
	var deps strings.Builder
	for _, d := range append([]string{p.CommonModule}, spec.Depends...) {
		fmt.Fprintf(&deps, "    \"implementation\"(project(\":%s\"))\n", d)
	}
	content := fmt.Sprintf("description = \"%s\"\n\ndependencies {\n%s}\n", spec.ID, deps.String())
	return os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte(content), 0o644)
}

func includeLineRe(id string) *regexp.Regexp {
	return regexp.MustCompile(`^[ \t]*include\([ \t]*":?` + regexp.QuoteMeta(id) + `"[ \t]*\)[ \t]*$`)
}

func projectDirLineRe(id string) *regexp.Regexp {
	return regexp.MustCompile(`^[ \t]*project\([ \t]*":` + regexp.QuoteMeta(id) + `"[ \t]*\)\.projectDir[ \t]*=`)
}

func projectDepRe(id string) *regexp.Regexp {
	return regexp.MustCompile(`project\([ \t]*":` + regexp.QuoteMeta(id) + `"[ \t]*\)`)
}

func (Gradle) Register(root string, p *meta.Project, id string, info Info) error {
	settings, err := readText(filepath.Join(root, "settings.gradle.kts"))
	if err != nil {
		return err
	}
	lines := strings.Split(settings.text, "\n")
	hasInclude, hasDir := false, false
	for _, l := range lines {
		hasInclude = hasInclude || includeLineRe(id).MatchString(l)
		hasDir = hasDir || projectDirLineRe(id).MatchString(l)
	}
	var add []string
	if !hasInclude {
		add = append(add, fmt.Sprintf("include(\"%s\")", id))
	}
	if !hasDir {
		add = append(add, fmt.Sprintf("project(\":%s\").projectDir = file(\"%s\")", id, path.Join(p.AppsDir, id)))
	}
	if len(add) > 0 {
		t := strings.TrimRight(settings.text, "\n")
		settings.text = t + "\n" + strings.Join(add, "\n") + "\n"
	}

	adminPath := filepath.Join(root, p.AdminModule, "build.gradle.kts")
	admin, err := readText(adminPath)
	if err != nil {
		return err
	}
	if !projectDepRe(id).MatchString(admin.text) {
		admin.text = insertProjectDependency(admin.text, id)
	}

	if err := settings.save(); err != nil {
		return err
	}
	return admin.save()
}

// insertProjectDependency 在 dependencies 块里最后一个 project(...) 依赖后面加一行。
func insertProjectDependency(text, id string) string {
	line := fmt.Sprintf("\"implementation\"(project(\":%s\"))", id)
	lines := strings.Split(text, "\n")
	open := -1
	for i, l := range lines {
		if depsOpenRe.MatchString(l) {
			open = i
			break
		}
	}
	if open < 0 {
		return strings.TrimRight(text, "\n") + "\n\ndependencies {\n    " + line + "\n}\n"
	}
	depth, end := 0, -1
	for i := open; i < len(lines) && end < 0; i++ {
		for _, r := range stripKotlinStrings(lines[i]) {
			if r == '{' {
				depth++
			} else if r == '}' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}
	}
	if end < 0 {
		end = len(lines) - 1
	}
	after, indent := -1, ""
	for i := open + 1; i < end; i++ {
		if strings.Contains(lines[i], "project(\":") {
			after = i
			indent = lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]
		}
	}
	if after < 0 {
		after = open
		base := lines[open][:len(lines[open])-len(strings.TrimLeft(lines[open], " \t"))]
		indent = base + "    "
	}
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:after+1]...)
	out = append(out, indent+line)
	out = append(out, lines[after+1:]...)
	return strings.Join(out, "\n")
}

// stripKotlinStrings 去掉行内字符串和注释，只留下能数大括号的部分。
func stripKotlinStrings(l string) string {
	if i := strings.Index(l, "//"); i >= 0 {
		l = l[:i]
	}
	var b strings.Builder
	in := false
	for i := 0; i < len(l); i++ {
		c := l[i]
		if c == '\\' && in {
			i++
			continue
		}
		if c == '"' {
			in = !in
			continue
		}
		if !in {
			b.WriteByte(c)
		}
	}
	return b.String()
}

func removeLines(text string, match func(string) bool) (string, bool) {
	lines := strings.Split(text, "\n")
	out := lines[:0]
	changed := false
	for _, l := range lines {
		if match(l) {
			changed = true
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n"), changed
}

func (Gradle) Unregister(root string, p *meta.Project, id string, info Info) error {
	settings, err := readText(filepath.Join(root, "settings.gradle.kts"))
	if err != nil {
		return err
	}
	inc, dir := includeLineRe(id), projectDirLineRe(id)
	settings.text, _ = removeLines(settings.text, func(l string) bool {
		return inc.MatchString(l) || dir.MatchString(l)
	})

	admin, err := readText(filepath.Join(root, p.AdminModule, "build.gradle.kts"))
	if err != nil {
		return err
	}
	dep := projectDepRe(id)
	admin.text, _ = removeLines(admin.text, func(l string) bool {
		return dep.MatchString(l) && !strings.HasPrefix(strings.TrimSpace(l), "//")
	})

	if err := settings.save(); err != nil {
		return err
	}
	return admin.save()
}

func (Gradle) Modules(root string, p *meta.Project) ([]string, error) {
	settings, err := readText(filepath.Join(root, "settings.gradle.kts"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range projectDirRe.FindAllStringSubmatch(settings.text, -1) {
		id, dir := m[1], normModule(m[2])
		if dir == path.Join(p.AppsDir, id) {
			out = append(out, id)
		}
	}
	return out, nil
}

func (g Gradle) Dependents(root string, p *meta.Project, id string, info Info) ([]string, error) {
	mods, err := g.Modules(root, p)
	if err != nil {
		return nil, err
	}
	dep := projectDepRe(id)
	var out []string
	for _, other := range mods {
		if other == id {
			continue
		}
		f, err := readText(filepath.Join(root, filepath.FromSlash(path.Join(p.AppsDir, other)), "build.gradle.kts"))
		if err != nil {
			continue
		}
		if dep.MatchString(f.text) {
			out = append(out, other)
		}
	}
	return out, nil
}

func (Gradle) VerifyCommand(p *meta.Project, id string) (string, []string) {
	task := "compileKotlin"
	if p.Language == config.LangJava {
		task = "compileJava"
	}
	return "gradle", []string{"--no-daemon", "-q", "--console=plain", ":" + id + ":" + task}
}
