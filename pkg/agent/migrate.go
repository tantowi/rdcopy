package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"rdcopy/pkg/core/dumper"
	"rdcopy/pkg/core/logger"
	"rdcopy/pkg/core/restore"
	"rdcopy/pkg/core/scanner"

	"github.com/spf13/cobra"
)

var parallelDumps, parallelRestores int
var overwrite bool

var migrateCmd = &cobra.Command{
	Use:   "migrate <source> <target>",
	Short: "Migrate keys from source instance to target instance by given pattern",
	Long:  "Migrate keys from source instance to target instance by given pattern",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start migration")
		ctx := context.Background()

		// create redis clients
		scannerClient, err := createClient(args[0])
		if err != nil {
			fmt.Println("Error creating scanner client")
			log.Fatal(err)
			return
		}

		defer scannerClient.Close()

		dumperClient, err := createClient(args[0])
		if err != nil {
			fmt.Println("Error creating dumper client")
			log.Fatal(err)
			return
		}

		defer dumperClient.Close()

		restorerClient, err := createClient(args[1])
		if err != nil {
			fmt.Println("Error creating restorer client")
			log.Fatal(err)
			return
		}

		defer restorerClient.Close()

		// init core services
		logger := logger.CreateService()
		scanner := scanner.CreateService(
			scannerClient,
			scanner.Options{
				SearchPattern:  pattern,
				RedisScanCount: scanCount,
				ParallelDumps:  parallelDumps,
			},
			logger,
		)
		dumper := dumper.CreateService(
			dumperClient,
			scanner.GetDumperChannel(),
			logger,
			parallelRestores,
		)
		restorer := restore.CreateService(restorerClient, dumper.GetRestorerChannel(), logger, overwrite)

		// start processing
		wgRestore := new(sync.WaitGroup)

		logger.Start(time.Second * time.Duration(logInterval))
		restorer.Start(ctx, wgRestore, parallelRestores)
		scanner.Start(ctx)
		dumper.Start(ctx, parallelDumps)

		// wait until all channels are closed
		wgRestore.Wait()
		logger.Stop()
		logger.Report()

		fmt.Println("Finish migration")
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVar(&pattern, "pattern", "*", "Matching pattern for keys")
	migrateCmd.Flags().IntVar(&scanCount, "scanCount", 1000, "COUNT parameter for redis SCAN command")
	migrateCmd.Flags().IntVar(&logInterval, "logInterval", 1, "Print current status every N seconds")
	migrateCmd.Flags().IntVar(&parallelDumps, "parallelDumps", 10, "Number of parallel dump goroutines")
	migrateCmd.Flags().IntVar(&parallelRestores, "parallelRestores", 10, "Number of parallel restore goroutines")
	migrateCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing keys in target instance")
}
