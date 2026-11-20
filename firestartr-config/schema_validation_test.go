package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSchemaVersionSelection(t *testing.T) {
	t.Run("Default version is v3.2.0", func(t *testing.T) {
		if DefaultSchemaVersion != "v3.2.0" {
			t.Fatalf("expected default schema version v3.2.0, got %s", DefaultSchemaVersion)
		}
	})

	t.Run("Accepts explicit pinned version from schema directory", func(t *testing.T) {
		p, err := newSchemaProvider("v3.2.0", "", testSchemaDirectory())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if p.version != "v3.2.0" {
			t.Fatalf("expected v3.2.0, got %s", p.version)
		}
		source := p.schemaSource("apps")
		if source != "v3.2.0/apps.schema.json" {
			t.Fatalf("unexpected source: %s", source)
		}
	})

	t.Run("Normalizes npm semver version with leading v", func(t *testing.T) {
		p, err := newSchemaProvider("v3.2.0", "npm", nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if p.npmVersion != "3.2.0" {
			t.Fatalf("expected npmVersion 3.2.0, got %s", p.npmVersion)
		}
	})

	t.Run("Accepts explicit latest from schema directory", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/dot-firestartr/v3.2.0/apps.schema.json")
		if err != nil {
			t.Fatal(err)
		}
		dir := dag.Directory().WithNewFile("latest/apps.schema.json", string(data))
		p, err := newSchemaProvider("latest", "", dir)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if p.version != "latest" {
			t.Fatalf("expected latest, got %s", p.version)
		}
		if _, err := p.loadSchema(context.Background(), "apps"); err != nil {
			t.Fatalf("expected to load schema for latest, got: %s", err)
		}
	})

	t.Run("Rejects invalid version", func(t *testing.T) {
		_, err := newSchemaProvider("v3.2.0/../../x", "", testSchemaDirectory())
		if err == nil {
			t.Fatal("expected error for invalid version")
		}
	})

	t.Run("Rejects caller-host path as schema base without directory", func(t *testing.T) {
		_, err := newSchemaProvider("v3.2.0", "./testdata/dot-firestartr", nil)
		if err == nil {
			t.Fatal("expected error for unsupported host path base")
		}
		if !strings.Contains(err.Error(), "schemaDirectory") {
			t.Fatalf("expected guidance to use schemaDirectory, got: %s", err)
		}
	})
}

func buildSchemasTarball(t *testing.T, contracts map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range contracts {
		hdr := &tar.Header{
			Name: "package/dot-firestartr/" + name + ".schema.json",
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
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

func loadTestSchemas(t *testing.T) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, c := range contractNames {
		b, err := os.ReadFile(filepath.Join("./testdata/dot-firestartr/v3.2.0", c+".schema.json"))
		if err != nil {
			t.Fatal(err)
		}
		out[c] = b
	}
	return out
}

type npmTestServer struct {
	server      *httptest.Server
	metaHits    int32
	tarballHits int32
}

func newNPMTestServer(t *testing.T, npmVersion, resolvedVersion string, tarball []byte, integrity string, shasum string, statusOverride map[string]int, metaOverride func(string) (int, []byte)) *npmTestServer {
	t.Helper()
	s := &npmTestServer{}
	mux := http.NewServeMux()
	pkgPath := "/" + encodeNPMPackage(DefaultSchemaPackage) + "/" + npmVersion
	tarballPath := "/pkg.tgz"

	mux.HandleFunc(pkgPath, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&s.metaHits, 1)
		if status, ok := statusOverride["metadata"]; ok {
			http.Error(w, "boom", status)
			return
		}
		if metaOverride != nil {
			code, body := metaOverride(s.server.URL + tarballPath)
			if code != 0 && code != 200 {
				w.WriteHeader(code)
				_, _ = w.Write(body)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
			return
		}
		meta := map[string]interface{}{
			"version": resolvedVersion,
			"dist": map[string]string{
				"tarball":   s.server.URL + tarballPath,
				"integrity": integrity,
				"shasum":    shasum,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
	})

	mux.HandleFunc(tarballPath, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&s.tarballHits, 1)
		if status, ok := statusOverride["tarball"]; ok {
			http.Error(w, "boom", status)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(tarball)
	})

	s.server = httptest.NewServer(mux)
	return s
}

func sriSHA512(data []byte) string {
	sum := sha512.Sum512(data)
	return "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
}

func sha1Hex(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

func TestSchemaNPMResolution(t *testing.T) {
	schemas := loadTestSchemas(t)
	tarball := buildSchemasTarball(t, schemas)

	t.Run("Resolves and loads all contracts from npm registry", func(t *testing.T) {
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, sriSHA512(tarball), "", nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, contract := range contractNames {
			if _, err := p.loadSchema(context.Background(), contract); err != nil {
				t.Fatalf("failed to load schema for %s: %s", contract, err)
			}
		}
	})

	t.Run("Supports leading v and dist-tag resolution", func(t *testing.T) {
		s := newNPMTestServer(t, "latest", "3.2.0", tarball, sriSHA512(tarball), "", nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("latest", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.loadSchema(context.Background(), "apps"); err != nil {
			t.Fatalf("failed to load schema for latest: %s", err)
		}
		if got := atomic.LoadInt32(&s.metaHits); got != 1 {
			t.Fatalf("expected 1 metadata hit, got %d", got)
		}
	})

	t.Run("Verifies shasum when integrity missing", func(t *testing.T) {
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, "", sha1Hex(tarball), nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.loadSchema(context.Background(), "apps"); err != nil {
			t.Fatalf("expected load with shasum, got: %s", err)
		}
	})

	t.Run("Rejects integrity mismatch", func(t *testing.T) {
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, sriSHA512([]byte("tampered")), "", nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "apps")
		if err == nil {
			t.Fatal("expected integrity error")
		}
		msg := err.Error()
		if !strings.Contains(msg, "integrity") || !strings.Contains(msg, DefaultSchemaPackage) {
			t.Fatalf("expected actionable integrity diagnostics, got: %s", msg)
		}
	})

	t.Run("Rejects missing contract in tarball", func(t *testing.T) {
		partial := map[string][]byte{}
		for k, v := range schemas {
			if k == "apps" {
				continue
			}
			partial[k] = v
		}
		badTar := buildSchemasTarball(t, partial)
		s := newNPMTestServer(t, "3.2.0", "3.2.0", badTar, sriSHA512(badTar), "", nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "platforms")
		if err == nil || !strings.Contains(err.Error(), "apps.schema.json") {
			t.Fatalf("expected missing contract error, got: %v", err)
		}
	})

	t.Run("Rejects missing dist.tarball in metadata", func(t *testing.T) {
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, sriSHA512(tarball), "", nil, func(tarURL string) (int, []byte) {
			b, _ := json.Marshal(map[string]interface{}{
				"version": "3.2.0",
				"dist":    map[string]string{},
			})
			return 200, b
		})
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "apps")
		if err == nil || !strings.Contains(err.Error(), "dist.tarball") {
			t.Fatalf("expected missing dist.tarball error, got: %v", err)
		}
	})
}

func TestSchemaFetchFailures(t *testing.T) {
	t.Run("Reports actionable diagnostics on npm metadata HTTP failure", func(t *testing.T) {
		schemas := loadTestSchemas(t)
		tarball := buildSchemasTarball(t, schemas)
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, sriSHA512(tarball), "", map[string]int{"metadata": 404}, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "apps")
		if err == nil {
			t.Fatal("expected fetch error")
		}
		msg := err.Error()
		for _, want := range []string{"apps", "v3.2.0", s.server.URL, "404", DefaultSchemaPackage} {
			if !strings.Contains(msg, want) {
				t.Errorf("expected %q in error, got: %s", want, msg)
			}
		}
	})

	t.Run("Reports actionable diagnostics on unreachable registry", func(t *testing.T) {
		s := newNPMTestServer(t, "3.2.0", "3.2.0", []byte{}, "", "", nil, nil)
		base := s.server.URL
		s.server.Close()

		p, err := newSchemaProvider("v3.2.0", base, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "platforms")
		if err == nil {
			t.Fatal("expected fetch error")
		}
		msg := err.Error()
		if !strings.Contains(msg, "platforms") || !strings.Contains(msg, base) {
			t.Errorf("expected actionable diagnostics, got: %s", msg)
		}
	})
}

func TestSchemaDirectorySource(t *testing.T) {
	t.Run("Loads from provided schema directory via Dagger API", func(t *testing.T) {
		p, err := newSchemaProvider("v3.2.0", "", testSchemaDirectory())
		if err != nil {
			t.Fatal(err)
		}
		for _, contract := range contractNames {
			if _, err := p.loadSchema(context.Background(), contract); err != nil {
				t.Fatalf("failed to load directory schema for %s: %s", contract, err)
			}
		}
	})

	t.Run("Directory takes precedence over schema base", func(t *testing.T) {
		p, err := newSchemaProvider("v3.2.0", "https://registry.invalid.example", testSchemaDirectory())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.loadSchema(context.Background(), "apps"); err != nil {
			t.Fatalf("expected directory source to be used, got: %s", err)
		}
	})

	t.Run("Reports actionable diagnostics when schema missing in directory", func(t *testing.T) {
		dir := dag.Directory().WithNewFile("v3.2.0/platforms.schema.json", "{}")
		p, err := newSchemaProvider("v3.2.0", "", dir)
		if err != nil {
			t.Fatal(err)
		}
		_, err = p.loadSchema(context.Background(), "apps")
		if err == nil {
			t.Fatal("expected missing schema error")
		}
		msg := err.Error()
		for _, want := range []string{"apps", "v3.2.0", "schema directory", "apps.schema.json"} {
			if !strings.Contains(msg, want) {
				t.Errorf("expected %q in error, got: %s", want, msg)
			}
		}
	})
}

func TestSchemaOfflineAndCaching(t *testing.T) {
	t.Run("Loads from schema directory without network", func(t *testing.T) {
		p, err := newSchemaProvider("v3.2.0", "", testSchemaDirectory())
		if err != nil {
			t.Fatal(err)
		}
		for _, contract := range contractNames {
			if _, err := p.loadSchema(context.Background(), contract); err != nil {
				t.Fatalf("failed to load schema for %s: %s", contract, err)
			}
		}
	})

	t.Run("Caches npm tarball resolution per provider load", func(t *testing.T) {
		schemas := loadTestSchemas(t)
		tarball := buildSchemasTarball(t, schemas)
		s := newNPMTestServer(t, "3.2.0", "3.2.0", tarball, sriSHA512(tarball), "", nil, nil)
		defer s.server.Close()

		p, err := newSchemaProvider("v3.2.0", s.server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 3; i++ {
			if _, err := p.loadSchema(context.Background(), "apps"); err != nil {
				t.Fatalf("load failed: %s", err)
			}
			if _, err := p.loadSchema(context.Background(), "platforms"); err != nil {
				t.Fatalf("load failed: %s", err)
			}
		}
		if got := atomic.LoadInt32(&s.metaHits); got != 1 {
			t.Fatalf("expected 1 metadata fetch, got %d", got)
		}
		if got := atomic.LoadInt32(&s.tarballHits); got != 1 {
			t.Fatalf("expected 1 tarball fetch, got %d", got)
		}
	})
}

func TestSchemaValidationDiagnostics(t *testing.T) {
	p, err := newSchemaProvider("v3.2.0", "", testSchemaDirectory())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Valid document passes", func(t *testing.T) {
		_, err := p.validateDocument(context.Background(), "apps", "apps/sample.yaml", "name: x\nstate_repo: r\nservices: []\n")
		if err != nil {
			t.Fatalf("expected valid, got: %s", err)
		}
	})

	t.Run("Invalid document reports file, contract, version and reason", func(t *testing.T) {
		_, err := p.validateDocument(context.Background(), "docker_registries", "docker_registries/bad.yaml", "name: x\nregistry: example.com:5000\nimage_types: [a]\ndefault: true\nauth_strategy: none\nbase_paths:\n  services: s\n")
		if err == nil {
			t.Fatal("expected validation error")
		}
		msg := err.Error()
		for _, want := range []string{"docker_registries/bad.yaml", "docker_registries", "v3.2.0", "schema directory"} {
			if !strings.Contains(msg, want) {
				t.Errorf("expected %q in error, got: %s", want, msg)
			}
		}
	})
}
