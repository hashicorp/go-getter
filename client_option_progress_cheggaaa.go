package getter

import (
	"io"
	"path/filepath"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// WithCheggaaaProgressBarV1 will have the downloads tracked
// by a github.com/cheggaaa/pb v2 multi progress bar.
func WithCheggaaaProgressBarV1() func(*Client) error {
	return func(c *Client) error {
		c.ProgressListener = cheggaaaProgressBar
		return nil
	}
}

// CheggaaaProgressBar just wraps
// a pb.Pool to display a progress
type CheggaaaProgressBar struct {
	pool *pb.Pool

	pbs int
}

var cheggaaaProgressBar ProgressTracker = &CheggaaaProgressBar{}

func defaultCheggaaaProgressBarConfigFN(bar *pb.ProgressBar, prefix string) {
	bar.SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.
func (cpb *CheggaaaProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	newPb := pb.New64(totalSize)
	newPb.Set64(currentSize)
	defaultCheggaaaProgressBarConfigFN(newPb, filepath.Base(src))
	if cpb.pool == nil {
		cpb.pool = pb.NewPool()
		cpb.pool.Start()
	}
	cpb.pool.Add(newPb)
	reader := newPb.NewProxyReader(stream)

	cpb.pbs++
	return &cheggaaaReadCloser{
		Reader: reader,
		close: func() error {
			cpb.pbs--
			if cpb.pbs <= 0 {
				cpb.pool.Stop()
			}
			newPb.Finish()
			return nil
		},
	}
}

type cheggaaaReadCloser struct {
	io.Reader
	close func() error
}

func (c *cheggaaaReadCloser) Close() error { return c.close() }
