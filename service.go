package tezos

import (
	"context"
	"net/http"
	"path"
)

// NetworkService is an interface for retrieving network stats from the Tezos RPC API.
type NetworkService interface {
	// GetStats implements https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat.
	GetStats(context.Context) (*NetworkStats, error)
	// GetConnections implements http://tezos.gitlab.io/mainnet/api/rpc.html#get-network-connections.
	GetConnections(context.Context) ([]NetworkConnection, error)
}

// ContractService is an interface for retrieving contract-related information from the Tezos RPC API.
type ContractService interface {
	// GetDelegateBalance implements http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-delegates-pkh-balance.
	GetDelegateBalance(ctx context.Context, chainID string, blockID string, pkh string) (string, error)
	// GetContractBalance implements http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-contracts-contract-id-balance.
	GetContractBalance(ctx context.Context, chainID string, blockID string, contractID string) (string, error)
}

// Service implements fetching of information from Tezos nodes via JSON.
type Service struct {
	client *RPCClient
}

var _ NetworkService = &Service{}
var _ ContractService = &Service{}

// GetStats returns current network stats https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
func (s *Service) GetStats(ctx context.Context) (*NetworkStats, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "/network/stat")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	stats := new(NetworkStats)
	//TODO make use of response, only works for GETs
	err = s.client.Get(ctx, req, stats)
	if err != nil {
		return nil, err
	}
	return stats, err
}

// GetConnections returns all network connections http://tezos.gitlab.io/mainnet/api/rpc.html#get-network-connections
func (s *Service) GetConnections(ctx context.Context) ([]NetworkConnection, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "/network/connections")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	conns := []NetworkConnection{}
	err = s.client.Get(ctx, req, &conns)
	if err != nil {
		return nil, err
	}
	return conns, err
}

// GetDelegateBalance returns a delegate's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-delegates-pkh-balance
func (s *Service) GetDelegateBalance(ctx context.Context, chainID string, blockID string, pkh string) (string, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "chains", chainID, "blocks", blockID, "context/delegates", pkh, "balance")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return "", err
	}

	var balance string
	err = s.client.Get(ctx, req, &balance)
	if err != nil {
		return "", err
	}
	return balance, err
}

// GetContractBalance returns a contract's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-contracts-contract-id-balance
func (s *Service) GetContractBalance(ctx context.Context, chainID string, blockID string, contractID string) (string, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "chains", chainID, "blocks", blockID, "context/contracts", contractID, "balance")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return "", err
	}

	var balance string
	err = s.client.Get(ctx, req, &balance)
	if err != nil {
		return "", err
	}
	return balance, err
}

// NetworkStats models global network bandwidth totals and usage in B/s.
type NetworkStats struct {
	TotalBytesSent int64 `json:"total_sent,string"`
	TotalBytesRecv int64 `json:"total_recv,string"`
	CurrentInflow  int64 `json:"current_inflow"`
	CurrentOutflow int64 `json:"current_outflow"`
}

// NetworkConnection models detailed information for one network connection.
type NetworkConnection struct {
	Incoming         bool             `json:"incoming"`
	PeerID           string           `json:"peer_id"`
	IDPoint          NetworkIDPoint   `json:"id_point"`
	RemoteSocketPort uint16           `json:"remote_socket_port"`
	Versions         []NetworkVersion `json:"versions"`
	Private          bool             `json:"private"`
	LocalMetadata    NetworkMetadata  `json:"local_metadata"`
	RemoteMetadata   NetworkMetadata  `json:"remote_metadata"`
}

// NetworkIDPoint models a point's address and port.
type NetworkIDPoint struct {
	Addr string `json:"addr"`
	Port uint16 `json:"port"`
}

// NetworkVersion models a network-layer version of a node.
type NetworkVersion struct {
	Name  string `json:"name"`
	Major uint16 `json:"major"`
	Minor uint16 `json:"minor"`
}

// NetworkMetadata models metadata of a node.
type NetworkMetadata struct {
	DisableMempool bool `json:"disable_mempool"`
	PrivateNode    bool `json:"private_node"`
}
