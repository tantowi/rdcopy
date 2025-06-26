# rdcopy

Command Line App to copy or move redis keys from one instance to another

Process thousand keys per minute using parallel processing via go routines


## Installation

1. Make sure `golang` is installed

2. Clone the repository
   
   `git clone https://github.com/tantowi/rdcopy rdcopy`

4. Change to the directory
   
   `cd rdcopy`

6. Compile
   
   `go build -v

8. Run the app
   
   `./rdcopy
   

## Migrate command
```
rdcopy migrate <source> <destination> --pattern="*" --sourcePassword="SourcePassword" --targetPassword="TargetPassword"
```

*Source*, *destination* - can be simple `<host>:<port>` or full URL format: `redis://[:<password>@]<host>:<port>[/<dbIndex>]`

*Pattern* - can be glob-style pattern supported by [Redis SCAN](https://redis.io/commands/scan) command.

#### Other flags:

```
  --logInterval int     "Print current status every N seconds" (default 1)
  --scanCount int       "COUNT parameter for redis SCAN command" (default 1000)
  --parallelDumps int   "Number of parallel dump goroutines" (default 100)
  --pushRoutines int    "Number of parallel restore goroutines" (default 100)
  --replaceExistingKeys bool    "Existing keys will be replaced" (default false)
```

## Delete command

```
rdcopy delete <source> --pattern="prefix:*" --password="Password" 
```

#### Other flags:
```
  --logInterval int       "Print current status every N seconds" (default 1)
  --scanCount int         "COUNT parameter for redis SCAN command" (default 1000)
  --parallelDeletes int   "Number of parallel delete goroutines" (default 100)
```

### Generate command

```
rdcopy generate <source> --password="Password" 
```

#### Other flags:
```
  --prefixes []string       "List of prefixes for generated keys" (default {"mykey:", "testkey:"})
  --prefixAmount []string   "Amount of keys to create for each prefix in one iteration" (default {"1", "2"})
  --entryCount int          "Iteration count to perform" (default 1)
```

## Notes

- Scan: Scanning is performed with a single goroutine, scanned keys are sent to channel

- Dump: X export goroutines are consuming keys and perform `DUMP` and `PTTL` as a pipeline command

- Restore: Results are sent to another channel, where another Y push goroutines are performing `RESTORE`/`REPLACE` command on the destination instance

- Monitor: A goroutine outputs status every T seconds 

