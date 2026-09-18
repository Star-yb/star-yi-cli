package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// Ask 提示输入；default 非空时回车采用默认值。
func Ask(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

// AskRequired 必填；空输入则重复提示。
func AskRequired(label string) string {
	for {
		v := Ask(label, "")
		if v != "" {
			return v
		}
		fmt.Println("  此项不能为空，请重新输入。")
	}
}

// AskYesNo 询问 y/n，defaultYes 为回车默认。
func AskYesNo(label string, defaultYes bool) bool {
	def := "n"
	if defaultYes {
		def = "y"
	}
	for {
		v := strings.ToLower(Ask(label+" (y/n)", def))
		if v == "" {
			return defaultYes
		}
		if v == "y" || v == "yes" || v == "是" {
			return true
		}
		if v == "n" || v == "no" || v == "否" {
			return false
		}
		fmt.Println("  请输入 y 或 n。")
	}
}

// Choose 数字菜单，返回所选序号；无效则重试。
func Choose(title string, options []string) int {
	for {
		fmt.Println()
		fmt.Println(title)
		for i, opt := range options {
			fmt.Printf("  %d) %s\n", i+1, opt)
		}
		fmt.Printf("请选择 [1-%d]: ", len(options))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(options) {
			fmt.Println("  无效选项，请输入列表中的数字。")
			continue
		}
		return n
	}
}

// AskOptionalList 逗号分隔列表，空则 nil。
func AskOptionalList(label string) []string {
	raw := Ask(label+"（多个用英文逗号分隔，可留空）", "")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
