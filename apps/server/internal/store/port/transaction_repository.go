package port

import "context"

// TransactionRunner is the service-level boundary for composite repository flows.
// Repository implementations decide how their storage operations join the transaction.
type TransactionRunner interface {
	WithTransaction(ctx context.Context, run func(ctx context.Context, repository Store) error) error
}
