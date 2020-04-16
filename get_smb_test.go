package getter

import (
	"context"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"testing"
)

func TestSmbGetter_impl(t *testing.T) {
	var _ Getter = new(SmbGetter)
}

// make the structure
func TestSmbGetter_Get(t *testing.T) {
	g := new(SmbGetter)
	ctx := context.Background()

	// correct url with auth data
	url, err := urlhelper.Parse("smb://vagrant:vagrant@samba/shared/file.txt")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req := &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: should not fail %s", err.Error())
	}

	//correct url with auth data and subdir
	url, err = urlhelper.Parse("smb://vagrant:vagrant@samba/shared/subdir/file.txt")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req = &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err != nil {
		t.Fatalf("err: should not fail: %s", err.Error())
	}

	// no hostname provided
	url, err = urlhelper.Parse("smb://")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req = &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}

	// no filepath provided
	url, err = urlhelper.Parse("smb://host")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req = &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}
}

//func TestSmbGetter_GetFile(t *testing.T) {
//	g := new(SmbGetter)
//	ctx := context.Background()
//
//	// no hostname provided
//	url, err := urlhelper.Parse("smb://")
//	if err != nil {
//		t.Fatalf("err: %s", err.Error())
//	}
//	req := &Request{
//		u: url,
//	}
//	if err := g.GetFile(ctx, req); err != nil && err.Error() != basePathError {
//		t.Fatalf("err: expected error: %s\n but error was: %s", basePathError, err.Error())
//	}
//
//	// no filepath provided
//	url, err = urlhelper.Parse("smb://host")
//	if err != nil {
//		t.Fatalf("err: %s", err.Error())
//	}
//	req = &Request{
//		u: url,
//	}
//	if err := g.GetFile(ctx, req); err != nil && err.Error() != basePathError {
//		t.Fatalf("err: expected error: %s\n but error was: %s", basePathError, err.Error())
//	}
//}

//func TestSmbGetter_Mode(t *testing.T) {
//	g := new(SmbGetter)
//	ctx := context.Background()
//
//	// no hostname provided
//	url, err := urlhelper.Parse("smb://")
//	if err != nil {
//		t.Fatalf("err: %s", err.Error())
//	}
//	if _, err := g.Mode(ctx, url); err == nil {
//		t.Fatalf("err: should fail when request url doesn't have a Host")
//	}
//	if _, err := g.Mode(ctx, url); err != nil && err.Error() != basePathError {
//		t.Fatalf("err: expected error: %s\n but error was: %s", basePathError, err.Error())
//	}
//
//	// no filepath provided
//	url, err = urlhelper.Parse("smb://")
//	if err != nil {
//		t.Fatalf("err: %s", err.Error())
//	}
//	if _, err := g.Mode(ctx, url); err == nil {
//		t.Fatalf("err: should fail when request url doesn't have a Host")
//	}
//	if _, err := g.Mode(ctx, url); err != nil && err.Error() != basePathError {
//		t.Fatalf("err: expected error: %s\n but error was: %s", basePathError, err.Error())
//	}
//}
