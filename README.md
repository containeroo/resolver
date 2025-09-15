# Resolver Package

The `resolver` package provides a flexible and extensible way to resolve configuration values from various sources including environment variables, plain files, `JSON`, `YAML`, `INI`, `TOML`, and key-value files.
It uses a **prefix-based system** to determine which resolver to use, and returns either the resolved value or an error.

## Installation

```bash
go get github.com/containeroo/resolver/resolver
```

## Usage

The primary entry point is the `ResolveVariable` function.
It takes a string and attempts to resolve it based on its prefix:

- **`env:`** - Environment variables.
  Example:

  ```text
  env:PATH
  ```

  → value of `$PATH`.

- **`file:`** - Simple key-value files. Supports `KEY=VAL` lines, with optional `export` prefixes and `#` comments.
  Example:

  ```text
  file:/config/app.txt//USERNAME
  ```

  → value of `USERNAME` in `app.txt`.

- **`json:`** - JSON files. Supports dot-notation for nested keys and array indexing.
  Examples:

  ```text
  json:/config/app.json//server.host
  json:/config/app.json//servers.0.host
  json:/config/app.json//servers.[name=api].port
  ```

- **`yaml:`** - YAML files. Same dot/array/filter notation as JSON.
  Example:

  ```text
  yaml:/config/app.yaml//servers.[host=example.org].port
  ```

- **`ini:`** - INI files. Supports section+key or default section.
  Examples:

  ```text
  ini:/config/app.ini//Database.User
  ini:/config/app.ini//Key1
  ```

- **`toml:`** - TOML files. Dot-notation for nested keys and array indexing.
  Example:

  ```text
  toml:/config/app.toml//server.host
  ```

- **No prefix** - Returns the value unchanged.
  Example:

  ```text
  just-a-literal
  ```

  → `"just-a-literal"`.

## String interpolation (`ResolveString`)

Interpolate `${...}` tokens inside a larger string and resolve each token with the same rules as `ResolveVariable`.

```go
// Replace ${...} tokens using the default registry (up to 8 passes).
func ResolveString(s string) (string, error)

// Registry method, if you use a custom registry:
func (*Registry) ResolveString(s string) (string, error)
```

**Features & rules**

- `${scheme:...}` tokens are resolved; the `${`...`}` wrapper is removed.
- `\${` emits a **literal** `"${"` (escape) and is **not** expanded.
- A bare `$` not followed by `{` is copied literally.
- Malformed tokens error with `ErrBadPath`:

  - missing closing `}` (e.g., `"${env:HOME"`)
  - empty token `"${}"`

- Multi-pass expansion: tokens that produce new `${...}` are expanded in subsequent passes (depth limit 8).
- Unknown schemes follow your registry policy:

  - Default (**PassThrough**): the token's **content** is inserted unchanged (e.g., `"${nosuch:x}" → "nosuch:x"`).
  - **ErrorOnUnknown**: unknown tokens yield `ErrNotFound`.

**Examples**

```go
os.Setenv("USER", "alice")

s, _ := resolver.ResolveString("db://u=${env:USER}@${json:/cfg/app.json//db.host}")
// → "db://u=alice@localhost"

s, _ = resolver.ResolveString(`literal \${env:USER}`)
// → "literal ${env:USER}"

s, _ = resolver.ResolveString("price is $$5 (not a token)")
// → "price is $$5 (not a token)"

// Multi-pass: a token that expands to another token
r := resolver.NewRegistry()
resolver.RegisterResolver("a:", resolver.ResolverFunc(func(_ string) (string, error) { return "${b:x}", nil }))
resolver.RegisterResolver("b:", resolver.ResolverFunc(func(_ string) (string, error) { return "OK", nil }))
s, _ = r.ResolveString("s=${a:any}")
// → "s=OK"
```

## Batch resolution

When you need to resolve a list of strings (e.g., CLI args, YAML arrays), use the slice helpers. Both preserve order, return a **new** slice, and leave inputs unchanged. Unknown schemes still **pass through** unchanged, just like `ResolveVariable`.

### `ResolveSlice` (strict)

```go
func ResolveSlice(values []string) ([]string, error)
```

Resolves each element using the default registry. If any element fails, the function **stops at the first error** and returns it. No partial results are returned.

### `ResolveSliceBestEffort`

```go
func ResolveSliceBestEffort(values []string) ([]string, []error)
```

Attempts to resolve **all** elements and never fails fast. Returns:

- `out`: resolved values (same length as input; failed items are `""`).
- `errs`: **per-index** errors you can inspect or log.

> Registry methods are also available: `(*Registry).ResolveSlice` and `(*Registry).ResolveSliceBestEffort`.

## Example

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/containeroo/resolver"
)

func main() {
    os.Setenv("MY_VAR", "HelloWorld")

    val, err := resolver.ResolveVariable("env:MY_VAR")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(val) // Output: HelloWorld

    // Example: resolve a JSON key
    // File: /config/app.json
    // {
    //   "server": {
    //     "host": "localhost",
    //     "port": 8080
    //   }
    // }
    host, err := resolver.ResolveVariable("json:/config/app.json//server.host")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(host) // Output: localhost

    // Interpolation example
    os.Setenv("USER", "alice")
    s, err := resolver.ResolveString("db://${env:USER}@${json:/config/app.json//server.host}")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(s) // db://alice@localhost
}
```

## Extensibility

You can register your own resolver schemes at runtime:

```go
resolver.RegisterResolver("secret:", myCustomResolver)
```

Resolvers must implement:

```go
type Resolver interface {
    Resolve(value string) (string, error)
}
```

You can also use the ergonomic adapter:

```go
resolver.RegisterResolver("secret:", resolver.ResolverFunc(func(v string) (string, error) {
    return fetchSecret(v), nil
}))
```

This allows you to plug in custom backends (e.g., Vault, Consul, HTTP endpoints).
