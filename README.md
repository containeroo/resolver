# Resolver

[github.com/containeroo/resolver](https://github.com/containeroo/resolver) resolves configuration values from environment variables, files, JSON, YAML, INI and TOML with explicit source prefixes. Unknown references pass through unchanged by default.

## Installation

```sh
go get github.com/containeroo/resolver
```

## Sources

| Syntax                                  | Meaning                              |
| --------------------------------------- | ------------------------------------ |
| `env:NAME`                              | Environment variable `NAME`          |
| `file:/path/token`                      | Entire file contents                 |
| `file:/path/config.env//KEY`            | Value of `KEY` from a key-value file |
| `json:/path/config.json//service.token` | JSON path                            |
| `yaml:/path/config.yaml//service.token` | YAML path                            |
| `ini:/path/config.ini//Section.Key`     | INI key                              |
| `toml:/path/config.toml//service.token` | TOML key                             |
| `literal:my:secret`                     | Literal string containing any colon  |
| `plain-secret`                          | Unprefixed literal, unchanged        |

```go
value, err := resolver.ResolveVariable("env:MY_TOKEN")
```

An unset environment variable returns `ErrNotFound`; **a set-but-empty variable is accepted by default**. The default whole-file resolver trims surrounding whitespace. To preserve whitespace, register a `KeyValueFileResolver` with an explicit `Whitespace` policy.

## Registry configuration

```go
r := resolver.NewDefaultRegistry(
    resolver.WithEmptyValuePolicy(resolver.RejectBlank),
    resolver.WithUnknownSchemePolicy(resolver.ErrorOnUnknown),
    resolver.WithMaxDepth(8),
)

secret, err := r.ResolveVariable("env:CLOUDFLARE_TOKEN")
```

Empty-value policies:

- `AllowEmpty` (default): accepts empty and whitespace-only strings.
- `RejectEmpty`: rejects only `""`.
- `RejectBlank`: rejects empty or whitespace-only strings, without trimming valid values.

Rejected empties produce `ErrEmptyValue`. Missing values still produce `ErrNotFound`.

Unknown-scheme policies:

- `PassThrough` (default): leaves unknown schemes unchanged.
- `ErrorOnUnknown`: rejects inputs starting with an unregistered scheme prefix such as `evn:KEY` or `https:`. Colons embedded in ordinary prose do not count as a scheme. To deliberately use an unknown scheme as a literal, use `literal:https://example.com`.

`NewRegistry(options...)` starts with no registered schemes; `NewDefaultRegistry(options...)` registers the built-in schemes. You can also use `SetEmptyValuePolicy` and `SetUnknownSchemePolicy` on a registry.

To allow only environment and file sources:

```go
r := resolver.NewRegistry(
    resolver.WithEmptyValuePolicy(resolver.RejectBlank),
    resolver.WithUnknownSchemePolicy(resolver.ErrorOnUnknown),
)
r.Register("env:", &resolver.EnvResolver{})
r.Register("file:", &resolver.KeyValueFileResolver{Whitespace: resolver.FileTrimNewlines})
r.Register("literal:", resolver.LiteralResolver{})
```

### File whitespace policies

- `FileTrimSpace` (default) trims surrounding whitespace, including newlines.
- `FileTrimNewlines` removes trailing carriage returns and line feeds only.
- `FilePreserveWhitespace` preserves bytes (excluding any leading UTF-8 BOM).

These policies apply when reading an entire file; key-value file parsing retains its existing whitespace rules. Values are resolved at call time, not watched for changes.

## String interpolation

Use `ResolveString` to interpolate `${scheme:value}` tokens in a string. Expansion can run multiple passes when one resolver produces another expression. The default maximum is eight passes and can be changed with `WithMaxDepth`.

```go
r := resolver.NewDefaultRegistry()
r.Register("a:", resolver.ResolverFunc(func(string) (string,error) {return "${b:x}", nil}))
r.Register("b:", resolver.ResolverFunc(func(string) (string,error) {return "OK", nil}))
value, err := r.ResolveString("${a:any}")  // "OK"
```

Escape a token as `\${env:NAME}` to leave it literal, even across multiple expansion passes. Use `ResolveVariable` for whole secret references rather than interpolation; interpolation is designed for assembling larger strings.

## Batch resolution

`ResolveSlice` returns a new slice or stops at the first error. `ResolveSliceBestEffort` attempts all values and returns a new slice plus per-index errors.

## Custom resolvers

```go
r := resolver.NewRegistry()
r.Register("custom:", resolver.ResolverFunc(func(value string) (string, error) {
    return fetchCustomValue(value)
}))
```

`RegisterResolver` registers on the package's global default registry. `Registry.Register` registers on a specific registry. `DefaultRegistry()` returns the global registry; avoid mutating it when different components need different source restrictions.
