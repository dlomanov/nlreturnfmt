package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"dlomanov/nlreturnfmt/pkg/nlreturnfmt"
)

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
	parallelism = flag.Int("parallelism", 0, "number of files to process in parallel (0 = auto, default = 15)")
)

func main() {
	//nolint: reassign
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [flags] [path ...]", formatterName)
		_, _ = fmt.Fprintf(os.Stderr, "\n%s", formatterDoc)
		_, _ = fmt.Fprintf(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion != nil && *showVersion {
		fmt.Printf("%s version %s (commit: %s, date: %s)\n", formatterName, version, commit, date)

		return
	}

	var opts []nlreturnfmt.Option
	if blockSize != nil {
		opts = append(opts, nlreturnfmt.WithBlockSize(*blockSize))
	}
	if write != nil && *write {
		opts = append(opts, nlreturnfmt.WithWrite())
	}
	if dryRun != nil && *dryRun {
		opts = append(opts, nlreturnfmt.WithDryRun())
	}
	if verbose != nil && *verbose {
		opts = append(opts, nlreturnfmt.WithVerbose())
	}
	if parallelism != nil {
		opts = append(opts, nlreturnfmt.WithParallelism(*parallelism))
	}
	formatter := nlreturnfmt.New(opts...)

	if err := process(formatter); err != nil {
		log.Fatal(err)
	}
}

func process(formatter *nlreturnfmt.Formatter) error {
	if flag.NArg() == 0 {
		if err := processSource(formatter); err != nil {
			return fmt.Errorf("processSource: %w", err)
		}
	} else {
		if err := processPaths(formatter, flag.Args()); err != nil {
			return fmt.Errorf("processPaths: %w", err)
		}
	}

	return nil
}

func processSource(formatter *nlreturnfmt.Formatter) error {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	result, modified, err := formatter.FormatFile("<stdin>", src)
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

func processPaths(formatter *nlreturnfmt.Formatter, paths []string) error {
	for _, path := range paths {
		if err := formatter.FormatPath(path); err != nil {
			return fmt.Errorf("formatter.FormatPath: %w", err)
		}
	}

	return nil
}
