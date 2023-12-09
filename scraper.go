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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type Scraper struct {
	client *http.Client
	host   string
}

func MustNewScraper(host string, username string, password string) *Scraper {
	return &Scraper{
		client: MustGetAuthenticatedClient(host, username, password),
		host:   host,
	}
}

func (s *Scraper) GetHosts() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?Object=Device.Hosts.Host&PhysAddress=&HostName=&Layer1Interface=&Active=&HostName=&IPAddress=", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get Hosts: %w", err)
	}
	datamap := dataToMap(data, "PhysAddress")
	return datamap, nil
}

func (s *Scraper) MustGetHosts() map[string]map[string]string {
	data, err := s.GetHosts()
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Scraper) GetAccessPoint() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?Object=Device.WiFi.AccessPoint", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get WiFi AccessPoint: %w", err)
	}
	datamap := dataToMap(data, "")
	return datamap, nil
}

func (s *Scraper) MustGetAccessPoint() map[string]map[string]string {
	data, err := s.GetAccessPoint()
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Scraper) GetSSID() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?Object=Device.WiFi.SSID.", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get Wifi SSID: %w", err)
	}
	datamap := dataToMap(data, "")
	return datamap, nil
}

func (s *Scraper) MustGetSSID() map[string]map[string]string {
	data, err := s.GetSSID()
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Scraper) GetRadio() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?Object=Device.WiFi.Radio&Channel=&OperatingFrequencyBand=", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get WiFi Radio: %w", err)
	}
	datamap := dataToMap(data, "")
	return datamap, nil
}

func (s *Scraper) MustGetRadio() map[string]map[string]string {
	data, err := s.GetRadio()
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Scraper) GetEthernet() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?/cgi/cgi_get?Object=Device.Ethernet.", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get Ethernet Interfaces: %w", err)
	}
	datamap := dataToMap(data, "")
	return datamap, nil
}

func (s *Scraper) MustGetEthernet() map[string]map[string]string {
	data, err := s.GetEthernet()
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Scraper) GetTemperatureStatus() (map[string]map[string]string, error) {
	data, err := GetData(s.client, fmt.Sprintf("https://%s/cgi/cgi_get?/cgi/cgi_get?Object=Device.DeviceInfo.TemperatureStatus.TemperatureSensor.", s.host))
	if err != nil {
		return nil, fmt.Errorf("Unable to get Ethernet Interfaces: %w", err)
	}
	datamap := dataToMap(data, "")
	return datamap, nil
}

func (s *Scraper) MustGetTemperatureStatus() map[string]map[string]string {
	data, err := s.GetTemperatureStatus()
	if err != nil {
		panic(err)
	}
	return data
}

type Data struct {
	Objects []Object
}

type Object struct {
	ObjName string
	Param   []Param
}

type Param struct {
	ParamName  string
	ParamValue string
}

func paramsToMap(params *[]Param) map[string]string {
	m := make(map[string]string)
	for _, p := range *params {
		m[p.ParamName] = p.ParamValue
	}
	return m
}

func dataToMap(data *Data, key string) map[string]map[string]string {
	m := make(map[string]map[string]string)
	for _, o := range data.Objects {
		params := paramsToMap(&o.Param)
		var keyValue string
		if key == "" {
			keyValue = o.ObjName
		} else {
			keyValue = params[key]
		}
		m[keyValue] = params
	}
	return m
}

func GetAuthenticatedClient(host string, username string, password string) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to create cookie jar: %w", err)
	}
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Must hand-craft the form, because the C4000XG croaks when url.Values generates the arguments in a different order(!)
	form := fmt.Sprintf("username=%s&password=%s", username, password)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/cgi/cgi_action", host), strings.NewReader(form))
	if err != nil {
		return nil, fmt.Errorf("Unable to make login request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to log in: %w", err)
	}
	return client, nil
}

func MustGetAuthenticatedClient(host string, username string, password string) *http.Client {
	client, err := GetAuthenticatedClient(host, username, password)
	if err != nil {
		panic(err)
	}
	return client
}

func GetData(client *http.Client, url string) (*Data, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to make request for %s: %w", url, err)
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to get %s: %w", url, err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var data Data
	err = dec.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse %s as JSON: %w", url, err)
	}
	return &data, nil

}
