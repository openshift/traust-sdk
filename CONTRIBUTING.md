# Contributing

## Setup

```bash
git clone <repo-url> && cd traust-sdk
make setup
```

Enables git hooks. One time per clone. Go SDK requires **Go 1.26+** (`toolchain` in `go/go.mod` pins the patched release for `govulncheck`).

For codegen from contracts, also clone `traust-contracts` as a sibling directory (see [README.md](README.md)).

## Commit messages

Conventional commits required. Format: `type(scope): subject`

Types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `chore`, `ci`, `build`, `style`, `revert`

Use the language id as scope when the change is SDK-specific:

```
feat(go): add harness provider interface
fix(go): correct enum code generation
ci: update shared gates              ← no language release required
```

## Hooks

| Hook | Runs | Speed |
|------|------|-------|
| `commit-msg` | subject format check | instant |
| `pre-commit` | gofmt on staged `.go` | <1s |
| `pre-push` | `go vet` + `go test ./...` | ~10s |

Bypass: `--no-verify` on commit or push. CI still enforces.

## Releasing

Each language versions independently. Tag format: `{tag_prefix}/vX.Y.Z` (see `ci/languages.json`).

If your PR has release-worthy commits (`feat`, `fix`, `perf`, or `!`) **for a language**:

```bash
make bump go minor
# add to go/CHANGELOG.md:
#   ## [X.Y.Z]
#   ### Added
#   - ...
make check-release
git add go/VERSION go/CHANGELOG.md
git commit -m "chore(go): release X.Y.Z"
```

Only bump languages touched by the PR.

CI/infra-only PRs: use `ci:` / `chore:` commits — no VERSION bump.

Registry: `ci/languages.json` (must stay aligned with `ci/lang/*.yml`).

## CI pipeline

Stages fan out across languages in parallel:

| Stage | Jobs (parallel) |
|-------|-----------------|
| `lint` | `lint:<lang>`, … + gates (MR only) |
| `test` | `test:<lang>`, … |
| `security` | `security:<lang>`, … |
| `release` | `release:tag` — tags each language with a pending `{lang}/VERSION` bump |

Jobs auto-enable when a language's marker file exists (`ci/languages.json`). Add a language: extend the table in `AGENTS.md`, register in `ci/languages.json`, add `ci/lang/<lang>.yml` + `ci/lang/<lang>/ci.py`.

**Main:** `release.py release-if-ready --all` compares each `{lang}/VERSION` to latest `{tag_prefix}/v*` tag.

## Running tests

```bash
make test
make security   # govulncheck + gosec (required in CI)
make -C go test
make -C go check-drift   # local only; needs sibling traust-contracts
```

## Architecture

See [README.md](README.md) for module layout and contract regeneration.
