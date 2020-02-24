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

func Group(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "group",
		ShortUsage: "lightctl group <subcommand> ...",
		ShortHelp:  "Interact with groups",
		Subcommands: []*ffcli.Command{
			GroupList(gateway, stdout, stderr),
			GroupGet(gateway, stdout, stderr),
			GroupSet(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func GroupList(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "lightctl group list",
		ShortHelp:  "List known groups",
		Exec: func(ctx context.Context, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			client, err := coap.NewClient(gateway.Scheme, gateway.Host, c.Username, c.PSK)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			groups, err := client.ListGroups()
			if err != nil {
				return fmt.Errorf("error listing groups: %w", err)
			}

			for _, g := range groups {
				fmt.Fprintf(os.Stdout, "%s\n", g.Short())
			}

			return nil
		},
	}
}

func GroupGet(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl group get", flag.ExitOnError)
	var (
		id = fs.Int("id", 0, "group ID")
	)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "lightctl group get",
		ShortHelp:  "Get detailed information about a group",
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

			g, err := client.GetGroup(*id)
			if err != nil {
				return fmt.Errorf("error listing groups: %w", err)
			}

			fmt.Fprintf(stdout, "%s\n", g.Long())

			return nil
		},
	}
}

func GroupSet(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "lightctl group set <subcommand>",
		ShortHelp:  "Set properties of a group",
		Subcommands: []*ffcli.Command{
			GroupSetLight(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func GroupSetLight(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	return &ffcli.Command{
		Name:       "light",
		ShortUsage: "lightctl group set light <subcommand>",
		ShortHelp:  "Set light control properties of a group",
		Subcommands: []*ffcli.Command{
			GroupSetLightState(gateway, stdout, stderr),
			GroupSetLightLevel(gateway, stdout, stderr),
			GroupSetLightWhite(gateway, stdout, stderr),
		},
		Exec: func(ctx context.Context, args []string) error { return flag.ErrHelp },
	}
}

func GroupSetLightState(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl group set light state", flag.ExitOnError)
	var (
		id    = fs.Int("id", 0, "group ID")
		state = fs.String("state", "", "on, off")
	)

	return &ffcli.Command{
		Name:       "state",
		ShortUsage: "lightctl group set light state [flags]",
		ShortHelp:  "Set light control state of a group",
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

			return client.SetLightControlState(coap.RootGroups, *id, *state == "on")
		},
	}
}

func GroupSetLightLevel(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl group set light level", flag.ExitOnError)
	var (
		id         = fs.Int("id", 0, "group ID")
		level      = fs.Int("level", 0, "0..100")
		transition = fs.Duration("transition", 0, "transition time")
	)

	return &ffcli.Command{
		Name:       "level",
		ShortUsage: "lightctl group set light level [flags]",
		ShortHelp:  "Set light control level of a group",
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
			return client.SetLightControlDimmer(coap.RootGroups, *id, dimmer, *transition)
		},
	}
}

func GroupSetLightWhite(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl group set light white", flag.ExitOnError)
	var (
		id         = fs.Int("id", 0, "group ID")
		white      = fs.Int("white", 0, "0..100 (0=red, 100=white)")
		transition = fs.Duration("transition", 0, "transition time")
	)

	return &ffcli.Command{
		Name:       "white",
		ShortUsage: "lightctl group set light white [flags]",
		ShortHelp:  "Set light control mireds (white spectrum color) of a group",
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

			return client.SetLightControlMireds(coap.RootGroups, *id, mireds, *transition)
		},
	}
}
