package testutil

import (
	"context"
	"sync"
)

// SchemaContext provides a per-test schema context that can be passed explicitly to repository methods
// This replaces the global variable approach to improve test isolation
type SchemaContext struct {
	// SchemaName is the name of the schema to use for this test
	SchemaName string

	// Context is the base context to use for this test, can carry schema information
	Context context.Context

	// mu protects the schema context from concurrent access
	mu sync.Mutex
}

// NewSchemaContext creates a new schema context with the given schema name
func NewSchemaContext(schemaName string) *SchemaContext {
	// Create a base context with the schema name as a value
	ctx := context.WithValue(context.Background(), contextKeySchema, schemaName)

	return &SchemaContext{
		SchemaName: schemaName,
		Context:    ctx,
	}
}

// WithContext returns a new context with schema information embedded
func (sc *SchemaContext) WithContext(ctx context.Context) context.Context {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// If the context already has a schema value, use that
	if existingSchema, ok := ctx.Value(contextKeySchema).(string); ok && existingSchema != "" {
		return ctx
	}

	// Otherwise, add the schema name to the context
	return context.WithValue(ctx, contextKeySchema, sc.SchemaName)
}

// GetSchemaName returns the current schema name
func (sc *SchemaContext) GetSchemaName() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.SchemaName
}

// SetSchemaName updates the schema name
func (sc *SchemaContext) SetSchemaName(schemaName string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.SchemaName = schemaName
	// Update the context with the new schema name
	sc.Context = context.WithValue(sc.Context, contextKeySchema, schemaName)
}

// context key for schema name
type contextKey string

const (
	contextKeySchema contextKey = "schema"
)

// GetSchemaFromContext extracts the schema name from a context
func GetSchemaFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if schemaName, ok := ctx.Value(contextKeySchema).(string); ok {
		return schemaName
	}
	return ""
}
