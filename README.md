# Minter Sentinel

Essential application for running validator in [Minter](https://minter.org) network. 
It will automatically turn off your masternode if it exceeds missed blocks threshold.

## Support us

If you find this project useful and want to thank us you can donate us some coins:

```
Mx0d13ae2af1548c60a7691ecf47ed680673d2fe2c
```

Or delegate coins to our [FriendsTRUST](https://friendstrust.club) validator:

```
Mpf64c5956078cd8e5db5b3e7ccfc4e2f9b596fa3801eb3dbfef691623df972279
```

## Installation

### From source

```bash
git clone https://github.com/friendstrust/minter-sentinel && \
cd minter-sentinel && \
make build
```

### Docker 

Docker image `friendstrust/minter-sentinel` is available on [Docker Hub](https://hub.docker.com/r/friendstrust/minter-sentinel).

### Binary

Download the latest release version from https://github.com/friendstrust/minter-sentinel/releases

## Configuration

Create config.yaml based on the config.example.yaml file.

## Usage

### Transaction Generator

Before starting watcher, you need to generate transaction to turn off your masternode. 

You can do it using `txgenerate` command:

```bash
./minter-sentinel txgenerate
```

### Watcher

```bash
./minter-sentinel start
```

If you don't want to turn off masternode if missed blocks threshold exceeds add `dry-run` flag to command:

```bash
./minter-sentinel start --dry-run
```

## Prometheus

In addition to the standard Go metrics, custom metrics by the application are exported:

```text
minter_sentinel_missed_blocks_threshold
minter_sentinel_sleep
minter_sentinel_blocks_signed
minter_sentinel_blocks_missed_total
minter_sentinel_blocks_missed_current
```
