package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/ctxutil"
)

type queryKey struct{}

type Txn interface {
	pgx.Tx
	Querier() Querier
	RollbackOnErr(err *error)
}

var _ Txn = (*txnImpl)(nil)

type txnImpl struct {
	pgx.Tx
	ctx context.Context
}

func (t *txnImpl) Querier() Querier {
	return New(t)
}

func (t *txnImpl) RollbackOnErr(err *error) {
	if *err != nil {
		_ = t.Rollback(t.ctx)
		return
	}

	*err = t.Commit(t.ctx)
}

type TxnProvider func(ctx context.Context) (context.Context, Txn, error)

type Config struct {
	DatabaseURL string `json:"database_url"`
}

//kibu:provider
func LoadConfig(ctx context.Context, store config.Store) (cfg Config, err error) {
	_, err = store.GetByKey(ctx, "database", &cfg)
	return
}

//kibu:provider
func NewConnPool(ctx context.Context, cfg Config) (pool *pgxpool.Pool, err error) {
	return pgxpool.New(ctx, cfg.DatabaseURL)
}

//kibu:provider
func NewQuerier(ctx context.Context, pool *pgxpool.Pool) (querier Querier, err error) {
	querier = New(pool)
	return
}

type txnKey struct{}

var ctxTxn = ctxutil.NewStore[pgx.Tx, txnKey]()
var ctxQuerier = ctxutil.NewStore[Querier, queryKey]()

func deriveTxnFunc(ctx context.Context, pool *pgxpool.Pool) ctxutil.DerivationFunc[pgx.Tx] {
	return func(parent pgx.Tx) (next pgx.Tx, err error) {
		if parent == nil {
			return pool.Begin(ctx)
		}

		return parent.Begin(ctx)
	}
}

//kibu:provider
func NewTxnProvider(ctx context.Context, pool *pgxpool.Pool) TxnProvider {
	return func(ctx context.Context) (context.Context, Txn, error) {
		txn, err := ctxTxn.Derive(ctx, deriveTxnFunc(ctx, pool))
		if err != nil {
			return ctx, nil, err
		}

		childCtx := ctxTxn.Save(ctx, txn)

		txnProvider := txnImpl{
			Tx:  txn,
			ctx: childCtx,
		}

		return childCtx, &txnProvider, nil
	}
}
