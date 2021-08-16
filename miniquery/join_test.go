package miniquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	assert.Equal(t, "", Join([]string{}))
	assert.Equal(t, "1=1", Join([]string{`1=1`}))
	assert.Equal(t, "(1=1) and (2=2)", Join([]string{`1=1`, `2=2`}))
}
