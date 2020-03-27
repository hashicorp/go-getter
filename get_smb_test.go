package getter

import (
	"context"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"testing"
)

func TestSmbGetter_impl(t *testing.T) {
	var _ Getter = new(SmbGetter)
}

func TestSmbGetter_Get(t *testing.T) {
	g := new(SmbGetter)
	ctx := context.Background()

	// no hostname provided
	url, err := urlhelper.Parse("smb://")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req := &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}
	if err := g.GetFile(ctx, req); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
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
	if err := g.GetFile(ctx, req); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
	}
}

func TestSmbGetter_GetFile(t *testing.T) {
	g := new(SmbGetter)
	ctx := context.Background()

	// no hostname provided
	url, err := urlhelper.Parse("smb://")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	req := &Request{
		u: url,
	}
	if err := g.Get(ctx, req); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}
	if err := g.GetFile(ctx, req); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
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
	if err := g.GetFile(ctx, req); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
	}
}

func TestSmbGetter_Mode(t *testing.T) {
	g := new(SmbGetter)
	ctx := context.Background()
	
	// no hostname provided
	url, err := urlhelper.Parse("smb://")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	if _, err := g.Mode(ctx, url); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}
	if _, err := g.Mode(ctx, url); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
	}

	// no filepath provided
	url, err = urlhelper.Parse("smb://")
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}
	if _, err := g.Mode(ctx, url); err == nil {
		t.Fatalf("err: should fail when request url doesn't have a Host")
	}
	if _, err := g.Mode(ctx, url); err != nil && err.Error() != pathError {
		t.Fatalf("err: expected error: %s\n but error was: %s", pathError, err.Error())
	}
}
