package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := context.Background()
	if assert.NotPanics(t, func() {
		ctx = WithContext(ctx, NewHTML("", "", false))
	}) {
		assert.NotNil(t, FromContextHTML(ctx))
	}

	if assert.NotPanics(t, func() {
		ctx = WithContext(ctx, NewPlain("", "", false))
	}) {
		assert.NotNil(t, FromContextPlain(ctx))
	}
}
