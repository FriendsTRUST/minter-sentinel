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

`docker-compose.yml` example:

```yaml
version: '3'

services:
  app:
    image: friendstrust/minter-sentinel:2
    ports:
      - "127.0.0.1:2112:2112"
    volumes:
      - ./config.yaml:/config.yaml
    restart: unless-stopped
    logging:
      options:
        max-size: "10m"
        max-file: "3"
    deploy:
      resources:
        limits:
          cpus: '0.50'
          memory: 50M
        reservations:
          cpus: '0.25'
          memory: 20M
```

### Binary

Download the latest release version from https://github.com/friendstrust/minter-sentinel/releases

## Configuration

Create config.yaml based on the [config.example.yaml](https://github.com/FriendsTRUST/minter-sentinel/blob/master/config.example.yaml) file.

## Usage

### "Turn off" transaction

In order to turn off validator when the missed blocks threshold exceeds,
you need either generate transaction manually beforehand or define private keys of controlling wallet in configuration file.

Seeds take precedence when turning off masternode, but, if watcher fails to generate transaction, `transaction_off` will be used.

#### Manual

In order to generate transaction, you can use `txgenerate` command:

```bash
./minter-sentinel txgenerate
```

The resulting transaction hash should be set in `transaction_off` parameter in configuration file.

#### Automatic

In order to automatically generate transactions, you need to get seed(s) using `seeds` command:

```bash
./minter-sentinel seeds
```

The resulting seeds paste in `private_keys` parameter as an array in configuration file.

Controlling wallet address will be fetched automatically from the Node API.

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
