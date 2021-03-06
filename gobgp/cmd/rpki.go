// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/packet"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

func showRPKIServer(args []string) error {
	arg := &api.Arguments{}

	stream, err := client.GetRPKI(context.Background(), arg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	format := "%-18s %-6s %-10s %s\n"
	fmt.Printf(format, "Session", "State", "Uptime", "#IPv4/IPv6 records")
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		s := "Down"
		uptime := "never"
		if r.State.Uptime != 0 {
			s = "Up"
			uptime = fmt.Sprint(formatTimedelta(int64(time.Now().Sub(time.Unix(r.State.Uptime, 0)).Seconds())))
		}

		fmt.Printf(format, fmt.Sprintf(r.Conf.Address), s, uptime, fmt.Sprintf("%d/%d", r.State.ReceivedIpv4, r.State.ReceivedIpv6))
	}
	return nil
}

func showRPKITable(args []string) error {
	family, err := checkAddressFamily(bgp.RouteFamily(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	arg := &api.Arguments{
		Family: uint32(family),
	}
	if len(args) > 0 {
		arg.Name = args[0]
	}
	stream, err := client.GetROA(context.Background(), arg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var format string
	afi, _ := bgp.RouteFamilyToAfiSafi(family)
	if afi == bgp.AFI_IP {
		format = "%-18s %-6s %-10s %s\n"
	} else {
		format = "%-42s %-6s %-10s %s\n"
	}
	fmt.Printf(format, "Network", "Maxlen", "AS", "Server")
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if len(args) > 0 && args[0] != r.Conf.Address {
			continue
		}

		server := net.JoinHostPort(r.Conf.Address, strconv.Itoa(int(r.Conf.RemotePort)))
		fmt.Printf(format, fmt.Sprintf("%s/%d", r.Prefix, r.Prefixlen), fmt.Sprint(r.Maxlen), fmt.Sprint(r.As), server)
	}
	return nil
}

func NewRPKICmd() *cobra.Command {
	rpkiCmd := &cobra.Command{
		Use: CMD_RPKI,
	}

	modRPKI := func(op api.Operation, address string) error {
		arg := &api.ModRpkiArguments{
			Operation: op,
			Address:   address,
			Port:      323,
		}
		_, err := client.ModRPKI(context.Background(), arg)
		return err
	}

	serverCmd := &cobra.Command{
		Use: CMD_RPKI_SERVER,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				showRPKIServer(args)
				return
			} else if len(args) != 2 {
				fmt.Println("usage: gobgp rpki server <ip address> [reset|softreset|enable]")
				os.Exit(1)
			}
			addr := net.ParseIP(args[0])
			if addr == nil {
				fmt.Println("invalid ip address:", args[0])
				os.Exit(1)
			}
			var op api.Operation
			switch args[1] {
			case "reset":
				op = api.Operation_RESET
			case "softreset":
				op = api.Operation_SOFTRESET
			case "enable":
				op = api.Operation_ENABLE
			default:
				fmt.Println("unknown operation:", args[1])
				os.Exit(1)
			}
			err := modRPKI(op, addr.String())
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	rpkiCmd.AddCommand(serverCmd)

	tableCmd := &cobra.Command{
		Use: CMD_RPKI_TABLE,
		Run: func(cmd *cobra.Command, args []string) {
			showRPKITable(args)
		},
	}
	tableCmd.PersistentFlags().StringVarP(&subOpts.AddressFamily, "address-family", "a", "", "address family")

	rpkiCmd.AddCommand(tableCmd)
	return rpkiCmd
}
