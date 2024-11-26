package sigctx

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name           string
		parentCtx      context.Context
		expectCanceled bool
	}{
		{
			name:           "returns valid context and cancel",
			parentCtx:      context.Background(),
			expectCanceled: false,
		},
		{
			name: "already canceled parent context",
			parentCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			expectCanceled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := New(tt.parentCtx)
			defer cancel()

			if tt.expectCanceled {
				assert.ErrorIs(t, ctx.Err(), context.Canceled)
			} else {
				assert.NoError(t, ctx.Err())
			}
		})
	}
}

func Test_New_CancelPropagation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(done chan struct{}) (context.Context, context.CancelFunc)
		expected error
	}{
		{
			name: "manual cancellation propagates",
			setup: func(done chan struct{}) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			expected: context.Canceled,
		},
		{
			name: "timeout propagates",
			setup: func(done chan struct{}) (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			expected: context.DeadlineExceeded,
		},
		{
			name: "deadline propagates",
			setup: func(done chan struct{}) (context.Context, context.CancelFunc) {
				return context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
			},
			expected: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			pCtx, pCancel := tt.setup(done)
			defer pCancel()

			ctx, cancel := New(pCtx)
			defer cancel()

			select {
			case <-ctx.Done():
				assert.ErrorIs(t, ctx.Err(), tt.expected)
			case <-time.After(1 * time.Second): // timeout for safety
				t.Fatal("test timed out waiting for context cancellation")
			}
		})
	}
}

func Test_New_SignalHandling(t *testing.T) {
	tests := []struct {
		name   string
		signal os.Signal
	}{
		{
			name:   "handles SIGTERM",
			signal: syscall.SIGTERM,
		},
		{
			name:   "handles SIGINT",
			signal: syscall.SIGINT,
		},
		{
			name:   "handles SIGHUP",
			signal: syscall.SIGHUP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := New(context.Background())
			defer cancel()

			proc, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)

			// Send signal after small delay to ensure handler is setup
			go func() {
				time.Sleep(50 * time.Millisecond)
				require.NoError(t, proc.Signal(tt.signal))
			}()

			select {
			case <-ctx.Done():
				assert.ErrorIs(t, ctx.Err(), context.Canceled)
			case <-time.After(500 * time.Millisecond):
				t.Fatal("context was not canceled after signal")
			}
		})
	}
}

func Test_New_MultipleSignals(t *testing.T) {
	ctx, cancel := New(context.Background())
	defer cancel()

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP}

	// Send multiple signals in sequence
	go func() {
		time.Sleep(50 * time.Millisecond)
		for _, sig := range signals {
			require.NoError(t, proc.Signal(sig))
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not canceled after multiple signals")
	}
}

func Test_New_ConcurrentCancellation(t *testing.T) {
	ctx, cancel := New(context.Background())
	defer cancel()

	// Attempt concurrent cancellations
	go cancel()
	go cancel()
	go cancel()

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context was not canceled")
	}
}
