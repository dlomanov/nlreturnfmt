package nlreturnfmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"path/filepath"
	"strings"
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
		verbose   bool

		modified bool
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
		verbose:   f.verbose,
	}
}

func (f *Formater) FormatBytes(filename string, src []byte) ([]byte, bool, error) {
	return f.formatBytes(filename, src)
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
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			return f.processFile(path)
		}

		return nil
	})
}

func (f *Formater) processFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	result, modified, err := f.formatBytes(filename, src)
	if err != nil {
		return fmt.Errorf("formatBytes: %w", err)
	}

	if !modified {
		if f.verbose {
			fmt.Printf("%s: no changes needed\n", filename)
		}

		return nil
	}

	if f.dryRun {
		fmt.Printf("%s: would be modified\n", filename)

		return nil
	}

	if f.write {
		if f.verbose {
			fmt.Printf("%s: formatted\n", filename)
		}

		return os.WriteFile(filename, result, 0644)
	}

	fmt.Printf("// %s - formatted:\n%s\n", filename, string(result))

	return nil
}

func (f *Formater) formatBytes(filename string, src []byte) ([]byte, bool, error) {
	ff := f.formater()

	file, err := parser.ParseFile(ff.fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, false, fmt.Errorf("parser.ParseFile: %w", err)
	}

	result := astutil.Apply(file, nil, ff.format)
	if !ff.modified {
		return src, false, nil
	}

	var buf bytes.Buffer
	if err = format.Node(&buf, ff.fset, result); err != nil {
		return nil, false, err
	}

	return buf.Bytes(), true, nil
}

func (f *formater) format(c *astutil.Cursor) bool {
	var name = "unknown"

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

		if f.verbose {
			pos := f.fset.Position(c.Node().Pos())
			fmt.Printf("- inserted blank line before %s at %s\n", name, pos)
		}
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

	if ret.Index() == 0 || f.line(ret.Node().Pos())-f.line(block[0].Pos()) < int(f.blockSize) {
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
