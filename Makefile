prometheus-centurylink-c4000xg-exporter: exporter.go main.go scraper.go go.sum Makefile
	go build

.PHONY: build
build: prometheus-centurylink-c4000xg-exporter

.PHONY: clean
clean:
	rm prometheus-centurylink-c4000xg-exporter

.PHONY: install
install:
	install -m 0755 -d /usr/local/bin
	install -m 0755 prometheus-centurylink-c4000xg-exporter /usr/local/bin/
	install -m 0755 -d /usr/local/lib/systemd
	install -m 0755 prometheus-centurylink-c4000xg-exporter.service /usr/local/lib/systemd/
	install -m 0755 -d /etc/prometheus
	[ -e /etc/prometheus/prometheus-centurylink-c4000xg-exporter.conf ] || install -m 0755 prometheus-centurylink-c4000xg-exporter.conf.example /etc/prometheus/prometheus-centurylink-c4000xg-exporter.conf
	systemctl link /usr/local/lib/systemd/prometheus-centurylink-c4000xg-exporter.service

.PHONY: uninstall
uninstall:
	systemctl stop prometheus-centurylink-c4000xg-exporter.service || true
	systemctl disable prometheus-centurylink-c4000xg-exporter.service || true
	systemctl unlink prometheus-centurylink-c4000xg-exporter.service || true
	rm /usr/local/bin/prometheus-centurylink-c4000xg-exporter
	rm /usr/local/lib/systemd/prometheus-centurylink-c4000xg-exporter.service

.PHONY: run
run:
	go run .

.PHONY: mod
mod:
	go mod tidy
