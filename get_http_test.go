package getter

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestHttpGetter_impl(t *testing.T) {
	var _ Getter = new(HttpGetter)
}

func TestHttpGetter_header(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/header"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_meta(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_metaSubdir(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-subdir"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_none(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/none"

	// Get it!
	if err := g.Get(dst, &u); err == nil {
		t.Fatal("should error")
	}
}

func TestHttpGetter_file(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempFile(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/file"

	// Get it!
	if err := g.GetFile(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "Hello\n")
}

func TestHttpGetter_auth(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-auth"
	u.User = url.UserPassword("foo", "bar")

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_basicAuth(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/file-basic-auth"
	// set HTTP basic auth with query params
	q := u.Query()
	q.Add("http_basic_auth_user", "basicUser")
	q.Add("http_basic_auth_pass", "basicPass")
	u.RawQuery = q.Encode()

	// Get it!
	if err := g.GetFile(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "HelloBasicAuth\n")
}

func TestHttpGetter_authNetrc(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	// Write the netrc file
	path, closer := tempFileContents(t, fmt.Sprintf(testHttpNetrc, ln.Addr().String()))
	defer closer()
	defer tempEnv(t, "NETRC", path)()

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testHttpServer(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/file", testHttpHandlerFile)
	mux.HandleFunc("/file-basic-auth", testHttpHandlerFileBasicAuth)
	mux.HandleFunc("/header", testHttpHandlerHeader)
	mux.HandleFunc("/meta", testHttpHandlerMeta)
	mux.HandleFunc("/meta-auth", testHttpHandlerMetaAuth)
	mux.HandleFunc("/meta-subdir", testHttpHandlerMetaSubdir)

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpHandlerFile(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello\n"))
}

func testHttpHandlerFileBasicAuth(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}

	if user != "basicUser" || pass != "basicPass" {
		w.WriteHeader(401)
		return
	}

	w.Write([]byte("HelloBasicAuth\n"))
}

func testHttpHandlerHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Terraform-Get", testModuleURL("basic").String())
	w.WriteHeader(200)
}

func testHttpHandlerMeta(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaAuth(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}

	if user != "foo" || pass != "bar" {
		w.WriteHeader(401)
		return
	}

	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaSubdir(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic//subdir").String())))
}

func testHttpHandlerNone(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(testHttpNoneStr))
}

const testHttpMetaStr = `
<html>
<head>
<meta name="terraform-get" content="%s">
</head>
</html>
`

const testHttpNoneStr = `
<html>
<head>
</head>
</html>
`

const testHttpNetrc = `
machine %s
login foo
password bar
`
