// Code generated from traust-contracts v0.1.0. DO NOT EDIT.

package types

type AdrRegistry struct {
	Note      *string                    `json:"note,omitempty"`
	Registers []AdrRegistryRegistersItem `json:"registers"`
	Version   int                        `json:"version"`
}

type AdrRegistryRegistersItem struct {
	DeclaredStatus *string  `json:"declared_status,omitempty"`
	Governs        []string `json:"governs,omitempty"`
	Name           string   `json:"name"`
	Note           *string  `json:"note,omitempty"`
	Paths          []string `json:"paths"`
	Pin            string   `json:"pin"`
	Repo           string   `json:"repo"`
}
