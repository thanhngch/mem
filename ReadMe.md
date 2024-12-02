## Show memory group by application in MacOS

## How it work
Call this command `ps x -o %cpu,rss,command -m -A` then group by `.app` name. 

## Build
```sh
go build -o ~/go/bin/mem
```

## Using
```sh
mem
# or
mem -n <SHOW_NUM_PROCESS> -p <SHOW_COUNT_PROCESS>
```
which setting
```go
const SHOW_NUM_PROCESS = 25
const SHOW_COUNT_PROCESS = 5
const GROUP_SYSTEM_APP = true
```

Note: when `GROUP_SYSTEM_APP` is `true` The command has prefix `/System/Library/`, `/usr/libexec/`, `/usr/sbin/` and `/Library/` is `System` app.

It will format 
```
%CPU      RSS COMMAND
0.7     2.9GB Google Chrome (24)
```

Command is application name (`<total process>`) or only application name.
It will show (`<total process>`) only `<total process>` is greater than `SHOW_COUNT_PROCESS`.
