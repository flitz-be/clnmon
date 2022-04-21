package lndmon

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/wtxmgr"
	clngrpc "github.com/flitz-be/cln-grpc-go/cln"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwallet"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/routing/route"
)

type CLN struct {
	cln clngrpc.NodeClient
}

func (cln *CLN) PayInvoice(ctx context.Context, invoice string, maxFee btcutil.Amount, outgoingChannel *uint64) chan lndclient.PaymentResult {
	return nil
}

func (cln *CLN) GetInfo(ctx context.Context) (*lndclient.Info, error) {
	info, err := cln.cln.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if err != nil {
		return nil, err
	}
	return &lndclient.Info{
		Version:             info.Version,
		BlockHeight:         info.Blockheight,
		IdentityPubkey:      convertPubkey(info.Id),
		Alias:               info.Alias,
		Network:             info.Network,
		Uris:                []string{},
		SyncedToChain:       true,        //todo: ?
		SyncedToGraph:       true,        //todo: ?
		BestHeaderTimeStamp: time.Time{}, //todo ?
		ActiveChannels:      info.NumActiveChannels,
		InactiveChannels:    info.NumInactiveChannels,
		PendingChannels:     info.NumPendingChannels,
	}, nil
}

func (cln *CLN) EstimateFeeToP2WSH(ctx context.Context, amt btcutil.Amount, confTarget int32) (btcutil.Amount, error) {
	return 0, fmt.Errorf("Not implemented")
}

// WalletBalance returns a summary of the node's wallet balance.
func (cln *CLN) WalletBalance(ctx context.Context) (*lndclient.WalletBalance, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (cln *CLN) AddInvoice(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error) {
	return lntypes.Hash{}, "", fmt.Errorf("Not implemented")
}

// LookupInvoice looks up an invoice by hash.
func (cln *CLN) LookupInvoice(ctx context.Context, hash lntypes.Hash) (*lndclient.Invoice, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ListTransactions returns all known transactions of the backing lnd
// node. It takes a start and end block height which can be used to
// limit the block range that we query over. These values can be left
// as zero to include all blocks. To include unconfirmed transactions
// in the query, endHeight must be set to -1.
func (cln *CLN) ListTransactions(ctx context.Context, startHeight int32, endHeight int32) ([]lndclient.Transaction, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ListChannels retrieves all channels of the backing lnd node.
func (cln *CLN) ListChannels(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error) {
	peers, err := cln.cln.ListPeers(ctx, &clngrpc.ListpeersRequest{})
	if err != nil {
		return nil, err
	}
	channels := []lndclient.ChannelInfo{}
	for _, p := range peers.Peers {
		for _, ch := range p.Channels {
			cap := btcutil.Amount(ch.TotalMsat.Msat / 1e3)
			ours := btcutil.Amount(ch.ToUsMsat.Msat / 1e3)
			receivable := btcutil.Amount(ch.ReceivableMsat.Msat / 1e3)
			chanId, err := convertChanId(ch.ShortChannelId)
			if err != nil {
				return nil, err
			}
			htlcs, err := convertHtlcs(ch.Htlcs, lnwire.NewShortChanIDFromInt(chanId))
			if err != nil {
				return nil, err
			}
			channels = append(channels, lndclient.ChannelInfo{
				ChannelPoint:      fmt.Sprintf("%s:%d", string(ch.FundingTxid), *ch.FundingOutnum),
				Active:            p.Connected,
				ChannelID:         chanId,
				PubKeyBytes:       convertPubkey(p.Id),
				Capacity:          cap,
				LocalBalance:      ours,
				RemoteBalance:     receivable,
				UnsettledBalance:  0,
				Initiator:         ch.Opener == clngrpc.ChannelSide_IN,
				Private:           *ch.Private,
				LifeTime:          0,
				Uptime:            0,
				TotalSent:         btcutil.Amount(ch.OutFulfilledMsat.Msat / 1e3), // Not accurate
				TotalReceived:     btcutil.Amount(ch.InFulfilledMsat.Msat / 1e3),  // Not accurate
				NumUpdates:        0,
				NumPendingHtlcs:   len(ch.Htlcs),
				PendingHtlcs:      htlcs,
				CSVDelay:          0,
				FeePerKw:          0,
				CommitWeight:      0,
				CommitFee:         0,
				LocalConstraints:  &lndclient.ChannelConstraints{},
				RemoteConstraints: &lndclient.ChannelConstraints{},
				CloseAddr:         nil,
			})
		}
	}
	return channels, nil
}

// PendingChannels returns a list of lnd's pending channels.
func (cln *CLN) PendingChannels(ctx context.Context) (*lndclient.PendingChannels, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ClosedChannels returns all closed channels of the backing lnd node.
func (cln *CLN) ClosedChannels(ctx context.Context) ([]lndclient.ClosedChannel, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ForwardingHistory makes a paginated call to our forwarding history
// endpoint.
func (cln *CLN) ForwardingHistory(ctx context.Context, req lndclient.ForwardingHistoryRequest) (*lndclient.ForwardingHistoryResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ListInvoices makes a paginated call to our list invoices endpoint.
func (cln *CLN) ListInvoices(ctx context.Context, req lndclient.ListInvoicesRequest) (*lndclient.ListInvoicesResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ListPayments makes a paginated call to our list payments endpoint.
func (cln *CLN) ListPayments(ctx context.Context, req lndclient.ListPaymentsRequest) (*lndclient.ListPaymentsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ChannelBackup retrieves the backup for a particular channel. The
// backup is returned as an encrypted chanbackup.Single payload.
func (cln *CLN) ChannelBackup(_ context.Context, _ wire.OutPoint) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ChannelBackups retrieves backups for all existing pending open and
// open channels. The backups are returned as an encrypted
// chanbackup.Multi payload.
func (cln *CLN) ChannelBackups(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

// SubscribeChannelBackups allows a client to subscribe to the
// most up to date information concerning the state of all channel
// backups.
func (cln *CLN) SubscribeChannelBackups(ctx context.Context) (<-chan lnrpc.ChanBackupSnapshot, <-chan error, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// SubscribeChannelEvents allows a client to subscribe to updates
// relevant to the state of channels. Events include new active
// channels, inactive channels, and closed channels.
func (cln *CLN) SubscribeChannelEvents(ctx context.Context) (<-chan *lndclient.ChannelEventUpdate, <-chan error, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// DecodePaymentRequest decodes a payment request.
func (cln *CLN) DecodePaymentRequest(ctx context.Context, payReq string) (*lndclient.PaymentRequest, error) {
	return nil, fmt.Errorf("Not implemented")
}

// OpenChannel opens a channel to the peer provided with the amounts
// specified.
func (cln *CLN) OpenChannel(ctx context.Context, peer route.Vertex, localSat btcutil.Amount, pushSat btcutil.Amount, private bool) (*wire.OutPoint, error) {
	return nil, fmt.Errorf("Not implemented")
}

// CloseChannel closes the channel provided.
func (cln *CLN) CloseChannel(ctx context.Context, channel *wire.OutPoint, force bool, confTarget int32, deliveryAddr btcutil.Address) (chan lndclient.CloseChannelUpdate, chan error, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// UpdateChanPolicy updates the channel policy for the passed chanPoint.
// If the chanPoint is nil, then the policy is applied for all existing
// channels.
func (cln *CLN) UpdateChanPolicy(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error {
	return fmt.Errorf("Not implemented")
}

// GetChanInfo returns the channel info for the passed channel,
// including the routing policy for both end.
func (cln *CLN) GetChanInfo(ctx context.Context, chanID uint64) (*lndclient.ChannelEdge, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ListPeers gets a list the peers we are currently connected to.
func (cln *CLN) ListPeers(ctx context.Context) ([]lndclient.Peer, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Connect attempts to connect to a peer at the host specified. If
// permanent is true then we'll attempt to connect to the peer
// permanently, meaning that the connection is maintained even if no
// channels exist between us and the peer.
func (cln *CLN) Connect(ctx context.Context, peer route.Vertex, host string, permanent bool) error {
	return fmt.Errorf("Not implemented")
}

// SendCoins sends the passed amount of (or all) coins to the passed
// address. Either amount or sendAll must be specified, while
// confTarget, satsPerByte are optional and may be set to zero in which
// case automatic conf target and fee will be used. Returns the tx id
// upon success.
func (cln *CLN) SendCoins(ctx context.Context, addr btcutil.Address, amount btcutil.Amount, sendAll bool, confTarget int32, satsPerByte int64, label string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

// ChannelBalance returns a summary of our channel balances.
func (cln *CLN) ChannelBalance(ctx context.Context) (*lndclient.ChannelBalance, error) {
	return nil, fmt.Errorf("Not implemented")
}

// GetNodeInfo looks up information for a specific node.
func (cln *CLN) GetNodeInfo(ctx context.Context, pubkey route.Vertex, includeChannels bool) (*lndclient.NodeInfo, error) {
	return nil, fmt.Errorf("Not implemented")
}

// DescribeGraph returns our view of the graph.
func (cln *CLN) DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error) {
	return nil, fmt.Errorf("Not implemented")
}

// SubscribeGraph allows a client to subscribe to gaph topology updates.
func (cln *CLN) SubscribeGraph(ctx context.Context) (<-chan *lndclient.GraphTopologyUpdate, <-chan error, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// NetworkInfo returns stats regarding our view of the network.
func (cln *CLN) NetworkInfo(ctx context.Context) (*lndclient.NetworkInfo, error) {
	return nil, fmt.Errorf("Not implemented")
}

// SubscribeInvoices allows a client to subscribe to updates
// of newly added/settled invoices.
func (cln *CLN) SubscribeInvoices(ctx context.Context, req lndclient.InvoiceSubscriptionRequest) (<-chan *lndclient.Invoice, <-chan error, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// ListPermissions returns a list of all RPC method URIs and the
// macaroon permissions that are required to access them.
func (cln *CLN) ListPermissions(ctx context.Context) (map[string][]lndclient.MacaroonPermission, error) {
	return nil, fmt.Errorf("Not implemented")
}

// ChannelAcceptor create a channel acceptor using the accept function
// passed in. The timeout provided will be used to timeout the passed
// accept closure when it exceeds the amount of time we allow. Note that
// this amount should be strictly less than lnd's chanacceptor timeout
// parameter.
func (cln *CLN) ChannelAcceptor(ctx context.Context, timeout time.Duration, accept lndclient.AcceptorFunction) (chan error, error) {
	return nil, fmt.Errorf("Not implemented")
}

// QueryRoutes can query LND to return a route (with fees) between two
// vertices.
func (cln *CLN) QueryRoutes(ctx context.Context, req lndclient.QueryRoutesRequest) (*lndclient.QueryRoutesResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

// CheckMacaroonPermissions allows a client to check the validity of a
// macaroon.
func (cln *CLN) CheckMacaroonPermissions(ctx context.Context, macaroon []byte, permissions []lndclient.MacaroonPermission, fullMethod string) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

// ListUnspent returns a list of all utxos spendable by the wallet with
// a number of confirmations between the specified minimum and maximum.
func (cln *CLN) ListUnspent(ctx context.Context, minConfs int32, maxConfs int32) ([]*lnwallet.Utxo, error) {
	return nil, fmt.Errorf("Not implemented")
}

// LeaseOutput locks an output to the given ID for the lease time
// provided, preventing it from being available for any future coin
// selection attempts. The absolute time of the lock's expiration is
// returned. The expiration of the lock can be extended by successive
// invocations of this call. Outputs can be unlocked before their
// expiration through `ReleaseOutput`.
func (cln *CLN) LeaseOutput(ctx context.Context, lockID wtxmgr.LockID, op wire.OutPoint, leaseTime time.Duration) (time.Time, error) {
	return time.Time{}, fmt.Errorf("Not implemented")
}

// ReleaseOutput unlocks an output, allowing it to be available for coin
// selection if it remains unspent. The ID should match the one used to
// originally lock the output.
func (cln *CLN) ReleaseOutput(ctx context.Context, lockID wtxmgr.LockID, op wire.OutPoint) error {
	return fmt.Errorf("Not implemented")
}

func (cln *CLN) DeriveNextKey(ctx context.Context, family int32) (*keychain.KeyDescriptor, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (cln *CLN) DeriveKey(ctx context.Context, locator *keychain.KeyLocator) (*keychain.KeyDescriptor, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (cln *CLN) NextAddr(ctx context.Context) (btcutil.Address, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (cln *CLN) PublishTransaction(ctx context.Context, tx *wire.MsgTx, label string) error {
	return fmt.Errorf("Not implemented")
}

func (cln *CLN) SendOutputs(ctx context.Context, outputs []*wire.TxOut, feeRate chainfee.SatPerKWeight, label string) (*wire.MsgTx, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (cln *CLN) EstimateFee(ctx context.Context, confTarget int32) (chainfee.SatPerKWeight, error) {
	return 0, fmt.Errorf("Not implemented")
}

// ListSweeps returns a list of sweep transaction ids known to our node.
// Note that this function only looks up transaction ids, and does not
// query our wallet for the full set of transactions.
func (cln *CLN) ListSweeps(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("Not implemented")
}

// BumpFee attempts to bump the fee of a transaction by spending one of
// its outputs at the given fee rate. This essentially results in a
// child-pays-for-parent (CPFP) scenario. If the given output has been
// used in a previous BumpFee call, then a transaction replacing the
// previous is broadcast, resulting in a replace-by-fee (RBF) scenario.
func (cln *CLN) BumpFee(_ context.Context, _ wire.OutPoint, _ chainfee.SatPerKWeight) error {
	return fmt.Errorf("Not implemented")
}

// ListAccounts retrieves all accounts belonging to the wallet by default.
// Optional name and addressType can be provided to filter through all of the
// wallet accounts and return only those matching.
func (cln *CLN) ListAccounts(ctx context.Context, name string, addressType walletrpc.AddressType) ([]*walletrpc.Account, error) {
	return nil, fmt.Errorf("Not implemented")
}

//todo check if this function is actually correct??
//probably not
func convertChanId(clnChanId *string) (lndChanId uint64, err error) {
	if clnChanId == nil {
		return 0, fmt.Errorf("Chan id is nil")
	}
	parts := strings.Split(*clnChanId, "x")
	if len(parts) != 3 {
		return 0, fmt.Errorf("Short chan id should have 3 parts but got %d parts", len(parts))
	}
	array := []byte{}
	blockHeight, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	binary.LittleEndian.PutUint64(array, uint64(blockHeight))

	blockIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	binary.LittleEndian.PutUint64(array, uint64(blockIndex))

	outputIndex, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	binary.LittleEndian.PutUint64(array, uint64(outputIndex))

	num := binary.LittleEndian.Uint64(array)
	return num, nil
}

func convertHash(clnKey []byte) (lndKey [32]byte) {
	pubkey := [32]byte{}
	copy(pubkey[:], lndKey[:])
	return pubkey
}

func convertPubkey(clnKey []byte) (lndKey [33]byte) {
	pubkey := [33]byte{}
	copy(pubkey[:], lndKey[:])
	return pubkey
}

func convertHtlcs(in []*clngrpc.ListpeersPeersChannelsHtlcs, chanId lnwire.ShortChannelID) (htlcs []lndclient.PendingHtlc, err error) {
	htlcs = make([]lndclient.PendingHtlc, len(in))
	for i, htlc := range in {
		htlcs[i] = lndclient.PendingHtlc{
			Incoming:          htlc.Direction == clngrpc.ListpeersPeersChannelsHtlcs_IN,
			Amount:            btcutil.Amount(htlc.AmountMsat.Msat) / 1e3,
			Hash:              convertHash(htlc.PaymentHash),
			Expiry:            htlc.Expiry,
			HtlcIndex:         htlc.Id,
			ForwardingChannel: chanId,
			ForwardingIndex:   htlc.Id, //?
		}
	}
	return nil, nil
}
