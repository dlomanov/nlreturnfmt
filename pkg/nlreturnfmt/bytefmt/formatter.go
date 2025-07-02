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
	Formater struct {
		fset      *token.FileSet
		blockSize uint

		modified bool
		details  *strings.Builder
	}
	Result struct {
		Filename string
		Value    []byte
		Modified bool
		Details  string
	}
)

func New(blockSize uint) *Formater {
	return &Formater{
		fset:      token.NewFileSet(),
		blockSize: blockSize,
		details:   &strings.Builder{},
	}
}

func (f *Formater) Format(filename string, src []byte) (Result, error) {
	file, err := parser.ParseFile(f.fset, filename, src, parser.ParseComments)
	if err != nil {
		return Result{}, fmt.Errorf("parser.ParseFile: %w", err)
	}

	res := astutil.Apply(file, nil, f.format)
	if !f.modified {
		return Result{}, nil
	}

	var buf bytes.Buffer
	if err = format.Node(&buf, f.fset, res); err != nil {
		return Result{}, err
	}

	return Result{
		Filename: filename,
		Value:    buf.Bytes(),
		Modified: true,
		Details:  f.details.String(),
	}, nil
}

func (f *Formater) format(c *astutil.Cursor) bool {
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

func (f *Formater) shouldInsert(ret *astutil.Cursor) bool {
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

func (f *Formater) line(pos token.Pos) int { return f.fset.Position(pos).Line }

func newBlankLine(node ast.Node) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.Ident{
			NamePos: node.Pos(),
			Name:    "", // Empty identifier creates line break
		},
	}
}
