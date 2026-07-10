// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"path/filepath"
	"testing"
)

func TestObjectDestination(t *testing.T) {
	dst := t.TempDir()
	tests := []struct {
		name    string
		prefix  string
		key     string
		want    string
		wantErr bool
	}{
		{name: "prefix root", prefix: "prefix", key: "prefix", want: dst},
		{name: "nested object", prefix: "prefix", key: filepath.Join("prefix", "module", "main.tf"), want: filepath.Join(dst, "module", "main.tf")},
		{name: "parent traversal", prefix: "prefix", key: filepath.Join("prefix", "..", "..", "target"), wantErr: true},
		{name: "sibling prefix", prefix: "prefix", key: filepath.Join("prefix-other", "main.tf"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objectDestination(dst, tt.prefix, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
