package app

import (
	"fmt"
	"strings"

	"github.com/star/star-yi-cli/pkg/pom"
)

func renderModulePOM(coords pom.Coordinates, moduleID string, depends []string) string {
	var depBlocks strings.Builder
	depBlocks.WriteString(fmt.Sprintf(`        <dependency>
            <groupId>%s</groupId>
            <artifactId>yi-common</artifactId>
        </dependency>
`, coords.GroupID))

	for _, d := range depends {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		depBlocks.WriteString(fmt.Sprintf(`        <dependency>
            <groupId>%s</groupId>
            <artifactId>%s</artifactId>
        </dependency>
`, coords.GroupID, d))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>%s</groupId>
        <artifactId>%s</artifactId>
        <version>%s</version>
        <relativePath>../../pom.xml</relativePath>
    </parent>

    <artifactId>%s</artifactId>
    <packaging>jar</packaging>
    <name>%s</name>

    <dependencies>
%s    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
                <configuration>
                    <classifier>exec</classifier>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
`, coords.GroupID, coords.ArtifactID, coords.Version, moduleID, moduleID, depBlocks.String())
}
