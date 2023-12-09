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
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load()
	modem_host := getenvOrDefault("MODEM_HOST", "192.168.0.1")
	modem_user := getenvOrDefault("MODEM_USER", "admin")
	modem_password := mustGetenv("MODEM_PASSWORD")
	metrics_namespace := getenvOrDefault("METRICS_NAMESPACE", "c4000xg")
	port, err := strconv.Atoi(getenvOrDefault("PORT", "9998"))
	if err != nil {
		panic(fmt.Errorf("Invalid value in $PORT: %w", err))
	}

	log.Print("Starting prometheus-centurylink-c4000xg-exporter")
	exporter := MustNewExporter(modem_host, modem_user, modem_password, metrics_namespace)
	prometheus.MustRegister(exporter)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func getenvOrDefault(key string, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func mustGetenv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Errorf("Environment variable %s must be defined", key))
}
