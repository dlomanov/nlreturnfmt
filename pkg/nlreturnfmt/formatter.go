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
	switch node := c.Node().(type) {
	case *ast.CaseClause:
		f.processBlock(c, node.Body)
	case *ast.CommClause:
		f.processBlock(c, node.Body)
	case *ast.BlockStmt:
		f.processBlock(c, node.List)
	}

	return true
}
func (f *formater) processBlock(c *astutil.Cursor, block []ast.Stmt) {
	// Process statements in reverse order to avoid index issues when inserting
	for i := len(block) - 1; i >= 0; i-- {
		stmt := block[i]

		switch stmt.(type) {
		case *ast.BranchStmt, *ast.ReturnStmt:
			if f.shouldInsertBlankLine(block, i) {
				f.insertBlankLineBefore(c, i)
			}
		}
	}
}

func (f *formater) shouldInsertBlankLine(block []ast.Stmt, returnIndex int) bool {
	// Rule 1: Return must not be the first statement
	if returnIndex == 0 {
		return false
	}

	// Rule 2: Check if previous statement is already empty (blank line exists)
	prevStmt := block[returnIndex-1]
	if _, isEmpty := prevStmt.(*ast.EmptyStmt); isEmpty {
		return false
	}

	// Rule 3: Check if return is "alone" in small block (block-size rule)
	if f.isReturnAloneInBlock(block, returnIndex) {
		return false
	}

	// Rule 4: Check line difference (similar to original linter logic)
	stmt := block[returnIndex]
	prevStmt = block[returnIndex-1]

	stmtLine := f.fset.Position(stmt.Pos()).Line
	prevLine := f.fset.Position(prevStmt.End()).Line

	// If there's already a blank line, don't add another
	if stmtLine-prevLine > 1 {
		return false
	}

	return true
}

func (f *formater) isReturnAloneInBlock(block []ast.Stmt, returnIndex int) bool {
	nonEmptyCount := uint(0)
	for i, stmt := range block {
		if i == returnIndex {
			continue // Skip the return statement itself
		}
		if _, isEmpty := stmt.(*ast.EmptyStmt); !isEmpty {
			nonEmptyCount++
		}
	}

	return nonEmptyCount <= f.blockSize
}

func (f *formater) insertBlankLineBefore(c *astutil.Cursor, returnIdx int) {
	// Get the current block node
	var stmtList *[]ast.Stmt

	switch node := c.Node().(type) {
	case *ast.BlockStmt:
		stmtList = &node.List
	case *ast.CaseClause:
		stmtList = &node.Body
	case *ast.CommClause:
		stmtList = &node.Body
	default:
		return // todo: Can't handle this node type
	}

	// Insert the empty statement before the return/branch statement
	newList := make([]ast.Stmt, 0, len(*stmtList)+1)
	newList = append(newList, (*stmtList)[:returnIdx]...)
	newList = append(newList, &ast.EmptyStmt{Semicolon: token.NoPos, Implicit: true})
	newList = append(newList, (*stmtList)[returnIdx:]...)

	*stmtList = newList
	f.modified = true

	if f.verbose {
		pos := f.fset.Position((*stmtList)[returnIdx+1].Pos())
		fmt.Printf("Inserted blank line before %s at %s\n", name((*stmtList)[returnIdx+1]), pos)
	}
}

func name(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.BranchStmt:
		return s.Tok.String()
	case *ast.ReturnStmt:
		return "return"
	default:
		return "unknown"
	}
}
