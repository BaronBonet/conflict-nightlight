package infrastructure

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCorrelationIDFunctions(t *testing.T) {
	// Create a context with a correlation ID using AddCorrelationIdToContext
	correlationID := "test-correlation-id"
	ctx, err := AddCorrelationIdToContext(context.Background(), &correlationID)
	assert.NoError(t, err)

	// Test if the function extracts the correct correlation ID
	extractedCorrelationID := ExtractCorrelationIDFromContext(ctx)
	if extractedCorrelationID != correlationID {
		t.Errorf("Expected correlation ID '%s', but got '%s'", correlationID, extractedCorrelationID)
	}

	// Test if the function returns "unknown" for a context without a correlation ID
	emptyCtx := context.Background()
	unknownCorrelationID := ExtractCorrelationIDFromContext(emptyCtx)
	if unknownCorrelationID != "unknown" {
		t.Errorf("Expected 'unknown', but got '%s'", unknownCorrelationID)
	}
}
