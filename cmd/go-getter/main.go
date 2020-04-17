package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	gcs "github.com/hashicorp/go-getter/gcs/v2"
	s3 "github.com/hashicorp/go-getter/s3/v2"
	getter "github.com/hashicorp/go-getter/v2"
)

func main() {
	modeRaw := flag.String("mode", "any", "get mode (any, file, dir)")
	progress := flag.Bool("progress", false, "display terminal progress")
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		log.Fatalf("Expected two args: URL and dst")
		os.Exit(1)
	}

	// Get the mode
	var mode getter.Mode
	switch *modeRaw {
	case "any":
		mode = getter.ModeAny
	case "file":
		mode = getter.ModeFile
	case "dir":
		mode = getter.ModeDir
	default:
		log.Fatalf("Invalid client mode, must be 'any', 'file', or 'dir': %s", *modeRaw)
		os.Exit(1)
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Build the client
	req := &getter.Request{
		Src:  args[0],
		Dst:  args[1],
		Pwd:  pwd,
		Mode: mode,
	}
	if *progress {
		req.ProgressListener = defaultProgressBar
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	client := getter.DefaultClient

	client.Getters["gcs"] = new(gcs.Getter)
	client.Getters["s3"] = new(s3.Getter)

	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		res, err := getter.DefaultClient.Get(ctx, req)
		if err != nil {
			errChan <- err
			return
		}
		log.Printf("-> %s", res.Dst)
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Printf("signal %v", sig)
	case <-ctx.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		log.Fatalf("Error downloading: %s", err)
	}
}
