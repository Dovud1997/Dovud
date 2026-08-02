package syncport

import "context"

type ctxKey int

const skipFanoutKey ctxKey = 1

// WithoutFanout marks a context so domain services skip RecordChange
// (used when sync push already writes the changelog).
func WithoutFanout(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipFanoutKey, true)
}

// ShouldFanout is false inside WithoutFanout contexts.
func ShouldFanout(ctx context.Context) bool {
	skip, _ := ctx.Value(skipFanoutKey).(bool)
	return !skip
}
