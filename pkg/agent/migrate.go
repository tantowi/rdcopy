package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/appit-online/redis-dumper/pkg/core/dumper"
	"github.com/appit-online/redis-dumper/pkg/core/logger"
	"github.com/appit-online/redis-dumper/pkg/core/restore"
	"github.com/appit-online/redis-dumper/pkg/core/scanner"
	"github.com/spf13/cobra"
)

var parallelDumps, parallelRestores int
var replaceExistingKeys bool

var migrateCmd = &cobra.Command{
	Use:   "migrate <source> <destination>",
	Short: "Migrate keys from source redis instance to destination by given pattern",
	Long: `Migrate keys from source redis instance to destination by given pattern <source> and <destination> 

Can be provided as just ""<host>:<port>""`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start migration")
		ctx := context.Background()

		// create redis clients
		scannerClient := createClient(args[0], sourcePassword)
		defer scannerClient.Close()
		dumperClient := createClient(args[0], sourcePassword)
		defer dumperClient.Close()
		restorerClient := createClient(args[1], targetPassword)
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
		restorer := restore.CreateService(restorerClient, dumper.GetRestorerChannel(), logger, replaceExistingKeys)

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
	migrateCmd.Flags().StringVar(&sourcePassword, "sourcePassword", "", "Password of source redis")
	migrateCmd.Flags().StringVar(&targetPassword, "targetPassword", "", "Password of target redis")
	migrateCmd.Flags().IntVar(&scanCount, "scanCount", 1000, "COUNT parameter for redis SCAN command")
	migrateCmd.Flags().IntVar(&logInterval, "logInterval", 1, "Print current status every N seconds")
	migrateCmd.Flags().IntVar(&parallelDumps, "parallelDumps", 100, "Number of parallel dump goroutines")
	migrateCmd.Flags().IntVar(&parallelRestores, "parallelRestores", 100, "Number of parallel restore goroutines")
	migrateCmd.Flags().BoolVar(&replaceExistingKeys, "replaceExistingKeys", false, "Existing keys will be replaced")
}
