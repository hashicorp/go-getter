package getter

import (
	"net/http"
	"strings"
	"testing"
)

const testBBUrl = "https://bitbucket.org/hashicorp/tf-test-git"

func TestBitBucketDetector(t *testing.T) {
	t.Parallel()

	if _, err := http.Get(testBBUrl); err != nil {
		t.Log("internet may not be working, skipping BB tests")
		t.Skip()
	}

	cases := []struct {
		Input  string
		Output string
		g      Getter
	}{
		// HTTP
		{
			"bitbucket.org/hashicorp/tf-test-git",
			"https://bitbucket.org/hashicorp/tf-test-git.git",
			new(GitGetter),
		},
		{
			"bitbucket.org/hashicorp/tf-test-git.git",
			"https://bitbucket.org/hashicorp/tf-test-git.git",
			new(GitGetter),
		},
		{
			"bitbucket.org/hashicorp/tf-test-hg",
			"https://bitbucket.org/hashicorp/tf-test-hg",
			new(HgGetter),
		},
	}

	pwd := "/pwd"
	for i, tc := range cases {
		var err error
		for i := 0; i < 3; i++ {
			var output string
			var ok bool
			output, ok, err = tc.g.Detect(tc.Input, pwd)
			if err != nil {
				if strings.Contains(err.Error(), "invalid character") {
					continue
				}

				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			if output != tc.Output {
				t.Fatalf("%d: bad: %#v", i, output)
			}

			break
		}
		if i >= 3 {
			t.Fatalf("failure from bitbucket: %s", err)
		}
	}
}
