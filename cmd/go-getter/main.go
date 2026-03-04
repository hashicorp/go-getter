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

	getter "github.com/hashicorp/go-getter"
)

func main() {
	modeRaw := flag.String("mode", "any", "get mode (any, file, dir)")
	progress := flag.Bool("progress", false, "display terminal progress")
	insecure := flag.Bool("insecure", false, "do not verify server's certificate chain (not recommended)")
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		log.Fatalf("Expected two args: URL and dst")
		os.Exit(1)
	}

	// For now, we only support HTTPS for security
	_ = modeRaw // Accept flag but enforce https only

	opts := []getter.ClientOption{}
	if *progress {
		opts = append(opts, getter.WithProgress(defaultProgressBar))
	}

	if *insecure {
		log.Println("WARNING: Using Insecure TLS transport!")
		opts = append(opts, getter.WithInsecure())
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	// Only HTTPS allowed for security
	client := &getter.Client{
		Ctx:  ctx,
		Src:  args[0],
		Dst:  args[1],
		Mode: getter.ClientModeAny,

		// Restrict allowed protocols - HTTPS only
		Getters: map[string]getter.Getter{
			"https": &getter.HttpGetter{},
		},

		// Disable symlink following
		DisableSymlinks: true,
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
