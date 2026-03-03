// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	getter "github.com/hashicorp/go-getter"
)

func main() {
	modeRaw := flag.String("mode", "any", "get mode (any, file, dir)")
	progress := flag.Bool("progress", false, "display terminal progress")
	insecure := flag.Bool("insecure", false, "do not verify server's certificate chain (not recommended)")
	allowLocal := flag.Bool("allow-local", false, "allow local file:// access (WARNING: enables local file inclusion attacks if input is untrusted)")
	allowRepo := flag.Bool("allow-repo", false, "allow git:// access (WARNING: enables arbitrary git repository directory reads)")
	useNetrc := flag.Bool("use-netrc", false, "use .netrc for HTTP authentication (WARNING: .netrc contains credentials in plain text)")
	maxSizeMB := flag.Int64("max-download-size-mb", 10240, "maximum download size in MB (0 = unlimited, default 10GB)")
	headTimeoutSec := flag.Int64("head-timeout-sec", 30, "timeout for HEAD requests in seconds (0 = no timeout)")
	readTimeoutSec := flag.Int64("read-timeout-sec", 300, "timeout for reading HTTP body in seconds (0 = no timeout)")
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		log.Fatalf("Expected two args: URL and dst")
		os.Exit(1)
	}

	// Get the mode
	var mode getter.ClientMode
	switch *modeRaw {
	case "any":
		mode = getter.ClientModeAny
	case "file":
		mode = getter.ClientModeFile
	case "dir":
		mode = getter.ClientModeDir
	default:
		log.Fatalf("Invalid client mode, must be 'any', 'file', or 'dir': %s", *modeRaw)
		os.Exit(1)
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	opts := []getter.ClientOption{}
	if *progress {
		opts = append(opts, getter.WithProgress(defaultProgressBar))
	}

	if *insecure {
		log.Println("WARNING: Using Insecure TLS transport!")
		opts = append(opts, getter.WithInsecure())
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Calculate max download size from flag, converting MB to bytes
	var maxDownloadSize int64
	if *maxSizeMB > 0 {
		maxDownloadSize = *maxSizeMB * 1024 * 1024
	}
	// Zero means unlimited

	// Convert timeout flags to time.Duration with mandatory minimums
	var headFirstTimeout, readTimeout time.Duration
	if *headTimeoutSec > 0 {
		headFirstTimeout = time.Duration(*headTimeoutSec) * time.Second
	} else {
		headFirstTimeout = 30 * time.Second // Mandatory minimum
	}
	if *readTimeoutSec > 0 {
		readTimeout = time.Duration(*readTimeoutSec) * time.Second
	} else {
		readTimeout = 300 * time.Second // Mandatory minimum
	}

	// Ensure maxDownloadSize has a safe default
	if maxDownloadSize == 0 {
		maxDownloadSize = 10 * 1024 * 1024 * 1024 // 10GB default
	}

	// Build getters map with security-conscious defaults
	// All HttpGetter instances have MaxBytes, HeadFirstTimeout, and ReadTimeout set for security
	getters := map[string]getter.Getter{
		"http": &getter.HttpGetter{ // secsync:ignore go/sast/internal/go-getter
			Netrc:            *useNetrc,
			MaxBytes:         maxDownloadSize,  // X-Terraform-Get limit for DoS prevention
			HeadFirstTimeout: headFirstTimeout, // HEAD request timeout
			ReadTimeout:      readTimeout,      // Body read timeout
		},
		"https": &getter.HttpGetter{ // secsync:ignore go/sast/internal/go-getter
			Netrc:            *useNetrc,
			MaxBytes:         maxDownloadSize,  // X-Terraform-Get limit for DoS prevention
			HeadFirstTimeout: headFirstTimeout, // HEAD request timeout
			ReadTimeout:      readTimeout,      // Body read timeout
		},
		"s3": new(getter.S3Getter),
	}

	// Only enable git repository access if explicitly requested
	if *allowRepo {
		getters["git"] = new(getter.GitGetter) // secsync:ignore go/sast/internal/go-getter
	}

	// Only enable local file access if explicitly requested
	if *allowLocal {
		getters["file"] = new(getter.FileGetter) // secsync:ignore go/sast/internal/go-getter
	}

	// Warn if potentially risky features are enabled
	if *useNetrc {
		log.Println("WARNING: Using .netrc for HTTP authentication (credentials stored in plain text)")
	}
	if *allowRepo {
		log.Println("WARNING: Enabling git:// access (enables arbitrary git repository directory reads)")
	}
	if *allowLocal {
		log.Println("WARNING: Enabling file:// access (enables local file inclusion attacks if input is untrusted)")
	}

	// Build the client with explicitly configured getters for security
	client := &getter.Client{
		Ctx:             ctx,
		Src:             args[0],
		Dst:             args[1],
		Pwd:             pwd,
		Mode:            mode,
		Options:         opts,
		DisableSymlinks: true, // Prevent symlink traversal attacks
		Getters:         getters,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Printf("signal %v", sig)
	case <-ctx.Done():
		wg.Wait()
		log.Printf("success!")
	case err := <-errChan:
		wg.Wait()
		log.Fatalf("Error downloading: %s", err)
	}
}
