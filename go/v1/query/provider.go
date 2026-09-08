package query

import "context"

// Provider performs read queries against the ledger service and returns the
// raw response bytes. Implementations (HTTP client, mock, etc.) live outside
// this package.
type Provider interface {
	Query(ctx context.Context, method, path string) ([]byte, error)
}
