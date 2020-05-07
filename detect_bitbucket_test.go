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
			var ok bool
			req := &Request{
				Src: tc.Input,
				Pwd: pwd,
			}
			ok, err = Detect(req, tc.g)
			if err != nil {
				if strings.Contains(err.Error(), "invalid character") {
					continue
				}

				t.Fatalf("err: %s", err)
			}
			if !ok {
				t.Fatal("not ok")
			}

			if req.Src != tc.Output {
				t.Fatalf("%d: bad: %#v", i, req.Src)
			}

			break
		}
		if i >= 3 {
			t.Fatalf("failure from bitbucket: %s", err)
		}
	}
}
