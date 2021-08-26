package getter

import (
	"net/url"
	"testing"
)

func TestRedactURL(t *testing.T) {
	cases := []struct {
		name string
		url  *url.URL
		want string
	}{
		{
			name: "non-blank Password",
			url: &url.URL{
				Scheme: "http",
				Host:   "host.tld",
				Path:   "this:that",
				User:   url.UserPassword("user", "password"),
			},
			want: "http://user:xxxxx@host.tld/this:that",
		},
		{
			name: "blank Password",
			url: &url.URL{
				Scheme: "http",
				Host:   "host.tld",
				Path:   "this:that",
				User:   url.User("user"),
			},
			want: "http://user@host.tld/this:that",
		},
		{
			name: "nil User",
			url: &url.URL{
				Scheme: "http",
				Host:   "host.tld",
				Path:   "this:that",
				User:   url.UserPassword("", "password"),
			},
			want: "http://:xxxxx@host.tld/this:that",
		},
		{
			name: "blank Username, blank Password",
			url: &url.URL{
				Scheme: "http",
				Host:   "host.tld",
				Path:   "this:that",
			},
			want: "http://host.tld/this:that",
		},
		{
			name: "empty URL",
			url:  &url.URL{},
			want: "",
		},
		{
			name: "nil URL",
			url:  nil,
			want: "",
		},
	}

	for _, tt := range cases {
		t := t
		t.Run(tt.name, func(t *testing.T) {
			if g, w := RedactURL(tt.url), tt.want; g != w {
				t.Fatalf("got: %q\nwant: %q", g, w)
			}
		})
	}
}
