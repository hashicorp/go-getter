package getter

import (
	"sync"
	"testing"
)

func TestGroup_serial(t *testing.T) {
	var g Group

	if v := g.Begin("foo"); v != "" {
		t.Fatalf("bad: %s", v)
	}

	g.End("foo", "bar")

	if v := g.Begin("foo"); v != "bar" {
		t.Fatalf("bad: %s", v)
	}
}

func TestGroup_emptyEnd(t *testing.T) {
	var g Group

	if v := g.Begin("foo"); v != "" {
		t.Fatalf("bad: %s", v)
	}

	g.End("foo", "bar")

	if v := g.Begin("foo"); v != "bar" {
		t.Fatalf("bad: %s", v)
	}

	g.End("foo", "")

	if v := g.Begin("foo"); v != "bar" {
		t.Fatalf("bad: %s", v)
	}
}

func TestGroup_block(t *testing.T) {
	var g Group

	// Begin on our own
	if v := g.Begin("foo"); v != "" {
		t.Fatalf("bad: %s", v)
	}

	// Start a bunch
	var wg sync.WaitGroup
	max := 5
	resultCh := make(chan string, max)
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val := g.Begin("foo")
			g.End("foo", "")
			resultCh <- val
		}()
	}

	// End ours, set our value
	g.End("foo", "bar")

	// Wait for the waitgroup to finish
	wg.Wait()
	close(resultCh)

	// Verify that we got the right result
	for v := range resultCh {
		if v != "bar" {
			t.Fatalf("bad: %s", v)
		}
	}
}
