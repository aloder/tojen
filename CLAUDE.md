# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`tojen` is a CLI that reads an existing `.go` file and emits Go source that, when run,
regenerates that file using the [jennifer](https://github.com/dave/jennifer) code-generation
library. In other words it produces *code that generates code* — a boilerplate starting point
you then hand-edit and extend for your own generators.

## Module layout & dependencies

This is a **GOPATH-era project — there is no `go.mod`**. The import path is hard-coded as
`github.com/aloder/tojen`, so the repo must live at `$GOPATH/src/github.com/aloder/tojen` for
builds and tests to resolve. External deps (`spf13/cobra`, `dave/jennifer`,
`stretchr/testify`, `pkg/errors`) are expected on the GOPATH, not vendored. If you add a
dependency you must `go get` it into the GOPATH; do not assume modules.

- `main.go` — thin entrypoint, calls `cmd.Execute()`.
- `cmd/root.go` — cobra command wiring for `tojen gen`.
- `gen/` — the translator (the heart of the project).
- `run/` — helper that writes a Go string to a temp dir and `go run`s it; used by tests.

## Commands

```bash
go build ./...          # build
go install              # install the `tojen` binary
go test ./...           # run all tests
go test ./gen -run TestFile/struct      # run one named subtest (table-driven, see below)
go vet ./...
gofmt -l .              # list unformatted files
```

Run the tool:

```bash
tojen gen path/to/file.go               # print generated jennifer code to stdout
tojen gen path/to/file.go out.go        # write generated code to out.go instead
tojen gen file.go -p mypkg              # set package name of generated file (default: main)
tojen gen file.go -m                    # also emit a main() that prints the regenerated source
tojen gen file.go -f                    # EXPERIMENTAL gofmt-style reformatting (see caveats)
```

## Architecture

### The translation pipeline (`gen/`)

`GenerateFileBytes` → `GenerateFile` is the top of the funnel. `GenerateFile`:
1. Parses the input bytes into a `*ast.File` via `go/parser` (`parseFile` **panics** on parse error).
2. Builds the import name→path map and anonymous (`_`) imports via `imports`.
3. Walks each top-level `ast.Decl`, turning it into a `genXxx() jen.Code` helper function
   (`makeJenCode`), and finally synthesizes a `genFile()` function that calls each helper and
   assembles a `*jen.File`.

The translation is a **node-type-dispatch recursion**: each Go AST node kind has a dedicated
function returning `jen.Code`. The dispatch tables are the switch statements in:
- `gen/expr.go` — `genExpr` dispatches every `ast.Expr` (idents, calls, literals, binary/unary
  ops, selectors, composite literals, channel/map/array/struct/interface/func types, etc.).
- `gen/stmt.go` — `stmt` dispatches every `ast.Stmt` (if/for/switch/select/range, assign,
  defer/go, return, branch, etc.), plus `fieldList`/`blockStmt` helpers shared widely.
- `gen/type.go` — `funcType`, `arrayType`, `structType`, `interfaceType`.
- `gen/main.go` — top-level decl handling (`funcDecl`, `genDecl`, `valueSpec`, `typeSpec`,
  `basicLit`) and the `genFile`/`main` assembly.

When extending support for a new Go construct, add a case to the relevant `switch` **and** the
function it dispatches to. The two dispatchers (`genExpr`, `stmt`) **panic with the
`reflect.TypeOf` name** for unhandled nodes — that panic message is the signal that a construct
isn't supported yet.

### The output idiom

Generated code is itself jennifer method chains, so the translator emits jennifer calls that
*emit jennifer calls*. Two recurring patterns to recognize:
- `jen.Dot("X").Call(...)` produces a chained `.X(...)` on the caller's statement.
- `jen.Id("jen").Add(genExpr(...))` starts a fresh `jen.<...>` chain for a nested expression
  (used when an argument needs to be its own standalone jennifer statement, e.g. inside
  `genExprsCode`, `fieldList`, call args).

`selectorExpr` (`gen/expr.go`) is where package-qualified references are decided: if the
selector's left side matches a known import name in the global `paths` map, it emits
`jen.Qual(path, sel)`; otherwise it's treated as a field/method access (`.Dot(sel)`).

### Global state

`gen/main.go` holds package-level globals `paths` (import name → import path) and `formating`.
`paths` is repopulated on every `GenerateFile` call and read by `selectorExpr`. Because it's
global, the translator is **not safe for concurrent generation** — keep that in mind before
parallelizing.

### Round-trip test strategy (`gen/file_test.go`)

`TestFile` is the primary correctness harness and the model to follow when adding features. It
is **table-driven**: `tests []tcg` holds named Go snippets. For each case it:
1. gofmt-normalizes the input snippet (the expected output).
2. Runs `GenerateFile(..., main=true)` to produce a jennifer program.
3. `run.Exec`-compiles-and-runs that program (this requires the `go` toolchain on PATH), capturing
   the source it prints.
4. gofmt-normalizes the printed source and asserts it equals the original.

So a passing test proves the *generated generator actually round-trips back to the source*. **To
add support for a new construct, add a snippet to the `tests` table** — if it round-trips, you're
done. Run a single case with `go test ./gen -run 'TestFile/<Name>'`.

`run.Exec` (`run/run.go`) shells out to `go run` in an `ioutil.TempDir`; tests will fail in an
environment without a working Go compiler or network access to fetch jennifer at run time.

## Conventions & gotchas

- **Parse and dispatch errors surface as panics**, not returned errors (`parseFile`, the
  `genExpr`/`stmt` default cases). Don't wrap them in error returns expecting graceful handling.
- The `-f`/`formatting.go` path is explicitly **EXPERIMENTAL**: it post-processes rendered bytes
  with regexes (`formatNulls`, `formatStructs`, `formatBlock`, `formatParams`) and still contains
  `println("find")` debug output. Treat it as unfinished; the default (unformatted) path is the
  reliable one.
- `.` (dot) imports are unsupported and `imports()` panics on them; `_` imports are collected
  separately and emitted via `ret.Anon(...)`.
- `basicLit` does not support imaginary numbers (panics) and has special-case handling for
  CHAR/STRING quoting — touch carefully when changing literal handling.
