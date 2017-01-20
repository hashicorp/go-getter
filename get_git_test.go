package getter

import (
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var testHasGit bool

func init() {
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}

func TestGitGetter_impl(t *testing.T) {
	var _ Getter = new(GitGetter)
}

func TestGitGetter(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// Git doesn't allow nested ".git" directories so we do some hackiness
	// here to get around that...
	moduleDir := filepath.Join(fixtureDir, "basic-git")
	oldName := filepath.Join(moduleDir, "DOTgit")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	// With a dir that doesn't exist
	if err := g.Get(dst, testModuleURL("basic-git")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_branch(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// Git doesn't allow nested ".git" directories so we do some hackiness
	// here to get around that...
	moduleDir := filepath.Join(fixtureDir, "basic-git")
	oldName := filepath.Join(moduleDir, "DOTgit")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	url := testModuleURL("basic-git")
	q := url.Query()
	q.Add("ref", "test-branch")
	url.RawQuery = q.Encode()

	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main_branch.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get again should work
	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "main_branch.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_branchUpdate(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// First setup the state with a fresh branch
	moduleDir := filepath.Join(fixtureDir, "git-branch-update")
	oldName := filepath.Join(moduleDir, "DOTgit-1")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	// Get the "test-branch" branch
	url := testModuleURL("git-branch-update")
	q := url.Query()
	q.Add("ref", "test-branch")
	url.RawQuery = q.Encode()
	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main_branch.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Swap the data to have a branch update
	if err := os.Rename(newName, oldName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(oldName, newName)
	oldName = filepath.Join(moduleDir, "DOTgit-2")
	newName = filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	// Get again should work
	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "main_branch_update.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_tag(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// Git doesn't allow nested ".git" directories so we do some hackiness
	// here to get around that...
	moduleDir := filepath.Join(fixtureDir, "basic-git")
	oldName := filepath.Join(moduleDir, "DOTgit")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	url := testModuleURL("basic-git")
	q := url.Query()
	q.Add("ref", "v1.0")
	url.RawQuery = q.Encode()

	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main_tag1.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get again should work
	if err := g.Get(dst, url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "main_tag1.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_GetFile(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempFile(t)

	// Git doesn't allow nested ".git" directories so we do some hackiness
	// here to get around that...
	moduleDir := filepath.Join(fixtureDir, "basic-git")
	oldName := filepath.Join(moduleDir, "DOTgit")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	// Download
	if err := g.GetFile(dst, testModuleURL("basic-git/foo.txt")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "Hello\n")
}

func TestGitGetter_gitVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "go-getter")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	script := filepath.Join(dir, "git")
	err = ioutil.WriteFile(
		script,
		[]byte("#!/bin/sh\necho git version 2.0\n"),
		0700)
	if err != nil {
		t.Fatal(err)
	}

	defer func(v string) {
		os.Setenv("PATH", v)
	}(os.Getenv("PATH"))

	os.Setenv("PATH", dir)

	// Asking for a higher version throws an error
	if err := checkGitVersion("2.3"); err == nil {
		t.Fatal("expect git version error")
	}

	// Passes when version is satisfied
	if err := checkGitVersion("1.9"); err != nil {
		t.Fatal(err)
	}
}

func TestGitGetter_sshKey(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	encodedKey := base64.StdEncoding.EncodeToString([]byte(testGitToken))

	u, err := url.Parse("ssh://git@github.com/hashicorp/test-private-repo" +
		"?sshkey=" + encodedKey)
	if err != nil {
		t.Fatal(err)
	}

	if err := g.Get(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	readmePath := filepath.Join(dst, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_submodule(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// Set up the parent
	moduleDir := filepath.Join(fixtureDir, "git-submodule-parent")
	oldName := filepath.Join(moduleDir, "DOTgit")
	newName := filepath.Join(moduleDir, ".git")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(newName, oldName)

	// Set up the child
	childModuleDir := filepath.Join(fixtureDir, "git-submodule-child")
	childOldName := filepath.Join(childModuleDir, "DOTgit")
	childNewName := filepath.Join(childModuleDir, ".git")
	if err := os.Rename(childOldName, childNewName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(childNewName, childOldName)

	// Set up the grandchild
	grandchildModuleDir := filepath.Join(fixtureDir, "git-submodule-grandchild")
	grandchildOldName := filepath.Join(grandchildModuleDir, "DOTgit")
	grandchildNewName := filepath.Join(grandchildModuleDir, ".git")
	if err := os.Rename(grandchildOldName, grandchildNewName); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Rename(grandchildNewName, grandchildOldName)

	// Clone the root repository
	if err := g.Get(dst, testModuleURL("git-submodule-parent")); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that the files exist
	for _, path := range []string{
		filepath.Join(dst, "parent.txt"),
		filepath.Join(dst, "child", "child.txt"),
		filepath.Join(dst, "child", "grandchild", "grandchild.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

// This is a read-only deploy key for an empty test repository.
var testGitToken = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEArGJ7eweUMiT58m424ZHLu6UordeoTcOTPEMeOjIL2GuVhPU+
Y6sdW3gMKEYFKo5ywXxVgNo8VCI8Ny8+PPfR+BNJaAI+VYNDU5rvD3ecfIjH3We4
VyRbT/PcxNK1XJcE260P6nVXrnNLJQBbsP6tjqSswwVy/9gCiI0aa4GxvK4R1ZPJ
H6ONYXzwgYR0QAH6jhyENe5skbH+40fT2u/I3z99HggqKOCJpgq9JkAWdXdqJPO7
kcGP6I6lTE1Cjpi7GEuVx6iWeflmX3uveOLTJohVkhAzGxIk5rIgbqkDoiNJ1RFl
MxFCc/LkmqdYiW6DgrWZJhlY9wB+YFWi3O/2BwIDAQABAoIBAQCE9LROcMsBXfmV
3SHhGqUrRjg41NOPnt+JpC7FLeJq+pdo5ApJrynGabHewhqr9xBVYUNFTY0oSvts
iLiVJ4K/tohwewJ+y+36ps3pfRSqDIkyoBPSykzPPsQw3l9ZWXU6xaE38Wc+Othj
YoJV4igUk7hX9nT7FSznCwWsk2x1m/w40PVDeWp0VOqGz407oPpirL8wS6yxwrcL
IR/XtEXOiOoJmHMdxlNwVOTdMz5mtCGJcl2IqjLZLP0az0SxAkTLrDeR+R9tTY/T
cbdZS3aBVi/9pXQ9yG+QcVrV1PKGdSzOoS1QB0746n9qW4pM93PoRkeENBAM44Gx
zJvanaqRAoGBANU7HbhkUzBiotEhFlf4uQ3cKFzlSMoJAX27OKR8MDD2vLEL0lBv
biYBntMBU/L3A7nr/oVHJRS3dGVEoJdmvoXB+eCpNhyYiZKDXrPfaY3ifRKvcIoq
XuWYkIGB0X1Djf7Sj6ruSxcm8y6M4l2kQq7bo7HXHvJuPRuG930OzAopAoGBAM72
A0+3xTQrzbHcffPJPw8GUvk8tVmypHojQyXdX283GDW7LYvHd+x6rCNDIdXiZ25L
M3YKEcZMPpjnjEH5CRUHyubocelyRiz7P2Hwj3MOSO5g11nLbSlkLYvoG4uuH8ck
2trIRJ81OnVwwIj61CNMCG3CyYk6GN5ShDCJNWSvAoGBAKScyKrrOJWn8A4GvxsW
9rXOepKMp47hOPd5q5bAEOwb7zu25pwWCjDpG1XGNqrhK01C9PCrJeNCZWcwfdGk
Df1w7JkVyKJ21+314Qx3syNH8EqWigkAANa62wQ/1hwgJOTOZP8Oi4XKGf6b4L1t
69TV1x+Z9Vgu5pnzregrnjVRAoGBAIm1KhjmB4KiTti1BN2sn5fItnb+jRClDEn0
op5UQUcIGsTNyg2C6Onh6h4AckgVwIqj4Rb+tjsCyngFQc83/HIQ4FJqgjk5/zW4
68CoR1rgO2jZ6RDnibgL3z6Db6iucJiajkEbFoX07fPs1T+P3o2p7sXR4TW9AYUU
1L5S3cMjAoGBAKd+zv8xjwN9bw9wGz3l/5lWni6muXpmJ7a43Hj562jspb+moMqM
thGypwYJHZX05VkSk8iXvZehE+Czj6xu9P5FtxKCWgMT6hc8qvCq4n41Ndx59zkN
yuFmGAiAN8bAZgSQYyIUnWENsqFJNkj/HHR4MA/O2gY1zPq/PFCvQ9Q4
-----END RSA PRIVATE KEY-----`
