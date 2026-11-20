package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

const (
	DefaultSchemaVersion = "v3.2.0"

	// DefaultSchemaPackage is the npm package that publishes dot-firestartr schemas.
	DefaultSchemaPackage = "@firestartr/firestartr-claims_schemas"
	// DefaultSchemaRegistry is the default npm registry base URL.
	DefaultSchemaRegistry = "https://registry.npmjs.org"
)

var contractNames = []string{"apps", "platforms", "docker_registries", "helm_repositories"}

type schemaSourceKind string

const (
	schemaSourceNPM   schemaSourceKind = "npm"
	schemaSourceLocal schemaSourceKind = "local"
)

type schemaProvider struct {
	kind schemaSourceKind

	// requested is the version/tag as provided (after defaulting).
	requested string
	// npmVersion is the version/tag sent to the npm registry (e.g. v3.2.0 -> 3.2.0).
	npmVersion string
	// version is the display/request key (requested).
	version string

	registry  string
	pkg       string
	localBase string

	client *http.Client

	mu          sync.Mutex
	loaded      bool
	loadErr     error
	raw         map[string][]byte
	resolved    string
	metadataURL string
	tarballURL  string

	cache map[string]*gojsonschema.Schema
}

func newSchemaProvider(version, base string) (*schemaProvider, error) {
	if version == "" {
		version = DefaultSchemaVersion
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("schema version must not be empty")
	}
	if strings.ContainsAny(version, " \t\n/\\") {
		return nil, fmt.Errorf("invalid schema version %q", version)
	}

	p := &schemaProvider{
		requested: version,
		version:   version,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string]*gojsonschema.Schema),
		raw:   make(map[string][]byte),
	}

	base = strings.TrimSpace(base)
	switch {
	case base == "" || base == "npm" || base == "npm:":
		p.kind = schemaSourceNPM
		p.registry = DefaultSchemaRegistry
		p.pkg = DefaultSchemaPackage
		p.npmVersion = normalizeNPMVersion(version)
	case isLocalSchemaBase(base):
		p.kind = schemaSourceLocal
		p.localBase = strings.TrimPrefix(base, "file://")
		if strings.TrimSpace(p.localBase) == "" {
			return nil, fmt.Errorf("schema base path must not be empty")
		}
	case strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://"):
		// Treat HTTP(S) base as an npm registry base. Docs raw URLs are no longer
		// the authoritative/published source of truth.
		p.kind = schemaSourceNPM
		p.registry = strings.TrimRight(base, "/")
		p.pkg = DefaultSchemaPackage
		p.npmVersion = normalizeNPMVersion(version)
	default:
		return nil, fmt.Errorf("unsupported schema base %q: use an npm registry URL (or empty for default npm), or a local path", base)
	}

	return p, nil
}

func isLocalSchemaBase(base string) bool {
	if strings.HasPrefix(base, "file://") {
		return true
	}
	if strings.HasPrefix(base, "/") || strings.HasPrefix(base, "./") || strings.HasPrefix(base, "../") {
		return true
	}
	if strings.Contains(base, `\`) {
		return true
	}
	// If it doesn't look like a URL scheme, treat as local path for offline overrides.
	if !strings.Contains(base, "://") {
		return true
	}
	return false
}

func normalizeNPMVersion(v string) string {
	// npm semver versions do not use a leading "v"; dist-tags (e.g. latest) pass through.
	if strings.HasPrefix(v, "v") && len(v) > 1 {
		if v[1] >= '0' && v[1] <= '9' {
			return strings.TrimPrefix(v, "v")
		}
	}
	return v
}

func encodeNPMPackage(name string) string {
	// Scoped packages must have '/' encoded as %2f in registry URLs.
	return strings.ReplaceAll(name, "/", "%2f")
}

func (p *schemaProvider) schemaSource(contract string) string {
	switch p.kind {
	case schemaSourceLocal:
		return filepath.Join(p.localBase, p.version, contract+".schema.json")
	case schemaSourceNPM:
		p.mu.Lock()
		resolved := p.resolved
		tarball := p.tarballURL
		p.mu.Unlock()
		ver := p.npmVersion
		if resolved != "" {
			ver = resolved
		}
		loc := fmt.Sprintf("npm:%s@%s", p.pkg, ver)
		if tarball != "" {
			loc += " (" + tarball + ")"
		}
		return fmt.Sprintf("%s#dot-firestartr/%s.schema.json", loc, contract)
	default:
		return contract
	}
}

func (p *schemaProvider) sourceDescription() string {
	switch p.kind {
	case schemaSourceLocal:
		return fmt.Sprintf("local base %s (version %s)", p.localBase, p.version)
	case schemaSourceNPM:
		p.mu.Lock()
		resolved := p.resolved
		meta := p.metadataURL
		tar := p.tarballURL
		p.mu.Unlock()
		desc := fmt.Sprintf("npm package %s (requested %s", p.pkg, p.requested)
		if resolved != "" {
			desc += ", resolved " + resolved
		} else {
			desc += ", registry version/tag " + p.npmVersion
		}
		desc += ", registry " + p.registry
		if meta != "" {
			desc += ", metadata " + meta
		}
		if tar != "" {
			desc += ", tarball " + tar
		}
		desc += ")"
		return desc
	default:
		return "unknown schema source"
	}
}

func (p *schemaProvider) loadSchema(ctx context.Context, contract string) (*gojsonschema.Schema, error) {
	if s, ok := p.cache[contract]; ok {
		return s, nil
	}

	data, err := p.contractBytes(ctx, contract)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema for contract %q (%s): %w", contract, p.sourceDescription(), err)
	}

	data = normalizeDraftSchema(data)

	loader := gojsonschema.NewBytesLoader(data)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema for contract %q (%s): %w", contract, p.sourceDescription(), err)
	}
	p.cache[contract] = schema
	return schema, nil
}

func (p *schemaProvider) contractBytes(ctx context.Context, contract string) ([]byte, error) {
	switch p.kind {
	case schemaSourceLocal:
		source := filepath.Join(p.localBase, p.version, contract+".schema.json")
		data, err := os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema for contract %q (version %s, source %s): %w", contract, p.version, source, err)
		}
		return data, nil
	case schemaSourceNPM:
		if err := p.ensureNPMLoaded(ctx); err != nil {
			return nil, err
		}
		p.mu.Lock()
		data, ok := p.raw[contract]
		p.mu.Unlock()
		if !ok {
			return nil, fmt.Errorf("schema for contract %q not found in npm package %s (%s)", contract, p.pkg, p.sourceDescription())
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported schema source")
	}
}

func (p *schemaProvider) ensureNPMLoaded(ctx context.Context) error {
	p.mu.Lock()
	if p.loaded || p.loadErr != nil {
		err := p.loadErr
		p.mu.Unlock()
		return err
	}
	p.mu.Unlock()

	err := p.loadFromNPM(ctx)

	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		p.loadErr = err
		return err
	}
	p.loaded = true
	return nil
}

type npmPackageMetadata struct {
	Version string `json:"version"`
	Dist    struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
		Shasum    string `json:"shasum"`
	} `json:"dist"`
}

func (p *schemaProvider) loadFromNPM(ctx context.Context) error {
	metadataURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(p.registry, "/"), encodeNPMPackage(p.pkg), p.npmVersion)
	p.mu.Lock()
	p.metadataURL = metadataURL
	p.mu.Unlock()

	metaBytes, err := p.fetchHTTP(ctx, metadataURL)
	if err != nil {
		return fmt.Errorf("failed to resolve npm schema package %s@%s from registry %s (metadata %s): %w", p.pkg, p.requested, p.registry, metadataURL, err)
	}

	var meta npmPackageMetadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return fmt.Errorf("failed to parse npm package metadata for %s@%s (metadata %s): %w", p.pkg, p.requested, metadataURL, err)
	}
	if meta.Dist.Tarball == "" {
		return fmt.Errorf("npm package metadata for %s@%s (metadata %s) is missing dist.tarball", p.pkg, p.requested, metadataURL)
	}

	p.mu.Lock()
	p.resolved = meta.Version
	p.tarballURL = meta.Dist.Tarball
	p.mu.Unlock()

	tarball, err := p.fetchHTTP(ctx, meta.Dist.Tarball)
	if err != nil {
		return fmt.Errorf("failed to download npm schema tarball for %s@%s (tarball %s): %w", p.pkg, displayResolved(p.requested, meta.Version), meta.Dist.Tarball, err)
	}

	if err := verifyNPMIntegrity(tarball, meta.Dist.Integrity, meta.Dist.Shasum); err != nil {
		return fmt.Errorf("failed to verify npm schema tarball for %s@%s (tarball %s): %w", p.pkg, displayResolved(p.requested, meta.Version), meta.Dist.Tarball, err)
	}

	schemas, err := extractDotFirestartrSchemas(tarball)
	if err != nil {
		return fmt.Errorf("failed to extract dot-firestartr schemas from npm package %s@%s (tarball %s): %w", p.pkg, displayResolved(p.requested, meta.Version), meta.Dist.Tarball, err)
	}

	for _, contract := range contractNames {
		if _, ok := schemas[contract]; !ok {
			return fmt.Errorf("npm package %s@%s (tarball %s) is missing contract schema dot-firestartr/%s.schema.json", p.pkg, displayResolved(p.requested, meta.Version), meta.Dist.Tarball, contract)
		}
	}

	p.mu.Lock()
	for k, v := range schemas {
		p.raw[k] = v
	}
	p.mu.Unlock()

	return nil
}

func displayResolved(requested, resolved string) string {
	if resolved == "" || resolved == requested || resolved == normalizeNPMVersion(requested) {
		return requested
	}
	return fmt.Sprintf("%s (resolved %s)", requested, resolved)
}

func (p *schemaProvider) fetchHTTP(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if len(body) > 0 {
			return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func verifyNPMIntegrity(data []byte, integrity, shasum string) error {
	integrity = strings.TrimSpace(integrity)
	if integrity != "" {
		return verifySRI(data, integrity)
	}
	shasum = strings.TrimSpace(shasum)
	if shasum == "" {
		return fmt.Errorf("missing dist.integrity and dist.shasum")
	}
	sum := sha1.Sum(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, shasum) {
		return fmt.Errorf("sha1 integrity mismatch: expected %s, got %s", shasum, got)
	}
	return nil
}

func verifySRI(data []byte, integrity string) error {
	parts := strings.Fields(integrity)
	if len(parts) == 0 {
		return fmt.Errorf("empty integrity value")
	}
	var verified bool
	var errs []string
	for _, part := range parts {
		alg, val, ok := strings.Cut(part, "-")
		if !ok || val == "" {
			errs = append(errs, fmt.Sprintf("malformed integrity entry %q", part))
			continue
		}
		switch strings.ToLower(alg) {
		case "sha512":
			h := sha512.Sum512(data)
			if base64.StdEncoding.EncodeToString(h[:]) == val || base64.RawStdEncoding.EncodeToString(h[:]) == val {
				verified = true
			} else {
				errs = append(errs, "sha512 mismatch")
			}
		case "sha256":
			h := sha256Bytes(data)
			if base64.StdEncoding.EncodeToString(h) == val || base64.RawStdEncoding.EncodeToString(h) == val {
				verified = true
			} else {
				errs = append(errs, "sha256 mismatch")
			}
		case "sha1":
			h := sha1.Sum(data)
			if base64.StdEncoding.EncodeToString(h[:]) == val || base64.RawStdEncoding.EncodeToString(h[:]) == val {
				verified = true
			} else {
				errs = append(errs, "sha1 mismatch")
			}
		default:
			errs = append(errs, fmt.Sprintf("unsupported integrity algorithm %q", alg))
		}
	}
	if verified {
		return nil
	}
	if len(errs) == 0 {
		return fmt.Errorf("no supported integrity hashes found")
	}
	return fmt.Errorf("integrity verification failed: %s", strings.Join(errs, "; "))
}

func extractDotFirestartrSchemas(tarball []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	out := map[string][]byte{}
	prefix := "package/dot-firestartr/"

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr == nil {
			continue
		}
		name := strings.TrimPrefix(hdr.Name, "./")
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".schema.json") {
			continue
		}
		contract := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".schema.json")
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != 0 {
			continue
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}
		for _, c := range contractNames {
			if c == contract {
				out[contract] = b
				break
			}
		}
	}

	return out, nil
}

func normalizeDraftSchema(data []byte) []byte {
	// The published contracts may declare draft 2020-12. gojsonschema v1.2.0
	// validates draft-04/06/07; the constraints used by these contracts are
	// compatible, so normalize the meta-schema identifier for validation.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}
	if v, ok := raw["$schema"].(string); ok && strings.Contains(v, "draft/2020-12") {
		raw["$schema"] = "http://json-schema.org/draft-07/schema#"
		if b, mErr := json.Marshal(raw); mErr == nil {
			return b
		}
	}
	return data
}

func (p *schemaProvider) validateDocument(ctx context.Context, contract string, filePath string, yamlContent string) (interface{}, error) {
	schema, err := p.loadSchema(ctx, contract)
	if err != nil {
		return nil, err
	}

	var doc interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s (%s schema, %s): %w", filePath, contract, p.sourceDescription(), err)
	}
	if doc == nil {
		return nil, fmt.Errorf("invalid configuration in %s (%s schema, %s): document is empty", filePath, contract, p.sourceDescription())
	}

	docLoader := gojsonschema.NewGoLoader(doc)
	result, err := schema.Validate(docLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to validate %s against %s schema (%s): %w", filePath, contract, p.sourceDescription(), err)
	}
	if !result.Valid() {
		errs := result.Errors()
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, fmt.Sprintf("- %s: %s", e.Field(), e.Description()))
		}
		return nil, fmt.Errorf("configuration %s is invalid against %s schema (%s):\n%s", filePath, contract, p.sourceDescription(), strings.Join(msgs, "\n"))
	}

	return doc, nil
}

func listConfigFiles(ctx context.Context, dir interface {
	Glob(context.Context, string) ([]string, error)
}, subdir string) ([]string, error) {
	files := []string{}
	seen := map[string]bool{}
	for _, ext := range []string{".yaml", ".yml"} {
		paths, err := dir.Glob(ctx, filepath.ToSlash(filepath.Join(subdir, "*"+ext)))
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			if seen[p] {
				continue
			}
			seen[p] = true
			files = append(files, p)
		}
	}
	sort.Strings(files)
	return files, nil
}

func sha256Bytes(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
