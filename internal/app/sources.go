package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/meta"
)

// writeSources 生成模块的源码目录、示例接口和说明。
func writeSources(dir string, p *meta.Project, spec Spec) error {
	pkg := p.BasePackage + "." + spec.Package
	pkgPath := filepath.Join(strings.Split(pkg, ".")...)

	var srcRoot string
	var subDirs []string
	if p.Language == config.LangKotlin {
		srcRoot = filepath.Join(dir, "src", "main", "kotlin", pkgPath)
		subDirs = []string{"entity", "controller", "admin"}
	} else {
		srcRoot = filepath.Join(dir, "src", "main", "java", pkgPath)
		subDirs = []string{"model/entity", "dao", "repository", "service/impl", "controller", "admin"}
	}

	keep := []string{filepath.Join(dir, "src", "main", "dto")}
	for _, s := range subDirs {
		if s != "controller" {
			keep = append(keep, filepath.Join(srcRoot, filepath.FromSlash(s)))
		}
	}
	for _, d := range keep {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(d, ".gitkeep"), nil, 0o644); err != nil {
			return err
		}
	}

	ctrlDir := filepath.Join(srcRoot, "controller")
	if err := os.MkdirAll(ctrlDir, 0o755); err != nil {
		return err
	}
	className := spec.ClassPrefix + "PingController"
	var file, content string
	if p.Language == config.LangKotlin {
		file = className + ".kt"
		content = fmt.Sprintf(`package %[1]s.controller

import cn.dev33.satoken.util.SaResult
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

/** 模块接入检查。启动后登录，带上 satoken 请求头访问 GET %[2]s/ping。 */
@RestController
@RequestMapping("%[2]s")
class %[3]s {

    @GetMapping("/ping")
    fun ping(): SaResult = SaResult.data("%[4]s")
}
`, pkg, spec.APIPath, className, spec.ID)
	} else {
		file = className + ".java"
		content = fmt.Sprintf(`package %[1]s.controller;

import cn.dev33.satoken.util.SaResult;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 模块接入检查。启动后登录，带上 satoken 请求头访问 GET %[2]s/ping。 */
@RestController
@RequestMapping("%[2]s")
public class %[3]s {

    @GetMapping("/ping")
    public SaResult ping() {
        return SaResult.data("%[4]s");
    }
}
`, pkg, spec.APIPath, className, spec.ID)
	}
	if err := os.WriteFile(filepath.Join(ctrlDir, file), []byte(content), 0o644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(moduleReadme(p, spec, pkg)), 0o644)
}

func moduleReadme(p *meta.Project, spec Spec, pkg string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", spec.ID)
	fmt.Fprintf(&b, "| 项 | 值 |\n|----|----|\n")
	fmt.Fprintf(&b, "| 包 | `%s` |\n", pkg)
	fmt.Fprintf(&b, "| 接口前缀 | `%s` |\n", spec.APIPath)
	if len(spec.Depends) > 0 {
		fmt.Fprintf(&b, "| 依赖的其他应用 | %s |\n", strings.Join(spec.Depends, "、"))
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "`%s` 已经依赖这个模块，启动后会扫描到 `%s` 下的代码。\n\n", p.AdminModule, pkg)
	if p.Language == config.LangKotlin {
		b.WriteString("写法参考 `yi-demo` 的演示公告：实体放 `entity/`，DTO 写在 `src/main/dto/*.dto`，" +
			"标准资源的 Controller 继承 `JimmerCrudResource` 并用注解开放接口。管理页 Controller 放 `admin/`，模板放 `src/main/resources/admin/`。\n\n")
	} else {
		b.WriteString("写法参考 `yi-demo` 的演示文章：实体放 `model/entity/`，DTO 写在 `src/main/dto/*.dto`，" +
			"再按 DAO、Repository、Service、Controller 分层。管理页 Controller 放 `admin/`，模板放 `src/main/resources/admin/`。\n\n")
	}
	b.WriteString("Jimmer 的 `database-validation-mode` 是 `ERROR`：新增实体后先建表，再启动。\n\n")
	b.WriteString("删除这个模块用项目根目录的命令：\n\n```text\nstar-yi.cmd remove-app --module-id " + spec.ID + "\n```\n")
	return b.String()
}
