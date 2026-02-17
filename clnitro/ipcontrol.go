package clnitro

import (
	"context"
	"fmt"
	"strings"
)

const (
	InvalidAccessList  AccessList = ""
	AllowAccessList    AccessList = "allow"
	BlockAccessList    AccessList = "block"
	NoneAccessList     AccessList = "none"
	DisabledAccessList AccessList = "disabled"
)

type AccessList string

func parseAccessList(s string) AccessList {
	switch s {
	case "allowlist":
		return AllowAccessList
	case "blocklist":
		return BlockAccessList
	case "disabled":
		return DisabledAccessList
	case "none":
		return NoneAccessList
	default:
		return InvalidAccessList
	}
}

type IpControlCidrBinding struct {
	stringmap   string
	vServer     string
	accessZone  AccessZone
	cidr        string
	description string
}

func (e IpControlCidrBinding) AccessZone() string {
	return string(e.accessZone)
}

func (e IpControlCidrBinding) Bind(ctx context.Context, c *Client) error {
	return c.bind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e IpControlCidrBinding) BindCommand() string {
	return bindCommand(e.stringmap, e.Key(), e.Value())
}

func (e IpControlCidrBinding) Cidr() string {
	return e.cidr
}

func (e IpControlCidrBinding) Description() string {
	return e.description
}

func (e IpControlCidrBinding) Key() string {
	return strings.ToLower(fmt.Sprintf("%s;%s;%s", e.vServer, string(e.accessZone), e.cidr))
}

func (e IpControlCidrBinding) Stringmap() string {
	return e.stringmap
}

func (e IpControlCidrBinding) Unbind(ctx context.Context, c *Client) error {
	return c.unbind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e IpControlCidrBinding) UnbindCommand() string {
	return unbindCommand(e.stringmap, e.Key())
}

func (e IpControlCidrBinding) Value() string {
	return fmt.Sprintf("%s", e.description)
}

func (e IpControlCidrBinding) VserverName() string {
	return e.vServer
}

type IpControlConfigBinding struct {
	stringmap  string
	vServer    string
	accessList AccessList
}

func (e IpControlConfigBinding) Bind(ctx context.Context, c *Client) error {
	return c.bind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e IpControlConfigBinding) BindCommand() string {
	return bindCommand(e.stringmap, e.Key(), e.Value())
}

func (e IpControlConfigBinding) Key() string {
	return strings.ToLower(fmt.Sprintf("%s", e.vServer))
}

func (e IpControlConfigBinding) ListType() AccessList {
	return e.accessList
}

func (e IpControlConfigBinding) Stringmap() string {
	return e.stringmap
}

func (e IpControlConfigBinding) Unbind(ctx context.Context, c *Client) error {
	return c.unbind(ctx, Binding{
		Stringmap: e.stringmap,
		Key:       e.Key(),
		Value:     e.Value(),
	})
}

func (e IpControlConfigBinding) UnbindCommand() string {
	return unbindCommand(e.stringmap, e.Key())
}

func (e IpControlConfigBinding) Value() string {
	return fmt.Sprintf("list=%slist;", string(e.accessList))
}

func (e IpControlConfigBinding) VserverName() string {
	return e.vServer
}
