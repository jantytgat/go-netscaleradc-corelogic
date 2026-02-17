package clnitro

import (
	"context"
	"fmt"
	"strings"
)

type CsControlBinding struct {
	stringmap   string
	csVserver   string
	accessZone  AccessZone
	rule        string
	lbVserver   string
	destination string
}

func (e CsControlBinding) AccessZone() string {
	return string(e.accessZone)
}

func (e CsControlBinding) Bind(ctx context.Context, c *Client) error {
	return c.bind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e CsControlBinding) BindCommand() string {
	return bindCommand(e.stringmap, e.Key(), e.Value())
}

func (e CsControlBinding) Key() string {
	return strings.ToLower(fmt.Sprintf("%s;%s;%s", e.csVserver, string(e.accessZone), e.rule))
}

func (e CsControlBinding) Stringmap() string {
	return e.stringmap
}

func (e CsControlBinding) Unbind(ctx context.Context, c *Client) error {
	return c.unbind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e CsControlBinding) UnbindCommand() string {
	return unbindCommand(e.stringmap, e.Key())
}

func (e CsControlBinding) Value() string {
	return fmt.Sprintf("vs=%s;dst=%s;", e.lbVserver, e.destination)
}

func (e CsControlBinding) CsVserverName() string {
	return e.csVserver
}
