package nlreturnfmt

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type (
	Formater struct {
		blockSize uint
		write     bool
		dryRun    bool
		verbose   bool
	}
	formater struct {
		fset      *token.FileSet
		blockSize uint

		modified bool
		details  *strings.Builder
	}
	result struct {
		value    []byte
		modified bool
		details  string
	}
)

type Option func(*Formater)

func WithBlockSize(blockSize uint) Option {
	return func(f *Formater) { f.blockSize = blockSize }
}

func WithWrite() Option {
	return func(f *Formater) { f.write = true }
}

func WithDryRun() Option {
	return func(f *Formater) { f.dryRun = true }
}

func WithVerbose() Option {
	return func(f *Formater) { f.verbose = true }
}

func New(opts ...Option) *Formater {
	f := &Formater{
		blockSize: 1,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *Formater) formater() *formater {
	return &formater{
		fset:      token.NewFileSet(),
		blockSize: f.blockSize,
		details:   &strings.Builder{},
	}
}

func (f *Formater) FormatBytes(filename string, src []byte) ([]byte, bool, error) {
	res, err := f.formatBytes(filename, src)
	if err != nil {
		return nil, false, fmt.Errorf("formatBytes: %w", err)
	}

	return res.value, res.modified, nil
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
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
				if err = f.processFile(path); err != nil {
					return fmt.Errorf("processFile: %w", err)
				}
			}

			return nil
		}()

		if errors.Is(err, filepath.SkipDir) && f.verbose {
			fmt.Printf("%s skipped\n", path)
		}

		return err
	})
}

func (f *Formater) processFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	res, err := f.formatBytes(filename, src)
	switch {
	case err != nil:
		return fmt.Errorf("formatBytes: %w", err)
	case !res.modified && f.verbose:
		fmt.Printf("%s: no changes needed\n", filename)
	case !res.modified:
	case f.dryRun && f.verbose:
		fmt.Printf("%s: would be modified\n%s", filename, res.details)
	case f.dryRun:
		fmt.Printf("%s: would be modified\n", filename)
	case f.write:
		if f.verbose {
			fmt.Printf("%s: formatted\n%s", filename, res.details)
		}
		//nolint: gosec
		if err = os.WriteFile(filename, res.value, 0o644); err != nil {
			return fmt.Errorf("os.WriteFile: %w", err)
		}
	default:
		fmt.Printf("// %s - formatted:\n%s\n", filename, string(res.value))
	}

	return nil
}

func (f *Formater) formatBytes(filename string, src []byte) (result, error) {
	ff := f.formater()

	file, err := parser.ParseFile(ff.fset, filename, src, parser.ParseComments)
	if err != nil {
		return result{}, fmt.Errorf("parser.ParseFile: %w", err)
	}

	res := astutil.Apply(file, nil, ff.format)
	if !ff.modified {
		return result{}, nil
	}

	var buf bytes.Buffer
	if err = format.Node(&buf, ff.fset, res); err != nil {
		return result{}, err
	}

	return result{
		value:    buf.Bytes(),
		modified: true,
		details:  ff.details.String(),
	}, nil
}

func (f *formater) format(c *astutil.Cursor) bool {
	var name string

	switch node := c.Node().(type) {
	case *ast.ReturnStmt:
		name = "return"
	case *ast.BranchStmt:
		name = node.Tok.String()
	default:
		return true
	}

	if f.shouldInsert(c) {
		c.InsertBefore(newBlankLine(c.Node()))
		f.modified = true

		pos := f.fset.Position(c.Node().Pos())
		_, _ = fmt.Fprintf(f.details, "- insert blank line before %s at %s\n", name, pos)
	}

	return true
}

func (f *formater) shouldInsert(ret *astutil.Cursor) bool {
	var block []ast.Stmt

	switch node := ret.Parent().(type) {
	case *ast.CaseClause:
		block = node.Body
	case *ast.CommClause:
		block = node.Body
	case *ast.BlockStmt:
		block = node.List
	default:
		return false
	}

	//nolint: gosec
	if ret.Index() == 0 || uint(f.line(ret.Node().Pos())-f.line(block[0].Pos())) < f.blockSize {
		return false
	}

	return f.line(ret.Node().Pos())-f.line(block[ret.Index()-1].End()) <= 1
}

func (f *formater) line(pos token.Pos) int { return f.fset.Position(pos).Line }

func newBlankLine(node ast.Node) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.Ident{
			NamePos: node.Pos(),
			Name:    "", // Empty identifier creates line break
		},
	}
}
