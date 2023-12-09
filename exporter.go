// Copyright 2022 Knut Grythe <knut@auvor.no>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	host          string
	username      string
	password      string
	metricDescs   map[string]*prometheus.Desc
	nameConverter *MetricNameConverter
}

func MustNewExporter(host string, username string, password string, namespace string) *Exporter {
	return &Exporter{
		host:          host,
		username:      username,
		password:      password,
		metricDescs:   make(map[string]*prometheus.Desc),
		nameConverter: MustNewMetricNameConverter(namespace),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// Descriptions are generated dynamically by Collect(),
	// just make sure we called it once since starting up
	if len(e.metricDescs) == 0 {
		//	e.Collect(make(chan<- prometheus.Metric))
	}
	for _, desc := range e.metricDescs {
		ch <- desc
	}
}

func (e *Exporter) GetDesc(prefix string, metric string, description string, tagKeys []string) *prometheus.Desc {
	desc, ok := e.metricDescs[metric]
	if !ok {
		desc = prometheus.NewDesc(e.nameConverter.Convert(prefix, metric), description, tagKeys, nil)
	}
	return desc
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	scraper := MustNewScraper(e.host, e.username, e.password)
	hosts := scraper.MustGetHosts()
	aps := scraper.MustGetAccessPoint()
	ssids := scraper.MustGetSSID()
	clientMatcher := regexp.MustCompile("\\.AssociatedDevice.\\d+$")

	macToAssociatedDevice := make(map[string]map[string]string)
	tagKeys := []string{
		"mac_address",
		"hostname",
		"ssid",
	}
	for key, associatedDevice := range aps {
		if clientMatcher.MatchString(key) && associatedDevice["Active"] == "true" {
			macAddress := associatedDevice["MACAddress"]
			macToAssociatedDevice[macAddress] = associatedDevice
			hostInfo := hosts[macAddress]
			ssidInfo := ssids[hostInfo["Layer1Interface"]]

			tagValues := []string{
				macAddress,
				hostInfo["HostName"],
				ssidInfo["SSID"],
			}

			stats := aps[key+".Stats"]

			e.CollectMetrics(ch, "client_", associatedDevice, tagKeys, tagValues)
			e.CollectMetrics(ch, "client_", stats, tagKeys, tagValues)
		}
	}

	e.CollectHostInfo(ch, hosts, macToAssociatedDevice, ssids, scraper.MustGetRadio())
	e.CollectNetworkMetrics(ch, scraper.MustGetEthernet())
	e.CollectTemperatureMetrics(ch, scraper.MustGetTemperatureStatus())
}

func (e *Exporter) CollectHostInfo(ch chan<- prometheus.Metric,
	hosts map[string]map[string]string,
	macToAssociatedDevice map[string]map[string]string,
	ssids map[string]map[string]string,
	radio map[string]map[string]string,
) {
	tagKeys := []string{
		"mac_address",
		"ip",
		"hostname",
		"ssid",
		"frequency_band",
		"wifi_standard",
		"vendor",
	}
	for _, hostInfo := range hosts {
		if hostInfo["Active"] != "1" {
			continue
		}
		macAddress := hostInfo["PhysAddress"]
		ssidInfo := ssids[hostInfo["Layer1Interface"]]
		associatedDevice := macToAssociatedDevice[macAddress]
		ssid, ok := ssidInfo["SSID"]
		if !ok {
			ssid = strings.Replace(hostInfo["Layer1Interface"], "Device.Ethernet.Interface.", "Ethernet ", 1)
		}
		tagValues := []string{
			macAddress,
			hostInfo["IPAddress"],
			hostInfo["HostName"],
			ssid,
			radio[removeTrailingDot(ssidInfo["LowerLayers"])]["OperatingFrequencyBand"],
			associatedDevice["OperatingStandard"],
			associatedDevice["X_GWS_VendorId"],
		}
		e.CollectMetrics(ch, "client_", map[string]string{"info": "1"}, tagKeys, tagValues)
	}
}

func (e *Exporter) CollectTemperatureMetrics(ch chan<- prometheus.Metric, metrics map[string]map[string]string) {
	tagKeys := []string{"name", "alias"}
	metric := "temperature"
	for _, entry := range metrics {
		floatValue, err := strconv.ParseFloat(entry["Value"], 64)
		if err == nil && entry["Enable"] == "true" {
			ch <- prometheus.MustNewConstMetric(
				e.GetDesc("", metric, "TemperatureSensor (number)", tagKeys),
				prometheus.GaugeValue,
				floatValue,
				entry["Name"], entry["Alias"])
		}
	}
}

func (e *Exporter) CollectNetworkMetrics(ch chan<- prometheus.Metric, metrics map[string]map[string]string) {
	statsMatcher := regexp.MustCompile("^Device\\.([^.]+)\\.([^.]+)\\.\\d+$")
	tagKeys := []string{"type", "name", "alias", "mac_address", "ssid"}
	for key, entry := range metrics {
		match := statsMatcher.FindStringSubmatch(key)
		if match != nil && entry["Enable"] == "true" {
			kind := fmt.Sprintf("%s_%s", strings.ToLower(match[1]), strings.ToLower(match[2]))
			tagValues := []string{kind, entry["Name"], entry["Alias"], entry["MACAddress"], entry["SSID"]}
			stats := metrics[key+".Stats"]
			e.CollectMetrics(ch, "", entry, tagKeys, tagValues)
			e.CollectMetrics(ch, "", stats, tagKeys, tagValues)
		}
	}
}

func (e *Exporter) CollectMetrics(ch chan<- prometheus.Metric, prefix string, stats map[string]string, tagKeys []string, tagValues []string) {
	for metric, value := range stats {
		if strings.HasSuffix(metric, "ID") {
			continue
		}
		if strings.HasSuffix(value, "MHz") {
			value = value[:len(value)-3]
		}
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			e.GetDesc(prefix, metric, metric, tagKeys),
			GetMetricType(metric),
			floatValue,
			tagValues...)
	}
}

func GetMetricType(metric string) prometheus.ValueType {
	counters := []string{"Failures", "Total", "Received", "Sent", "Time", "Count"}
	for _, suffix := range counters {
		if strings.HasSuffix(metric, suffix) {
			return prometheus.CounterValue
		}
	}
	return prometheus.GaugeValue
}

type MetricNameConverter struct {
	namespace string
	matchers  []*regexp.Regexp
}

func MustNewMetricNameConverter(namespace string) *MetricNameConverter {
	return &MetricNameConverter{
		namespace: namespace,
		matchers: []*regexp.Regexp{
			regexp.MustCompile("([^_A-Z])([A-Z])"),
			regexp.MustCompile("([A-Z]+)([A-Z][a-z])"),
			regexp.MustCompile("([A-Za-z])([0-9]+)")},
	}
}

func (c *MetricNameConverter) Convert(prefix string, s string) string {
	snaked := s
	for _, matcher := range c.matchers {
		snaked = matcher.ReplaceAllString(snaked, "${1}_${2}")
	}
	return c.namespace + "_" + prefix + strings.ToLower(snaked)
}

func removeTrailingDot(s string) string {
	if len(s) > 0 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}
