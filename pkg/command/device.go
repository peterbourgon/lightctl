package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/peterbourgon/lightctl/pkg/coap"
	"github.com/peterbourgon/lightctl/pkg/config"
)

func Device(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "device",
		ShortUsage: "lightctl device <subcommand> ...",
		ShortHelp:  "Interact with devices",
		Subcommands: []*ffcli.Command{
			DeviceList(gateway, stdout, stderr),
			DeviceGet(gateway, stdout, stderr),
			DeviceSet(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func DeviceList(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "lightctl device list",
		ShortHelp:  "List known devices",
		Exec: func(ctx context.Context, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			devices, err := client.ListDevices()
			if err != nil {
				return fmt.Errorf("error listing devices: %w", err)
			}

			for _, d := range devices {
				fmt.Fprintf(os.Stdout, "%s\n", d.Short())
			}

			return nil
		},
	}
}

func DeviceGet(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl device get", flag.ExitOnError)
	var (
		id = fs.Int("id", 0, "device ID")
	)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "lightctl device get",
		ShortHelp:  "Get detailed information about a device",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return flag.ErrHelp
			}

			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			d, err := client.GetDevice(*id)
			if err != nil {
				return fmt.Errorf("error getting device %d: %w", *id, err)
			}

			fmt.Fprintf(stdout, "%+v\n", d.Long())

			return nil
		},
	}
}

func DeviceSet(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "lightctl device set <subcommand>",
		ShortHelp:  "Set properties of a device",
		Subcommands: []*ffcli.Command{
			DeviceSetLight(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func DeviceSetLight(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "light",
		ShortUsage: "lightctl device set light <subcommand>",
		ShortHelp:  "Set light control properties of a device",
		Subcommands: []*ffcli.Command{
			DeviceSetLightState(gateway, stdout, stderr),
			DeviceSetLightLevel(gateway, stdout, stderr),
			DeviceSetLightWhite(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func DeviceSetLightState(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl device set light state", flag.ExitOnError)
	var (
		id    = fs.Int("id", 0, "device ID")
		state = fs.String("state", "", "on, off")
	)

	return &ffcli.Command{
		Name:       "state",
		ShortUsage: "lightctl device set light state [flags]",
		ShortHelp:  "Set light control state of a device",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			return client.SetLightControlState(coap.RootDevices, *id, *state == "on")
		},
	}
}

func DeviceSetLightLevel(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl device set light level", flag.ExitOnError)
	var (
		id         = fs.Int("id", 0, "device ID")
		level      = fs.Int("level", 0, "0..100")
		transition = fs.Duration("transition", 0, "transition time")
	)

	return &ffcli.Command{
		Name:       "level",
		ShortUsage: "lightctl device set light level [flags]",
		ShortHelp:  "Set light control level of a device",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			if *level < 0 {
				*level = 0
			}
			if *level > 100 {
				*level = 100
			}

			dimmer := int((float64(*level) / 100) * 255.0)
			return client.SetLightControlDimmer(coap.RootDevices, *id, dimmer, *transition)
		},
	}
}

func DeviceSetLightWhite(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl device set light white", flag.ExitOnError)
	var (
		id         = fs.Int("id", 0, "device ID")
		white      = fs.Int("white", 0, "0..100 (0=red, 100=white)")
		transition = fs.Duration("transition", 0, "transition time")
	)

	return &ffcli.Command{
		Name:       "white",
		ShortUsage: "lightctl device set light white [flags]",
		ShortHelp:  "Set light control miwhites (white spectrum color) of a device",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			if *white < 0 {
				*white = 0
			}
			if *white > 100 {
				*white = 100
			}

			red := 100 - *white
			mireds := 250 + int((float64(red)/100)*(454-250))
			return client.SetLightControlMireds(coap.RootDevices, *id, mireds, *transition)
		},
	}
}
