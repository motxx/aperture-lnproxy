// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"context"
	"database/sql"
)

type Querier interface {
	DeleteOnionPrivateKey(ctx context.Context) error
	DeleteSecretByIdHash(ctx context.Context, macaroonIDHash []byte) (int64, error)
	GetSecretByIdHash(ctx context.Context, macaroonIDHash []byte) ([]byte, error)
	GetSession(ctx context.Context, passphraseEntropy []byte) (LncSession, error)
	GetSettledAtByPaymentHash(ctx context.Context, paymentHash []byte) (sql.NullTime, error)
	InsertSecret(ctx context.Context, arg InsertSecretParams) (int32, error)
	InsertSession(ctx context.Context, arg InsertSessionParams) error
	SelectOnionPrivateKey(ctx context.Context) ([]byte, error)
	SetExpiry(ctx context.Context, arg SetExpiryParams) error
	SetRemotePubKey(ctx context.Context, arg SetRemotePubKeyParams) error
	SetSettledAtByPaymentHash(ctx context.Context, arg SetSettledAtByPaymentHashParams) error
	UpsertOnion(ctx context.Context, arg UpsertOnionParams) error
}

var _ Querier = (*Queries)(nil)
