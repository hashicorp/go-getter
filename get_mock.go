package getter

import (
	"context"
	"net/url"
)

// MockGetter is an implementation of Getter that can be used for tests.
type MockGetter struct {
	getter

	// Proxy, if set, will be called after recording the calls below.
	// If it isn't set, then the *Err values will be returned.
	Proxy Getter

	GetCalled bool
	GetDst    string
	GetURL    *url.URL
	GetErr    error

	GetFileCalled bool
	GetFileDst    string
	GetFileURL    *url.URL
	GetFileErr    error
}

func (g *MockGetter) Get(ctx context.Context, dst string, u *url.URL) error {
	g.GetCalled = true
	g.GetDst = dst
	g.GetURL = u

	if g.Proxy != nil {
		return g.Proxy.Get(ctx, dst, u)
	}

	return g.GetErr
}

func (g *MockGetter) GetFile(ctx context.Context, dst string, u *url.URL) error {
	g.GetFileCalled = true
	g.GetFileDst = dst
	g.GetFileURL = u

	if g.Proxy != nil {
		return g.Proxy.GetFile(ctx, dst, u)
	}
	return g.GetFileErr
}

func (g *MockGetter) ClientMode(u *url.URL) (ClientMode, error) {
	if l := len(u.Path); l > 0 && u.Path[l-1:] == "/" {
		return ClientModeDir, nil
	}
	return ClientModeFile, nil
}
