package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//type txKey struct{}

type executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

//func withTx(ctx context.Context, tx pgx.Tx) context.Context {
//	return context.WithValue(ctx, txKey{}, tx)
//}
//
//func txFromContext(ctx context.Context) (pgx.Tx, bool) {
//	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
//	return tx, ok
//}
