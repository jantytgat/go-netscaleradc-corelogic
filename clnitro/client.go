package clnitro

import (
	"context"
	"fmt"
	"iter"
	"slices"

	"github.com/jantytgat/go-netscaleradc/adcconfig"
	"github.com/jantytgat/go-netscaleradc/adcnitro"
	"github.com/jantytgat/go-netscaleradc/adcresource"
)

func NewClient(name string, address string, credentials adcnitro.ApiCredentials, settings adcnitro.ApiConnectionSettings, mode adcresource.SerializationMode) (*Client, error) {
	var err error
	var nitroClient *adcnitro.Client
	if nitroClient, err = adcnitro.NewClient(name, address, credentials, settings, mode); err != nil {
		return nil, err
	}

	var cl *Client
	cl = &Client{
		nitro: nitroClient,
	}

	return cl, nil
}

type Client struct {
	nitro *adcnitro.Client
}

func (c *Client) bind(ctx context.Context, b Binding) error {
	var err error
	if err = c.isPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return err
	}

	var binding = adcconfig.PolicyStringmapPatternBinding{
		Name:  b.Stringmap,
		Key:   b.Key,
		Value: b.Value,
	}

	if err = c.nitro.PolicyStringmapPatternBinding.Add(ctx, binding); err != nil {
		// TODO customize error handling
		return err
	}
	return nil
}

func (c *Client) getBindings(ctx context.Context) (iter.Seq[Binding], error) {
	var err error
	if err = c.isPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var liveBindings []adcconfig.PolicyStringmapPatternBinding
	if liveBindings, err = c.nitro.PolicyStringmapPatternBinding.List(ctx, nil, nil); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var output = make([]Binding, len(liveBindings))
	for i, b := range liveBindings {
		output[i] = parseBinding(b)
	}

	// Remove invalid bindings
	output = slices.DeleteFunc(output, func(b Binding) bool {
		return b.BindingType() == invalidBindingType
	})
	return slices.Values(output), nil
}

func (c *Client) getCsVservers(ctx context.Context) (iter.Seq[CsVserver], error) {
	var err error
	if err = c.isPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var liveCs []adcconfig.CsVserver
	if liveCs, err = c.nitro.CsVserver.List(ctx, []string{"name"}, nil); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var output = make([]CsVserver, len(liveCs))
	for i, cs := range liveCs {
		output[i] = CsVserver{
			Name: cs.Name,
		}
	}
	return slices.Values(output), nil
}

func (c *Client) getLbVservers(ctx context.Context) (iter.Seq[LbVserver], error) {
	var err error
	if err = c.isPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var liveLb []adcconfig.LbVserver
	if liveLb, err = c.nitro.LbVserver.List(ctx, []string{"name"}, nil); err != nil {
		// TODO customize error handling
		return nil, err
	}

	var output = make([]LbVserver, len(liveLb))
	for i, lb := range liveLb {
		output[i] = LbVserver{
			Name: lb.Name,
		}
	}
	return slices.Values(output), nil
}

func (c *Client) GetState(ctx context.Context) (State, error) {
	var err error
	var csVservers iter.Seq[CsVserver]
	if csVservers, err = c.getCsVservers(ctx); err != nil {
		// TODO improve error handling
		return State{}, err
	}

	var lbVservers iter.Seq[LbVserver]
	if lbVservers, err = c.getLbVservers(ctx); err != nil {
		return State{}, err
	}

	var bindings iter.Seq[Binding]
	if bindings, err = c.getBindings(ctx); err != nil {
		return State{}, err
	}

	return State{
		CsVserver: parseCsVserverState(csVservers, bindings),
		LbVserver: parseLbVserverState(lbVservers, bindings),
	}, nil
}

func (c *Client) isPrimaryNode(ctx context.Context) error {
	var err error
	var isPrimaryNode bool
	if isPrimaryNode, err = c.nitro.IsPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return err
	}

	if !isPrimaryNode {
		// TODO customize error handling
		return fmt.Errorf("not a primary node")
	}

	return nil
}

func (c *Client) unbind(ctx context.Context, b Binding) error {
	var err error
	if err = c.isPrimaryNode(ctx); err != nil {
		// TODO customize error handling
		return err
	}

	if err = c.nitro.PolicyStringmapPatternBinding.Delete(ctx, b.Stringmap, b.Key); err != nil {
		// TODO customize error handling
		return err
	}
	return nil
}

func (c *Client) SaveConfig(ctx context.Context) error {
	return c.nitro.SaveConfig(ctx)
}
