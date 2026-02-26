package service

import (
	"context"
	"fmt"
	"sync"
	"tuno_backend/internal/domain"
)

// Event Bus Interface
type EventBus interface {
	Subscribe(eventName string, handler EventHandlerFunc)
	Publish(ctx context.Context, event domain.Event) error
}

type EventHandlerFunc func(ctx context.Context, event domain.Event) error

// Simple In-Memory Event Bus (Synchronous)
type InMemoryEventBus struct {
	handlers map[string][]EventHandlerFunc
	mu       sync.RWMutex
}

func NewEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandlerFunc),
	}
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *InMemoryEventBus) Publish(ctx context.Context, event domain.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, ok := b.handlers[event.EventName()]
	if !ok {
		// No handlers registered, which might be okay depending on the use case
		// But for critical events, we might want to log this.
		return nil
	}

	// Synchronous dispatch for now to ensure consistency as per requirements
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("event handler failed for %s: %w", event.EventName(), err)
		}
	}
	return nil
}
