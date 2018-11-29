package tezos

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// RawBlockHeader is a part of the Tezos block data
type RawBlockHeader struct {
	Level           int32     `json:"level"`
	Proto           byte      `json:"proto"`
	Predecessor     string    `json:"predecessor"`
	Timestamp       time.Time `json:"timestamp"`
	ValidationPass  byte      `json:"validation_pass"`
	OperationsHash  string    `json:"operations_hash"`
	Fitness         [][]byte  `json:"fitness"`
	Context         string    `json:"context"`
	Priority        int16     `json:"priority"`
	ProofOfWorkNone []byte    `json:"proof_of_work_nonce"`
	SeedNonceHash   string    `json:"seed_nonce_hash"`
	Signature       string    `json:"signature"`
}

// TestChainStatusType is a variable structure depending on the Status field
type TestChainStatusType interface {
	GetStatus() string
}

// GenericTestChainStatus holds the common values among all TestChainStatusType variants
type GenericTestChainStatus struct {
	Status string `json:"status"`
}

// GetStatus gets the TestChainStatusType's Status field
func (tcs GenericTestChainStatus) GetStatus() string {
	return tcs.Status
}

// NotRunningTestChainStatus is a TestChainStatusType variant for Status=not_running
type NotRunningTestChainStatus struct {
	GenericTestChainStatus
}

// ForkingTestChainStatus is a TestChainStatusType variant for Status=forking
type ForkingTestChainStatus struct {
	GenericTestChainStatus
	Protocol   string `json:"protocol"`
	Expiration string `json:"expiration"`
}

// RunningTestChainStatus is a TestChainStatusType variant for Status=running
type RunningTestChainStatus struct {
	GenericTestChainStatus
	ChainID    string `json:"chain_id"`
	Genesis    string `json:"genesis"`
	Protocol   string `json:"protocol"`
	Expiration string `json:"expiration"`
}

// MaxOperationListLengthType is a part of the BlockHeaderMetadata
type MaxOperationListLengthType struct {
	MaxSize int32 `json:"max_size"`
	MaxOp   int32 `json:"max_op"`
}

// LevelType is a part of BlockHeaderMetadata
type LevelType struct {
	Level                int32 `json:"level"`
	LevelPosition        int32 `json:"level_position"`
	Cycle                int32 `json:"cycle"`
	CyclePosition        int32 `json:"cycle_position"`
	VotingPeriod         int32 `json:"voting_period"`
	VotingPeriodPosition int32 `json:"voting_period_position"`
	ExpectedCommitment   bool  `json:"expected_commitment"`
}

// BalanceUpdateType is a variable structure depending on the Kind field
type BalanceUpdateType interface {
	GetKind() string
}

// GenericBalanceUpdate holds the common values among all BalanceUpdatesType variants
type GenericBalanceUpdate struct {
	Kind   string `json:"kind"`
	Change string `json:"change"`
}

// GetKind returns the BalanceUpdateType's Kind field
func (gbu GenericBalanceUpdate) GetKind() string {
	return gbu.Kind
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
	Level    int32  `json:"level"`
}

type BalanceUpdatesType []BalanceUpdateType

func (bus *BalanceUpdatesType) UnmarshalJSON(data []byte) error {
	aTemp := []struct {
		Kind string `json:"kind"`
	}{}
	if err := json.Unmarshal(data, &aTemp); err != nil {
		return err
	}

	aRaw := []json.RawMessage{}
	if err := json.Unmarshal(data, &aRaw); err != nil {
		return err
	}

	*bus = make(BalanceUpdatesType, len(aTemp))
	for i, bu := range aTemp {
		// Resolve the multi-variant BalanceUpdates field
		switch bu.Kind {
		case "contract":
			cbu := ContractBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: bu.Kind}}
			if err := json.Unmarshal(aRaw[i], &cbu); err != nil {
				return err
			}
			(*bus)[i] = cbu
		case "freezer":
			fbu := FreezerBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: bu.Kind}}
			if err := json.Unmarshal(aRaw[i], &fbu); err != nil {
				return err
			}
			(*bus)[i] = fbu
		default:
			return fmt.Errorf("Unknown BalanceUpdates.Kind: %v", bu.Kind)
		}
	}

	return nil
}

// BlockHeaderMetadata is a part of the Tezos block data
type BlockHeaderMetadata struct {
	Protocol               string                       `json:"protocol"`
	NextProtocol           string                       `json:"next_protocol"`
	TestChainStatus        TestChainStatusType          `json:"test_chain_status"`
	MaxOperationsTTL       int32                        `json:"max_operations_ttl"`
	MaxOperationDataLength int32                        `json:"max_operation_data_length"`
	MaxBlockHeaderLength   int32                        `json:"max_block_header_length"`
	MaxOperationListLength []MaxOperationListLengthType `json:"max_operation_list_length"`
	Baker                  string                       `json:"baker"`
	Level                  LevelType                    `json:"level"`
	VotingPeriodKind       string                       `json:"voting_period_kind"`
	NonceHash              string                       `json:"nonce_hash"`
	ConsumedGas            string                       `json:"consumed_gas"` // TODO: replace with bigIntStr when merged
	Deactivated            []string                     `json:"deactivated"`
	BalanceUpdates         BalanceUpdatesType           `json:"balance_updates"`
}

// UnmarshalJSON unmarshals the BlockHeaderMetadata JSON
func (bhm *BlockHeaderMetadata) UnmarshalJSON(data []byte) error {
	var temp struct {
		TestChainStatus struct {
			Status string `json:"status"`
		} `json:"test_chain_status"`
		BalanceUpdates []struct {
			Kind string `json:"kind"`
		} `json:"balance_updates"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Resolve the multi-variant TestChainStatus field
	switch temp.TestChainStatus.Status {
	case "not_running":
		bhm.TestChainStatus = &NotRunningTestChainStatus{}
	case "forking":
		bhm.TestChainStatus = &ForkingTestChainStatus{}
	case "running":
		bhm.TestChainStatus = &RunningTestChainStatus{}
	default:
		return fmt.Errorf("Unknown TestChainStatus.Status: %v", temp.TestChainStatus.Status)
	}
	log.Println("BlockHeaderMetadata")

	type tempBHM BlockHeaderMetadata
	err := json.Unmarshal(data, (*tempBHM)(bhm))

	log.Println(jsonifyWhatever(bhm))
	return err
}

// Block holds information about a Tezos block
type Block struct {
	Protocol string              `json:"protocol"`
	ChainID  string              `json:"chain_id"`
	Hash     string              `json:"hash"`
	Header   RawBlockHeader      `json:"header"`
	Metadata BlockHeaderMetadata `json:"metadata"`
	//Operations []Operation `json:"operations"`
}
