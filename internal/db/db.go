package db

import "context"

type Database interface {
	SetupDatabase(ctx context.Context) error
	SetupTable(ctx context.Context, tableName string) error
	Upsert(ctx context.Context, tableName string, records interface{}) error
	Close() error
}
