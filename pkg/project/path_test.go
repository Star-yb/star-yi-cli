package project

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestProjectDir(t *testing.T) {
	got, err := ProjectDir(`D:\JAVA_File\111`, "sss")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(`D:\JAVA_File\111`, "sss")
	if runtime.GOOS == "windows" {
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	} else if filepath.Base(got) != "sss" {
		t.Fatalf("unexpected: %s", got)
	}
}
