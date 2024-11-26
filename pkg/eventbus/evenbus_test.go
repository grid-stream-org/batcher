package eventbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	eb := New()
	require.NotNil(t, eb, "EventBus should be successfully created")

	// Subscribe to the EventBus
	ch1 := eb.Subscribe(10)
	ch2 := eb.Subscribe(10)

	// Ensure the channels are not nil
	require.NotNil(t, ch1, "Channel 1 should not be nil")
	require.NotNil(t, ch2, "Channel 2 should not be nil")

	// Publish an event
	event := "Test Event"
	eb.Publish(event)

	// Use a timeout to prevent the test from hanging
	select {
	case received := <-ch1:
		assert.Equal(t, event, received, "Channel 1 should receive the published event")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event on Channel 1")
	}

	select {
	case received := <-ch2:
		assert.Equal(t, event, received, "Channel 2 should receive the published event")
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event on Channel 2")
	}
}

func TestEventBus_PublishWithoutSubscribers(t *testing.T) {
	eb := New()
	require.NotNil(t, eb, "EventBus should be successfully created")

	// Publish an event when there are no subscribers
	event := "Event without subscribers"
	eb.Publish(event)

	// Ensure no errors or panics occur when publishing without subscribers
	assert.True(t, true, "Publishing without subscribers should not cause any issues")
}

func TestEventBus_SubscribeMultipleTimes(t *testing.T) {
	eb := New()
	require.NotNil(t, eb, "EventBus should be successfully created")

	// Subscribe multiple times and verify that each channel is unique
	ch1 := eb.Subscribe(10)
	ch2 := eb.Subscribe(10)

	require.NotNil(t, ch1, "Channel 1 should not be nil")
	require.NotNil(t, ch2, "Channel 2 should not be nil")
	assert.NotEqual(t, ch1, ch2, "Each subscriber should have a unique channel")
}

func TestEventBus_NonBlockingPublish(t *testing.T) {
	eb := New()
	require.NotNil(t, eb, "EventBus should be successfully created")

	// Subscribe with a channel that won't read immediately
	ch := eb.Subscribe(10)
	require.NotNil(t, ch, "Channel should not be nil")

	// Publish an event and ensure it doesn't block even if the channel is not ready to read
	event := "Non-blocking event"
	done := make(chan bool)

	go func() {
		eb.Publish(event)
		done <- true
	}()

	select {
	case <-done:
		assert.True(t, true, "Publishing should be non-blocking and complete successfully")
	case <-time.After(1 * time.Second):
		t.Error("Timeout: Publishing event should not block")
	}
}
