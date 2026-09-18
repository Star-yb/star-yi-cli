package pom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleParentPOM = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.star</groupId>
    <artifactId>Star-Yi</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <packaging>pom</packaging>
    <modules>
        <module>yi-common</module>
        <module>yi-admin</module>
    </modules>
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>com.star</groupId>
                <artifactId>yi-common</artifactId>
                <version>0.0.1-SNAPSHOT</version>
            </dependency>
        </dependencies>
    </dependencyManagement>
</project>
`

const sampleAdminPOM = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>com.star</groupId>
        <artifactId>Star-Yi</artifactId>
        <version>0.0.1-SNAPSHOT</version>
    </parent>
    <artifactId>yi-admin</artifactId>
    <dependencies>
        <dependency>
            <groupId>com.star</groupId>
            <artifactId>yi-common</artifactId>
            <version>0.0.1-SNAPSHOT</version>
        </dependency>
    </dependencies>
</project>
`

func TestRegisterAppModule(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "yi-admin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(sampleParentPOM), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "yi-admin", "pom.xml"), []byte(sampleAdminPOM), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "yi-common"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "yi-common", "pom.xml"), []byte("<project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	coords := Coordinates{GroupID: "com.star", ArtifactID: "Star-Yi", Version: "0.0.1-SNAPSHOT"}
	if err := RegisterAppModule(dir, "order-service", coords); err != nil {
		t.Fatal(err)
	}

	parent, err := os.ReadFile(filepath.Join(dir, "pom.xml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(parent)
	if !strings.Contains(s, "<module>apps/order-service</module>") {
		t.Fatalf("parent pom missing module: %s", s)
	}
	if !strings.Contains(s, "<artifactId>order-service</artifactId>") {
		t.Fatalf("parent pom missing dependencyManagement entry: %s", s)
	}

	admin, err := os.ReadFile(filepath.Join(dir, "yi-admin", "pom.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(admin), "order-service") {
		t.Fatalf("admin pom missing dependency: %s", admin)
	}
}

func TestLoadCoordinates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(sampleParentPOM), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadCoordinates(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.ArtifactID != "Star-Yi" || c.GroupID != "com.star" {
		t.Fatalf("unexpected coords: %+v", c)
	}
}
