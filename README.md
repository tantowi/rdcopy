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
   
   `go build -v`

8. Run the app
   
   `./rdcopy`
   

## Migrate command

### Usage

```
  rdcopy migrate <source> <destination> [flags]
```

### Parameters

*Source* : Redis source

*Destination* : Redis destination

*Source* and *Destination* can be simple `<host>:<port>` or full URL format: `redis://[:<password>@]<host>:<port>[/<dbIndex>]`

### Flags:

```
  -h, --help                    help for migrate
      --logInterval int         Print current status every N seconds (default 1)
      --parallelDumps int       Number of parallel dump goroutines (default 100)
      --parallelRestores int    Number of parallel restore goroutines (default 100)
      --pattern string          Matching pattern for keys (default "*")
      --replaceExistingKeys     Existing keys will be replaced
      --scanCount int           COUNT parameter for redis SCAN command (default 1000)
      --sourcePassword string   Password of source redis
      --targetPassword string   Password of target redis
```

For matching pattern, see [Redis SCAN](https://redis.io/commands/scan) command.

## Delete command

### Usage

```
  rdcopy delete <source> [flags]
```

### Parameter

*Source* : redis source, can be simple `<host>:<port>` or full URL format: `redis://[:<password>@]<host>:<port>[/<dbIndex>]`

### Flags:

```
  -h, --help                  help for delete
      --logInterval int       Log current status every N seconds (default 1)
      --parallelDeletes int   Number of parallel delete goroutines (default 100)
      --password string       Password for redis
      --pattern string        Matching pattern for keys (default "*")
      --scanCount int         COUNT parameter for redis SCAN command (default 1000)
```

For matching pattern, see [Redis SCAN](https://redis.io/commands/scan) command.


## Generate command

### Usage

```
rdcopy generate <source> [flags]
```

### Parameter

*Source* : redis source, can be simple `<host>:<port>` or full URL format: `redis://[:<password>@]<host>:<port>[/<dbIndex>]`

### Flags:

```
      --entryCount int             Iteration count to perform (default 1)
  -h, --help                       help for generate
      --password string            Password for redis
      --prefixAmount stringArray   Amount of keys to create for each prefix in one iteration (default [1,2])
      --prefixes stringArray       List of prefixes for generated keys (default [mykey:,testkey:])
```

## Notes

- Scan: Scanning is performed with a single goroutine, scanned keys are sent to channel

- Dump: X export goroutines are consuming keys and perform `DUMP` and `PTTL` as a pipeline command

- Restore: Results are sent to another channel, where another Y push goroutines are performing `RESTORE`/`REPLACE` command on the destination instance

- Monitor: A goroutine outputs status every T seconds 

