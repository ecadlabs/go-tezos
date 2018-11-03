package tezos

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceGetMethods(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		get             func(s *Service) (interface{}, error)
		respFixture     string
		respStatus      int
		respContentType string
		expectedPath    string
		expectedValue   interface{}
		errMsg          string
		errType         interface{}
	}{
		{
			get:             func(s *Service) (interface{}, error) { return s.GetStats(ctx) },
			respFixture:     "fixtures/network/stat.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			expectedValue: &NetworkStats{
				TotalBytesSent: 291690080,
				TotalBytesRecv: 532639553,
				CurrentInflow:  23596,
				CurrentOutflow: 14972,
			},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetConnections(ctx) },
			respFixture:     "fixtures/network/connections.json",
			respContentType: "application/json",
			expectedPath:    "/network/connections",
			expectedValue:   []NetworkConnection{NetworkConnection{Incoming: false, PeerID: "idt5qvkLiJ15rb6yJU1bjpGmdyYnPJ", IDPoint: NetworkIDPoint{Addr: "::ffff:34.253.64.43", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: false, LocalMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}, RemoteMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}}, NetworkConnection{Incoming: true, PeerID: "ids8VJTHEuyND6B8ahGgXPAJ3BDp1c", IDPoint: NetworkIDPoint{Addr: "::ffff:176.31.255.202", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: true, LocalMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}, RemoteMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetDelegateBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:     "fixtures/contract/delegate_balance.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/context/delegates/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue:   "13490453135591",
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetContractBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:     "fixtures/contract/contract_balance.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/context/contracts/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue:   "4700354460878",
		},
		// Handling 5xx errors from the Tezos node with RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: RPC error (kind = "permanent", id = "proto.002-PsYLVpVv.context.storage_error")`,
			errType:         (*rpcErrors)(nil),
		},
		// Handling 5xx errors from the Tezos node with empty RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/empty_error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: empty error response`,
			errType:         (*plainError)(nil),
		},
		// Handling 5xx errors from the Tezos node with malformed RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/malformed_error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: error decoding RPC error: invalid character ',' looking for beginning of value`,
			errType:         (*plainError)(nil),
		},
		// Handling unexpected response status codes.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:   404,
			respFixture:  "fixtures/empty.json",
			expectedPath: "/network/stat",
			errMsg:       `tezos: HTTP status 404`,
			errType:      (*httpError)(nil),
		},
	}

	for _, test := range tests {
		// Start a test HTTP server that responds as specified in the test case parameters.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, test.expectedPath, r.URL.Path)
			require.Equal(t, http.MethodGet, r.Method)

			buf, err := ioutil.ReadFile(test.respFixture)
			require.NoError(t, err, "error reading fixture %q", test.respFixture)

			if test.respContentType != "" {
				w.Header().Set("Content-Type", test.respContentType)
			}

			if test.respStatus != 0 {
				w.WriteHeader(test.respStatus)
			}
			_, err = w.Write(buf)
			require.NoError(t, err, "error writing HTTP response")
		}))

		c, err := NewRPCClient(nil, srv.URL)
		require.NoError(t, err, "error creating client")

		s := &Service{Client: c}

		value, err := test.get(s)

		if test.errType != nil {
			require.IsType(t, test.errType, err)
		}

		if test.errMsg == "" {
			require.NoError(t, err, "error getting value")
			require.Equal(t, test.expectedValue, value, "unexpected value")
		} else {
			require.EqualError(t, err, test.errMsg, "unexpected error string")
		}

		srv.Close()
	}
}
