package main

import (
	"dlomanov/nlreturnfmt/pkg/nlreturnfmt"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	formatterName = "nlreturn-formatter"
	formatterDoc  = `A Go code formatter that inserts blank lines before return and branch statements except when the return is alone inside a statement group (such as an if statement) to increase code clarity.`
)

var (
	blockSize = flag.Uint("block-size", 1, "set block size that is still ok")
	write     = flag.Bool("w", false, "write result to (source) file instead of stdout")
	dryRun    = flag.Bool("n", false, "don't modify files, just print what would be changed")
	verbose   = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [flags] [path ...]", formatterName)
		_, _ = fmt.Fprintf(os.Stderr, "\n%s", formatterDoc)
		_, _ = fmt.Fprintf(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

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
	formatter := nlreturnfmt.New(opts...)

	if err := process(formatter); err != nil {
		log.Fatal(err)
	}
}

func process(formatter *nlreturnfmt.Formater) error {
	if flag.NArg() == 0 {
		if err := processSource(formatter); err != nil {
			return fmt.Errorf("processSource: %v", err)
		}
	} else {
		if err := processPaths(formatter, flag.Args()); err != nil {
			return fmt.Errorf("processPaths: %v", err)
		}
	}

	return nil
}

func processSource(formatter *nlreturnfmt.Formater) error {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	result, modified, err := formatter.FormatBytes("<stdin>", src)
	if err != nil {
		return fmt.Errorf("formatter.FormatBytes: %w", err)
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

func processPaths(formatter *nlreturnfmt.Formater, paths []string) error {
	for _, path := range paths {
		if err := formatter.FormatPath(path); err != nil {
			return fmt.Errorf("formatter.FormatPath: %w", err)
		}
	}

	return nil
}
