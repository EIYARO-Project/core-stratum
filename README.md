# Core Stratum Solo

A stratum server for `GPU` mining in solo mode.

## Install

### Get the source code

```bash
git clone https://github.com/EIYARO-Project/core-stratum.git
```

### Build source code

```bash
cd core-stratum/stratum/eiyaro/cmd
go build -o eiyaro_stratum
```

## Run

### Configure parameters

```bash
cd core-stratum/stratum/eiyaro/conf
vim prod.yml
```

Set `node.url` with the eiyaro-class node url, then leave other parameters with default value.

### Run It

```bash
cd core-stratum/stratum/eiyaro/cmd
./eiyaro_stratum -config=../conf/prod.yml
```

## Parameter interpretation

```yaml
mode: prod # run mode, defines logger level and so on

# server
stratum.id: 0 # session offset id for different miner
stratum.port: 9119 # miner connection
stratum.max_conn: 32768 # max connection of miner
stratum.default_ban_period: 10m # ban malicious miner, 0s means disable

# session
session_timeout: 5m # connection timeout
session.sched_interval: 0 # work braodcast interval, 0 means braodcast when new work coming
session.diff: 1050000 # diff for miner

# node
node.url: http://127.0.0.1:9888 # eiyaro node url
node.name: eiyaro_mainnet # eiyaronode name, set with default
node.sync_interval: 100ms # interval of getting work from node

service.port: 11002 # gin server port
```
