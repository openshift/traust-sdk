package query_test

import (
	"context"
	"fmt"

	"github.com/openshift/traust-sdk/go/v1/query"
	"github.com/openshift/traust-sdk/go/v1/query/querytest"
)

func ExampleNewClient() {
	provider := querytest.NewStaticProvider().
		WithFindingsResponse("repo-a", querytest.FixtureFindingsResponse())
	client := query.NewClient(provider)

	findings, err := client.GetFindings(context.Background(), "repo-a")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("findings: %d\n", len(findings.Findings))
	// Output: findings: 1
}

func ExampleNewHTTPClient() {
	client := query.NewHTTPClient("https://ledger.example.com",
		query.WithBearerToken("my-token"),
	)

	_, _ = client.GetFindings(context.Background(), "repo-a")
}
