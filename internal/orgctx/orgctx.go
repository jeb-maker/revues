package orgctx

import "context"

type ctxKey struct{}

// WithOrganizationID returns a context carrying the active organization id.
func WithOrganizationID(ctx context.Context, organizationID int64) context.Context {
	return context.WithValue(ctx, ctxKey{}, organizationID)
}

// OrganizationID returns the active organization id from context, if set.
func OrganizationID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxKey{}).(int64)
	return id, ok && id > 0
}
