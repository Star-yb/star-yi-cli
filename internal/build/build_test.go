package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/source"
	"github.com/star/star-yi-cli/internal/xmledit"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "proj")
	if err := source.CopyTree(filepath.Join("testdata", name), dst, false); err != nil {
		t.Fatal(err)
	}
	return dst
}

func project(build, lang string) *meta.Project {
	return &meta.Project{
		Build: build, Language: lang, AppsDir: "apps", BasePackage: "com.star",
		AdminModule: "yi-admin", CommonModule: "yi-common",
	}
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func snapshot(t *testing.T, root string, files ...string) map[string]string {
	m := map[string]string{}
	for _, f := range files {
		m[f] = read(t, filepath.Join(root, f))
	}
	return m
}

func TestPOMRoundTripIsByteIdentical(t *testing.T) {
	for _, f := range []string{"pom.xml", "yi-admin/pom.xml", "yi-common/pom.xml", "yi-demo/pom.xml"} {
		p := filepath.Join("testdata", "maven", f)
		orig, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := xmledit.Parse(p, orig)
		if err != nil {
			t.Fatal(err)
		}
		out, err := doc.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != string(orig) {
			t.Errorf("%s 原样读写后内容变了", f)
		}
	}
}

func TestMavenRename(t *testing.T) {
	root := fixture(t, "maven")
	m := Maven{}
	p := project(config.BuildMaven, config.LangJava)
	if err := m.Validate(root, p); err != nil {
		t.Fatal(err)
	}
	before, _ := m.Info(root)
	if before.Name != "Star-Yi" || before.GroupID != "com.star" {
		t.Fatalf("骨架坐标不对: %+v", before)
	}
	err := m.Rename(root, p, RenameSpec{Name: "My-Platform", DisplayName: "我的平台", GroupID: "com.acme", Version: "1.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	info, _ := m.Info(root)
	if info != (Info{Name: "My-Platform", DisplayName: "我的平台", GroupID: "com.acme", Version: "1.2.0"}) {
		t.Fatalf("改名后坐标: %+v", info)
	}
	rootPom := read(t, filepath.Join(root, "pom.xml"))
	if !strings.Contains(rootPom, "<artifactId>spring-boot-starter-parent</artifactId>\n        <version>4.0.3</version>") {
		t.Error("Spring Boot 父工程不应被改动")
	}
	if strings.Contains(rootPom, "<groupId>com.star</groupId>") {
		t.Error("根 pom 里还有旧 groupId")
	}
	for _, mod := range []string{"yi-admin", "yi-common", "yi-demo"} {
		s := read(t, filepath.Join(root, mod, "pom.xml"))
		if !strings.Contains(s, "<artifactId>My-Platform</artifactId>") {
			t.Errorf("%s 的 parent 没改", mod)
		}
		if strings.Contains(s, "<groupId>com.star</groupId>") || strings.Contains(s, "0.0.1-SNAPSHOT") {
			t.Errorf("%s 还有旧坐标:\n%s", mod, s)
		}
	}
	admin := read(t, filepath.Join(root, "yi-admin", "pom.xml"))
	if !strings.Contains(admin, "<groupId>org.springframework.boot</groupId>") {
		t.Error("第三方依赖被误改")
	}
}

func TestMavenRegisterUnregisterRestoresFiles(t *testing.T) {
	root := fixture(t, "maven")
	m := Maven{}
	p := project(config.BuildMaven, config.LangJava)
	info, _ := m.Info(root)
	before := snapshot(t, root, "pom.xml", "yi-admin/pom.xml")

	if err := os.MkdirAll(filepath.Join(root, "apps", "order-service"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := m.WriteModuleBuild(root, p, ModuleSpec{ID: "order-service"}, info); err != nil {
		t.Fatal(err)
	}
	modPom := read(t, filepath.Join(root, "apps", "order-service", "pom.xml"))
	if !strings.Contains(modPom, "<relativePath>../../pom.xml</relativePath>") {
		t.Errorf("relativePath 不对:\n%s", modPom)
	}
	if err := m.Register(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	rootPom := read(t, filepath.Join(root, "pom.xml"))
	if strings.Count(rootPom, "<module>apps/order-service</module>") != 1 {
		t.Errorf("module 应登记一次:\n%s", rootPom)
	}
	if !strings.Contains(rootPom, "        <module>yi-common</module>\n        <module>apps/order-service</module>\n    </modules>") {
		t.Errorf("module 缩进不对:\n%s", rootPom)
	}
	if !strings.Contains(rootPom, "            <dependency>\n                <groupId>com.star</groupId>\n                <artifactId>order-service</artifactId>\n                <version>0.0.1-SNAPSHOT</version>\n            </dependency>") {
		t.Errorf("dependencyManagement 格式不对:\n%s", rootPom)
	}
	admin := read(t, filepath.Join(root, "yi-admin", "pom.xml"))
	if strings.Count(admin, "<artifactId>order-service</artifactId>") != 1 {
		t.Errorf("yi-admin 依赖应加一次:\n%s", admin)
	}
	mods, _ := m.Modules(root, p)
	if len(mods) != 1 || mods[0] != "order-service" {
		t.Errorf("Modules = %v", mods)
	}

	if err := m.Unregister(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	for f, want := range before {
		if got := read(t, filepath.Join(root, f)); got != want {
			t.Errorf("注销后 %s 和原文件不一致:\n%s", f, got)
		}
	}
}

func TestMavenNestedAppsDir(t *testing.T) {
	root := fixture(t, "maven")
	m := Maven{}
	p := project(config.BuildMaven, config.LangJava)
	p.AppsDir = "modules/biz"
	info, _ := m.Info(root)
	dir := filepath.Join(root, "modules", "biz", "pay")
	_ = os.MkdirAll(dir, 0o755)
	if err := m.WriteModuleBuild(root, p, ModuleSpec{ID: "pay"}, info); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, filepath.Join(dir, "pom.xml")), "<relativePath>../../../pom.xml</relativePath>") {
		t.Error("嵌套目录的 relativePath 不对")
	}
	if err := m.Register(root, p, "pay", info); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, filepath.Join(root, "pom.xml")), "<module>modules/biz/pay</module>") {
		t.Error("嵌套目录没登记")
	}
}

func TestMavenDependents(t *testing.T) {
	root := fixture(t, "maven")
	m := Maven{}
	p := project(config.BuildMaven, config.LangJava)
	info, _ := m.Info(root)
	for _, s := range []ModuleSpec{{ID: "user-center"}, {ID: "order", Depends: []string{"user-center"}}} {
		_ = os.MkdirAll(filepath.Join(root, "apps", s.ID), 0o755)
		if err := m.WriteModuleBuild(root, p, s, info); err != nil {
			t.Fatal(err)
		}
		if err := m.Register(root, p, s.ID, info); err != nil {
			t.Fatal(err)
		}
	}
	deps, err := m.Dependents(root, p, "user-center", info)
	if err != nil || len(deps) != 1 || deps[0] != "order" {
		t.Fatalf("Dependents = %v, %v", deps, err)
	}
}

func TestGradleRename(t *testing.T) {
	root := fixture(t, "gradle")
	g := Gradle{}
	p := project(config.BuildGradle, config.LangKotlin)
	if err := g.Validate(root, p); err != nil {
		t.Fatal(err)
	}
	before, _ := g.Info(root)
	if before.Name != "Star-Yi-Arc" || before.GroupID != "com.star" {
		t.Fatalf("骨架坐标不对: %+v", before)
	}
	if err := g.Rename(root, p, RenameSpec{Name: "my-arc", DisplayName: `我的 "Arc" $平台`, GroupID: "com.acme", Version: "2.0.0"}); err != nil {
		t.Fatal(err)
	}
	info, _ := g.Info(root)
	if info.Name != "my-arc" || info.GroupID != "com.acme" || info.Version != "2.0.0" {
		t.Fatalf("改名后坐标: %+v", info)
	}
	bf := read(t, filepath.Join(root, "build.gradle.kts"))
	if !strings.Contains(bf, `description = "我的 \"Arc\" \$平台"`) {
		t.Errorf("description 没转义:\n%s", bf)
	}
	if !strings.Contains(bf, `version "4.1.1" apply false`) {
		t.Error("插件版本被误改")
	}
	if strings.Count(bf, `version = "2.0.0"`) != 1 {
		t.Error("项目版本应只改一处")
	}
}

func TestGradleRegisterUnregisterRestoresFiles(t *testing.T) {
	root := fixture(t, "gradle")
	g := Gradle{}
	p := project(config.BuildGradle, config.LangKotlin)
	info, _ := g.Info(root)
	before := snapshot(t, root, "settings.gradle.kts", "yi-admin/build.gradle.kts")

	if err := g.Register(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	if err := g.Register(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	settings := read(t, filepath.Join(root, "settings.gradle.kts"))
	if strings.Count(settings, `include("order-service")`) != 1 ||
		!strings.Contains(settings, `project(":order-service").projectDir = file("apps/order-service")`) {
		t.Errorf("settings.gradle.kts:\n%s", settings)
	}
	admin := read(t, filepath.Join(root, "yi-admin", "build.gradle.kts"))
	want := "    \"implementation\"(project(\":yi-demo\"))\n    \"implementation\"(project(\":order-service\"))\n"
	if !strings.Contains(admin, want) {
		t.Errorf("yi-admin 依赖位置不对:\n%s", admin)
	}
	mods, _ := g.Modules(root, p)
	if len(mods) != 1 || mods[0] != "order-service" {
		t.Errorf("Modules = %v", mods)
	}

	if err := g.Unregister(root, p, "order-service", info); err != nil {
		t.Fatal(err)
	}
	for f, want := range before {
		got := read(t, filepath.Join(root, f))
		if strings.TrimRight(got, "\n") != strings.TrimRight(want, "\n") {
			t.Errorf("注销后 %s 和原文件不一致:\n%s", f, got)
		}
	}
}

func TestInsertProjectDependencyWithoutProjectDeps(t *testing.T) {
	src := "plugins {\n    id(\"x\")\n}\n\ndependencies {\n    \"implementation\"(\"a:b:1\")\n}\n"
	got := insertProjectDependency(src, "pay")
	if !strings.Contains(got, "dependencies {\n    \"implementation\"(project(\":pay\"))\n    \"implementation\"(\"a:b:1\")") {
		t.Errorf("got:\n%s", got)
	}
	got = insertProjectDependency("plugins {\n}\n", "pay")
	if !strings.HasSuffix(got, "dependencies {\n    \"implementation\"(project(\":pay\"))\n}\n") {
		t.Errorf("没有 dependencies 块时应新建:\n%s", got)
	}
}
