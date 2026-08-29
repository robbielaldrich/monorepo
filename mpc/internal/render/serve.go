package render

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const upstream = "https://github.com/Investigamer/cardconjurer.git"

// CacheDir is where a fetched Card Conjurer checkout lives by default.
func CacheDir() string {
	if d := os.Getenv("MPC_CARDCONJURER"); d != "" {
		return d
	}
	base, err := os.UserCacheDir()
	if err != nil {
		base = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(base, "mpc", "cardconjurer")
}

// Fetch clones Card Conjurer into dir, or pulls if it is already there.
func Fetch(dir string, stderr *os.File) error {
	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Stdout, cmd.Stderr = stderr, stderr
		return cmd.Run()
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return run("-C", dir, "pull", "--ff-only")
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return err
	}
	return run("clone", "--depth", "1", upstream, dir)
}

// Check reports whether dir looks like a Card Conjurer checkout.
func Check(dir string) error {
	for _, want := range []string{"creator/index.html", "js/creator-23.js", "js/frames"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			return fmt.Errorf("%s is not a Card Conjurer checkout (missing %s); run `mpc fetch`", dir, want)
		}
	}
	return nil
}

// Server serves a Card Conjurer checkout plus the deck's art directory, and
// collects rendered images the page posts back.
//
// The page uploads its canvases rather than returning them inline because a
// 2010x2814 PNG is several megabytes; round-tripping that through a CDP
// evaluate result as base64 is slow at best and kills the tab at worst.
type Server struct {
	Addr string
	ln   net.Listener
	srv  *http.Server

	mu      sync.Mutex
	uploads map[string][]byte
}

// UploadPath is the endpoint prefix the page posts finished images to.
const UploadPath = "/_mpc/put/"

// Take returns and clears an uploaded image.
func (s *Server) Take(name string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.uploads[name]
	delete(s.uploads, name)
	return b, ok
}

// Reset drops anything left over from a previous card.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.uploads)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, UploadPath)
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.uploads[name] = body
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// Serve starts a loopback HTTP server exposing ccDir at / and artDir at /art/
// on an arbitrary free port.
func Serve(ccDir, artDir string) (*Server, error) { return ServeOn(ccDir, artDir, 0) }

// ServeOn is Serve on a fixed port, for `mpc serve`.
func ServeOn(ccDir, artDir string, port int) (*Server, error) {
	if err := Check(ccDir); err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		Addr:    "http://" + ln.Addr().String(),
		ln:      ln,
		uploads: map[string][]byte{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc(UploadPath, s.upload)
	mux.Handle("/", noCache(http.FileServer(http.Dir(ccDir))))
	if artDir != "" {
		mux.Handle("/art/", http.StripPrefix("/art/", http.FileServer(http.Dir(artDir))))
	}
	s.srv = &http.Server{Handler: mux}
	go s.srv.Serve(ln)
	return s, nil
}

func (s *Server) Close() error { return s.srv.Close() }

// noCache keeps the creator's HTML and JS fresh; Card Conjurer ships far-future
// cache headers in its nginx config that get in the way of a local checkout.
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".html") || strings.HasSuffix(r.URL.Path, "/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		h.ServeHTTP(w, r)
	})
}
