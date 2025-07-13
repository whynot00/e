package e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whynot00/e"
)

func TestWrapRecovered_WithStack(t *testing.T) {
	err := e.WrapRecovered(&e.RecoverOpts{}, "something went wrong")

	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "something went wrong")

	wrapped, ok := err.(*e.ErrorWrapper)
	require.True(t, ok)
	assert.NotEmpty(t, wrapped.StackTrace())
}

func TestWrapRecovered_WithoutStack(t *testing.T) {
	err := e.WrapRecovered(&e.RecoverOpts{WithoutStack: true}, "oops")

	require.NotNil(t, err)
	wrapped, ok := err.(*e.ErrorWrapper)
	require.True(t, ok)
	assert.Empty(t, wrapped.StackTrace())
}

func TestRecover_CallsCallback(t *testing.T) {
	var called bool
	var recoveredErr error

	func() {
		defer e.Recover(nil, func(err error) {
			called = true
			recoveredErr = err
		})
		panic("test panic")
	}()

	require.True(t, called)
	require.NotNil(t, recoveredErr)
	assert.Contains(t, recoveredErr.Error(), "test panic")
}

func TestRecover_WithoutStack(t *testing.T) {
	var frames int

	func() {
		defer e.Recover(&e.RecoverOpts{WithoutStack: true}, func(err error) {
			wrapped := err.(*e.ErrorWrapper)
			frames = len(wrapped.StackTrace())
		})
		panic("no stack")
	}()

	assert.Equal(t, 0, frames)
}

func TestRecover_Suppressed(t *testing.T) {
	var called bool

	func() {
		defer e.Recover(&e.RecoverOpts{RecoverOnly: true}, func(err error) {
			called = true
		})
		panic("should not be called")
	}()

	assert.False(t, called)
}

func TestRecoverToChannel(t *testing.T) {
	ch := make(chan error, 1)

	go func() {
		defer e.RecoverToChannel(nil, ch)
		panic("from goroutine")
	}()

	err := <-ch
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "from goroutine")
}

func TestRecoverToChannel_Suppressed(t *testing.T) {
	ch := make(chan error, 1)

	go func() {
		defer e.RecoverToChannel(&e.RecoverOpts{RecoverOnly: true}, ch)
		panic("no send")
	}()

	select {
	case err := <-ch:
		t.Errorf("unexpected error sent: %v", err)
	default:
		// OK: nothing sent
	}
}
