package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/services"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"text/tabwriter"
)

type CliHandler struct {
	app *cli.App
}

func NewCLIHandler(ctx context.Context, productService *services.OrchestratorService) *CliHandler {
	app := &cli.App{
		Name:                 "Map Controller",
		EnableBashCompletion: true,
		Usage:                "A cli tool for the conflict nightlight application.",
		Commands: []*cli.Command{
			{
				Name:  "listRawMaps",
				Usage: "List all raw maps",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "json",
					},
				},
				Action: func(c *cli.Context) error {
					maps, err := productService.ListRawInternalMaps(ctx)
					if err != nil {
						return err
					}
					if c.String("json") == "true" {
						return printMapsAsJson(maps)
					} else {
						return printMapsAsTable(maps)
					}
				},
			},
			{
				Name:  "listProcessedMaps",
				Usage: "List all processed maps",
				Action: func(c *cli.Context) error {
					maps, err := productService.ListProcessedInternalMaps(ctx)
					if err != nil {
						return err
					}
					if c.String("json") == "true" {
						return printMapsAsJson(maps)
					} else {
						return printMapsAsTable(maps)
					}
				},
			},
			{
				Name:  "listPublishedMaps",
				Usage: "List all published maps",
				Action: func(c *cli.Context) error {
					maps, err := productService.ListPublishedMaps(ctx)
					if err != nil {
						return err
					}
					return printPublishedMapsToTerminal(maps)
				},
			},
			{
				Name:      "publishMap",
				Usage:     "publish a map, pass in a map that is copied from the output of listProcessedMaps",
				ArgsUsage: "[map]",
				Action: func(c *cli.Context) error {
					m, err := getMapFromArgs(c)
					if err != nil {
						return err
					}
					if err := productService.PublishMap(ctx, m); err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:      "deleteMap",
				Usage:     "Delete a map from our entire system, including the tile server",
				ArgsUsage: "[map]",
				Action: func(c *cli.Context) error {
					m, err := getMapFromArgs(c)
					if err != nil {
						return err
					}
					productService.DeleteMap(ctx, m)
					return nil
				},
			},

			{
				Name:      "invokeFullPipeline",
				Usage:     "Runs the sync map request with will subsequently put messages on sqs and invoke the entire pipeline",
				ArgsUsage: "[syncMapRequest]",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("the map argument is required")
					}
					syncMapRequest := conflict_nightlightv1.SyncMapRequest{}
					if err := json.Unmarshal([]byte(c.Args().Get(0)), &syncMapRequest); err != nil {
						return err
					}
					_, err := productService.SyncInternalWithExternalMaps(ctx, prototransformers.ProtoToSyncMapsRequest(&syncMapRequest))
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	return &CliHandler{app: app}
}

func (h *CliHandler) Run(args []string) error {
	return h.app.Run(args)
}

func getMapFromArgs(c *cli.Context) (domain.Map, error) {
	if c.NArg() < 1 {
		return domain.Map{}, fmt.Errorf("the map argument is required")
	}
	m := domain.Map{}
	if err := json.Unmarshal([]byte(c.Args().Get(0)), &m); err != nil {
		return domain.Map{}, err
	}
	return m, nil
}

func printMapsAsJson(maps []domain.Map) error {
	for _, m := range maps {
		jsonData, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonData))
		fmt.Println()
	}
	return nil
}

func printPublishedMapsToTerminal(maps []domain.PublishedMap) error {
	for _, m := range maps {
		jsonData, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonData))
		fmt.Println()
	}
	return nil
}

func printMapsAsTable(maps []domain.Map) error {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)

	_, err := fmt.Fprintln(writer, "Date\tMap Type\tBounds\tSource_MapProvider")
	if err != nil {
		return err
	}

	for _, m := range maps {
		mapTypeStr := m.MapType.String()
		boundsStr := m.Bounds.String()
		_, err := fmt.Fprintf(writer, "%d-%02d-%02d\t%s\t%s\t%s\n", m.Date.Year, m.Date.Month, m.Date.Day, mapTypeStr, boundsStr, m.Source.MapProvider.String())
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
