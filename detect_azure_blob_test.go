package getter

import (
	"testing"
)

func TestDetectURL(t *testing.T) {
	d := &AzureBlobDetector{} // Assuming this struct exists

	tests := []struct {
		name    string
		input   string
		want    string
		wantOk  bool
		wantErr bool
	}{
		// Valid Cases
		{
			name:   "Valid HTTPS URL",
			input:  "https://myaccount.blob.core.windows.net/mycontainer",
			want:   "azureblob::https://myaccount.blob.core.windows.net/mycontainer",
			wantOk: true,
		},
		{
			name:   "Valid HTTP URL",
			input:  "http://myaccount.blob.core.windows.net/mycontainer",
			want:   "azureblob::https://myaccount.blob.core.windows.net/mycontainer",
			wantOk: true,
		},
		{
			name:   "Valid URL with blob path",
			input:  "https://myaccount.blob.core.windows.net/mycontainer/mypath/file.txt",
			want:   "azureblob::https://myaccount.blob.core.windows.net/mycontainer/mypath/file.txt",
			wantOk: true,
		},

		// Invalid Cases
		{
			name:    "Invalid Scheme",
			input:   "ftp://myaccount.blob.core.windows.net/mycontainer",
			wantErr: true,
		},
		{
			name:    "Invalid Hostname",
			input:   "https://myaccount.blob.azure.com/mycontainer",
			wantErr: true,
		},
		{
			name:    "Missing Container Name",
			input:   "https://myaccount.blob.core.windows.net/",
			wantErr: true,
		},
		{
			name:    "Completely Invalid URL",
			input:   "not_a_url",
			wantErr: true,
		},
		{
			name:    "Empty Input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := d.detectURL(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("Expected %q, got %q", tt.want, got)
				}
				if ok != tt.wantOk {
					t.Errorf("Expected ok = %v, got %v", tt.wantOk, ok)
				}
			}
		})
	}
}
