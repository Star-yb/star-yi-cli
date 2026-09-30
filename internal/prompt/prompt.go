// Package prompt 终端问答。
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// IsTerminal 标准输入是否是交互终端。
func IsTerminal() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func readLine() string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// Ask 提示输入；回车采用默认值。
func Ask(label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	if v := readLine(); v != "" {
		return v
	}
	return def
}

// AskValid 反复询问，直到 check 通过。
func AskValid(label, def string, check func(string) error) string {
	for {
		v := Ask(label, def)
		if err := check(v); err != nil {
			fmt.Println("  " + err.Error())
			continue
		}
		return v
	}
}

// YesNo 询问是否。
func YesNo(label string, def bool) bool {
	d := "y/N"
	if def {
		d = "Y/n"
	}
	for {
		fmt.Printf("%s (%s): ", label, d)
		switch strings.ToLower(readLine()) {
		case "":
			return def
		case "y", "yes", "是":
			return true
		case "n", "no", "否":
			return false
		}
		fmt.Println("  请输入 y 或 n。")
	}
}

// Choose 数字菜单，返回 1 开始的序号。
func Choose(title string, options []string) int {
	for {
		fmt.Println(title)
		for i, o := range options {
			fmt.Printf("  %d) %s\n", i+1, o)
		}
		fmt.Printf("请选择 [1-%d]: ", len(options))
		n, err := strconv.Atoi(readLine())
		if err == nil && n >= 1 && n <= len(options) {
			return n
		}
		fmt.Println("  请输入列表里的数字。")
	}
}

// List 逗号分隔的列表。
func List(label string) []string {
	var out []string
	for _, p := range strings.Split(Ask(label, ""), ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
