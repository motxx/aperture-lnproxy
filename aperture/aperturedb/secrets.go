package aperturedb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"

	"github.com/lightningnetwork/lnd/clock"
	"github.com/motxx/aperture-lnproxy/aperture/aperturedb/sqlc"
	"github.com/motxx/aperture-lnproxy/aperture/lsat"
	"github.com/motxx/aperture-lnproxy/aperture/mint"
)

type (
	// NewSecret is a struct that contains the parameters required to insert
	// a new secret into the database.
	NewSecret                       = sqlc.InsertSecretParams
	SetSettledAtByPaymentHashParams = sqlc.SetSettledAtByPaymentHashParams
	NullTime                        = sql.NullTime
)

// SecretsDB is an interface that defines the set of operations that can be
// executed against the secrets database.
type SecretsDB interface {
	// InsertSecret inserts a new secret into the database.
	InsertSecret(ctx context.Context, arg NewSecret) (int32, error)

	// SetSettledTime sets the settled_at that corresponds to the given hash.
	SetSettledAtByPaymentHash(ctx context.Context, arg SetSettledAtByPaymentHashParams) error

	// GetSecretByIdHash returns the secret that corresponds to the given hash.
	GetSecretByIdHash(ctx context.Context, idHash []byte) ([]byte, error)

	// GetSettledAtByPaymentHash returns the settled_at that corresponds to the given
	// hash.
	GetSettledAtByPaymentHash(ctx context.Context, paymentHash []byte) (NullTime, error)

	// DeleteSecretByIdHash removes the secret that corresponds to the given
	// hash.
	DeleteSecretByIdHash(ctx context.Context, idHash []byte) (int64, error)
}

// SecretsTxOptions defines the set of db txn options the SecretsStore
// understands.
type SecretsDBTxOptions struct {
	// readOnly governs if a read only transaction is needed or not.
	readOnly bool
}

// ReadOnly returns true if the transaction should be read only.
//
// NOTE: This implements the TxOptions
func (a *SecretsDBTxOptions) ReadOnly() bool {
	return a.readOnly
}

// NewSecretsDBReadTx creates a new read transaction option set.
func NewSecretsDBReadTx() SecretsDBTxOptions {
	return SecretsDBTxOptions{
		readOnly: true,
	}
}

// BatchedSecretsDB is a version of the SecretsDB that's capable of batched
// database operations.
type BatchedSecretsDB interface {
	SecretsDB

	BatchedTx[SecretsDB]
}

// SecretsStore represents a storage backend.
type SecretsStore struct {
	db    BatchedSecretsDB
	clock clock.Clock
}

// NewSecretsStore creates a new SecretsStore instance given a open
// BatchedSecretsDB storage backend.
func NewSecretsStore(db BatchedSecretsDB) *SecretsStore {
	return &SecretsStore{
		db:    db,
		clock: clock.NewDefaultClock(),
	}
}

// NewSecret creates a new cryptographically random secret which is
// keyed by the given hash.
func (s *SecretsStore) NewSecret(ctx context.Context,
	idHash [sha256.Size]byte, paymentHash [sha256.Size]byte) ([lsat.SecretSize]byte, error) {

	var secret [lsat.SecretSize]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return [lsat.SecretSize]byte{}, err
	}

	var writeTxOpts SecretsDBTxOptions
	err := s.db.ExecTx(ctx, &writeTxOpts, func(tx SecretsDB) error {
		_, err := tx.InsertSecret(ctx, NewSecret{
			MacaroonIDHash: idHash[:],
			PaymentHash:    paymentHash[:],
			Secret:         secret[:],
			CreatedAt:      s.clock.Now().UTC(),
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return [lsat.SecretSize]byte{}, fmt.Errorf("unable to insert "+
			"new secret for idHash(%x): %w", idHash, err)
	}

	return secret, nil
}

// SetSettledAtByPaymentHash sets the settled_at time for the secret that
// corresponds to the given hash.
func (s *SecretsStore) SetSettledAtByPaymentHash(ctx context.Context,
	paymentHash [sha256.Size]byte, settledAt NullTime) error {

	var writeTxOpts SecretsDBTxOptions
	err := s.db.ExecTx(ctx, &writeTxOpts, func(tx SecretsDB) error {
		err := tx.SetSettledAtByPaymentHash(ctx, SetSettledAtByPaymentHashParams{
			PaymentHash: paymentHash[:],
			SettledAt:   settledAt,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to set settled at for paymentHash(%x): %w",
			paymentHash, err)
	}

	return nil
}

// GetSecret returns the cryptographically random secret that
// corresponds to the given hash. If there is no secret, then
// ErrSecretNotFound is returned.
func (s *SecretsStore) GetSecret(ctx context.Context,
	idHash [sha256.Size]byte) ([lsat.SecretSize]byte, error) {

	var secret [lsat.SecretSize]byte
	readOpts := NewSecretsDBReadTx()
	err := s.db.ExecTx(ctx, &readOpts, func(db SecretsDB) error {
		secretRow, err := db.GetSecretByIdHash(ctx, idHash[:])
		switch {
		case err == sql.ErrNoRows:
			return mint.ErrSecretNotFound

		case err != nil:
			return err
		}

		copy(secret[:], secretRow)

		return nil
	})

	if err != nil {
		return [lsat.SecretSize]byte{}, fmt.Errorf("unable to get "+
			"secret for hash(%x): %w", idHash, err)
	}

	return secret, nil
}

// GetSettledAtByPaymentHash returns the settled_at time for the secret that
// corresponds to the given hash.
func (s *SecretsStore) GetSettledAtByPaymentHash(ctx context.Context,
	paymentHash [sha256.Size]byte) (NullTime, error) {

	var settledAt NullTime
	readOpts := NewSecretsDBReadTx()
	err := s.db.ExecTx(ctx, &readOpts, func(db SecretsDB) error {
		settledAtRow, err := db.GetSettledAtByPaymentHash(ctx, paymentHash[:])
		switch {
		case err == sql.ErrNoRows:
			return mint.ErrSecretNotFound

		case err != nil:
			return err
		}

		settledAt = settledAtRow

		return nil
	})

	if err != nil {
		return NullTime{}, fmt.Errorf("unable to get settled_at "+
			"for paymentHash(%x): %w", paymentHash, err)
	}

	return settledAt, nil
}

// RevokeSecret removes the cryptographically random secret that
// corresponds to the given hash. This acts as a NOP if the secret does
// not exist.
func (s *SecretsStore) RevokeSecret(ctx context.Context,
	idHash [sha256.Size]byte) error {

	var writeTxOpts SecretsDBTxOptions
	err := s.db.ExecTx(ctx, &writeTxOpts, func(tx SecretsDB) error {
		nRows, err := tx.DeleteSecretByIdHash(ctx, idHash[:])
		if err != nil {
			return err
		}

		if nRows != 1 {
			log.Info("deleting secret(%x) did not affect %w rows",
				idHash, nRows)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("unable to revoke secret for hash(%x): %w",
			idHash, err)
	}

	return nil
}
