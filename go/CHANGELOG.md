# Changelog

All notable changes to the Go SDK are documented here.

## [0.11.0]

### Changed
- bumped contracts and ran code gen
- renamed to traust-sdk

## [0.10.1]

**Open-source readiness clean-up.**

### Changed
- Go module path: `github.com/openshift/traust-sdk/go`.
- Regenerated types, enums, and embedded schemas from traust-contracts 0.7.1.
- Genericized internal references across all source comments and doc strings.
- Added `SECURITY.md`.

## [0.10.0]

Fix report-lane fingerprint submission (triage, validation, verification).

### Fixed

- Report-lane submits (`SubmitTriageReport`, `SubmitValidationReport`,
  `SubmitVerificationReport`) no longer fail with
  `finding "X" missing fingerprint`. Triage/validation/verification reports do
  not carry fingerprints on their findings — the fingerprint is a scan-lane
  identity — so the pre-conversion report check could never pass. The check now
  runs post-conversion at the event level.

### Added

- `FingerprintIndex map[string]string` (finding_ref → fingerprint) on
  `TriageReportInput`, `ValidationReportInput`, and `VerificationReportInput`.
  Callers supply the scan-lane fingerprints; conversion stamps each derived
  event. The field is not serialized (`json:"-"`); only the derived events
  travel to the ledger. The SDK never computes fingerprints and never sets
  `fingerprint_algo` (the ledger owns both).

### Changed

- **BREAKING (direct converter callers):** `ConvertTriageReport`,
  `ConvertValidationReport`, and `ConvertVerificationReport` take an additional
  `fpIndex map[string]string` argument. Callers using the `Submit*Report`
  client methods are unaffected — pass the index via the input's
  `FingerprintIndex`.

## [0.9.1]

### Added

- `ConvertVerificationReport` — maps verification report verdicts to
  `disposition.resolution` events (or `validity` for `false_positive`),
  including `cross_repo.propagation == pending` → `fix_in_progress` override.
- `SubmitVerificationReport` on ingest `Client`.
- `VerificationReportInput` type in `v1/ingest`.
- `KindVerification` and `SchemaVerification` constants.
- Empty `OriginalId` findings now reported in `Skipped` with reason
  (previously silently dropped).

## [0.9.0]

**Align with ledger v0.18.x API.**

### Added

- `SignLayer` method on ingest `Client` — request server-side merkle signing via
  `POST /v1/ledger/layers/{id}/sign`. Supports optional `?rekor=true` query param.
- `SignResponse` and `SignOpts` types in `v1/ingest`.
- `ingesttest.FixtureSignResponse` and `WithSignResponse` test helpers.

### Changed

- `SubmitResponse` gains `EventIDs []string` and `QueueAdded int` fields, matching
  the batch submit response from ledger 0.17+.
- `ingesttest.FixtureTriageResponse` now includes `EventIDs` and `EventCount`.

### Fixed

- README: removed references to deleted report kinds (`code-audit`, `cloud-config`,
  etc.) and stale `@v0.3.0` install instructions.

## [0.8.0]

**Breaking: ledger v0.16.0 API alignment.**

### Removed

- `SubmitCodeAuditReport`, `SubmitCloudConfigReport`, `SubmitContainerAuditReport`,
  `SubmitThreatModelReport`, `SubmitPQCReport`, `SubmitInventoryReport`,
  `SubmitRemediationReport`, `SubmitVerificationReport` — these methods targeted
  `POST /v1/ledger/reports` which no longer exists. They never produced events
  server-side (the ledger returned `event_count=0` for non-verdict kinds).
- `POST /v1/ledger/reports` transport path removed from HTTP provider.
- Input types: `CodeAuditReportInput`, `CloudConfigReportInput`,
  `ContainerAuditReportInput`, `RawReportInput`, `RemediationReportInput`,
  `VerificationReportInput`.
- Dead `IngestKind` constants for removed report types.

### Changed

- `SubmitTriageReport` and `SubmitValidationReport` now convert reports to
  events client-side and batch-submit via `POST /v1/ledger/layers/{id}/submit`.
  The server no longer runs converters — the SDK owns the report→event transform.
- `Provider` interface gains `Submit(ctx, layerID, payload)` and
  `Post(ctx, path, payload)` methods (breaking for custom provider implementations).
- `NewHTTPClient` signature simplified (no longer needs `EventLaneKindStrings`).

### Added

- `BatchSubmit` — submit pre-formed events directly to a layer.
- `ResolveReviewItem` — resolve a needs_review queue item.
- `ComputeFingerprints` — stamp finding fingerprints via the service.
- `ConvertTriageReport` / `ConvertValidationReport` — public converter functions
  for callers who want the events without submitting.
- `ConvertResult` type for converter output.
- `ListLayers` query method (`GET /v1/ledger/layers`).
- `FindingDisposition.Fingerprint` and `FindingDisposition.Orphan` fields.
- `PhaseConvert` error phase for converter failures.

## [0.7.2]

**Contracts 0.7.0 codegen sync.**

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts
  **0.7.0** (`ContractsVersion` now `0.7.0`).
- `LayerReviewQueueReasonNeedsIdentity` enum value for unstamped findings that
  queue instead of recording.

## [0.7.1]

**Contracts 0.6.1 codegen sync.**

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts
  **0.6.1** (`ContractsVersion` now `0.6.1`).
- `Layer` gains optional `baseline_claims`; `LayerMetadata` gains `artifact_digests`
  and `external_refs` (CVE/provenance stamping from contracts 0.5.4+).
- Embedded JSON schemas updated (identity description corrections from 0.6.1).

## [0.7.0]

### Added

- `ListEvents` query method — `GET /v1/ledger/layers/{id}/events` with
  `finding_ref`, `source_type`, `limit`, and `offset` filters.
- `EventsResponse` and `ListEventsOpts` types in `v1/query`.
- `querytest.WithEventsResponse` test helper for mocking event responses.

## [0.6.0]

**Read/query client for the ledger.**

### Added

- `v1/query` package — typed read client for ledger service GET endpoints
  (`GetLayer`, `GetFindings`, `ListFindings`, `VerifyLayer`, `Health`).
- Built-in HTTP transport (`NewHTTPClient`, bearer-token and custom-client options).
- `v1/query/querytest` — `StaticProvider` test double and fixture helpers.

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts
  **0.5.3** (`ContractsVersion` now `0.5.3`). No schema shape changes; version stamp
  only.

## [0.5.4]

### Added

- `SubmitRemediationReport`, `SubmitVerificationReport` typed ingest ops with
  schema validation against embedded `remediation` / `verification` contracts.
- `RemediationReportInput`, `VerificationReportInput` typed input structs
  (replace `RawReportInput` catch-all for these kinds).
- `KindRemediation`, `KindVerification` ingest kind constants.

## [0.5.3]

**Align ingest API with the ledger service (0.7.0).**

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts **0.5.2**
  (`ContractsVersion` now `0.5.2`). Adds OIDC-era actor identity fields on `Actor`,
  content-addressed `audit_report_sha256` / `audit_report_ref` on `LayerMetadata`, and
  `RepoScopePath` enum.
- **Ingest kinds** now match ledger report/event kinds (`triage`, `validation`,
  `code-audit`, …; `countersign`, `severity`) instead of legacy `scan-result` /
  `triage-result` names.
- **Report inputs** use ledger service payload shape: `layer_id`, `source_ref`, `recorded_at`,
  `report`.
- **Event inputs** require `layer_id` and `recorded_at`; countersign accepts
  `finding_ref` (or `finding_fingerprint` alias server-side).
- `SubmitResponse` includes optional `event_count` and `merkle_root`.

### Added

- `SubmitValidationReport`, `SubmitCodeAuditReport`, `SubmitCloudConfigReport`,
  `SubmitContainerAuditReport`, `SubmitThreatModelReport`, `SubmitPQCReport`,
  `SubmitSeverity`.
- `ReportMeta`, `EventMeta`, `RawReportInput` shared envelope types.

### Removed

- `SubmitScanResult`, `SubmitTriageResult`, `SubmitRemediationResult`,
  `SubmitVerificationResult`, `SubmitInventory` and their legacy input structs.
- Legacy ingest kind constants (`scan-result`, `triage-result`, …).

### Fixed

- Enum codegen treats `:` as a word separator (fixes `repo:maintenance`-style values).
- Event-lane HTTP envelope includes required `kind` field for the ledger service.

## [0.5.2]

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts **0.4.4**
  (`ContractsVersion` now `0.4.4`). Adds `EventAlias`, `EventFinding` on layer events,
  `SourceTypeVulnScanReport`, and report `DependencyProvenance` types.

## [0.5.1] — 2026-08-17

### Added

- `v1/skills/names.go` — harness skill identifier constants (`SkillSecureCodeAudit`,
  `SkillTriage`, `SkillVulnScan`, etc.) for `SkillMeta` construction and matching.

## [0.5.0] — 2026-08-17

**Prune to transport + contract surface; drop ledger-semantic client APIs.**

### Added

- `v1/ingest/http.go` and `v1/ingest/internal/transport` — built-in HTTP client for
  the ledger service (`NewHTTPClient`, bearer-token and custom-client options).

### Removed

- `v1/identity` package (finding fingerprint helpers and golden vectors).
- `v1/ingest/event` package — lifecycle DTOs, actor factories, source mapping,
  `ReadDisposition`, and event-id golden vectors.
- Ledger-semantic ingest helpers: `event.New`, `Emit*`, `SubmitLedgerEvent`.

### Fixed

- `FingerprintError` deduplication in ingest error handling.
- Unexported internal `Op` type; README and package docs updated for the slimmer API.

### Compatibility

Breaking release. Consumers on 0.4.x that used lifecycle events, actor factories, or
the identity package must migrate to contract types + transport-only ingest, or stay
pinned below 0.5.0.

## [0.4.0] — 2026-08-17

**Close ledger-ingest gaps for lifecycle events and disposition reads.**

### Added

- `v1/ingest/event/lifecycle.go` — lifecycle event construction helpers.
- `v1/ingest/event/actor.go` — actor factory helpers for ingest events.
- `v1/ingest/event/source.go` — evidence source-type mapping.
- `v1/ingest/event/read.go` — read current disposition from ledger state.
- Verification emit path on the ingest client.

## [0.3.0] — 2026-08-17

**Async skill execution and contracts 0.4.1 codegen sync.**

### Added

- `AsyncProvider` interface (`Dispatch` / `Collect`) and `JobRef` handle type.
- `AsyncClient` and `SyncAdapter` (wraps async providers as synchronous `Provider`).
- `skillstest` async provider helpers for tests.
- `version.go` generation — `types.ContractsVersion` constant from contracts `VERSION`.

### Changed

- Regenerated `v1/types`, `v1/enums`, and `v1/validate` from traust-contracts
  **0.4.1**.

## [0.2.0] — 2026-08-14

### Added

- `v1/ingest` package — typed submission of scan results and human events to the
  ledger (`Provider` interface, `ingesttest` fixtures, fingerprint checks).
- `v1/ingest/event` — event construction and event-id golden vectors.
- Enum codegen for `actor_kind`, `disposition_embargo`, `event_validity`, `source_type`.

### Fixed

- Codegen integer-type mapping bug in `internal/generate/types.go`.

## [0.1.0] — 2026-08-13

### Added

- Initial Go SDK: `v1/skills` typed skill runner (`Provider`, `Client`, `skillstest`).
- `v1/types` — generated structs for every contracts `schemas/v1` document.
- `v1/enums` — generated shared vocabularies (severity, validity, verdicts, …).
- `v1/validate` — validate raw JSON against embedded v1 schemas.
- `internal/generate` — codegen from sibling `traust-contracts` schemas and enums.
