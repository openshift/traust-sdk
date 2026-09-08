# traust-sdk

Multi-language SDKs for security harness skills. Go module under `go/` today.

- **Before done:** for each language you touch, run its lint + test + security targets (see table); CI enforces all configured gates

| Language | Lint | Test | Security (hard) | Generate | Packages |
|----------|------|------|-----------------|----------|----------|
| Go | `go vet`, `gofmt` (`make lint`) | `go test ./...` | `govulncheck`, `gosec` (`make security`) | `make -C go generate` (sibling `traust-contracts`) | go modules |

Add a row when a new SDK lands; wire matching `lint:<lang>`, `test:<lang>`, `security:<lang>` jobs under `ci/lang/`.

- **Commits:** conventional `type(scope): subject` — scope = language id when SDK-specific (e.g. `feat(go): …`)
- **Releases:** per-language `{lang}/VERSION` + `## [X.Y.Z]` in `{lang}/CHANGELOG.md`; tag `{tag_prefix}/vX.Y.Z` from `ci/languages.json`
- **CI:** per-language jobs in parallel per stage; registry in `ci/languages.json`
- **Setup:** `make setup` once per clone (hooks). More: [CONTRIBUTING.md](CONTRIBUTING.md), [README.md](README.md)
