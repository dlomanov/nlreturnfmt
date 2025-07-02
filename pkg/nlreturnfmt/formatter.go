package nlreturnfmt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dlomanov/nlreturnfmt/pkg/nlreturnfmt/bytefmt"
)

/*
* разобраться c отношением Formatter и formatter
* понять где можно параллелить и как это правильно сделать (сначала сделать, потом понять)
* кажется нужно вытащить отдельно формат байтов, потому что там лежит сам алгоритм, можно вынести в пакет
 */

const blockSizeDefault = 1

type (
	Formater struct {
		blockSize uint
		write     bool
		dryRun    bool
		verbose   bool
	}
)

func New(opts ...Option) *Formater {
	f := &Formater{
		blockSize: blockSizeDefault,
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
	return filepath.Walk(dir, f.processDirWalk)
}

func (f *Formater) processDirWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("filepath.Walk: %w", err)
	}

	name := info.Name()
	if err = func() error {
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
			if err = f.processFile(path); err != nil {
				return fmt.Errorf("processFile: %w", err)
			}
		}

		return nil
	}(); err != nil {
		if errors.Is(err, filepath.SkipDir) && f.verbose {
			fmt.Printf("%s skipped\n", path)
		}

		return err
	}

	return nil
}

func (f *Formater) processFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	res, err := f.format(filename, src)
	switch {
	case err != nil:
		return fmt.Errorf("bfmt.Format: %w", err)
	case !res.Modified && f.verbose:
		fmt.Printf("%s: no changes needed\n", filename)
	case !res.Modified:
	case f.dryRun && f.verbose:
		fmt.Printf("%s: would be modified\n%s", filename, res.Details)
	case f.dryRun:
		fmt.Printf("%s: would be modified\n", filename)
	case f.write:
		if f.verbose {
			fmt.Printf("%s: formatted\n%s", filename, res.Details)
		}
		//nolint: gosec
		if err = os.WriteFile(filename, res.Value, 0o644); err != nil {
			return fmt.Errorf("os.WriteFile: %w", err)
		}
	default:
		fmt.Printf("// %s - formatted:\n%s\n", filename, string(res.Value))
	}

	return nil
}

func (f *Formater) format(filename string, src []byte) (bytefmt.Result, error) {
	ff := bytefmt.New(f.blockSize)

	return ff.Format(filename, src)
}
