// Code generated from traust-contracts v0.7.1. DO NOT EDIT.

package types

type ModelRegistry struct {
	Escalation *ModelRegistryEscalation               `json:"escalation,omitempty"`
	Providers  map[string]ModelRegistryProvidersEntry `json:"providers"`
	Roles      map[string]ModelRegistryRolesEntry     `json:"roles"`
	Tiers      map[string]ModelRegistryTiersEntry     `json:"tiers"`
	Updated    string                                 `json:"updated"`
	Version    int                                    `json:"version"`
}

type ModelRegistryEscalation struct {
	Triggers []string `json:"triggers,omitempty"`
}

type ModelRegistryProvidersEntry struct {
	DataHandling string                                            `json:"data_handling"`
	Models       map[string]ModelRegistryProvidersEntryModelsEntry `json:"models"`
}

type ModelRegistryProvidersEntryModelsEntry struct {
	BatchEligible   bool        `json:"batch_eligible"`
	PricePerMtokIn  interface{} `json:"price_per_mtok_in,omitempty"`
	PricePerMtokOut interface{} `json:"price_per_mtok_out,omitempty"`
	Tier            string      `json:"tier"`
}

type ModelRegistryRolesEntry struct {
	Approved             []string                      `json:"approved"`
	Candidates           []string                      `json:"candidates"`
	ClaudeCodeOnly       []string                      `json:"claude_code_only,omitempty"`
	Evals                *ModelRegistryRolesEntryEvals `json:"evals,omitempty"`
	Floor                string                        `json:"floor"`
	LedgerValidityWriter bool                          `json:"ledger_validity_writer"`
	Note                 *string                       `json:"note,omitempty"`
}

type ModelRegistryRolesEntryEvals struct {
	BenchmarkedOn interface{} `json:"benchmarked_on,omitempty"`
	Consistency   interface{} `json:"consistency,omitempty"`
	Note          *string     `json:"note,omitempty"`
	RecallDelta   interface{} `json:"recall_delta,omitempty"`
}

type ModelRegistryTiersEntry struct {
	Description string `json:"description"`
}
