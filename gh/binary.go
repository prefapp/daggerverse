package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	dst := fmt.Sprintf("./gh_%s_%s_%s", version, goos, goarch)

	fmt.Printf(
		"Getting gh version %s for OS %s and architecture %s. URL: %s\n",
		version, goos, goarch, url,
	)

	bearerToken, err := token.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	bearer := fmt.Sprintf("Bearer %s", strings.TrimSpace(bearerToken))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error on response.\n[ERROR] - %s\n", err)
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp(".", "test")
	if err != nil {
		log.Fatal("cannot open temp file", err)
	}
	defer f.Close()
	io.Copy(f, resp.Body)

	err = b.ungzip(dst, f.Name(), 0755)
	if err != nil {
		return nil, err
	}

	return dag.CurrentModule().WorkdirFile(path.Join(dst, "bin/gh")), nil
}

func (b GHBinary) ungzip(dst, src string, umask os.FileMode) error {
	// If we're going into a directory we should make that first
	gzipDst := "./ungzip_file"
	if err := os.MkdirAll(filepath.Dir(gzipDst), umask); err != nil {
		return err
	}

	// File first
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// gzip compression is second
	gzipR, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gzipR.Close() }()

	// Explicitly chmod; the process umask is unconditionally applied otherwise.
	// We'll mask the mode with our own umask, but that may be different than
	// the process umask
	return untar(gzipR, ".", "gzip", true, 0755)
}

func untar(input io.Reader, dst, src string, dir bool, umask os.FileMode) error {
	tarR := tar.NewReader(input)
	done := false
	dirHdrs := []*tar.Header{}
	now := time.Now()

	for {
		hdr, err := tarR.Next()
		if err == io.EOF {
			if !done {
				// Empty archive
				return fmt.Errorf("empty archive: %s", src)
			}

			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			// don't unpack extended headers as files
			continue
		}

		path := dst
		if dir {
			path = filepath.Join(path, hdr.Name)
		}

		fileInfo := hdr.FileInfo()

		if fileInfo.IsDir() {
			if !dir {
				return fmt.Errorf("expected a single file: %s", src)
			}

			// A directory, just make the directory and continue unarchiving...
			if err := os.MkdirAll(path, umask); err != nil {
				return err
			}

			// Record the directory information so that we may set its attributes
			// after all files have been extracted
			dirHdrs = append(dirHdrs, hdr)

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

		// We have a file. If we already decoded, then it is an error
		if !dir && done {
			return fmt.Errorf("expected a single file, got multiple: %s", src)
		}

		// Mark that we're done so future in single file mode errors
		done = true

		dstF, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, umask)
		if err != nil {
			return err
		}
		defer func() { _ = dstF.Close() }()

		_, err = io.Copy(dstF, tarR)
		if err != nil {
			return err
		}

		// Set the access and modification time if valid, otherwise default to current time
		aTime := now
		mTime := now
		if hdr.AccessTime.Unix() > 0 {
			aTime = hdr.AccessTime
		}
		if hdr.ModTime.Unix() > 0 {
			mTime = hdr.ModTime
		}
		if err := os.Chtimes(path, aTime, mTime); err != nil {
			return err
		}
	}

	// Perform a final pass over extracted directories to update metadata
	for _, dirHdr := range dirHdrs {
		path := filepath.Join(dst, dirHdr.Name)
		// Chmod the directory since they might be created before we know the mode flags
		if err := os.Chmod(path, umask); err != nil {
			return err
		}
		// Set the mtime/atime attributes since they would have been changed during extraction
		aTime := now
		mTime := now
		if dirHdr.AccessTime.Unix() > 0 {
			aTime = dirHdr.AccessTime
		}
		if dirHdr.ModTime.Unix() > 0 {
			mTime = dirHdr.ModTime
		}
		if err := os.Chtimes(path, aTime, mTime); err != nil {
			return err
		}
	}

	return nil
}
