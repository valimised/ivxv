package storage

import (
	"bytes"
	"context"
	"time"

	"ivxv.ee/common/collector/errors"
	"ivxv.ee/common/collector/log"
	"ivxv.ee/common/collector/q11n"
)

// Transaction makes it possible to call "Txn" prefixed storage services
// within a single transaction. This is the new approach, since traditionally
// IVXV supported only 1 operation per a single transaction.
type Transaction interface {
	// Begin a transaction, all subsequent calls to the "Txn"
	// prefixed storage services, will be added to a single transaction.
	Begin(context.Context) (TxnOp, error)

	// Commit guarantees that transaction is committed, and therefore all
	// operations that have been added to the transaction in a "Txn" prefixed
	// storage services will be applied with "All or nothing" principle.
	Commit(context.Context, TxnOp) error

	// AutoCommit allows a caller to assume that once transaction has begun,
	// it will be auto committed, without manual Commit operation.
	//
	// Please note, that AutoCommit doesn't guarantee that transaction is
	// committed as well as context is deadlined. This guarantee completely
	// relies on the "Txn" prefixed storage methods.
	//
	// Roughly speaking, the main difference between Commit and AutoCommit
	// is that former allows a caller to commit a transaction and the
	// latter delegates committing of a transaction to the "Txn" prefixed
	// storage services.
	AutoCommit(context.Context, TxnOp)
}

// TxnOp represents an operation that is done in a single Transaction.
type TxnOp interface {
	// PutAll calls Put for each req in reqs.
	PutAll(reqs ...*PutAllRequest)

	// Put creates a single INSERT request a for key-value pair and
	// adds it to the Transaction.
	Put(key string, value []byte)

	// PutForce is almost same as Put, the only difference between
	// them is that former should not do any conditional checks against
	// the underlying storage before making an INSERT request.
	PutForce(key string, value []byte)

	// CAS performs "compare-and-swap", replacing old with new
	// for a given key.
	CAS(key string, old, new []byte) //nolint:revive

	// Ready reports that TxnOp is ready to be committed.
	Ready(ctx context.Context) error
}

// Txn returns a Transaction to the caller.
func (c *Client) Txn() Transaction {
	return c.prot.(Transaction)
}

// TxnStoreQualifyingProperty stores qualifying properties (ocsp/tspreg)
// for a vote in the underlying storage.
func (c *Client) TxnStoreQualifyingProperty(ctx context.Context,
	voteID []byte, protocol q11n.Protocol, property []byte, op TxnOp) {
	prefix := voteIDPrefix(voteID)
	op.Put(prefix+string(protocol), property)

	log.Debug(ctx, TxnStoreQualifyingPropertyAddPropertyToTxn{Property: protocol,
		Description: _STORAGE_TXN_PUT})
}

// TxnSetVoted notifies storage that voteID was successful, meaning all configured
// qualifying properties for it are received and stored. ctime is the canonical
// time of the vote, which is usually the time of a qualifying property, e.g.,
// vote registration timestamp.
func (c *Client) TxnSetVoted(ctx context.Context,
	txnOp TxnOp, voteID []byte, voterName string, ctime time.Time,
	testVote bool) error {
	// voteID was successful: refresh indexes related to the voter. Keep in
	// mind that voteID is not guaranteed to be the latest vote from the
	// voter, so be careful when updating the information.

	// Critical error, if occurs then vote cannot be stored
	var unexpected UnexpectedValueError

	// Get vote metadata in a single batch call.
	idVoterKey := voteIDPrefix(voteID) + voterKey     // Voter assigned to voteID.
	idVersionKey := voteIDPrefix(voteID) + versionKey // Effective voter list version.
	data, err := c.getAllStrict(ctx, idVoterKey, idVersionKey)
	if err != nil {
		unexpected.Err = TxnSetVotedGetAllStrictError{
			VoteID:      voteID,
			Err:         err,
			Description: _STORAGE_GET,
		}
		return unexpected
	}

	// Get the administrative unit code of voter in voter list version.
	idVoter := string(data[idVoterKey])
	idAdminCode, district, err := c.GetVoter(ctx, string(data[idVersionKey]), idVoter)
	switch {
	case err == nil:
	case errors.CausedBy(err, new(NotExistError)) != nil:
		// The voter has voted, yet is not in the voter list. Assume
		// that the voting service is ignoring voter lists (see
		// ivxv.ee/common/collector/conf.Election.IgnoreVoterList). Use an empty
		// administrative unit code.
		idAdminCode = ""
		district = ""
	default:
		unexpected.Err = TxnSetVotedGetVoterError{
			Voter:       idVoter,
			Err:         err,
			Description: _STORAGE_GET,
		}
		return unexpected
	}

	ctimeStr := ctime.Format(timefmt)
	if !testVote { // Only count non-test votes in statistics.
		// Update the voted stats index: if the administrative unit codes
		// match, then use the older timestamp; if they differ, then use the
		// newer timestamp. See GetVotedStats for the reasoning behind this.
		votedStatsValue := encodePair(idAdminCode, ctimeStr)
		key := votedStatsPrefix + idVoter
		existing, err := c.prot.Get(ctx, key)

		// err == NotExistError, i.e key doesn't exist in db
		if errors.CausedBy(err, new(NotExistError)) != nil {
			// Add votedStatsValue to transaction
			txnOp.Put(key, votedStatsValue)

			log.Debug(ctx, TxnSetVotedAddedNewVotedStatsToTxn{
				VoterID:     idVoter,
				AdminCode:   idAdminCode,
				VotedAt:     ctimeStr,
				Description: _STORAGE_TXN_PUT,
			})
		} else {
			oldAdminCode, oldTimeStr, err := decodePair(existing)
			if err != nil {
				unexpected.Err = TxnSetVotedDecodePairError{
					Key:         key,
					Value:       string(existing),
					Voter:       idVoter,
					Err:         err,
					Description: _STORAGE_DECODE_PAIR,
				}
				return unexpected
			}

			oldTime, err := time.Parse(timefmt, oldTimeStr)
			if err != nil {
				unexpected.Err = TxnSetVotedParseTimeError{
					Voter:       idVoter,
					Err:         err,
					Description: _STORAGE_TS_PARSE,
				}
				return unexpected
			}

			switch {
			case bytes.Equal(existing, votedStatsValue):
				txnOp.PutForce(key, existing) // Keep already existing vote.
			case idAdminCode == oldAdminCode && ctime.After(oldTime):
				txnOp.PutForce(key, existing) // Keep older vote in same admin.
			case idAdminCode != oldAdminCode && ctime.Before(oldTime):
				txnOp.PutForce(key, existing) // Keep newer vote in different admin.
			default:
				txnOp.PutForce(key, votedStatsValue) // Update the index.

				log.Debug(ctx, TxnSetVotedOverrideOldVotedStatsAndAddToTxn{
					VoterID:     idVoter,
					AdminCode:   idAdminCode,
					VotedAt:     ctimeStr,
					Description: _STORAGE_TXN_PUT,
				})
			}
		}

		log.Debug(ctx, TxnSetVotedStartTxnCommit{Description: _STORAGE_TXN_FINISHING})

		// Send ready signal to AutoCommit and wait until committed
		err = txnOp.Ready(ctx)
		if err != nil {
			// Transaction had been committed but other errors has arisen.
			// Additional info in error is needed for restoring AddVoteOrder data.
			if errors.CausedBy(err, new(UnexpectedValueError)) != nil {
				return TxnSetVotedAutoCommitUnsuccessfullTxnError{
					VoterName:   voterName,
					VoterID:     idVoter,
					AdminCode:   idAdminCode,
					District:    district,
					Key:         key,
					Err:         err, // UnexpectedValueError is nested here
					Description: _STORAGE_TXN_READY,
				}
			}
			unexpected.Err = TxnSetVotedAutoCommitError{
				Key:         key,
				Err:         err,
				Description: _STORAGE_TXN_READY,
			}
			return unexpected
		}

		log.Debug(ctx, TxnSetVotedTxnSuccessfullyCommited{Description: _STORAGE_TXN_COMMITTED})

		// voterName is empty when rebuilding voted stats
		if voterName != "" {
			err = c.AddVoteOrder(ctx, voterName, idVoter, district, idAdminCode)
			if err != nil {
				return err
			}
		}
	} else {
		log.Debug(ctx, TxnSetVotedStartTxnCommitTestVote{Description: _STORAGE_TEST_VOTE})

		// Send ready signal to AutoCommit and wait until committed
		err = txnOp.Ready(ctx)
		if err != nil {
			unexpected.Err = TxnSetVotedTestVoteAutoCommitError{Err: err, Description: _STORAGE_TXN_READY}
			return unexpected
		}

		log.Debug(ctx, TxnSetVotedTxnSuccessfullyCommitedTestVote{Description: _STORAGE_TXN_COMMITTED})
	}
	// Update the voted latest index: store the vote identifier with the
	// newer timestamp, unless they match in which case corrupt the
	// identifier, since we have no way of ordering the two votes.
	votedLatestValue := encodePair(ctimeStr, string(voteID))
	if err := c.update(ctx, votedLatestPrefix+idVoter, func(existing []byte) ([]byte, error) {
		if existing == nil {
			return votedLatestValue, nil // Initial value.
		}
		if bytes.Equal(votedLatestValue, existing) {
			return nil, nil // Fast path if already same value.
		}

		oldTimeStr, oldVoteID, err := decodePair(existing)
		if err != nil {
			return nil, log.Alert(TxnSetVotedVotedLatestDecodePairError{
				Voter:       idVoter,
				Err:         err,
				Description: _STORAGE_DECODE_PAIR,
			})
		}
		oldTime, err := time.Parse(timefmt, oldTimeStr)
		if err != nil {
			return nil, log.Alert(TxnSetVotedVotedLatestParseTimeError{
				Voter:       idVoter,
				Err:         err,
				Description: _STORAGE_TS_PARSE,
			})
		}

		switch {
		case oldTime.After(ctime):
			return nil, nil // Keep newer vote.
		case oldTime.Equal(ctime):
			// Corrupt the index until a newer vote comes in. Keep
			// both vote identifiers for manual debugging.
			voteID = encodePair(string(voteID), oldVoteID)
			return encodePair(ctimeStr, string(voteID)), nil
		default:
			return votedLatestValue, nil // Update the index.
		}
	}); err != nil {
		// Only log error because "err" may have nested UnexpectedValueError
		// which may result in undesired error message to the client
		log.Log(ctx, TxnSetVotedUpdateVotedLatestPrefixError{Err: err, Description: _STORAGE_UPDATE})
		return TxnSetVotedUpdateVotedLatestError{Description: _STORAGE_UPDATE}
	}

	return nil
}
