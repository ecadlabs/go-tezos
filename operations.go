package tezos

import (
	"encoding/json"
	"strconv"
)

// OperationElem must be implemented by all operation elements
type OperationElem interface {
	OperationElemKind() string
}

// GenericOperationElem is a most generic element type
type GenericOperationElem struct {
	Kind string `json:"kind"`
}

// OperationElemKind implements OperationElem
func (e *GenericOperationElem) OperationElemKind() string {
	return e.Kind
}

// OperationElements is a slice of OperationElem with custom JSON unmarshaller
type OperationElements []OperationElem

// UnmarshalJSON implements json.Unmarshaler
func (e *OperationElements) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = make(OperationElements, len(raw))

opLoop:
	for i, r := range raw {
		var tmp GenericOperationElem
		if err := json.Unmarshal(r, &tmp); err != nil {
			return err
		}

		switch tmp.Kind {
		case "endorsement":
			(*e)[i] = &EndorsementOperationElem{}
		case "transaction":
			(*e)[i] = &TransactionOperationElem{}

		default:
			(*e)[i] = &tmp
			continue opLoop

			// TODO: add other item kinds
		}

		if err := json.Unmarshal(r, (*e)[i]); err != nil {
			return err
		}
	}

	return nil
}

// EndorsementOperationElem represents an endorsement operation
type EndorsementOperationElem struct {
	GenericOperationElem
	Level    int                           `json:"level"`
	Metadata *EndorsementOperationMetadata `json:"metadata"`
}

// EndorsementOperationMetadata represents an endorsement operation metadata
type EndorsementOperationMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
	Delegate       string         `json:"delegate"`
	Slots          []int          `json:"slots"`
}

// TransactionOperationElem represents a transaction operation
type TransactionOperationElem struct {
	GenericOperationElem
	Source       string                        `json:"source"`
	Fee          bigIntStr                     `json:"fee"`
	Counter      bigIntStr                     `json:"counter"`
	GasLimit     bigIntStr                     `json:"gas_limit"`
	StorageLimit bigIntStr                     `json:"storage_limit"`
	Amount       bigIntStr                     `json:"amount"`
	Destination  string                        `json:"destination"`
	Parameters   map[string]interface{}        `json:"parameters"`
	Metadata     *EndorsementOperationMetadata `json:"metadata"`
}

// BalanceUpdate is a variable structure depending on the Kind field
type BalanceUpdate interface {
	BalanceUpdateKind() string
}

// GenericBalanceUpdate holds the common values among all BalanceUpdatesType variants
type GenericBalanceUpdate struct {
	Kind   string        `json:"kind"`
	Change BalanceChange `json:"change"`
}

// BalanceUpdateKind returns the BalanceUpdateType's Kind field
func (g *GenericBalanceUpdate) BalanceUpdateKind() string {
	return g.Kind
}

// ContractBalanceUpdate is a BalanceUpdatesType variant for Kind=contract
type ContractBalanceUpdate struct {
	GenericBalanceUpdate
	Contract string `json:"contract"`
}

// FreezerBalanceUpdate is a BalanceUpdatesType variant for Kind=freezer
type FreezerBalanceUpdate struct {
	GenericBalanceUpdate
	Category string `json:"category"`
	Delegate string `json:"delegate"`
	Level    int    `json:"level"`
}

// BalanceUpdates is a list of balance update operations
type BalanceUpdates []BalanceUpdate

// UnmarshalJSON implements json.Unmarshaler
func (b *BalanceUpdates) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*b = make(BalanceUpdates, len(raw))

opLoop:
	for i, r := range raw {
		var tmp GenericBalanceUpdate
		if err := json.Unmarshal(r, &tmp); err != nil {
			return err
		}

		switch tmp.Kind {
		case "contract":
			(*b)[i] = &ContractBalanceUpdate{}

		case "freezer":
			(*b)[i] = &FreezerBalanceUpdate{}

		default:
			(*b)[i] = &tmp
			continue opLoop
		}

		if err := json.Unmarshal(r, (*b)[i]); err != nil {
			return err
		}
	}

	return nil
}

// BalanceChange is a string encoded int64
type BalanceChange int64

// UnmarshalJSON implements json.Unmarshaler
func (b *BalanceChange) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}

	*(*int64)(b), err = strconv.ParseInt(s, 0, 64)

	return err
}

// Operation represents an operation included into block
type Operation struct {
	Protocol  string            `json:"protocol"`
	ChainID   string            `json:"chain_id"`
	Hash      string            `json:"hash"`
	Branch    string            `json:"branch"`
	Contents  OperationElements `json:"contents"`
	Signature string            `json:"signature"`
}

/*
OperationAlt is a heterogeneously encoded Operation with hash as a first array member, i.e.
	[
		"...", // hash
		{
			"protocol": "...",
			...
		}
	]
instead of
	{
		"protocol": "...",
		"hash": "...",
		...
	}
*/
type OperationAlt Operation

// UnmarshalJSON implements json.Unmarshaler
func (o *OperationAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &o.Hash, (*Operation)(o))
}

// OperationWithError represents unsuccessful operation
type OperationWithError struct {
	Operation
	Error Errors `json:"error"`
}

// OperationWithErrorAlt is a heterogeneously encoded OperationWithError with hash as a first array member.
// See OperationAlt for details
type OperationWithErrorAlt OperationWithError

// UnmarshalJSON implements json.Unmarshaler
func (o *OperationWithErrorAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &o.Hash, (*OperationWithError)(o))
}
