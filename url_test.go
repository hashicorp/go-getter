// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

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
			want: "http://user:redacted@host.tld/this:that",
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
			want: "http://:redacted@host.tld/this:that",
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
		{
			name: "non-blank SSH key in URL query parameter",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey=LS0tLS1CRUdJTiBPUE",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted",
		},
		{
			name: "blank SSH key in URL query parameter",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey=",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted",
		},
		{
			name: "multiple SSH keys with no and non-empty values",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey&sshkey=secretkey",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted&sshkey=redacted",
		},
		{
			name: "multiple SSH keys with all empty values",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey&sshkey",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted&sshkey=redacted",
		},
		{
			name: "multiple SSH keys with mixed empty and blank values",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey=&sshkey=secretkey",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted&sshkey=redacted",
		},
		{
			name: "multiple SSH keys in URL query parameter",
			url: &url.URL{
				Scheme:   "ssh",
				User:     url.User("git"),
				Host:     "github.com",
				Path:     "hashicorp/go-getter-test-private.git",
				RawQuery: "sshkey=secretkey1&sshkey=secretkey2",
			},
			want: "ssh://git@github.com/hashicorp/go-getter-test-private.git?sshkey=redacted&sshkey=redacted",
		},
		{
			name: "S3 URL with aws_access_key_id",
			url: &url.URL{
				Scheme:   "s3",
				Host:     "bucket.s3.amazonaws.com",
				Path:     "/key",
				RawQuery: "aws_access_key_id=AKIAIOSFODNN7EXAMPLE",
			},
			want: "s3://bucket.s3.amazonaws.com/key?aws_access_key_id=redacted",
		},
		{
			name: "S3 URL with aws_access_key_secret",
			url: &url.URL{
				Scheme:   "s3",
				Host:     "bucket.s3.amazonaws.com",
				Path:     "/key",
				RawQuery: "aws_access_key_secret=wJalrXUtnFEMI%2FK7MDENG%2FbPxRfiCYEXAMPLEKEY",
			},
			want: "s3://bucket.s3.amazonaws.com/key?aws_access_key_secret=redacted",
		},
		{
			name: "S3 URL with aws_access_token",
			url: &url.URL{
				Scheme:   "s3",
				Host:     "bucket.s3.amazonaws.com",
				Path:     "/key",
				RawQuery: "aws_access_token=AQoXnyc4lcK4w",
			},
			want: "s3://bucket.s3.amazonaws.com/key?aws_access_token=redacted",
		},
		{
			name: "S3 URL with all three AWS credential params",
			url: &url.URL{
				Scheme:   "s3",
				Host:     "bucket.s3.amazonaws.com",
				Path:     "/key",
				RawQuery: "aws_access_key_id=AKID&aws_access_key_secret=SECRET&aws_access_token=TOKEN",
			},
			want: "s3://bucket.s3.amazonaws.com/key?aws_access_key_id=redacted&aws_access_key_secret=redacted&aws_access_token=redacted",
		},
		{
			name: "S3 URL with AWS credentials and non-sensitive params preserved",
			url: &url.URL{
				Scheme:   "s3",
				Host:     "bucket.s3.amazonaws.com",
				Path:     "/key",
				RawQuery: "aws_access_key_id=AKID&aws_access_key_secret=SECRET&region=us-east-1",
			},
			want: "s3://bucket.s3.amazonaws.com/key?aws_access_key_id=redacted&aws_access_key_secret=redacted&region=us-east-1",
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
