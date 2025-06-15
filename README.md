# nlreturnfmt

A Go code formatter that automatically inserts blank lines before return and branch statements except when the return is alone inside a statement group (such as an if statement) to increase code clarity.

Based on https://github.com/ssgreg/nlreturn

## Installation

```bash
go install github.com/dlomanov/nlreturnfmt@latest
```

## Usage

```bash
nlreturnfmt [flags] [path ...]
```

### Flags

* `-w` write result to (source) file instead of stdout
* `-n` don't modify files, just print what would be changed (dry-run)
* `-v` verbose output
* `-block-size n` set block size that is still ok (default: 1)

### Examples

**Format and print to stdout:**
```bash
nlreturnfmt file.go
```

**Format and write to file:**
```bash
nlreturnfmt -w file.go
```

**Format all Go files in directory:**
```bash
nlreturnfmt -w ./...
```

**Dry-run to see what would be changed:**
```bash
nlreturnfmt -n -v file.go
```

**Read from stdin:**
```bash
cat file.go | nlreturnfmt
```

## Example

### Before formatting (incorrect):

```go
func foo() int {
    a := 0
    _ = a
    return a
}

func bar() int {
    a := 0
    if a == 0 {
        _ = a
        return
    }
    return a
}

func baz() {
    for i := 0; i < 10; i++ {
        if i == 5 {
            break
        }
        continue
    }
}
```

### After formatting (correct):

```go
func foo() int {
    a := 0
    _ = a

    return a
}

func bar() int {
    a := 0
    if a == 0 {
        _ = a

        return
    }

    return a
}

func baz() {
    for i := 0; i < 10; i++ {
        if i == 5 {

            break
        }

        continue
    }
}
```

## Block Size

The `-block-size` parameter controls the minimum number of statements required in a block before blank lines are enforced. With the default value of 1, blank lines are required before return/branch statements in blocks with more than 1 non-empty statement.

### Example with `-block-size 2`:

```go
// This would NOT be formatted (block size <= 2)
func small() {
    x := 1
    return x
}

// This WOULD be formatted (block size > 2)
func large() {
    x := 1
    y := 2
    z := 3
    return x + y + z  // <- blank line inserted here
}
```

## Integration

### With gofmt pipeline:
```bash
nlreturnfmt -w . && gofmt -w .
```

### With go generate:
```go
//go:generate nlreturnfmt -w .
```

### With Make:
```makefile
format:
	nlreturnfmt -w ./...
	go fmt ./...
```

## Supported Statements

The formatter handles the following statements:
- `return` statements
- `break` statements
- `continue` statements
- `fallthrough` statements
- `goto` statements

## Exclusions

Blank lines are NOT inserted when:
- The return/branch statement is the first statement in a block
- A blank line already exists before the statement
- The statement is alone in a small block (controlled by `-block-size`)
- The block contains fewer statements than the block-size threshold

## License

MIT License