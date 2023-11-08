package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var keyPrefix, prefixAmount []string
var entryCount int

var generatorCmd = &cobra.Command{
	Use:   "generate <redis>",
	Short: "Create random entries in redis instance",
	Long: `Create random entries in redis instance
Url can be provided as just "<host>:<port>""`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start generating keys")
		ctx := context.Background()

		randomMap, err := createRandomMap(keyPrefix, prefixAmount)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Generated random values: ", args[0])
		generatorClient := createClient(args[0], sourcePassword)

		rand.Seed(time.Now().UTC().UnixNano())
		for j := 0; j < entryCount; j++ {
			for prefix, number := range randomMap {
				for i := 0; i < number; i++ {
					randVal := strconv.Itoa(rand.Int())
					err := generatorClient.Do(ctx, "SET", prefix+randVal, randVal).Err()
					if err != nil {
						fmt.Println(fmt.Errorf("could not create entry: %w", err))
					}
				}
			}

			fmt.Printf("Generation: %d done\n", j)
		}
	},
}

func createRandomMap(prefix []string, prefixAmount []string) (map[string]int, error) {
	randomMap := make(map[string]int)
	for key, val := range prefix {
		randomMap[val] = 1

		if key < len(prefixAmount) {
			// parse to int because int array not possible via cli
			countForPrefix, err := strconv.Atoi(prefixAmount[key])
			if err != nil {
				return nil, err
			}

			if countForPrefix <= 0 {
				return nil, errors.New("count cannot be zero or negative")
			}

			randomMap[val] = countForPrefix
		}
	}

	return randomMap, nil
}

func init() {
	RootCmd.AddCommand(generatorCmd)

	generatorCmd.Flags().StringVar(&sourcePassword, "password", "", "Password for redis")
	generatorCmd.Flags().StringArrayVar(&keyPrefix, "prefixes", []string{"mykey:", "testkey:"}, "List of prefixes for generated keys")
	generatorCmd.Flags().StringArrayVar(&prefixAmount, "prefixAmount", []string{"1", "2"}, "Amount of keys to create for each prefix in one iteration")
	generatorCmd.Flags().IntVar(&entryCount, "entryCount", 1, "Iteration count to perform")

}
