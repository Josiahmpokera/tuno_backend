package service

import (
	"context"
	"fmt"
)

// Simplified Command Bus

type Command interface {
	CommandName() string
}

type CommandHandlerFunc func(ctx context.Context, cmd interface{}) (interface{}, error)

type CommandBus struct {
	handlers map[string]CommandHandlerFunc
}

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandlerFunc),
	}
}

func (b *CommandBus) Register(commandName string, handler CommandHandlerFunc) {
	b.handlers[commandName] = handler
}

func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) (interface{}, error) {
	handler, ok := b.handlers[cmd.CommandName()]
	if !ok {
		return nil, fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
	}
	return handler(ctx, cmd)
}
