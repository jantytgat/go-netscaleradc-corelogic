package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jantytgat/go-netscaleradc/adcnitro"
	"github.com/jantytgat/go-netscaleradc/adcresource"

	"github.com/jantytgat/go-netscaleradc-corelogic/clnitro"
)

var (
	NITRO_ENV_NAME     string
	NITRO_ENV_ADDRESS  string
	NITRO_ENV_USERNAME string
	NITRO_ENV_PASSWORD string
)

func main() {
	var err error
	var file *os.File

	if file, err = os.Open(".apiClient.env"); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		input := scanner.Text()
		if strings.HasPrefix(input, "#") {
			continue
		}
		values := strings.Split(input, "=")
		switch values[0] {
		case "NITRO_ENV_NAME":
			NITRO_ENV_NAME = values[1]
			fmt.Println("NITRO_ENV_NAME:", NITRO_ENV_NAME)
		case "NITRO_ENV_ADDRESS":
			NITRO_ENV_ADDRESS = values[1]
			fmt.Println("NITRO_ENV_ADDRESS:", NITRO_ENV_ADDRESS)
		case "NITRO_ENV_USERNAME":
			NITRO_ENV_USERNAME = values[1]
			fmt.Println("NITRO_ENV_USERNAME:", NITRO_ENV_USERNAME)
		case "NITRO_ENV_PASSWORD":
			NITRO_ENV_PASSWORD = values[1]
			if NITRO_ENV_PASSWORD == "" {
				panic(errors.New("NITRO_ENV_PASSWORD is empty"))
			}
		default:
			fmt.Println("Unknown variable:", input)
			panic(values)
		}
		line++
	}

	//var client *clnitro.client
	client, clientErr := clnitro.NewClient(
		NITRO_ENV_NAME,
		NITRO_ENV_ADDRESS,
		adcnitro.ApiCredentials{
			Username: NITRO_ENV_USERNAME,
			Password: NITRO_ENV_PASSWORD,
		},
		adcnitro.ApiConnectionSettings{
			UseSsl:                    false,
			Timeout:                   0,
			UserAgent:                 "dev-nitro",
			ValidateServerCertificate: false,
			LogTlsSecrets:             false,
			LogTlsSecretsDestination:  "",
			AutoLogin:                 false,
		},
		adcresource.StrictSerializationMode)
	if clientErr != nil {
		panic(errors.New("error creating client: " + clientErr.Error()))
	}

	ctx := context.Background()
	var s clnitro.State
	if s, err = client.GetState(ctx); err != nil {
		panic(err)
	}

	for entry := range s.CsVserver {
		fmt.Println(entry.Name)
		fmt.Println("\tFlows:")
		if entry.State.Flows != nil {
			for v := range entry.State.Flows {
				fmt.Println(" \t\t", v.BindCommand())
			}
		}
		fmt.Println("\tConfiguration:")
		if entry.State.Config != nil {
			for v := range entry.State.Config {
				fmt.Println(" \t\t", v.BindCommand())
			}
		}
		fmt.Println("\tZones:")
		if entry.State.Zone != nil {
			for z := range entry.State.Zone {
				fmt.Println(" \t\t", z.BindCommand())
			}
		}
		fmt.Println("\tACL:")
		if entry.State.Acl != nil {
			for a := range entry.State.Acl {
				fmt.Println(" \t\t", a.Cidr(), a.Value())
			}
		}
		fmt.Println()
	}

	for entry := range s.LbVserver {
		fmt.Println(entry.Name)
		fmt.Println("\tConfiguration:")
		if entry.State.Config != nil {
			for v := range entry.State.Config {
				fmt.Println(" \t\t", v.BindCommand())
			}
		}
		fmt.Println("\tACL:")
		if entry.State.Acl != nil {
			for a := range entry.State.Acl {
				fmt.Println(" \t\t", a.Cidr(), a.Value())
			}
		}
		fmt.Println()
	}

	for missing := range s.MissingVservers() {
		fmt.Println("Missing vserver:", missing)
	}
}
