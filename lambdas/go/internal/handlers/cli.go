package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	conflict_nightlightv1 "github.com/BaronBonet/conflict-nightlight/generated/conflict_nightlight/v1"
	"github.com/BaronBonet/conflict-nightlight/internal/core/domain"
	"github.com/BaronBonet/conflict-nightlight/internal/core/ports"
	"github.com/BaronBonet/conflict-nightlight/internal/infrastructure/prototransformers"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

type CliHandler struct {
	app *cli.App
}

func NewCLIHandler(ctx context.Context, productService ports.OrchestratorService) *CliHandler {
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
					}
					return printMapsAsTable(maps)
				},
			},
			{
				Name:  "listProcessedMaps",
				Usage: "List all processed maps",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "json",
					},
				},
				Action: func(c *cli.Context) error {
					maps, err := productService.ListProcessedInternalMaps(ctx)
					if err != nil {
						return err
					}
					if c.String("json") == "true" {
						return printMapsAsJson(maps)
					}
					return printMapsAsTable(maps)
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
				Usage:     "publish a map, pass in a map that is copied from the output of listProcessedMaps --json, i.e. you need to use the json flag and copy the whole object you want to publish.",
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
				Name:      "publishMaps",
				Usage:     "publish a list of maps, pass in an array of maps that is copied from the output of listProcessedMaps --json, i.e. you need to use the json flag and copy all the objects you want to publish and wrap them in brackets [].",
				ArgsUsage: "[maps]",
				Action: func(c *cli.Context) error {
					maps, err := getMapsFromArgs(c)
					if err != nil {
						return err
					}

					errChan := make(chan error, len(maps))
					var wg sync.WaitGroup

					for _, m := range maps {
						wg.Add(1)
						go func(m domain.Map) {
							defer wg.Done()
							if err := productService.PublishMap(ctx, m); err != nil {
								errChan <- fmt.Errorf("failed to publish map %v: %w", m, err)
							}
						}(m)
					}

					wg.Wait()
					close(errChan)

					var errors []error
					for err := range errChan {
						if err != nil {
							errors = append(errors, err)
						}
					}

					if len(errors) > 0 {
						// Combine all errors into a single error message
						errorMsg := fmt.Sprintf("Failed to publish some maps (%d errors):\n", len(errors))
						for i, err := range errors {
							errorMsg += fmt.Sprintf("Error %d: %v\n", i+1, err)
						}
						return fmt.Errorf(errorMsg)
					}

					fmt.Printf("Successfully published %d maps\n", len(maps))
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
				Usage:     "Runs the sync map request which will subsequently put messages on sqs and invoke the entire pipeline",
				ArgsUsage: "[syncMapRequest]",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("the syncMapRequest argument is required")
					}
					syncMapRequest := conflict_nightlightv1.SyncMapRequest{}
					if err := protojson.Unmarshal([]byte(c.Args().Get(0)), &syncMapRequest); err != nil {
						return err
					}
					_, err := productService.SyncInternalWithExternalMaps(
						ctx,
						prototransformers.ProtoToSyncMapsRequest(&syncMapRequest),
					)
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
func getMapsFromArgs(c *cli.Context) ([]domain.Map, error) {
	if c.NArg() < 1 {
		return nil, fmt.Errorf("the maps argument is required")
	}
	var maps []domain.Map
	if err := json.Unmarshal([]byte(c.Args().Get(0)), &maps); err != nil {
		return nil, fmt.Errorf("failed to parse maps json: %w", err)
	}
	return maps, nil
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

func printSliceAsJson[T any](items []T) error {
	jsonData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func printMapsAsJson(maps []domain.Map) error {
	return printSliceAsJson(maps)
}

func printPublishedMapsToTerminal(maps []domain.PublishedMap) error {
	return printSliceAsJson(maps)
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
		_, err := fmt.Fprintf(
			writer,
			"%d-%02d-%02d\t%s\t%s\t%s\n",
			m.Date.Year,
			m.Date.Month,
			m.Date.Day,
			mapTypeStr,
			boundsStr,
			m.Source.MapProvider.String(),
		)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
