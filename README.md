# prometheus-centurylink-c4000xg-exporter

This is a prometheus exporter for the CenturyLink C4000XG, a router provided to most CenturyLink Fiber customers.

## Comparison to prometheus-c4000xg-exporter

`prometheus-c4000xg-exporter` is [a different exporter for the CenturyLink
C4000XG written in Node.js](https://github.com/mariotacke/prometheus-c4000xg-exporter).
At the time of writing this exporter, `prometheus-c4000xg-exporter` only
exported data on LAN traffic, not WLAN.

This exporter is mostly a superset of the data exported from
`prometheus-c4000xg-exporter`, except some of the more obscure metrics are not
spelled exactly the same. The other module was missing underscores between
words if the next word started with a number, while this module includes the
underscore. This is unlikely to affect you, because the metrics in question are
not particularly interesting.

## How to install

Note: The install script is pretty crude, and just copies the data to
`/usr/local` before adding links to systemd. Use at your own risk.

```sh
make
sudo make install
```

## Grafana dashboard

[An example Grafana dashboard](grafana/prometheus-centurylink-c4000xg-exporter.grafana.json)
has been provided to get you started with graphing your network.

## Docker images

No docker images are provided, but considering this is a Go project you could
easily make one and copy the binary into it if you were so inclined.

## Project status

I no longer use this router, so I will be unable to test any incoming patches.
Still feel free to submit them, or to fork the project if you want to.
