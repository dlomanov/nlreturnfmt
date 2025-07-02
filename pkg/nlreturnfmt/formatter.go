package nlreturnfmt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"dlomanov/nlreturnfmt/pkg/nlreturnfmt/bytefmt"
)

const blockSizeDefault = 1

type (
	Formater struct {
		blockSize   uint
		write       bool
		dryRun      bool
		verbose     bool
		parallelism int
	}
)

func New(opts ...Option) *Formater {
	f := &Formater{
		blockSize:   blockSizeDefault,
		parallelism: runtime.NumCPU(),
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *Formater) FormatFile(filename string, src []byte) ([]byte, bool, error) {
	res, err := f.format(filename, src)
	if err != nil {
		return nil, false, fmt.Errorf("format: %w", err)
	}

	return res.Value, res.Modified, nil
}

func (f *Formater) FormatPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}

	if info.IsDir() {
		return f.processDir(path)
	}

	return f.processFile(path)
}

func (f *Formater) processDir(dir string) error {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(f.parallelism)

	resch := make(chan bytefmt.Result, f.parallelism)

	processFile := func(filename string) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		src, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("os.ReadFile: %w", err)
		}

		g.Go(func() error {
			res, innerr := f.format(filename, src)
			if innerr != nil {
				return fmt.Errorf("format: %w", innerr)
			}

			select {
			case <-ctx.Done():
			case resch <- res:
			}

			return nil
		})

		return nil
	}

	g.Go(func() error {
		return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			return f.processDirWalk(path, info, err, processFile)
		})
	})
	go func() {
		defer close(resch)

		_ = g.Wait()
	}()

	var errs error
	for res := range resch {
		if err := f.processFileResult(res); err != nil {
			errs = errors.Join(errs, fmt.Errorf("processFileResult: %w", err))
		}
	}

	return errors.Join(errs, g.Wait())
}

func (f *Formater) processDirWalk(path string, info os.FileInfo, err error, fn func(string) error) error {
	if err != nil {
		return fmt.Errorf("filepath.Walk: %w", err)
	}
	name := info.Name()

	err = func() error {
		switch {
		case info.IsDir():
			switch {
			case strings.HasPrefix(name, "vendor"):
				return filepath.SkipDir
			case strings.HasPrefix(name, "testdata"):
				return filepath.SkipDir
			case name != "." && strings.HasPrefix(name, "."):
				return filepath.SkipDir
			default:
				return nil
			}
		case strings.HasSuffix(name, "_test.go"):
		case strings.HasSuffix(name, ".go"):
			if err = fn(path); err != nil {
				return err
			}
		}

		return nil
	}()

	if errors.Is(err, filepath.SkipDir) && f.verbose {
		fmt.Printf("%s skipped\n", path)
	}

	return err
}

func (f *Formater) processFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	res, err := f.format(filename, src)
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	if err = f.processFileResult(res); err != nil {
		return fmt.Errorf("processFileResult: %w", err)
	}

	return nil
}

func (f *Formater) processFileResult(res bytefmt.Result) error {
	switch {
	case !res.Modified && f.verbose:
		fmt.Printf("%s: no changes needed\n", res.Filename)
	case !res.Modified:
	case f.dryRun && f.verbose:
		fmt.Printf("%s: would be modified\n%s", res.Filename, res.Details)
	case f.dryRun:
		fmt.Printf("%s: would be modified\n", res.Filename)
	case f.write:
		if f.verbose {
			fmt.Printf("%s: formatted\n%s", res.Filename, res.Details)
		}
		//nolint: gosec
		if err := os.WriteFile(res.Filename, res.Value, 0o644); err != nil {
			return fmt.Errorf("os.WriteFile: %w", err)
		}
	default:
		fmt.Printf("// %s - formatted:\n%s\n", res.Filename, string(res.Value))
	}

	return nil
}

func (f *Formater) format(filename string, src []byte) (bytefmt.Result, error) {
	ff := bytefmt.New(f.blockSize)

	return ff.Format(filename, src)
}
