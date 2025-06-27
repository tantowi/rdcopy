package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"rdcopy/pkg/core/deleter"
	"rdcopy/pkg/core/logger"
	"rdcopy/pkg/core/scanner"

	"github.com/spf13/cobra"
)

var parallelDeletes int

var deleteCmd = &cobra.Command{
	Use:   "delete <source>",
	Short: "Delete keys from redis instance by given pattern",
	Long:  "Delete keys from redis instance by given pattern",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start deleting keys")
		ctx := context.Background()

		// create redis clients
		scannerClient, err := createClient(args[0])
		if err != nil {
			fmt.Println("Error creating scanner client")
			log.Fatal(err)
			return
		}

		defer scannerClient.Close()

		deleterClient, err := createClient(args[0])
		if err != nil {
			fmt.Println("Error creating deleter client")
			log.Fatal(err)
			return
		}

		defer deleterClient.Close()

		// init core services
		logger := logger.CreateService()
		scanner := scanner.CreateService(
			scannerClient,
			scanner.Options{
				SearchPattern:  pattern,
				RedisScanCount: scanCount,
				ParallelDumps:  parallelDeletes,
			},
			logger,
		)
		deleter := deleter.CreateService(
			deleterClient,
			scanner.GetDumperChannel(),
			logger,
		)

		// start processing
		logger.Start(time.Second * time.Duration(logInterval))
		scanner.Start(ctx)
		deleter.Start(ctx, parallelDeletes)

		logger.Stop()
		logger.Report()

		fmt.Println("Finish deleting keys")
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVar(&pattern, "pattern", "*", "Matching pattern for keys")
	deleteCmd.Flags().IntVar(&scanCount, "scanCount", 1000, "COUNT parameter for redis SCAN command")
	deleteCmd.Flags().IntVar(&logInterval, "logInterval", 1, "Log current status every N seconds")
	deleteCmd.Flags().IntVar(&parallelDeletes, "parallelDeletes", 10, "Number of parallel delete goroutines")
}
