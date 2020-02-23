package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/peterbourgon/ff/v2"
	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/peterbourgon/lightctl/pkg/command"
)

// https://github.com/glenndehaan/ikea-tradfri-coap-docs
// https://github.com/home-assistant/home-assistant/issues/10252
// https://github.com/obgm/libcoap/blob/develop/man/coap-client.txt.in
// https://github.com/ggravlingen/pytradfri/blob/master/pytradfri/const.py

func main() {
	if err := run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var (
		options    = []ff.Option{ff.WithEnvVarNoPrefix()}
		rootfs     = flag.NewFlagSet("lightctl", flag.ExitOnError)
		gateway    = rootfs.String("gateway", "udp://10.0.1.11:5684", "TRÅDFRI gateway address")
		gatewayURL url.URL

		auth   = command.Auth(&gatewayURL, stdout, stderr)
		device = command.Device(&gatewayURL, stdout, stderr)
		group  = command.Group(&gatewayURL, stdout, stderr)
	)

	root := &ffcli.Command{
		ShortUsage:  "lightctl <subcommand> ...",
		Subcommands: []*ffcli.Command{auth, device, group},
		FlagSet:     rootfs,
		Options:     options,
		Exec:        func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}

	if err := root.Parse(args[1:]); err != nil {
		return fmt.Errorf("error during Parse: %w", err)
	}

	u, err := url.Parse(*gateway)
	if err != nil {
		return fmt.Errorf("error parsing gateway: %w", err)
	}
	gatewayURL = *u

	err = root.Run(context.Background())
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
