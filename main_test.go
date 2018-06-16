package main

import (
	. "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this test assumes that you have 3 local instances of mongo running in a
// replica set where the first one is the primary and can't write to the local
// disk
func TestSpin(t *T) {
	sess1, err := connect("127.0.0.1:27017", 5*time.Second)
	require.NoError(t, err)
	sess2, err := connect("127.0.0.1:27018", 5*time.Second)
	require.NoError(t, err)
	sess3, err := connect("127.0.0.1:27019", 5*time.Second)
	require.NoError(t, err)

	// make sure that all work and that 1 is the primary
	// ignore the error from the first one since it can't write
	status, _ := upsert(sess1)
	require.True(t, status.Repl.IsMaster)

	_, err = upsert(sess2)
	require.NoError(t, err)

	_, err = upsert(sess3)
	require.NoError(t, err)

	// start up the spinner
	go spin(sess1, time.Second, 15*time.Second)

	var switched bool
	for i := 0; i < 6; i++ {
		status, err := upsert(sess2)
		if err == nil && status.Repl.IsMaster {
			switched = true
			break
		} else if err != nil {
			t.Logf("error from 27018: %v", err)
		}
		status, err = upsert(sess3)
		if err == nil && status.Repl.IsMaster {
			switched = true
			break
		} else if err != nil {
			t.Logf("error from 27018: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
	assert.True(t, switched)
}
