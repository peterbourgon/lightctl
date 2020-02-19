package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/go-ocf/go-coap"
	"github.com/peterbourgon/ff/v2"
	"github.com/peterbourgon/ff/v2/ffcli"
)

func main() {
	if err := run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var (
		options = []ff.Option{ff.WithEnvVarNoPrefix()}
		rootfs  = flag.NewFlagSet("lightctl", flag.ExitOnError)
		gateway = rootfs.String("gateway", "udp://10.0.1.11:5864", "TRÅDFRI gateway address")
		conn    *coap.ClientConn
	)

	debug := &ffcli.Command{
		Name:       "status",
		ShortUsage: "lightctl debug",
		ShortHelp:  "Debug current light information",
		Exec: func(ctx context.Context, args []string) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			msg, err := conn.GetWithContext(ctx, "/15001")
			if err != nil {
				return fmt.Errorf("error making list: %w", err)
			}

			fmt.Fprintf(stdout, "%s\n", string(msg.Payload()))
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "lightctl <subcommand> ...",
		Subcommands: []*ffcli.Command{debug},
		FlagSet:     rootfs,
		Options:     options,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.Parse(args[1:]); err != nil {
		return fmt.Errorf("error during Parse: %w", err)
	}

	{
		u, err := url.Parse(*gateway)
		if err != nil {
			return fmt.Errorf("error parsing gateway: %w", err)
		}
		fmt.Fprintf(stdout, "network %s, address %s\n", u.Scheme, u.Host)

		// TODO(pb): this isn't right
		// https://github.com/glenndehaan/ikea-tradfri-coap-docs
		// https://github.com/home-assistant/home-assistant/issues/10252
		// https://github.com/obgm/libcoap/blob/develop/man/coap-client.txt.in
		conn, err = coap.DialTimeout(u.Scheme, u.Host, 3*time.Second)
		if err != nil {
			return fmt.Errorf("error during Dial: %w", err)
		}
	}

	err := root.Run(context.Background())
	switch {
	case err == nil:
		return nil
	case errors.Is(err, flag.ErrHelp):
		fmt.Fprintln(stderr)
		return nil
	default:
		return err
	}
}
