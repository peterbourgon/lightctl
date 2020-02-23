package command

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/peterbourgon/lightctl/pkg/coap"
	"github.com/peterbourgon/lightctl/pkg/config"
)

func Auth(gateway *url.URL, stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl auth", flag.ExitOnError)
	var (
		username = fs.String("username", "", "username of your choice")
		code     = fs.String("code", "", "16-character security code on the bottom of the TRÅDFRI gateway")
	)

	return &ffcli.Command{
		Name:       "auth",
		ShortUsage: "lightctl auth [flags]",
		ShortHelp:  "Authenticate with the TRÅDFRI gateway",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			client, err := coap.NewClient(gateway.Scheme, gateway.Host, "Client_identity", *code)
			if err != nil {
				return fmt.Errorf("error dialing gateway: %w", err)
			}

			psk, err := client.Auth(*username)
			if err != nil {
				return fmt.Errorf("error performing auth: %w", err)
			}

			buf, err := json.Marshal(config.Config{
				Username: *username,
				PSK:      psk,
			})
			if err != nil {
				return fmt.Errorf("error marshaling config: %w", err)
			}

			if err := os.MkdirAll(filepath.Dir(config.FilePath), 0700); err != nil {
				return fmt.Errorf("error creating config directory: %w", err)
			}

			if err := ioutil.WriteFile(config.FilePath, buf, 0600); err != nil {
				return fmt.Errorf("error writing config file: %w", err)
			}

			return nil
		},
	}
}
