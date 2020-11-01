package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_errDone(t *testing.T) {
	assert.EqualError(t, errDone, "done")
}
