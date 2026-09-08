# traust-sdk

Typed SDKs for invoking security harness skills with pluggable execution backends.

## Setup

```bash
make setup   # enable git hooks (once per clone)
make test    # go test ./...
make status  # per-language versions and tags
```

Each language has its own `{lang}/VERSION` and git tag (`{tag_prefix}/vX.Y.Z` in `ci/languages.json`).

## Language SDKs

| Language | Path | Module |
|---|---|---|
| Go | [`go/`](go/) | `github.com/openshift/traust-sdk/go` |

## Architecture

```
traust-contracts        ← schemas, enums (source of truth)
traust-sdk              ← typed SDKs that consumers import (this repo)
```

The SDK is opinionated on **contracts** (input/output shapes validated against
schemas from `traust-contracts`) and unopinionated on **execution** (you
implement a Provider interface to run skills however you want).

## Regenerating from contracts

```bash
cd go/
make generate   # reads schemas/enums from sibling traust-contracts repo
make test       # verify everything passes
```

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
