package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActiveMembers(t *testing.T) {
	active := newActiveMembers()
	active.add("1", "A")

	assert.Equal(t, "A", active.get("1"))

	assert.True(t, active.keyIn("1"))
	assert.False(t, active.keyIn("2"))

	active.delete("1")
	assert.False(t, active.keyIn("1"))

}
