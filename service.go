package tezos

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Service implements fetching of information from Tezos nodes via JSON.
type Service struct {
	Client *RPCClient
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
	IDPoint          NetworkAddress   `json:"id_point"`
	RemoteSocketPort uint16           `json:"remote_socket_port"`
	Versions         []NetworkVersion `json:"versions"`
	Private          bool             `json:"private"`
	LocalMetadata    NetworkMetadata  `json:"local_metadata"`
	RemoteMetadata   NetworkMetadata  `json:"remote_metadata"`
}

// NetworkAddress models a point's address and port.
type NetworkAddress struct {
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

// BootstrappedBlock represents bootstrapped block stream message
type BootstrappedBlock struct {
	Block     string    `json:"block"`
	Timestamp time.Time `json:"timestamp"`
}

// NetworkConnectionTime represents peer address with timestamp added
type NetworkConnectionTime struct {
	NetworkAddress
	Time time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (n *NetworkConnectionTime) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &n.NetworkAddress, &n.Time)
}

// NetworkPeer represents peer info
type NetworkPeer struct {
	PeerID                    string                 `json:"-"`
	Score                     int64                  `json:"score"`
	Trusted                   bool                   `json:"trusted"`
	ConnMetadata              *NetworkMetadata       `json:"conn_metadata"`
	State                     string                 `json:"state"`
	ReachableAt               *NetworkAddress        `json:"reachable_at"`
	Stat                      NetworkStats           `json:"stat"`
	LastEstablishedConnection *NetworkConnectionTime `json:"last_established_connection"`
	LastSeen                  *NetworkConnectionTime `json:"last_seen"`
	LastFailedConnection      *NetworkConnectionTime `json:"last_failed_connection"`
	LastRejectedConnection    *NetworkConnectionTime `json:"last_rejected_connection"`
	LastDisconnection         *NetworkConnectionTime `json:"last_disconnection"`
	LastMiss                  *NetworkConnectionTime `json:"last_miss"`
}

type networkPeerWithID NetworkPeer

func (n *networkPeerWithID) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &n.PeerID, (*NetworkPeer)(n))
}

// NetworkPeerLogEntry represents peer log entry
type NetworkPeerLogEntry struct {
	NetworkAddress
	Kind      string    `json:"kind"`
	Timestamp time.Time `json:"timestamp"`
}

// GetNetworkStats returns current network stats https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
func (s *Service) GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/stat", nil)
	if err != nil {
		return nil, err
	}

	var stats NetworkStats
	if err := s.Client.Get(req, &stats); err != nil {
		return nil, err
	}
	return &stats, err
}

// GetNetworkConnections returns all network connections http://tezos.gitlab.io/mainnet/api/rpc.html#get-network-connections
func (s *Service) GetNetworkConnections(ctx context.Context) ([]*NetworkConnection, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/connections", nil)
	if err != nil {
		return nil, err
	}

	var conns []*NetworkConnection
	if err := s.Client.Get(req, &conns); err != nil {
		return nil, err
	}
	return conns, err
}

// GetNetworkPeers returns all network peers https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers
func (s *Service) GetNetworkPeers(ctx context.Context, filter string) ([]*NetworkPeer, error) {
	u := url.URL{
		Path: "/network/peers",
	}

	if filter != "" {
		q := url.Values{
			"filter": []string{filter},
		}
		u.RawQuery = q.Encode()
	}

	req, err := s.Client.NewRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var peers []*networkPeerWithID
	if err := s.Client.Get(req, &peers); err != nil {
		return nil, err
	}

	ret := make([]*NetworkPeer, len(peers))
	for i, p := range peers {
		ret[i] = (*NetworkPeer)(p)
	}

	return ret, err
}

// GetNetworkPeer returns peer info https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id
func (s *Service) GetNetworkPeer(ctx context.Context, peerID string) (*NetworkPeer, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID, nil)
	if err != nil {
		return nil, err
	}

	var peer NetworkPeer
	if err := s.Client.Get(req, &peer); err != nil {
		return nil, err
	}
	peer.PeerID = peerID

	return &peer, err
}

// BanNetworkPeer bans the peer https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-ban
func (s *Service) BanNetworkPeer(ctx context.Context, peerID string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/ban", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Get(req, nil); err != nil {
		return err
	}
	return nil
}

// TrustNetworkPeer turns peer into trust mode https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-trust
func (s *Service) TrustNetworkPeer(ctx context.Context, peerID string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/trust", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Get(req, nil); err != nil {
		return err
	}
	return nil
}

// GetNetworkPeerBanned returns true if the peer is banned https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-banned
func (s *Service) GetNetworkPeerBanned(ctx context.Context, peerID string) (bool, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/banned", nil)
	if err != nil {
		return false, err
	}

	var banned bool
	if err := s.Client.Get(req, &banned); err != nil {
		return false, err
	}

	return banned, err
}

// GetNetworkPeerLog returns peer's log https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) GetNetworkPeerLog(ctx context.Context, peerID string) ([]*NetworkPeerLogEntry, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/log", nil)
	if err != nil {
		return nil, err
	}

	var log []*NetworkPeerLogEntry
	if err := s.Client.Get(req, &log); err != nil {
		return nil, err
	}

	return log, err
}

// MonitorNetworkPeerLog returns peer's log as a stream https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) MonitorNetworkPeerLog(ctx context.Context, peerID string, results chan<- []*NetworkPeerLogEntry) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/log?monitor", nil)
	if err != nil {
		return err
	}

	return s.Client.Get(req, results)
}

// GetDelegateBalance returns a delegate's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-delegates-pkh-balance
func (s *Service) GetDelegateBalance(ctx context.Context, chainID string, blockID string, pkh string) (string, error) {
	u := "/chains/" + chainID + "/blocks/" + blockID + "/context/delegates/" + pkh + "/balance"
	req, err := s.Client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	var balance string
	if err := s.Client.Get(req, &balance); err != nil {
		return "", err
	}

	return balance, err
}

// GetContractBalance returns a contract's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-contracts-contract-id-balance
func (s *Service) GetContractBalance(ctx context.Context, chainID string, blockID string, contractID string) (string, error) {
	u := "/chains/" + chainID + "/blocks/" + blockID + "/context/contracts/" + contractID + "/balance"
	req, err := s.Client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	var balance string
	if err := s.Client.Get(req, &balance); err != nil {
		return "", err
	}

	return balance, err
}

// GetBootstrapped reads from the bootstrapped blocks stream http://tezos.gitlab.io/mainnet/api/rpc.html#get-monitor-bootstrapped
func (s *Service) GetBootstrapped(ctx context.Context, results chan<- *BootstrappedBlock) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/monitor/bootstrapped", nil)
	if err != nil {
		return err
	}

	return s.Client.Get(req, results)
}
