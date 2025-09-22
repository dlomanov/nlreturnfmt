package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"dlomanov/nlreturnfmt/pkg/nlreturnfmt"
)

// Unix: 128 + signal number (SIGINT = 2).
const exitCodeCanceled = 130

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

const (
	formatterName = "nlreturnfmt"
	formatterDoc  = `A Go code formatter that inserts blank lines before return and branch statements to increase code clarity.`
)

var (
	blockSize   = flag.Int("block-size", 1, "set block size that is still ok")
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	dryRun      = flag.Bool("n", false, "don't modify files, just print what would be changed")
	verbose     = flag.Bool("v", false, "verbose output")
	showVersion = flag.Bool("version", false, "show version information")
	parallelism = flag.Int("parallelism", 0, "number of files to process in parallel (0 = NumCPU)")
)

func main() {
	err := run()
	switch {
	case errors.Is(err, context.Canceled):
		_, _ = fmt.Fprintln(os.Stderr, "operation canceled")
		os.Exit(exitCodeCanceled)
	case err != nil:
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	//nolint: reassign
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [flags] [path ...]", formatterName)
		_, _ = fmt.Fprintf(os.Stderr, "\n%s", formatterDoc)
		_, _ = fmt.Fprintf(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s (commit: %s, date: %s)\n", formatterName, version, commit, date)

		return nil
	}

	opts := []nlreturnfmt.Option{
		nlreturnfmt.WithBlockSize(*blockSize),
		nlreturnfmt.WithParallelism(*parallelism),
	}
	if *write {
		opts = append(opts, nlreturnfmt.WithWrite())
	}
	if *dryRun {
		opts = append(opts, nlreturnfmt.WithDryRun())
	}
	if *verbose {
		opts = append(opts, nlreturnfmt.WithVerbose())
	}
	formatter := nlreturnfmt.New(opts...)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return process(ctx, formatter)
}

func process(ctx context.Context, formatter *nlreturnfmt.Formatter) error {
	if flag.NArg() == 0 {
		if *write {
			return errors.New("-w flag is not supported when processing from stdin")
		}
		if err := processSource(ctx, formatter); err != nil {
			return fmt.Errorf("processSource: %w", err)
		}
	} else {
		if err := processPaths(ctx, formatter, flag.Args()); err != nil {
			return fmt.Errorf("processPaths: %w", err)
		}
	}

	return nil
}

func processSource(ctx context.Context, formatter *nlreturnfmt.Formatter) error {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	result, modified, err := formatter.FormatFile(ctx, "<stdin>", src)
	if err != nil {
		return fmt.Errorf("formatter.FormatFile: %w", err)
	}

	if !modified {
		if *verbose {
			_, _ = fmt.Fprintln(os.Stderr, "No changes needed")
		}
		fmt.Print(string(src))
	} else {
		fmt.Print(string(result))
	}

	return nil
}

func processPaths(ctx context.Context, formatter *nlreturnfmt.Formatter, paths []string) error {
	for _, path := range paths {
		if err := formatter.FormatPath(ctx, path); err != nil {
			return fmt.Errorf("formatter.FormatPath: %w", err)
		}
	}

	return nil
}
