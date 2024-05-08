# influxdb-alerter

Sends an alert when there are no new values posted to a given measurement in InfluxDB in a given time.

## Usage

First, create a copy of the example config:

```shell
cp example-config.yaml config.yaml
```

Fill in the desired alert threshold age (time since last measurement, config value `--max_age`) and your InfluxDB credentials. If you want email notifications, also fill out your SMTP credentials. Place the config file either in the project directory, `$HOME/.influxdb-alerter/config.yaml` or `/etc/influxdb-alerter/config.yaml`. Or you can specify the config file location with the `--config <path>` CLI argument.

You can run the alerter either via Docker:

```shell
./build.sh
./run.sh
```

or directly with:
```
go run main.go check
```

In case of an InfluxDB connection error, the program will exit with status code 1, otherwise it will exit with code 0 and send the configured notifications.
