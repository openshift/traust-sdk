# Changelog

All notable changes to the Go SDK are documented here.

## [0.1.1]

## Changes

- `occurred_at` no longer gets a midnight suffix appended to a value that
  already carries a time. Three duplicate `*OccurredAt` helpers collapse into
  one `eventOccurredAt`; a report date with a time component passes through
  unchanged, and only a bare date is padded. Previously a timestamped report
  date produced `"2026-01-16T00:00:00ZT00:00:00+00:00"`, which the ledger's
  dict-only write path stores without complaint and then cannot read back.

- The schema compiler now calls `AssertFormat()`. Without it `format` is
  annotation-only, so `format: date` and `format: date-time` accepted any
  string and report validation passed values the Python side rejects.

### Note

`traust-ledger` >= 0.1.1 enforces RFC 3339 on event timestamps when reading a
layer, but its write path does not validate, so a malformed `occurred_at`
submitted by an older SDK is accepted and only fails on the next read. Upgrade
any Go producer before it writes.

## [0.1.0]

**First public release.**

Typed Go client for the traust-ledger service: convert and submit
triage/validation/verification reports as disposition events, request
layer merkle signing, and query layers/findings/events. Generated
`v1/types`, `v1/enums`, and `v1/validate` track `traust-contracts` schemas.
