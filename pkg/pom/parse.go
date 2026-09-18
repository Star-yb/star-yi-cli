package pom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

// Coordinates 父 POM 坐标。
type Coordinates struct {
	GroupID    string
	ArtifactID string
	Version    string
}

// LoadCoordinates 从项目根 pom.xml 读取 groupId、artifactId、version。
func LoadCoordinates(projectRoot string) (Coordinates, error) {
	pomPath := filepath.Join(projectRoot, "pom.xml")
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return Coordinates{}, fmt.Errorf("读取 pom.xml 失败: %w", err)
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return Coordinates{}, fmt.Errorf("解析 pom.xml 失败: %w", err)
	}
	root := doc.SelectElement("project")
	if root == nil {
		return Coordinates{}, fmt.Errorf("pom.xml 缺少 <project> 根节点")
	}
	c := Coordinates{
		GroupID:    textOf(root, "groupId"),
		ArtifactID: textOf(root, "artifactId"),
		Version:    textOf(root, "version"),
	}
	if c.GroupID == "" {
		if parent := root.SelectElement("parent"); parent != nil {
			c.GroupID = textOf(parent, "groupId")
		}
	}
	if c.Version == "" {
		if parent := root.SelectElement("parent"); parent != nil {
			c.Version = textOf(parent, "version")
		}
	}
	if c.ArtifactID == "" {
		return Coordinates{}, fmt.Errorf("无法从 pom.xml 解析 artifactId")
	}
	if c.GroupID == "" {
		c.GroupID = "com.star"
	}
	if c.Version == "" {
		c.Version = "0.0.1-SNAPSHOT"
	}
	return c, nil
}

func textOf(parent *etree.Element, tag string) string {
	el := parent.SelectElement(tag)
	if el == nil {
		return ""
	}
	return strings.TrimSpace(el.Text())
}

// ValidateStarYiProject 校验是否为 Star-Yi 根目录。
func ValidateStarYiProject(projectRoot string) error {
	pom := filepath.Join(projectRoot, "pom.xml")
	if _, err := os.Stat(pom); err != nil {
		return fmt.Errorf("找不到 %s，请确认 --project 指向 Star-Yi 项目根目录", pom)
	}
	admin := filepath.Join(projectRoot, "yi-admin", "pom.xml")
	if _, err := os.Stat(admin); err != nil {
		return fmt.Errorf("找不到 yi-admin/pom.xml，请确认当前目录是 Star-Yi 项目根目录")
	}
	common := filepath.Join(projectRoot, "yi-common", "pom.xml")
	if _, err := os.Stat(common); err != nil {
		return fmt.Errorf("找不到 yi-common/pom.xml，请确认当前目录是 Star-Yi 项目根目录")
	}
	return nil
}
