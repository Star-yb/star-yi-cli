package projecttool

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/star/star-yi-cli/pkg/pom"
)

// InitOptions 生成项目专用脚本参数。
type InitOptions struct {
	ProjectRoot string
	CLICommand  string
	Force       bool
	AllPlatforms bool
}

// InitResult 脚本生成结果。
type InitResult struct {
	ProjectRoot string
	Generated   []string
	CLIPath     string
}

// InitProjectTool 在项目根目录生成项目专用脚本（默认仅当前系统）。
func InitProjectTool(opts InitOptions) (InitResult, error) {
	root, err := filepath.Abs(strings.TrimSpace(opts.ProjectRoot))
	if err != nil {
		return InitResult{}, err
	}
	if err := pom.ValidateStarYiProject(root); err != nil {
		return InitResult{}, err
	}

	cliPath, err := CopyCLIToProject(root)
	if err != nil {
		return InitResult{}, fmt.Errorf("复制 CLI 到项目目录失败: %w", err)
	}

	cli := strings.TrimSpace(opts.CLICommand)
	if cli == "" {
		cli = "star-yi-cli"
	}

	scriptMap := map[string]string{
		"cmd": filepath.Join(root, "star-yi-project-tool.cmd"),
		"ps1": filepath.Join(root, "star-yi-project-tool.ps1"),
		"sh":  filepath.Join(root, "star-yi-project-tool.sh"),
	}

	targetKinds := []string{}
	if opts.AllPlatforms {
		targetKinds = []string{"cmd", "ps1", "sh"}
	} else {
		switch runtime.GOOS {
		case "windows":
			targetKinds = []string{"cmd"}
		default:
			targetKinds = []string{"sh"}
		}
	}

	for _, kind := range targetKinds {
		p := scriptMap[kind]
		if !opts.Force {
			if _, err := os.Stat(p); err == nil {
				return InitResult{}, fmt.Errorf("文件已存在: %s（可使用 --force 覆盖）", p)
			}
		}
	}

	sh := renderSh(root, cli)
	ps1 := renderPs1(root, cli)
	cmd := renderCmd(root, cli)

	generated := make([]string, 0, len(targetKinds))
	for _, kind := range targetKinds {
		path := scriptMap[kind]
		var content string
		mode := os.FileMode(0o644)
		switch kind {
		case "cmd":
			content = cmd
		case "ps1":
			content = ps1
		case "sh":
			content = sh
			mode = 0o755
		}
		if kind == "cmd" {
			content = toCRLF(content)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			return InitResult{}, err
		}
		generated = append(generated, path)
	}

	return InitResult{
		ProjectRoot: root,
		Generated:   generated,
		CLIPath:     cliPath,
	}, nil
}

func renderSh(projectRoot, cli string) string {
	_ = projectRoot
	_ = cli
	return `#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
CLI_CMD="star-yi-cli"
if [[ -n "${STAR_YI_CLI:-}" ]]; then CLI_CMD="$STAR_YI_CLI"; fi
if [[ -x "$PROJECT_ROOT/common-core/star-yi-cli" ]]; then CLI_CMD="$PROJECT_ROOT/common-core/star-yi-cli"; fi

usage() {
  echo "用法:"
  echo "  ./star-yi-project-tool.sh add-app --module-id <id> --package-suffix <suffix>"
  echo "  ./star-yi-project-tool.sh remove-app --module-id <id>"
  echo "  # 其它原生命令透传:"
  echo "  ./star-yi-project-tool.sh add app --module-id <id> --package-suffix <suffix>"
}

if [[ $# -lt 1 ]]; then
  echo "请选择操作："
  echo "  1) add-app"
  echo "  2) remove-app"
  echo "  3) 退出"
  read -r -p "输入选项 [1-3]: " choice
  if [[ "$choice" == "1" ]]; then
    read -r -p "模块名(module-id): " module_id
    read -r -p "包后缀(package-suffix, 默认 xxx): " package_suffix
    if [[ -z "$package_suffix" ]]; then package_suffix="xxx"; fi
    exec "$CLI_CMD" add app --project "$PROJECT_ROOT" --module-id "$module_id" --package-suffix "$package_suffix"
  elif [[ "$choice" == "2" ]]; then
    read -r -p "模块名(module-id): " module_id
    exec "$CLI_CMD" remove app --project "$PROJECT_ROOT" --module-id "$module_id"
  else
    exit 0
  fi
fi

action="$1"
shift || true

if [[ "$action" == "add-app" ]]; then
  exec "$CLI_CMD" add app --project "$PROJECT_ROOT" "$@"
elif [[ "$action" == "remove-app" ]]; then
  exec "$CLI_CMD" remove app --project "$PROJECT_ROOT" "$@"
else
  exec "$CLI_CMD" "$action" "$@"
fi
`
}

func renderPs1(projectRoot, cli string) string {
	return fmt.Sprintf(`param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
)

$ProjectRoot = %q
$CliCmd = if ($env:STAR_YI_CLI) { $env:STAR_YI_CLI } else { %q }
$BundledCli = Join-Path $ProjectRoot "common-core\star-yi-cli.exe"
if (Test-Path $BundledCli) { $CliCmd = $BundledCli }

function Show-Usage {
    Write-Host "用法:"
    Write-Host "  .\star-yi-project-tool.ps1 add-app --module-id <id> --package-suffix <suffix>"
    Write-Host "  .\star-yi-project-tool.ps1 remove-app --module-id <id>"
    Write-Host "  # 其它原生命令透传:"
    Write-Host "  .\star-yi-project-tool.ps1 add app --module-id <id> --package-suffix <suffix>"
}

if (-not $Args -or $Args.Count -lt 1) {
    Write-Host "请选择操作："
    Write-Host "  1) add-app"
    Write-Host "  2) remove-app"
    Write-Host "  3) 退出"
    $choice = Read-Host "输入选项 [1-3]"
    if ($choice -eq "1") {
        $moduleId = Read-Host "模块名(module-id)"
        $packageSuffix = Read-Host "包后缀(package-suffix, 默认 xxx)"
        if ([string]::IsNullOrWhiteSpace($packageSuffix)) { $packageSuffix = "xxx" }
        & $CliCmd add app --project $ProjectRoot --module-id $moduleId --package-suffix $packageSuffix
        exit $LASTEXITCODE
    }
    elseif ($choice -eq "2") {
        $moduleId = Read-Host "模块名(module-id)"
        & $CliCmd remove app --project $ProjectRoot --module-id $moduleId
        exit $LASTEXITCODE
    }
    else {
        exit 0
    }
}

$action = $Args[0]
$rest = @()
if ($Args.Count -gt 1) {
    $rest = $Args[1..($Args.Count - 1)]
}

if ($action -eq "add-app") {
    & $CliCmd add app --project $ProjectRoot @rest
    exit $LASTEXITCODE
}
elseif ($action -eq "remove-app") {
    & $CliCmd remove app --project $ProjectRoot @rest
    exit $LASTEXITCODE
}
else {
    & $CliCmd $action @rest
    exit $LASTEXITCODE
}
`, projectRoot, cli)
}

func renderCmd(projectRoot, cli string) string {
	_ = projectRoot
	_ = cli
	return `@echo off
setlocal EnableDelayedExpansion

set "PROJECT_ROOT=%~dp0"
if "%PROJECT_ROOT:~-1%"=="\" set "PROJECT_ROOT=%PROJECT_ROOT:~0,-1%"

set "CLI_CMD=star-yi-cli"
if defined STAR_YI_CLI set "CLI_CMD=%STAR_YI_CLI%"
if exist "%~dp0common-core\star-yi-cli.exe" set "CLI_CMD=%~dp0common-core\star-yi-cli.exe"

if not "%~1"=="" goto run_cmd
goto menu

:menu
echo Select action:
echo   1) add-app
echo   2) remove-app
echo   3) exit
set "CHOICE="
set /p "CHOICE=Enter choice [1-3]: "
if "%CHOICE%"=="1" goto do_add
if "%CHOICE%"=="2" goto do_remove
if "%CHOICE%"=="3" exit /b 0
echo Invalid choice.
goto menu

:do_add
set "MODULE_ID="
set /p "MODULE_ID=module-id: "
if "!MODULE_ID!"=="" goto module_required
set "PACKAGE_SUFFIX="
echo package-suffix, press Enter for default xxx
set /p "PACKAGE_SUFFIX="
if "!PACKAGE_SUFFIX!"=="" set "PACKAGE_SUFFIX=xxx"
call "%CLI_CMD%" add app --project "%PROJECT_ROOT%" --module-id "!MODULE_ID!" --package-suffix "!PACKAGE_SUFFIX!"
set "EC=!ERRORLEVEL!"
exit /b !EC!

:do_remove
set "MODULE_ID="
set /p "MODULE_ID=module-id: "
if "!MODULE_ID!"=="" goto module_required
call "%CLI_CMD%" remove app --project "%PROJECT_ROOT%" --module-id "!MODULE_ID!"
set "EC=!ERRORLEVEL!"
exit /b !EC!

:module_required
echo module-id is required.
exit /b 1

:run_cmd
set "ACTION=%~1"
shift

if /I "%ACTION%"=="add-app" goto run_add
if /I "%ACTION%"=="remove-app" goto run_remove
call "%CLI_CMD%" %ACTION% %*
set "EC=!ERRORLEVEL!"
exit /b !EC!

:run_add
call "%CLI_CMD%" add app --project "%PROJECT_ROOT%" %*
set "EC=!ERRORLEVEL!"
exit /b !EC!

:run_remove
call "%CLI_CMD%" remove app --project "%PROJECT_ROOT%" %*
set "EC=!ERRORLEVEL!"
exit /b !EC!
`
}

func toCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

