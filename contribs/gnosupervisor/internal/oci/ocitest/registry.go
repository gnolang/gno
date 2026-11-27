// Package ocitest is an in-memory OCI registry for tests: it serves image
// indexes, manifests and layers built from file maps, and can be told to
// damage what it serves.
package ocitest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// File is one tar entry of a layer.
type File struct {
	Name     string
	Body     string
	Mode     int64
	Type     byte   // tar.TypeReg when zero
	Linkname string // symlinks and hard links
}

// Registry is an httptest server speaking the read side of the
// distribution API.
type Registry struct {
	*httptest.Server
	mu    sync.Mutex
	blobs map[string][]byte // digest -> body
	types map[string]string // manifest digest -> media type

	// Token, when set, makes the registry answer a bearer challenge and
	// require "Bearer <Token>".
	Token string
	// Tamper, when set, may replace a body before it is served; attempt
	// counts the requests for that path so far, from 1.
	Tamper func(path string, attempt int, body []byte) []byte
	// ContentType, when set, replaces the Content-Type of every manifest.
	ContentType string
	// Requests counts the requests per path.
	requests sync.Map
	Total    atomic.Int64
}

// New starts a registry and stops it with the test.
func New(t *testing.T) *Registry {
	t.Helper()
	r := &Registry{blobs: map[string][]byte{}, types: map[string]string{}}
	r.Server = httptest.NewServer(http.HandlerFunc(r.serve))
	t.Cleanup(r.Close)
	return r
}

// Digest is the sha256 digest of b.
func Digest(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Layer builds a gzipped tar layer.
func Layer(t *testing.T, files ...File) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, f := range files {
		typ := f.Type
		if typ == 0 {
			typ = tar.TypeReg
		}
		mode := f.Mode
		if mode == 0 {
			mode = 0o644
			if typ == tar.TypeDir {
				mode = 0o755
			}
		}
		h := &tar.Header{Name: f.Name, Mode: mode, Typeflag: typ, Linkname: f.Linkname}
		if typ == tar.TypeReg {
			h.Size = int64(len(f.Body))
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			if _, err := tw.Write([]byte(f.Body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// PushImage stores a multi-platform image whose every platform has the
// given layers, plus an attestation manifest like the ones BuildKit
// attaches, and returns the index digest.
func (r *Registry) PushImage(t *testing.T, platforms []string, layers ...[]byte) string {
	t.Helper()
	type desc struct {
		MediaType   string            `json:"mediaType"`
		Digest      string            `json:"digest"`
		Size        int               `json:"size"`
		Platform    map[string]string `json:"platform,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
	}
	var layerDescs []desc
	for _, l := range layers {
		layerDescs = append(layerDescs, desc{MediaType: "application/vnd.oci.image.layer.v1.tar+gzip", Digest: r.put(l, ""), Size: len(l)})
	}
	cfg := []byte(`{"architecture":"amd64","os":"linux"}`)
	cfgDesc := desc{MediaType: "application/vnd.oci.image.config.v1+json", Digest: r.put(cfg, ""), Size: len(cfg)}

	var manifests []desc
	for _, p := range platforms {
		osName, arch, _ := strings.Cut(p, "/")
		m := r.mustJSON(t, map[string]any{
			"schemaVersion": 2,
			"mediaType":     "application/vnd.oci.image.manifest.v1+json",
			"config":        cfgDesc,
			"layers":        layerDescs,
		})
		manifests = append(manifests, desc{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Digest:    r.put(m, "application/vnd.oci.image.manifest.v1+json"),
			Size:      len(m),
			Platform:  map[string]string{"os": osName, "architecture": arch},
		})
	}
	att := r.mustJSON(t, map[string]any{"schemaVersion": 2, "mediaType": "application/vnd.oci.image.manifest.v1+json", "config": cfgDesc, "layers": []desc{}})
	manifests = append(manifests, desc{
		MediaType:   "application/vnd.oci.image.manifest.v1+json",
		Digest:      r.put(att, "application/vnd.oci.image.manifest.v1+json"),
		Size:        len(att),
		Platform:    map[string]string{"os": "unknown", "architecture": "unknown"},
		Annotations: map[string]string{"vnd.docker.reference.type": "attestation-manifest"},
	})
	index := r.mustJSON(t, map[string]any{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.oci.image.index.v1+json",
		"manifests":     manifests,
	})
	return r.put(index, "application/vnd.oci.image.index.v1+json")
}

// Requests is how many requests asked for path so far.
func (r *Registry) Requests(path string) int64 {
	v, ok := r.requests.Load(path)
	if !ok {
		return 0
	}
	return v.(*atomic.Int64).Load()
}

func (r *Registry) mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (r *Registry) put(body []byte, manifestType string) string {
	d := Digest(body)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blobs[d] = body
	if manifestType != "" {
		r.types[d] = manifestType
	}
	return d
}

func (r *Registry) serve(w http.ResponseWriter, req *http.Request) {
	r.Total.Add(1)
	if req.URL.Path == "/token" {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": r.Token})
		return
	}
	if r.Token != "" && req.Header.Get("Authorization") != "Bearer "+r.Token {
		w.Header().Set("WWW-Authenticate", `Bearer realm="`+r.URL+`/token",service="test",scope="repository:x:pull"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	v, _ := r.requests.LoadOrStore(req.URL.Path, new(atomic.Int64))
	attempt := int(v.(*atomic.Int64).Add(1))

	i := strings.LastIndex(req.URL.Path, "/")
	d := req.URL.Path[i+1:]
	r.mu.Lock()
	body, ok := r.blobs[d]
	mt := r.types[d]
	r.mu.Unlock()
	if !ok {
		http.NotFound(w, req)
		return
	}
	if r.Tamper != nil {
		body = r.Tamper(req.URL.Path, attempt, body)
	}
	if strings.Contains(req.URL.Path, "/manifests/") {
		if r.ContentType != "" {
			mt = r.ContentType
		}
		w.Header().Set("Content-Type", mt)
	}
	_, _ = w.Write(body)
}
