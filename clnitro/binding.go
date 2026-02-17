package clnitro

import (
	"context"
	"fmt"
	"strings"

	"github.com/jantytgat/go-netscaleradc/adcconfig"
)

const (
	invalidBindingType         BindingType = "invalid"
	csControlBindingType       BindingType = "cs_control"
	ipControlBindingType       BindingType = "ip_control"
	ipControlConfigBindingType BindingType = "ip_control_config"
	ipControlZoneBindingType   BindingType = "ip_control_zone"
	ipControlAclBindingType    BindingType = "ip_control_acl"
)

type BindingType string

func parseBinding(b adcconfig.PolicyStringmapPatternBinding) Binding {
	return Binding{
		Stringmap: b.Name,
		Key:       b.Key,
		Value:     b.Value,
	}
}

type Binding struct {
	Stringmap string
	Key       string
	Value     string
}

func (b Binding) BindingType() BindingType {
	switch calcBindingFromStringmapName(b.Stringmap) {
	case csControlBindingType:
		return csControlBindingType
	case ipControlBindingType:
		return calcBindingFromIpControlBinding(b.Key, b.Value)
	default:
		return invalidBindingType
	}
}

func (b Binding) Update(ctx context.Context, c *Client) error {
	// TODO improve error handling
	return c.nitro.PolicyStringmapPatternBinding.Add(ctx, adcconfig.PolicyStringmapPatternBinding{
		Key:   b.Key,
		Name:  b.Stringmap,
		Value: b.Value,
	})
}

func (b Binding) Delete(ctx context.Context, c *Client) error {
	// TODO improve error handling
	return c.nitro.PolicyStringmapPatternBinding.Delete(ctx, b.Stringmap, b.Key)
}

func calcBindingFromStringmapName(name string) BindingType {
	if strings.Contains(strings.ToLower(name), "ip_control") {
		return ipControlBindingType
	}
	if strings.Contains(strings.ToLower(name), "cs_control") {
		return csControlBindingType
	}
	return invalidBindingType
}

func calcBindingFromIpControlBinding(key, value string) BindingType {
	if !strings.Contains(key, ";") && strings.HasPrefix(value, "list=") {
		return ipControlConfigBindingType
	}

	splitKey := strings.Split(key, ";")
	if len(splitKey) != 3 {
		return invalidBindingType
	}

	switch splitKey[1] {
	case string(AnyAccessZone):
		return ipControlAclBindingType
	case string(LanAccessZone):
		return ipControlZoneBindingType
	default:
		fmt.Println("Invalid binding type for key:", key)
		return invalidBindingType
	}
}
