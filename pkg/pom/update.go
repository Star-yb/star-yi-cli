package pom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

// RegisterAppModule 在父 pom 注册 apps 模块，并在 yi-admin 添加依赖。
func RegisterAppModule(projectRoot string, moduleID string, coords Coordinates) error {
	parentPom := filepath.Join(projectRoot, "pom.xml")
	if err := addModuleToParent(parentPom, moduleID, coords); err != nil {
		return fmt.Errorf("更新父 pom.xml: %w", err)
	}
	adminPom := filepath.Join(projectRoot, "yi-admin", "pom.xml")
	if err := addDependencyToAdmin(adminPom, moduleID, coords); err != nil {
		return fmt.Errorf("更新 yi-admin/pom.xml: %w", err)
	}
	return nil
}

// UnregisterAppModule 从父 pom 与 yi-admin 中移除模块引用。
func UnregisterAppModule(projectRoot string, moduleID string) error {
	parentPom := filepath.Join(projectRoot, "pom.xml")
	if err := removeModuleFromParent(parentPom, moduleID); err != nil {
		return fmt.Errorf("更新父 pom.xml: %w", err)
	}
	adminPom := filepath.Join(projectRoot, "yi-admin", "pom.xml")
	if err := removeDependencyFromAdmin(adminPom, moduleID); err != nil {
		return fmt.Errorf("更新 yi-admin/pom.xml: %w", err)
	}
	return nil
}

func addModuleToParent(pomPath, moduleID string, coords Coordinates) error {
	doc, err := readPOM(pomPath)
	if err != nil {
		return err
	}
	project := doc.Root()

	modules := project.SelectElement("modules")
	if modules == nil {
		modules = project.CreateElement("modules")
	}
	modulePath := "apps/" + moduleID
	hasModule := false
	for _, m := range modules.SelectElements("module") {
		if strings.TrimSpace(m.Text()) == modulePath {
			hasModule = true
			break
		}
	}
	if !hasModule {
		mod := modules.CreateElement("module")
		mod.SetText(modulePath)
	}

	dm := project.SelectElement("dependencyManagement")
	if dm == nil {
		dm = project.CreateElement("dependencyManagement")
	}
	deps := dm.SelectElement("dependencies")
	if deps == nil {
		deps = dm.CreateElement("dependencies")
	}
	for _, d := range deps.SelectElements("dependency") {
		if textOf(d, "artifactId") == moduleID {
			return writePOM(pomPath, doc)
		}
	}
	dep := deps.CreateElement("dependency")
	dep.CreateElement("groupId").SetText(coords.GroupID)
	dep.CreateElement("artifactId").SetText(moduleID)
	dep.CreateElement("version").SetText(coords.Version)

	return writePOM(pomPath, doc)
}

func addDependencyToAdmin(pomPath, moduleID string, coords Coordinates) error {
	doc, err := readPOM(pomPath)
	if err != nil {
		return err
	}
	project := doc.Root()
	deps := project.SelectElement("dependencies")
	if deps == nil {
		deps = project.CreateElement("dependencies")
	}
	for _, d := range deps.SelectElements("dependency") {
		if textOf(d, "artifactId") == moduleID {
			return nil
		}
	}
	dep := deps.CreateElement("dependency")
	dep.CreateElement("groupId").SetText(coords.GroupID)
	dep.CreateElement("artifactId").SetText(moduleID)
	dep.CreateElement("version").SetText(coords.Version)
	scope := dep.CreateElement("scope")
	scope.SetText("compile")
	return writePOM(pomPath, doc)
}

func removeModuleFromParent(pomPath, moduleID string) error {
	doc, err := readPOM(pomPath)
	if err != nil {
		return err
	}
	project := doc.Root()
	modulePath := "apps/" + moduleID

	if modules := project.SelectElement("modules"); modules != nil {
		for _, m := range modules.SelectElements("module") {
			if strings.TrimSpace(m.Text()) == modulePath {
				modules.RemoveChild(m)
			}
		}
	}

	if dm := project.SelectElement("dependencyManagement"); dm != nil {
		if deps := dm.SelectElement("dependencies"); deps != nil {
			for _, d := range deps.SelectElements("dependency") {
				if textOf(d, "artifactId") == moduleID {
					deps.RemoveChild(d)
				}
			}
		}
	}

	return writePOM(pomPath, doc)
}

func removeDependencyFromAdmin(pomPath, moduleID string) error {
	doc, err := readPOM(pomPath)
	if err != nil {
		return err
	}
	project := doc.Root()
	deps := project.SelectElement("dependencies")
	if deps == nil {
		return nil
	}
	for _, d := range deps.SelectElements("dependency") {
		if textOf(d, "artifactId") == moduleID {
			deps.RemoveChild(d)
		}
	}
	return writePOM(pomPath, doc)
}

func readPOM(path string) (*etree.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	return doc, nil
}

func writePOM(path string, doc *etree.Document) error {
	doc.Indent(4)
	data, err := doc.WriteToBytes()
	if err != nil {
		return err
	}
	// 保留 XML 声明
	if !strings.HasPrefix(string(data), "<?xml") {
		data = append([]byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"), data...)
	}
	return os.WriteFile(path, data, 0o644)
}
