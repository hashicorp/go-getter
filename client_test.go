package getter

import "testing"

func TestGet_git(t *testing.T) {
	c := &Client{
		Src: "github.com/hashicorp/go-getter",
		Dst: tempDir(t),
		Dir: true,
		Getters: map[string]Getter{
			"git": new(GitGetter),
		},
	}

	if err := c.Get(); err != nil {
		t.Fatalf("\nFatal getting go-getter: %s", err)
	}
}

func TestGet_git_detect(t *testing.T) {
	if err := Get(tempDir(t), "github.com/hashicorp/go-getter"); err != nil {
		t.Fatalf("\nFatal getting go-getter: %s", err)
	}
}
