package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"gh/internal/dagger"

	"github.com/google/go-github/v61/github"
	"github.com/samber/lo"
)

// GHBinary is the configuration for the Github CLI binary.
type GHBinary struct {
	// Version of the Github CLI
	Version string

	// Operating system of the Github CLI
	GOOS string

	// Architecture of the Github CLI
	GOARCH string
}

// WithVersion returns the GHBinary with the given version.
func (b GHBinary) WithVersion(version string) GHBinary {
	return GHBinary{
		Version: version,
		GOOS:    b.GOOS,
		GOARCH:  b.GOARCH,
	}
}

// WithArch returns the GHBinary with the given architecture.
func (b GHBinary) WithArch(goarch string) GHBinary {
	return GHBinary{
		Version: b.Version,
		GOOS:    b.GOOS,
		GOARCH:  goarch,
	}
}

// WithOS returns the GHBinary with the given operating system.
func (b GHBinary) WithOS(goos string) GHBinary {
	return GHBinary{
		Version: b.Version,
		GOOS:    goos,
		GOARCH:  b.GOARCH,
	}
}

// getLatestCliVersion returns the latest version of the Github CLI.
func (b GHBinary) getLatestCliVersion(ctx context.Context) (string, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, "cli", "cli")
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}

// binary returns the Github CLI binary.
func (b GHBinary) binary(ctx context.Context, runnerGh *dagger.File, token *dagger.Secret) (*dagger.File, error) {
	if runnerGh != nil {
		fmt.Printf("Using specified local gh binary\n")
		return runnerGh, nil
	}

	if b.Version == "" {
		fmt.Printf("Latest version specified, resolving...\n")
		version, err := b.getLatestCliVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest GitHub CLI version: %w", err)
		}

		fmt.Printf("The latest version avaliable is %s\n", version)
		b.Version = version
	}

	var (
		goos    = lo.Ternary(b.GOOS != "", b.GOOS, runtime.GOOS)
		goarch  = lo.Ternary(b.GOARCH != "", b.GOARCH, runtime.GOARCH)
		version = strings.TrimPrefix(b.Version, "v")
		suffix  = lo.Ternary(goos == "linux", "tar.gz", "zip")
	)

	if goos != "linux" && goos != "darwin" {
		return nil, fmt.Errorf("unsupported operating system: %s", goos)
	}

	// github releases use "macOS" instead of "darwin"
	if goos == "darwin" {
		goos = "macOS"
	}

	url := fmt.Sprintf(
		"https://github.com/cli/cli/releases/download/v%s/gh_%s_%s_%s.%s",
		version, version, goos, goarch, suffix,
	)
	dst := fmt.Sprintf("gh_%s_%s_%s", version, goos, goarch)

	bearerToken, err := token.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Printf(
		"Getting gh version %s for OS %s and architecture %s (authenticating with --token param value). URL: %s\n",
		version, goos, goarch, url,
	)

	bearer := fmt.Sprintf("Bearer %s", strings.TrimSpace(bearerToken))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error on response.\n[ERROR] - %s\n", err)
	}
	defer resp.Body.Close()

	err = b.untargz(resp.Body, 0755)
	if err != nil {
		return nil, err
	}

	return dag.CurrentModule().WorkdirFile(path.Join(dst, "bin/gh")), nil
}

func (b GHBinary) untargz(src io.Reader, umask os.FileMode) error {
	fmt.Printf("Ungzipping gh release tar...\n")

	gzipR, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer func() { _ = gzipR.Close() }()

	return untar(gzipR, "gzip", true, umask)
}

func untar(input io.Reader, src string, dir bool, umask os.FileMode) error {
	fmt.Printf("Untaring gh release...\n")

	tarR := tar.NewReader(input)

	for {
		hdr, err := tarR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			// don't unpack extended headers as files
			continue
		}

		path := hdr.Name
		fileInfo := hdr.FileInfo()

		if fileInfo.IsDir() {
			if !dir {
				return fmt.Errorf("expected a single file: %s", src)
			}

			// A directory, just make the directory and continue unarchiving...
			if err := os.MkdirAll(path, umask); err != nil {
				return err
			}

			continue
		} else {
			// There is no ordering guarantee that a file in a directory is
			// listed before the directory
			dstPath := filepath.Dir(path)

			// Check that the directory exists, otherwise create it
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				if err := os.MkdirAll(dstPath, umask); err != nil {
					return err
				}
			}
		}

		dstF, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, umask)
		if err != nil {
			return err
		}
		defer func() { _ = dstF.Close() }()

		_, err = io.Copy(dstF, tarR)
		if err != nil {
			return err
		}
	}

	return nil
}
