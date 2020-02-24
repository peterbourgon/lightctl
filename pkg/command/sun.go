package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/peterbourgon/ff/v2/ffcli"
	"github.com/sixdouglas/suncalc"
)

func Sun(stdout, stderr io.Writer) *ffcli.Command {
	fs := flag.NewFlagSet("lightctl sun", flag.ExitOnError)
	var (
		latitude  = fs.Float64("latitude", 52.520008, "latitude in decimal form")
		longitude = fs.Float64("longitude", 13.404954, "longitude in decimal form")
		dateStr   = fs.String("date", time.Now().Truncate(time.Minute).Format(time.RFC3339), "date to calculate for (RFC3339)")
	)
	return &ffcli.Command{
		Name:       "sun",
		ShortUsage: "lightctl sun [flags]",
		ShortHelp:  "Print height of sun",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			date, err := time.Parse(time.RFC3339, *dateStr)
			if err != nil {
				return fmt.Errorf("error parsing date: %w", err)
			}

			var (
				times   = suncalc.GetTimes(date, *latitude, *longitude)
				sunrise = times[suncalc.Sunrise].Time
				noon    = times[suncalc.SolarNoon].Time
				sunset  = times[suncalc.Sunset].Time
				percent float64
			)
			switch {
			case date.Before(sunrise):
				percent = 0.0
			case date.Before(noon):
				percent = float64(date.Sub(sunrise)) / float64(noon.Sub(sunrise))
			case date.Equal(noon):
				percent = 1.0
			case date.Before(sunset):
				percent = 1 - (float64(date.Sub(noon)) / float64(sunset.Sub(noon)))
			case date.After(sunset):
				percent = 0.0
			default:
				return fmt.Errorf("invalid date")
			}

			fmt.Fprintf(stdout, "%d\n", int(percent*100))

			return nil
		},
	}
}
