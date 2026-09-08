// Package query is the read surface for the ledger service.
//
// Use it to retrieve layers, resolve finding dispositions, paginate across
// the ledger, verify Merkle integrity, or probe service health. As the
// ledger gains new read endpoints they will be added here, keeping all
// query logic in one place for consumers.
//
// The package is transport-agnostic: callers supply a [Provider]
// implementation that performs the actual HTTP call. The SDK owns response
// decoding; Provider owns delivery.
//
//	client := query.NewHTTPClient("https://ledger.example.com",
//	    query.WithBearerToken(os.Getenv("LEDGER_TOKEN")),
//	)
//	findings, err := client.GetFindings(ctx, "repo-a")
package query
