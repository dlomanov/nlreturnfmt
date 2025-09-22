package bytefmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type (
	Formatter struct {
		fset      *token.FileSet
		blockSize int
	}
	Result struct {
		Filename string
		Value    []byte
		Modified bool
		Details  string
	}
)

func New(blockSize int) *Formatter {
	return &Formatter{
		fset:      token.NewFileSet(),
		blockSize: blockSize,
	}
}

func (f *Formatter) Format(filename string, src []byte) (Result, error) {
	file, err := parser.ParseFile(f.fset, filename, src, parser.ParseComments)
	if err != nil {
		return Result{}, fmt.Errorf("parser.ParseFile: %w", err)
	}

	var (
		modified bool
		details  = &strings.Builder{}
	)

	res := astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
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
			modified = true

			pos := f.fset.Position(c.Node().Pos())
			_, _ = fmt.Fprintf(details, "- insert blank line before %s at %s\n", name, pos)
		}

		return true
	})

	var buf bytes.Buffer
	if err = format.Node(&buf, f.fset, res); err != nil {
		return Result{}, err
	}

	return Result{
		Filename: filename,
		Value:    buf.Bytes(),
		Modified: modified,
		Details:  details.String(),
	}, nil
}

func (f *Formatter) shouldInsert(ret *astutil.Cursor) bool {
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

	// Do not add a newline if the statement is the first in the block,
	// or if the block is too short (fewer lines than blockSize).
	if ret.Index() == 0 || f.line(ret.Node().Pos())-f.line(block[0].Pos()) < f.blockSize {
		return false
	}

	return f.line(ret.Node().Pos())-f.line(block[ret.Index()-1].End()) <= 1
}

func (f *Formatter) line(pos token.Pos) int { return f.fset.Position(pos).Line }

func newBlankLine(node ast.Node) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.Ident{
			NamePos: node.Pos(),
			Name:    "", // Empty identifier creates line break
		},
	}
}
