package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/xmledit"
)

// Maven 对应 Star-Yi：父 POM 登记 module 与 dependencyManagement，yi-admin 的 pom.xml 加依赖。
type Maven struct{}

func (Maven) Name() string      { return "maven" }
func (Maven) BuildFile() string { return "pom.xml" }

func (Maven) Validate(root string, p *meta.Project) error {
	for _, rel := range []string{
		"pom.xml",
		filepath.Join(p.AdminModule, "pom.xml"),
		filepath.Join(p.CommonModule, "pom.xml"),
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			return fmt.Errorf("找不到 %s，这不是 Maven 版的 Star-Yi 骨架", rel)
		}
	}
	return nil
}

func (Maven) Info(root string) (Info, error) {
	doc, err := xmledit.Read(filepath.Join(root, "pom.xml"))
	if err != nil {
		return Info{}, err
	}
	return mavenInfo(doc.Root()), nil
}

func mavenInfo(project *etree.Element) Info {
	parent := project.SelectElement("parent")
	info := Info{
		Name:        xmledit.Text(project, "artifactId"),
		DisplayName: xmledit.Text(project, "name"),
		GroupID:     xmledit.Text(project, "groupId"),
		Version:     xmledit.Text(project, "version"),
	}
	if info.GroupID == "" {
		info.GroupID = xmledit.Text(parent, "groupId")
	}
	if info.Version == "" {
		info.Version = xmledit.Text(parent, "version")
	}
	if info.DisplayName == "" {
		info.DisplayName = info.Name
	}
	return info
}

func (m Maven) Rename(root string, p *meta.Project, spec RenameSpec) error {
	rootPom, err := xmledit.Read(filepath.Join(root, "pom.xml"))
	if err != nil {
		return err
	}
	project := rootPom.Root()
	old := mavenInfo(project)
	if old.Name == "" {
		return fmt.Errorf("根 pom.xml 没有 artifactId")
	}

	docs := []*xmledit.Doc{rootPom}
	internal := map[string]bool{old.Name: true}
	if mods := project.SelectElement("modules"); mods != nil {
		for _, mod := range mods.SelectElements("module") {
			rel := filepath.FromSlash(strings.TrimSpace(mod.Text()))
			d, err := xmledit.Read(filepath.Join(root, rel, "pom.xml"))
			if err != nil {
				return fmt.Errorf("读取模块 %s: %w", rel, err)
			}
			docs = append(docs, d)
			if id := xmledit.Text(d.Root(), "artifactId"); id != "" {
				internal[id] = true
			}
		}
	}

	xmledit.SetText(project, "artifactId", spec.Name)
	setOrKeep(project, "groupId", spec.GroupID)
	setOrKeep(project, "version", spec.Version)
	xmledit.SetText(project, "name", spec.DisplayName)
	if desc := xmledit.Text(project, "description"); desc == old.Name || desc == old.DisplayName {
		xmledit.SetText(project, "description", spec.DisplayName)
	}

	for _, d := range docs {
		pr := d.Root()
		if parent := pr.SelectElement("parent"); parent != nil &&
			xmledit.Text(parent, "groupId") == old.GroupID && xmledit.Text(parent, "artifactId") == old.Name {
			xmledit.SetText(parent, "groupId", spec.GroupID)
			xmledit.SetText(parent, "artifactId", spec.Name)
			xmledit.SetText(parent, "version", spec.Version)
		}
		if d != rootPom && xmledit.Text(pr, "groupId") == old.GroupID {
			xmledit.SetText(pr, "groupId", spec.GroupID)
		}
		for _, dep := range allDependencies(pr) {
			if xmledit.Text(dep, "groupId") != old.GroupID || !internal[xmledit.Text(dep, "artifactId")] {
				continue
			}
			xmledit.SetText(dep, "groupId", spec.GroupID)
			if xmledit.Text(dep, "version") == old.Version {
				xmledit.SetText(dep, "version", spec.Version)
			}
		}
	}
	for _, d := range docs {
		if err := d.Save(); err != nil {
			return err
		}
	}
	return nil
}

func setOrKeep(e *etree.Element, tag, value string) {
	if value != "" {
		xmledit.SetText(e, tag, value)
	}
}

func allDependencies(project *etree.Element) []*etree.Element {
	var out []*etree.Element
	if deps := project.SelectElement("dependencies"); deps != nil {
		out = append(out, deps.SelectElements("dependency")...)
	}
	if dm := project.SelectElement("dependencyManagement"); dm != nil {
		if deps := dm.SelectElement("dependencies"); deps != nil {
			out = append(out, deps.SelectElements("dependency")...)
		}
	}
	return out
}

func modulePath(p *meta.Project, id string) string {
	return path.Join(p.AppsDir, id)
}

func normModule(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, `\`, "/"))
	return strings.TrimSuffix(path.Clean(s), "/")
}

func (Maven) WriteModuleBuild(root string, p *meta.Project, spec ModuleSpec, info Info) error {
	dir := filepath.Join(root, filepath.FromSlash(modulePath(p, spec.ID)))
	rel, err := filepath.Rel(dir, root)
	if err != nil {
		return err
	}
	relPom := filepath.ToSlash(rel) + "/pom.xml"

	var deps strings.Builder
	for _, d := range append([]string{p.CommonModule}, spec.Depends...) {
		fmt.Fprintf(&deps, `        <dependency>
            <groupId>%s</groupId>
            <artifactId>%s</artifactId>
            <version>%s</version>
            <scope>compile</scope>
        </dependency>
`, xmledit.Escape(info.GroupID), xmledit.Escape(d), xmledit.Escape(info.Version))
	}

	pom := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>%[1]s</groupId>
        <artifactId>%[2]s</artifactId>
        <version>%[3]s</version>
        <relativePath>%[4]s</relativePath>
    </parent>

    <artifactId>%[5]s</artifactId>
    <packaging>jar</packaging>
    <name>%[5]s</name>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>

    <dependencies>
%[6]s    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
                <configuration>
                    <!-- 普通 jar 留给 yi-admin 依赖，可执行 jar 另带 exec 后缀 -->
                    <classifier>exec</classifier>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
`, xmledit.Escape(info.GroupID), xmledit.Escape(info.Name), xmledit.Escape(info.Version), relPom,
		xmledit.Escape(spec.ID), deps.String())
	return os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(pom), 0o644)
}

func depFragment(group, id, version string, withScope bool) string {
	scope := ""
	if withScope {
		scope = "\n    <scope>compile</scope>"
	}
	return fmt.Sprintf("<dependency>\n    <groupId>%s</groupId>\n    <artifactId>%s</artifactId>\n    <version>%s</version>%s\n</dependency>",
		xmledit.Escape(group), xmledit.Escape(id), xmledit.Escape(version), scope)
}

// lastInternal 最后一个同 groupId 的依赖，新依赖放在它后面，和骨架模块挨在一起。
func lastInternal(deps *etree.Element, group string) *etree.Element {
	var last *etree.Element
	for _, d := range deps.SelectElements("dependency") {
		if xmledit.Text(d, "groupId") == group {
			last = d
		}
	}
	return last
}

func findDep(deps *etree.Element, group, id string) *etree.Element {
	if deps == nil {
		return nil
	}
	for _, d := range deps.SelectElements("dependency") {
		if xmledit.Text(d, "artifactId") == id && (xmledit.Text(d, "groupId") == group || xmledit.Text(d, "groupId") == "") {
			return d
		}
	}
	return nil
}

func (Maven) Register(root string, p *meta.Project, id string, info Info) error {
	rootPom, err := xmledit.Read(filepath.Join(root, "pom.xml"))
	if err != nil {
		return err
	}
	project := rootPom.Root()

	mods, err := xmledit.Ensure(project, "modules")
	if err != nil {
		return err
	}
	mp := modulePath(p, id)
	exists := false
	var lastMod *etree.Element
	for _, m := range mods.SelectElements("module") {
		lastMod = m
		if normModule(m.Text()) == mp {
			exists = true
		}
	}
	if !exists {
		if _, err := xmledit.Append(mods, "<module>"+xmledit.Escape(mp)+"</module>", lastMod); err != nil {
			return err
		}
	}

	dm, err := xmledit.Ensure(project, "dependencyManagement")
	if err != nil {
		return err
	}
	dmDeps, err := xmledit.Ensure(dm, "dependencies")
	if err != nil {
		return err
	}
	if findDep(dmDeps, info.GroupID, id) == nil {
		frag := depFragment(info.GroupID, id, info.Version, false)
		if _, err := xmledit.Append(dmDeps, frag, lastInternal(dmDeps, info.GroupID)); err != nil {
			return err
		}
	}

	adminPom, err := xmledit.Read(filepath.Join(root, p.AdminModule, "pom.xml"))
	if err != nil {
		return err
	}
	adminDeps, err := xmledit.Ensure(adminPom.Root(), "dependencies")
	if err != nil {
		return err
	}
	if findDep(adminDeps, info.GroupID, id) == nil {
		frag := depFragment(info.GroupID, id, info.Version, true)
		if _, err := xmledit.Append(adminDeps, frag, lastInternal(adminDeps, info.GroupID)); err != nil {
			return err
		}
	}

	if err := rootPom.Save(); err != nil {
		return err
	}
	return adminPom.Save()
}

func (Maven) Unregister(root string, p *meta.Project, id string, info Info) error {
	rootPom, err := xmledit.Read(filepath.Join(root, "pom.xml"))
	if err != nil {
		return err
	}
	project := rootPom.Root()
	mp := modulePath(p, id)
	if mods := project.SelectElement("modules"); mods != nil {
		for _, m := range mods.SelectElements("module") {
			if normModule(m.Text()) == mp {
				xmledit.Remove(m)
			}
		}
	}
	if dm := project.SelectElement("dependencyManagement"); dm != nil {
		deps := dm.SelectElement("dependencies")
		for d := findDep(deps, info.GroupID, id); d != nil; d = findDep(deps, info.GroupID, id) {
			xmledit.Remove(d)
		}
	}

	adminPom, err := xmledit.Read(filepath.Join(root, p.AdminModule, "pom.xml"))
	if err != nil {
		return err
	}
	deps := adminPom.Root().SelectElement("dependencies")
	for d := findDep(deps, info.GroupID, id); d != nil; d = findDep(deps, info.GroupID, id) {
		xmledit.Remove(d)
	}

	if err := rootPom.Save(); err != nil {
		return err
	}
	return adminPom.Save()
}

func (Maven) Modules(root string, p *meta.Project) ([]string, error) {
	doc, err := xmledit.Read(filepath.Join(root, "pom.xml"))
	if err != nil {
		return nil, err
	}
	var out []string
	prefix := p.AppsDir + "/"
	if mods := doc.Root().SelectElement("modules"); mods != nil {
		for _, m := range mods.SelectElements("module") {
			mp := normModule(m.Text())
			if rest, ok := strings.CutPrefix(mp, prefix); ok && rest != "" && !strings.Contains(rest, "/") {
				out = append(out, rest)
			}
		}
	}
	return out, nil
}

func (m Maven) Dependents(root string, p *meta.Project, id string, info Info) ([]string, error) {
	mods, err := m.Modules(root, p)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, other := range mods {
		if other == id {
			continue
		}
		doc, err := xmledit.Read(filepath.Join(root, filepath.FromSlash(modulePath(p, other)), "pom.xml"))
		if err != nil {
			continue
		}
		if findDep(doc.Root().SelectElement("dependencies"), info.GroupID, id) != nil {
			out = append(out, other)
		}
	}
	return out, nil
}

func (Maven) VerifyCommand(p *meta.Project, id string) (string, []string) {
	return "mvn", []string{"-q", "compile", "-pl", modulePath(p, id), "-am"}
}
