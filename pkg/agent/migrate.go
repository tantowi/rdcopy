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

var migrateCmd = &cobra.Command{
	Use:   "migrate <source> <destination>",
	Short: "Migrate keys from source redis instance to destination by given pattern",
	Long: `Migrate keys from source redis instance to destination by given pattern <source> and <destination> 

Can be provided as just ""<host>:<port>" or in Redis URL format: "redis://[:<password>@]<host>:<port>[/<dbIndex>]"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start migration")
		ctx := context.Background()

		// create redis clients
		scannerClient := createClient(ctx, sourcePassword, args[0], sourceUseTLS)
		dumperClient := createClient(ctx, sourcePassword, args[0], sourceUseTLS)
		restorerClient := createClient(ctx, targetPassword, args[1], targetUseTLS)

		// init core services
		logger := logger.CreateService()
		scanner := scanner.CreateService(
			scannerClient,
			scanner.Options{
				SearchPattern:  pattern,
				RedisScanCount: scanCount,
			},
			logger,
		)
		dumper := dumper.CreateService(
			dumperClient,
			scanner.GetScanChannel(),
			logger,
		)
		restorer := restore.CreateService(restorerClient, dumper.GetDumpChannel(), logger)

		// start processing
		wgRestore := new(sync.WaitGroup)

		logger.Start(time.Second * time.Duration(logInterval))
		restorer.Start(ctx, wgRestore, parallelRestores)
		fmt.Println(123)

		scanner.Start(ctx)
		fmt.Println(123)

		dumper.Start(ctx, parallelDumps)
		fmt.Println(456)

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
	migrateCmd.Flags().BoolVar(&sourceUseTLS, "sourceUseTLS", true, "Enable TLS for source redis- default true")
	migrateCmd.Flags().StringVar(&targetPassword, "targetPassword", "", "Password of target redis")
	migrateCmd.Flags().BoolVar(&targetUseTLS, "targetUseTLS", true, "Enable TLS for target redis- default true")
	migrateCmd.Flags().IntVar(&scanCount, "scanCount", 1000, "COUNT parameter for redis SCAN command")
	migrateCmd.Flags().IntVar(&logInterval, "logInterval", 1, "Print current status every N seconds")
	migrateCmd.Flags().IntVar(&parallelDumps, "parallelDumps", 100, "Number of parallel dump goroutines")
	migrateCmd.Flags().IntVar(&parallelRestores, "parallelRestores", 100, "Number of parallel restore goroutines")
}
