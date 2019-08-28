package tezos

import (
	"encoding/json"
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
		case "ballot":
			(*e)[i] = &BallotOperationElem{}
		case "proposals":
			(*e)[i] = &ProposalOperationElem{}
		case "seed_nonce_revelation":
			(*e)[i] = &SeedNonceRevelationOperationElem{}
		case "double_endorsement_evidence":
			(*e)[i] = &DoubleEndorsementEvidenceOperationElem{}
		case "double_baking_evidence":
			(*e)[i] = &DoubleBakingEvidenceOperationElem{}
		case "activate_account":
			(*e)[i] = &ActivateAccountOperationElem{}
		case "reveal":
			(*e)[i] = &RevealOperationElem{}
		case "origination":
			(*e)[i] = &OriginationOperationElem{}
		case "delegation":
			(*e)[i] = &DelegationOperationElem{}
		default:
			(*e)[i] = &tmp
			continue opLoop
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
	Level    int                          `json:"level"`
	Metadata EndorsementOperationMetadata `json:"metadata"`
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
	Source       string                       `json:"source"`
	Fee          BigInt                       `json:"fee"`
	Counter      BigInt                       `json:"counter"`
	GasLimit     BigInt                       `json:"gas_limit"`
	StorageLimit BigInt                       `json:"storage_limit"`
	Amount       BigInt                       `json:"amount"`
	Destination  string                       `json:"destination"`
	Parameters   map[string]interface{}       `json:"parameters,omitempty"`
	Metadata     TransactionOperationMetadata `json:"metadata"`
}

// TransactionOperationMetadata represents a transaction operation metadata
type TransactionOperationMetadata struct {
	BalanceUpdates  BalanceUpdates             `json:"balance_updates"`
	OperationResult TransactionOperationResult `json:"operation_result"`
}

// TransactionOperationResult represents a transaction operation result
type TransactionOperationResult struct {
	Status              string                 `json:"status"`
	Storage             map[string]interface{} `json:"storage,omitempty"`
	BalanceUpdates      BalanceUpdates         `json:"balance_updates,omitempty"`
	OriginatedContracts []string               `json:"originated_contracts,omitempty"`
	ConsumedGas         *BigInt                `json:"consumed_gas,omitempty"`
	StorageSize         *BigInt                `json:"storage_size,omitempty"`
	PaidStorageSizeDiff *BigInt                `json:"paid_storage_size_diff,omitempty"`
	Errors              Errors                 `json:"errors,omitempty"`
}

// BallotOperationElem represents a ballot operation
type BallotOperationElem struct {
	GenericOperationElem
	Source   string                 `json:"source"`
	Period   int                    `json:"period"`
	Proposal string                 `json:"proposal"`
	Ballot   string                 `json:"ballot"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ProposalOperationElem represents a proposal operation
type ProposalOperationElem struct {
	GenericOperationElem
	Source    string                 `json:"source"`
	Period    int                    `json:"period"`
	Proposals []string               `json:"proposals"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SeedNonceRevelationOperationElem represents seed_nonce_revelation operation
type SeedNonceRevelationOperationElem struct {
	GenericOperationElem
	Level    int32                           `json:"level"`
	Nonce    string                          `json:"nonce"`
	Metadata BalanceUpdatesOperationMetadata `json:"metadata"`
}

// BalanceUpdatesOperationMetadata contains balance updates only
type BalanceUpdatesOperationMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
}

// InlinedEndorsement corresponds to $inlined.endorsement
type InlinedEndorsement struct {
	Branch     string                     `json:"branch"`
	Operations InlinedEndorsementContents `json:"operations"`
	Signature  string                     `json:"signature"`
}

// InlinedEndorsementContents corresponds to $inlined.endorsement.contents
type InlinedEndorsementContents struct {
	Kind  string `json:"endorsement"`
	Level int    `json:"level"`
}

// DoubleEndorsementEvidenceOperationElem represents double_endorsement_evidence operation
type DoubleEndorsementEvidenceOperationElem struct {
	GenericOperationElem
	Operation1 InlinedEndorsement              `json:"op1"`
	Operation2 InlinedEndorsement              `json:"op2"`
	Metadata   BalanceUpdatesOperationMetadata `json:"metadata"`
}

// DoubleBakingEvidenceOperationElem represents double_baking_evidence operation
type DoubleBakingEvidenceOperationElem struct {
	GenericOperationElem
	BlockHeader1 RawBlockHeader                  `json:"bh1"`
	BlockHeader2 RawBlockHeader                  `json:"bh2"`
	Metadata     BalanceUpdatesOperationMetadata `json:"metadata"`
}

// ActivateAccountOperationElem represents activate_account operation
type ActivateAccountOperationElem struct {
	GenericOperationElem
	PKH      string                          `json:"pkh"`
	Secret   string                          `json:"secret"`
	Metadata BalanceUpdatesOperationMetadata `json:"metadata"`
}

// RevealOperationElem represents a reveal operation
type RevealOperationElem struct {
	GenericOperationElem
	Source       string                  `json:"source"`
	Fee          BigInt                  `json:"fee"`
	Counter      BigInt                  `json:"counter"`
	GasLimit     BigInt                  `json:"gas_limit"`
	StorageLimit BigInt                  `json:"storage_limit"`
	PublicKey    string                  `json:"public_key"`
	Metadata     RevealOperationMetadata `json:"metadata"`
}

// RevealOperationMetadata represents a reveal operation metadata
type RevealOperationMetadata DelegationOperationMetadata

// OriginationOperationElem represents a origination operation
type OriginationOperationElem struct {
	GenericOperationElem
	Source        string                       `json:"source"`
	Fee           BigInt                       `json:"fee"`
	Counter       BigInt                       `json:"counter"`
	GasLimit      BigInt                       `json:"gas_limit"`
	StorageLimit  BigInt                       `json:"storage_limit"`
	ManagerPubKey string                       `json:"managerPubkey"`
	Balance       BigInt                       `json:"balance"`
	Spendable     *bool                        `json:"spendable,omitempty"`
	Delegatable   *bool                        `json:"delegatable,omitempty"`
	Delegate      string                       `json:"delegate,omitempty"`
	Script        *ScriptedContracts           `json:"script,omitempty"`
	Metadata      OriginationOperationMetadata `json:"metadata"`
}

// ScriptedContracts corresponds to $scripted.contracts
type ScriptedContracts struct {
	Code    map[string]interface{} `json:"code"`
	Storage map[string]interface{} `json:"storage"`
}

// OriginationOperationMetadata represents a origination operation metadata
type OriginationOperationMetadata struct {
	BalanceUpdates  BalanceUpdates             `json:"balance_updates"`
	OperationResult OriginationOperationResult `json:"operation_result"`
}

// OriginationOperationResult represents a origination operation result
type OriginationOperationResult struct {
	Status              string         `json:"status"`
	BalanceUpdates      BalanceUpdates `json:"balance_updates,omitempty"`
	OriginatedContracts []string       `json:"originated_contracts,omitempty"`
	ConsumedGas         *BigInt        `json:"consumed_gas,omitempty"`
	StorageSize         *BigInt        `json:"storage_size,omitempty"`
	PaidStorageSizeDiff *BigInt        `json:"paid_storage_size_diff,omitempty"`
	Errors              Errors         `json:"errors,omitempty"`
}

// DelegationOperationElem represents a delegation operation
type DelegationOperationElem struct {
	GenericOperationElem
	Source        string                      `json:"source"`
	Fee           BigInt                      `json:"fee"`
	Counter       BigInt                      `json:"counter"`
	GasLimit      BigInt                      `json:"gas_limit"`
	StorageLimit  BigInt                      `json:"storage_limit"`
	ManagerPubKey string                      `json:"managerPubkey"`
	Balance       BigInt                      `json:"balance"`
	Spendable     *bool                       `json:"spendable,omitempty"`
	Delegatable   *bool                       `json:"delegatable,omitempty"`
	Delegate      string                      `json:"delegate,omitempty"`
	Script        *ScriptedContracts          `json:"script,omitempty"`
	Metadata      DelegationOperationMetadata `json:"metadata"`
}

// DelegationOperationMetadata represents a delegation operation metadata
type DelegationOperationMetadata struct {
	BalanceUpdates  BalanceUpdates            `json:"balance_updates"`
	OperationResult DelegationOperationResult `json:"operation_result"`
}

// DelegationOperationResult represents a delegation operation result
type DelegationOperationResult struct {
	Status string `json:"status"`
	Errors Errors `json:"errors"`
}

// BalanceUpdate is a variable structure depending on the Kind field
type BalanceUpdate interface {
	BalanceUpdateKind() string
}

// GenericBalanceUpdate holds the common values among all BalanceUpdatesType variants
type GenericBalanceUpdate struct {
	Kind   string `json:"kind"`
	Change int64  `json:"change,string"`
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
