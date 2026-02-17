package clnitro

import (
	"context"
	"fmt"
)

type Stringmapper interface {
	Bind(ctx context.Context, h *Client) error
	BindCommand() string
	Key() string
	Stringmap() string
	Unbind(ctx context.Context, h *Client) error
	UnbindCommand() string
	Value() string
}

func bindCommand(stringmap, key, value string) string {
	return fmt.Sprintf("bind policy stringmap %s \"%s\" \"%s\"", stringmap, key, value)
}

func unbindCommand(stringmap, key string) string {
	return fmt.Sprintf("unbind policy stringmap %s \"%s\"", stringmap, key)
}
