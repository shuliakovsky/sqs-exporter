# sqs-exporter
AWS SQS exporter for prometheus
```html
exported metrics:
    sqs_message_count - gauge
    sqs_message_age   - gauge
```
#### how to run in container
```shell
edit config.yaml
docker compose up
```
#### how to build

```shell
# example build for apple macOS & Apple Silicon Chip
./build.sh darwin arm64
```
```shell
# example build for apple macOS & Intel Chip
./build.sh darwin amd64
```
```shell
# build with no args for help
./build.sh 
```
```shell
# build with Makefile in docker for apple macOS & Apple Silicon Chip
make docker-build OS=darwin ARCH=arm64
```

####  systemd service example

```unit file (systemd)
[Unit]
  Description=SQS exporter
  Wants=network-online.target
  After=network-online.target

[Service]

  ExecStart=/usr/local/bin/sqs-exporter --config /etc/sqs-exporter/config.yml
  SyslogIdentifier=sqs-exporter
  Restart=always

[Install]
  WantedBy=multi-user.target

```
####  ./config.yml example
```yaml
aws_region: us-east-1                                                             # Default SQS region
listen_ip: 0.0.0.0                                                                # Listening IP address, default 0.0.0.0
port: 9090                                                                        # Listening port, default 9090
queues:
  - url: "https://sqs.us-east-1.amazonaws.com/0123456789012/MyQueue"
  - url: "https://sqs.us-west-2.amazonaws.com/123456789012/MyQueueInUsWest2"
    region: "us-west-2"
```
