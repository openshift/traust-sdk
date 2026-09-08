// Package ingest provides typed submission of analysis results to the ledger
// service. It enforces contractual guarantees (schema validation,
// fingerprint presence, contracts version stamping) client-side so that
// producers cannot submit malformed data.
//
// The package is transport-agnostic: callers supply a Provider implementation
// that performs the actual HTTP call. The SDK owns validation; Provider owns
// delivery.
//
//	client := ingest.NewHTTPClient("https://ledger.example.com",
//	    ingest.WithBearerToken(os.Getenv("LEDGER_TOKEN")),
//	)
//	resp, err := client.SubmitTriageReport(ctx, ingest.TriageReportInput{...})
package ingest
