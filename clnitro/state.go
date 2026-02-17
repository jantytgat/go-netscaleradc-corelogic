package clnitro

import (
	"fmt"
	"iter"
	"regexp"
	"slices"
	"sort"
	"strings"
)

type CsState struct {
	Flows  iter.Seq[CsControlBinding]
	Config iter.Seq[IpControlConfigBinding]
	Zone   iter.Seq[IpControlCidrBinding]
	Acl    iter.Seq[IpControlCidrBinding]
}

type CsVserver struct {
	Name  string
	State CsState
}

func (v CsVserver) IsConfigured() bool {
	return v.State.Config != nil && len(slices.Collect(v.State.Config)) > 0
}

type LbState struct {
	Config iter.Seq[IpControlConfigBinding]
	Acl    iter.Seq[IpControlCidrBinding]
}

type LbVserver struct {
	Name  string
	State LbState
}

func (v LbVserver) IsConfigured() bool {
	return v.State.Config != nil && len(slices.Collect(v.State.Config)) > 0
}

type State struct {
	CsVserver iter.Seq[CsVserver]
	LbVserver iter.Seq[LbVserver]
}

func (s State) MissingVservers() iter.Seq[string] {
	iCsv := slices.Collect(s.CsVserver)
	iLbv := slices.Collect(s.LbVserver)

	iCsv = slices.DeleteFunc(iCsv, func(v CsVserver) bool { return v.IsConfigured() })
	iLbv = slices.DeleteFunc(iLbv, func(v LbVserver) bool { return v.IsConfigured() })

	var output []string
	for _, csv := range iCsv {
		output = append(output, csv.Name)
	}

	for _, lbv := range iLbv {
		output = append(output, lbv.Name)
	}
	return slices.Values(output)
}

var (
	regCsControlKey         = regexp.MustCompile("^(.*);(any|lan);(.*)")
	regCsControlVsValue     = regexp.MustCompile("^vs=(.*);")
	regCsControlDstValue    = regexp.MustCompile("^.*;dst=(.*);")
	regIpControlConfigValue = regexp.MustCompile("^list=(.*);")
	regIpControlCidrKey     = regexp.MustCompile("^(.*);(any|lan);(.*)")
)

func parseCsControlEntry(b Binding) CsControlBinding {
	keys := regCsControlKey.FindStringSubmatch(b.Key)
	vs := regCsControlVsValue.FindStringSubmatch(b.Value)
	var dst string
	if regCsControlDstValue.FindStringSubmatch(b.Value) != nil {
		dst = regCsControlDstValue.FindStringSubmatch(b.Value)[1]
	}
	return CsControlBinding{
		stringmap:   b.Stringmap,
		csVserver:   keys[1],
		accessZone:  parseAccessZone(keys[2]),
		rule:        keys[3],
		lbVserver:   vs[1],
		destination: dst,
	}
}

func parseIpControlConfigEntry(b Binding) IpControlConfigBinding {
	return IpControlConfigBinding{
		stringmap:  b.Stringmap,
		vServer:    b.Key,
		accessList: parseAccessList(regIpControlConfigValue.FindStringSubmatch(b.Value)[1]),
	}
}
func parseIpControlCidrEntry(b Binding) IpControlCidrBinding {
	matches := regIpControlCidrKey.FindStringSubmatch(b.Key)
	return IpControlCidrBinding{
		stringmap:   b.Stringmap,
		vServer:     matches[1],
		accessZone:  parseAccessZone(matches[2]),
		cidr:        matches[3],
		description: b.Value,
	}
}

func parseCsVserverBindings(cs CsVserver, bindings iter.Seq[Binding], ch chan CsVserver) {
	var csControlBindings []CsControlBinding
	var configBindings []IpControlConfigBinding
	var zoneBindings []IpControlCidrBinding
	var aclBindings []IpControlCidrBinding

	for b := range bindings {
		if b.BindingType() == csControlBindingType && strings.HasPrefix(b.Key, strings.ToLower(fmt.Sprintf("%s;", cs.Name))) {
			if regCsControlKey.MatchString(b.Key) && regCsControlVsValue.MatchString(b.Value) {
				csControlBindings = append(csControlBindings, parseCsControlEntry(b))
			}
		}
		if b.BindingType() == ipControlConfigBindingType && strings.ToLower(cs.Name) == strings.ToLower(b.Key) {
			if regIpControlConfigValue.MatchString(b.Value) {
				configBindings = append(configBindings, parseIpControlConfigEntry(b))
				continue
			}
		}
		if strings.HasPrefix(b.Key, strings.ToLower(fmt.Sprintf("%s;", cs.Name))) {
			switch b.BindingType() {
			case ipControlZoneBindingType:
				if regIpControlCidrKey.MatchString(b.Key) {
					zoneBindings = append(zoneBindings, parseIpControlCidrEntry(b))
				}
			case ipControlAclBindingType:
				if regIpControlCidrKey.MatchString(b.Key) {
					aclBindings = append(aclBindings, parseIpControlCidrEntry(b))
				}
			default:
				continue
			}
		}
	}
	sort.Slice(csControlBindings, func(i, j int) bool {
		return csControlBindings[i].Key() < csControlBindings[j].Key()
	})
	sort.Slice(configBindings, func(i, j int) bool {
		return configBindings[i].Key() < configBindings[j].Key()
	})
	sort.Slice(zoneBindings, func(i, j int) bool {
		return zoneBindings[i].Key() < zoneBindings[j].Key()
	})
	sort.Slice(aclBindings, func(i, j int) bool {
		return aclBindings[i].Key() < aclBindings[j].Key()
	})

	cs.State = CsState{
		Flows:  slices.Values(csControlBindings),
		Config: slices.Values(configBindings),
		Zone:   slices.Values(zoneBindings),
		Acl:    slices.Values(aclBindings),
	}
	ch <- cs
}

func parseCsVserverState(csVservers iter.Seq[CsVserver], bindings iter.Seq[Binding]) iter.Seq[CsVserver] {
	var lenCsVservers = len(slices.Collect(csVservers))
	var output = make([]CsVserver, lenCsVservers)
	var chCsVserver = make(chan CsVserver, lenCsVservers)

	for cs := range csVservers {
		go parseCsVserverBindings(cs, bindings, chCsVserver)
	}

	var done bool
	var processed int
	for {
		if done {
			break
		}
		select {
		case cs := <-chCsVserver:
			output[processed] = cs
			processed++
		default:
			if processed == lenCsVservers {
				done = true
			}
		}
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].Name < output[j].Name
	})
	return slices.Values(output)
}

func parseLbVserverBindings(lb LbVserver, bindings iter.Seq[Binding], ch chan LbVserver) {
	var configBindings []IpControlConfigBinding
	var aclBindings []IpControlCidrBinding

	for b := range bindings {
		if b.BindingType() == ipControlConfigBindingType && strings.ToLower(lb.Name) == strings.ToLower(b.Key) {
			if regIpControlConfigValue.MatchString(b.Value) {
				configBindings = append(configBindings, parseIpControlConfigEntry(b))
			}
			continue
		}
		if strings.HasPrefix(b.Key, strings.ToLower(fmt.Sprintf("%s;", lb.Name))) {
			switch b.BindingType() {
			case ipControlAclBindingType:
				if regIpControlCidrKey.MatchString(b.Key) {
					aclBindings = append(aclBindings, parseIpControlCidrEntry(b))
				}
			default:
				continue
			}
		}
	}
	sort.Slice(configBindings, func(i, j int) bool {
		return configBindings[i].Key() < configBindings[j].Key()
	})
	sort.Slice(aclBindings, func(i, j int) bool {
		return aclBindings[i].Key() < aclBindings[j].Key()
	})

	lb.State = LbState{
		Config: slices.Values(configBindings),
		Acl:    slices.Values(aclBindings),
	}
	ch <- lb
}

func parseLbVserverState(lbVservers iter.Seq[LbVserver], bindings iter.Seq[Binding]) iter.Seq[LbVserver] {
	var lenLbVservers = len(slices.Collect(lbVservers))
	var output = make([]LbVserver, lenLbVservers)
	var chLbVserver = make(chan LbVserver, lenLbVservers)

	for lb := range lbVservers {
		go parseLbVserverBindings(lb, bindings, chLbVserver)
	}

	var done bool
	var processed int
	for {
		if done {
			break
		}
		select {
		case lb := <-chLbVserver:
			output[processed] = lb
			processed++
		default:
			if processed == lenLbVservers {
				done = true
			}
		}
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].Name < output[j].Name
	})
	return slices.Values(output)
}
