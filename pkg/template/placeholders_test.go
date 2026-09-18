package template

import "testing"

func TestReplaceLine(t *testing.T) {
	pairs := Vars{
		ProjectArtifactID: "Education-Yi",
		ProjectName:       "教育平台",
		GroupID:           "com.star",
		Version:           "0.0.1-SNAPSHOT",
	}.Pairs()

	got := ReplaceLine("<artifactId>{projectArtifactId}</artifactId>", pairs)
	want := "<artifactId>Education-Yi</artifactId>"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	line := "name={projectName} group={groupId} ver={version}"
	got = ReplaceLine(line, pairs)
	if got != "name=教育平台 group=com.star ver=0.0.1-SNAPSHOT" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestReplaceBytes(t *testing.T) {
	pairs := map[string]string{"{projectArtifactId}": "My-App"}
	out := string(ReplaceBytes([]byte("id={projectArtifactId}"), pairs))
	if out != "id=My-App" {
		t.Fatalf("got %q", out)
	}
}

func TestIsBinaryPath(t *testing.T) {
	if !IsBinaryPath("lib/foo.jar") {
		t.Fatal("expected jar binary")
	}
	if IsBinaryPath("pom.xml") {
		t.Fatal("pom.xml should be text")
	}
}
