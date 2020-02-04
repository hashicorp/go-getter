package getter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type MockProgressTracking struct {
	sync.Mutex
	downloaded map[string]int
}

func (p *MockProgressTracking) TrackProgress(src string,
	currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	p.Lock()
	defer p.Unlock()

	if p.downloaded == nil {
		p.downloaded = map[string]int{}
	}

	v, _ := p.downloaded[src]
	p.downloaded[src] = v + 1
	return stream
}

func TestGet_progress(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// all good
		rw.Header().Add("X-Terraform-Get", "something")
	}))
	defer s.Close()
	ctx := context.Background()

	{ // dl without tracking
		dst := tempTestFile(t)
		defer os.RemoveAll(filepath.Dir(dst))
		if _, err := GetFile(ctx, dst, s.URL+"/file?thig=this&that"); err != nil {
			t.Fatalf("download failed: %v", err)
		}
	}

	{ // tracking
		p := &MockProgressTracking{}
		dst := tempTestFile(t)
		defer os.RemoveAll(filepath.Dir(dst))
		req := &Request{
			Dst:              dst,
			Src:              s.URL + "/file?thig=this&that",
			ProgressListener: p,
		}
		if _, err := DefaultClient.Get(ctx, req); err != nil {
			t.Fatalf("download failed: %v", err)
		}
		req = &Request{
			Dst:              dst,
			Src:              s.URL + "/otherfile?thig=this&that",
			ProgressListener: p,
		}
		if _, err := DefaultClient.Get(ctx, req); err != nil {
			t.Fatalf("download failed: %v", err)
		}

		if p.downloaded["file"] != 1 {
			t.Error("Expected a file download")
		}
		if p.downloaded["otherfile"] != 1 {
			t.Error("Expected a otherfile download")
		}
	}
}
