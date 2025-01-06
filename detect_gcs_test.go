// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"testing"
)

func TestGCSDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"www.googleapis.com/storage/v1/bucket/foo",
			"gcs::https://www.googleapis.com/storage/v1/bucket/foo",
		},
		{
			"www.googleapis.com/storage/v1/bucket/foo/bar",
			"gcs::https://www.googleapis.com/storage/v1/bucket/foo/bar",
		},
		{
			"www.googleapis.com/storage/v1/foo/bar.baz",
			"gcs::https://www.googleapis.com/storage/v1/foo/bar.baz",
		},
		{
			"www.googleapis.com/storage/v2/foo/bar/toor.baz",
			"gcs::https://www.googleapis.com/storage/v2/foo/bar/toor.baz",
		},
	}

	pwd := "/pwd"
	f := new(GCSDetector)
	for i, tc := range cases {
		output, ok, err := f.Detect(tc.Input, pwd)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if output != tc.Output {
			t.Fatalf("%d: bad: %#v", i, output)
		}
	}
}

func TestGCSDetector_MalformedDetectHTTP(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected string
		Output   string
	}{
		{
			"valid url",
			"www.googleapis.com/storage/v1/my-bucket/foo/bar",
			"",
			"gcs::https://www.googleapis.com/storage/v1/my-bucket/foo/bar",
		},
		{
			"not valid url length",
			"www.googleapis.com.invalid/storage/v1/",
			"URL is not a valid GCS URL",
			"",
		},
		{
			"not valid version",
			"www.googleapis.com/storage/invalid-version/my-bucket/foo",
			"GCS URL version is not valid",
			"",
		},
		{
			"not valid bucket",
			"www.googleapis.com/storage/v1/127.0.0.1/foo",
			"GCS URL bucket name is not valid",
			"",
		},
		{
			"not valid object",
			"www.googleapis.com/storage/v1/my-bucket/.well-known/acme-challenge/foo",
			"GCS URL object name is not valid",
			"",
		},
		{
			"path traversal",
			"www.googleapis.com/storage/v1/my-bucket/../../../foo/bar",
			"URL is not a valid GCS URL",
			"",
		},
	}

	pwd := "/pwd"
	f := new(GCSDetector)
	for _, tc := range cases {
		output, _, err := f.Detect(tc.Input, pwd)
		if err != nil {
			if err.Error() != tc.Expected {
				t.Fatalf("expected error %s, got %s for %s", tc.Expected, err.Error(), tc.Name)
			}
		}

		if output != tc.Output {
			t.Fatalf("expected %s, got %s", tc.Output, output)
		}
	}
}

func TestIsValidGCSVersion(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{
			"valid version",
			"v1",
			true,
		},
		{
			"invalid version",
			"invalid1",
			false,
		},
	}

	for _, tc := range cases {
		output := isValidGCSVersion(tc.Input)
		if output != tc.Expected {
			t.Fatalf("expected %t, got %t for test %s", tc.Expected, output, tc.Name)
		}
	}
}

func TestIsValidGCSBucketName(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{
			"valid bucket name",
			"my-bucket",
			true,
		},
		{
			"invalid bucket name",
			"..",
			false,
		},
	}

	for _, tc := range cases {
		output := isValidGCSBucketName(tc.Input)
		if output != tc.Expected {
			t.Fatalf("expected %t, got %t for test %s", tc.Expected, output, tc.Name)
		}
	}
}

func TestIsValidGCSObjectName(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{
			"valid object name",
			"my-object",
			true,
		},
		{
			"invalid object name",
			"..",
			false,
		},
	}

	for _, tc := range cases {
		output := isValidGCSObjectName(tc.Input)
		if output != tc.Expected {
			t.Fatalf("expected %t, got %t for test %s", tc.Expected, output, tc.Name)
		}
	}
}
