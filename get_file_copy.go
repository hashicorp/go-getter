package getter

import (
	"context"
	"io"
	"os"
)

// readerFunc is syntactic sugar for read interface.
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

// Copy is a io.Copy cancellable by context
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	return io.Copy(dst, readerFunc(func(p []byte) (int, error) {

		select {
		case <-ctx.Done():
			// context has been canceled
			// stop process and propagate "context canceled" error
			return 0, ctx.Err()
		default:
			// otherwise just run default io.Reader implementation
			return src.Read(p)
		}
	}))
}

// copyReader copies from an io.Reader into a file, using umask to create the dst file
func copyReader(dst string, src io.Reader, mode os.FileMode) error {
	dstF, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstF.Close()

	_, err = io.Copy(dstF, src)
	return err
}

// copyFile copies a file in chunks from src path to dst path, using umask to create the dst file
func copyFile(ctx context.Context, dst string, src string, mode os.FileMode) (int64, error) {
	srcF, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcF.Close()

	dstF, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return 0, err
	}
	defer dstF.Close()

	return Copy(ctx, dstF, srcF)
}
