// Copyright 2016 Sisa-Tech Pty Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package configEditor

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sisatech/vcli/editor"
	"github.com/sisatech/vcli/shared"
)

type Config struct {
	Metadata        Metadata     `yaml:"metadata" json:"metadata" nav:"metadata"`
	AppSettings     AppSettings  `yaml:"appsettings" json:"appsettings" nav:"app settings"`
	NetworkSettings Network      `yaml:"networksettings" json:"networksettings" nav:"network settings"`
	DiskSettings    Disk         `yaml:"disksettings" json:"disksettings" nav:"disk settings"`
	TimeServers     TimeServers  `yaml:"serversettings" json:"serversettings" nav:"server settings"`
	FileForwards    FileForwards `yaml:"fileforwarding" json:"fileforwarding" nav:"file forwarding"`
	// SaveAndExport   bool
	// arg             string
}

type Metadata struct {
	Name        string `yaml:"name" json:"name" nav:"name"`
	Description string `yaml:"description" json:"name" nav:"description"`
	Author      string `yaml:"author" json:"author" nav:"author"`
	// Release Date
	Version string `yaml:"version" json:"version" nav:"version"`
	AppURL  string `yaml:"appurl" json:"appurl" nav:"app url"`
}

type AppSettings struct {
	// BinaryType string
	BinaryArgs string `yaml:"binaryargs" json:"binaryargs" nav:"binary args"`
	SystemEnvs string `yaml:"systemenvs" json:"systemenvs" nav:"system envs"`
}

type Network struct {
	DNS          DNS          `yaml:"dns" json:"dns" nav:"dns"`
	NetworkCards NetworkCards `yaml:"networkcards" json:"networkcards" nav:"network cards"`
}

type DNS struct {
	DNS0 string `yaml:"dns0" json:"dns0" nav:"dns 0"`
	DNS1 string `yaml:"dns1" json:"dns1" nav:"dns 1"`
	DNS2 string `yaml:"dns2" json:"dns2" nav:"dns 2"`
	DNS3 string `yaml:"dns3" json:"dns3" nav:"dns 3"`
}

type NetworkCards struct {
	Card0 Card `yaml:"card0" json:"card0" nav:"card 0"`
	Card1 Card `yaml:"card1" json:"card1" nav:"card 1"`
	Card2 Card `yaml:"card2" json:"card2" nav:"card 2"`
	Card3 Card `yaml:"card3" json:"card3" nav:"card 3"`
}

type Card struct {
	IP      string `yaml:"ip" json:"ip" nav:"ip address"`
	Mask    string `yaml:"mask" json:"mask" nav:"mask"`
	Gateway string `yaml:"gateway" json:"gateway" nav:"gateway"`
}

type Disk struct {
	FileSystem string `yaml:"filesystem" json:"filesystem" nav:"file system"`
	MaxFD      int    `yaml:"maxfd" json:"maxfd" nav:"max fd"`
	DiskSize   int    `yaml:"disksize" json:"disksize" nav:"disk size"`
}

type TimeServers struct {
	Hostname string `nav:"hostname"`
	Server0  string `nav:"server 0"`
	Server1  string `nav:"server 1"`
	Server2  string `nav:"server 2"`
	Server3  string `nav:"server 3"`
}

type FileForwards struct {
	Rule0 shared.Redirect `nav:"rule 0"`
	Rule1 shared.Redirect `nav:"rule 1"`
	Rule2 shared.Redirect `nav:"rule 2"`
	Rule3 shared.Redirect `nav:"rule 3"`
}

var config Config
var ErrQ = errors.New("user manually terminated the editor, changes have not been saved")

func readFile(file *shared.BuildConfig) error {
	// Metadata
	config.Metadata.Name = file.Name
	config.Metadata.Description = file.Description
	config.Metadata.Author = file.Author
	config.Metadata.Version = file.Version
	config.Metadata.AppURL = file.AppURL

	// App Settings
	for i := 0; i < len(file.App.BinaryArgs); i++ {
		if i == 0 || i == len(file.App.BinaryArgs)-1 {
			config.AppSettings.BinaryArgs += file.App.BinaryArgs[i] + " "
		} else {
			config.AppSettings.BinaryArgs += file.App.BinaryArgs[i] + " "
		}
	}
	for i := 0; i < len(file.App.SystemEnvs); i++ {
		if i == 0 || i == len(file.App.SystemEnvs)-1 {
			config.AppSettings.SystemEnvs += file.App.SystemEnvs[i] + " "
		} else {
			config.AppSettings.SystemEnvs += file.App.SystemEnvs[i] + ":"
		}
	}

	// Disk Settings
	config.DiskSettings.FileSystem = file.Disk.FileSystem
	config.DiskSettings.MaxFD = file.Disk.MaxFD
	config.DiskSettings.DiskSize = file.Disk.DiskSize

	// Network Settings
	// DNS
	for i := 0; i < len(file.Network.DNS); i++ {
		switch i {
		case 0:
			config.NetworkSettings.DNS.DNS0 = file.Network.DNS[i]
		case 1:
			config.NetworkSettings.DNS.DNS1 = file.Network.DNS[i]
		case 2:
			config.NetworkSettings.DNS.DNS2 = file.Network.DNS[i]
		case 3:
			config.NetworkSettings.DNS.DNS3 = file.Network.DNS[i]
		}
	}
	// Network Cards
	for i := 0; i < len(file.Network.NetworkCards); i++ {
		switch i {
		case 0:
			config.NetworkSettings.NetworkCards.Card0.IP = file.Network.NetworkCards[i].IP
			config.NetworkSettings.NetworkCards.Card0.Mask = file.Network.NetworkCards[i].Mask
			config.NetworkSettings.NetworkCards.Card0.Gateway = file.Network.NetworkCards[i].Gateway
		case 1:
			config.NetworkSettings.NetworkCards.Card1.IP = file.Network.NetworkCards[i].IP
			config.NetworkSettings.NetworkCards.Card1.Mask = file.Network.NetworkCards[i].Mask
			config.NetworkSettings.NetworkCards.Card1.Gateway = file.Network.NetworkCards[i].Gateway
		case 2:
			config.NetworkSettings.NetworkCards.Card2.IP = file.Network.NetworkCards[i].IP
			config.NetworkSettings.NetworkCards.Card2.Mask = file.Network.NetworkCards[i].Mask
			config.NetworkSettings.NetworkCards.Card2.Gateway = file.Network.NetworkCards[i].Gateway
		case 3:
			config.NetworkSettings.NetworkCards.Card3.IP = file.Network.NetworkCards[i].IP
			config.NetworkSettings.NetworkCards.Card3.Mask = file.Network.NetworkCards[i].Mask
			config.NetworkSettings.NetworkCards.Card3.Gateway = file.Network.NetworkCards[i].Gateway
		}
	}

	// TimeServers ...
	config.TimeServers.Hostname = file.NTP.Hostname
	for i := 0; i < len(file.NTP.Servers); i++ {
		switch i {
		case 0:
			config.TimeServers.Server0 = file.NTP.Servers[i]
		case 1:
			config.TimeServers.Server1 = file.NTP.Servers[i]
		case 2:
			config.TimeServers.Server2 = file.NTP.Servers[i]
		case 3:
			config.TimeServers.Server3 = file.NTP.Servers[i]
		}
	}

	// File Forwarding Rules
	for i := 0; i < len(file.Redirects.Rules); i++ {
		switch i {
		case 0:
			config.FileForwards.Rule0 = file.Redirects.Rules[i]
		case 1:
			config.FileForwards.Rule1 = file.Redirects.Rules[i]
		case 2:
			config.FileForwards.Rule2 = file.Redirects.Rules[i]
		case 3:
			config.FileForwards.Rule3 = file.Redirects.Rules[i]
		}
	}

	return nil
}

func Edit(file *shared.BuildConfig, isNew bool) error {
	readFile(file)
	ed, err := editor.New(config)
	if err != nil {
		// fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer ed.Cleanup()

	ed.Title("VCLI Configuration Client")

	// Top-Level Callbacks
	ed.DisplayCallback(metadataDisplay, 0)
	ed.DisplayCallback(appSettingsDisplay, 1)
	ed.DisplayCallback(netSettingsDisplay, 2)
	ed.DisplayCallback(diskSettingsDisplay, 3)
	ed.DisplayCallback(timeserversDisplay, 4)
	ed.DisplayCallback(ffDisplay, 5)
	// ed.DisplayCallback(saveDisplay, 4)
	// ed.EditCallback(saveEdit, "", 4)

	// Metadata Callbacks
	ed.DisplayCallback(nameDisplay, 0, 0)
	ed.DisplayCallback(descDisplay, 0, 1)
	ed.DisplayCallback(authorDisplay, 0, 2)
	ed.DisplayCallback(versionDisplay, 0, 3)
	ed.DisplayCallback(appurlDisplay, 0, 4)
	ed.EditCallback(nameEdit, "", config.Metadata.Name, 0, 0)
	ed.EditCallback(descEdit, "", config.Metadata.Description, 0, 1)
	ed.EditCallback(authorEdit, "", config.Metadata.Author, 0, 2)
	ed.EditCallback(versionEdit, "", config.Metadata.Version, 0, 3)
	ed.EditCallback(urlEdit, "", config.Metadata.AppURL, 0, 4)

	// App Settings Callbacks
	// ed.DisplayCallback(bTypeDisplay, 1, 0)
	ed.DisplayCallback(bArgsDisplay, 1, 0)
	ed.DisplayCallback(sysEnvsDisplay, 1, 1)
	// ed.EditCallback(bTypeEdit, 1, 0)
	ed.EditCallback(bArgsEdit, "Each individual argument must be separated by a space.", strings.TrimSpace(config.AppSettings.BinaryArgs), 1, 0)
	ed.EditCallback(sysEnvsEdit, "Each individual argument must be in the format of key=value, separated by a space.", strings.TrimSpace(config.AppSettings.SystemEnvs), 1, 1)

	// Network Settings Callbacks
	ed.DisplayCallback(dnsDisplay, 2, 0)
	ed.DisplayCallback(netCardDisplay, 2, 1)

	// DNS Callbacks
	ed.DisplayCallback(dns0display, 2, 0, 0)
	ed.DisplayCallback(dns1display, 2, 0, 1)
	ed.DisplayCallback(dns2display, 2, 0, 2)
	ed.DisplayCallback(dns3display, 2, 0, 3)
	ed.EditCallback(dns0Edit, "", config.NetworkSettings.DNS.DNS0, 2, 0, 0)
	ed.EditCallback(dns1Edit, "", config.NetworkSettings.DNS.DNS1, 2, 0, 1)
	ed.EditCallback(dns2Edit, "", config.NetworkSettings.DNS.DNS2, 2, 0, 2)
	ed.EditCallback(dns3Edit, "", config.NetworkSettings.DNS.DNS3, 2, 0, 3)

	// Network Card Callbacks
	ed.DisplayCallback(card0display, 2, 1, 0)
	ed.DisplayCallback(card1display, 2, 1, 1)
	ed.DisplayCallback(card2display, 2, 1, 2)
	ed.DisplayCallback(card3display, 2, 1, 3)

	ed.DisplayCallback(card0IPdisplay, 2, 1, 0, 0)
	ed.DisplayCallback(card0Maskdisplay, 2, 1, 0, 1)
	ed.DisplayCallback(card0Gatewaydisplay, 2, 1, 0, 2)
	ed.EditCallback(card0IPEdit, "", config.NetworkSettings.NetworkCards.Card0.IP, 2, 1, 0, 0)
	ed.EditCallback(card0MaskEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card0.Mask, 2, 1, 0, 1)
	ed.EditCallback(card0GatewayEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card0.Gateway, 2, 1, 0, 2)

	ed.DisplayCallback(card1IPdisplay, 2, 1, 1, 0)
	ed.DisplayCallback(card1Maskdisplay, 2, 1, 1, 1)
	ed.DisplayCallback(card1Gatewaydisplay, 2, 1, 1, 2)
	ed.EditCallback(card1IPEdit, "", config.NetworkSettings.NetworkCards.Card1.IP, 2, 1, 1, 0)
	ed.EditCallback(card1MaskEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card1.Mask, 2, 1, 1, 1)
	ed.EditCallback(card1GatewayEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card1.Gateway, 2, 1, 1, 2)

	ed.DisplayCallback(card2IPdisplay, 2, 1, 2, 0)
	ed.DisplayCallback(card2Maskdisplay, 2, 1, 2, 1)
	ed.DisplayCallback(card2Gatewaydisplay, 2, 1, 2, 2)
	ed.EditCallback(card2IPEdit, "", config.NetworkSettings.NetworkCards.Card2.IP, 2, 1, 2, 0)
	ed.EditCallback(card2MaskEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card2.Mask, 2, 1, 2, 1)
	ed.EditCallback(card2GatewayEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card2.Gateway, 2, 1, 2, 2)

	ed.DisplayCallback(card3IPdisplay, 2, 1, 3, 0)
	ed.DisplayCallback(card3Maskdisplay, 2, 1, 3, 1)
	ed.DisplayCallback(card3Gatewaydisplay, 2, 1, 3, 2)
	ed.EditCallback(card3IPEdit, "", config.NetworkSettings.NetworkCards.Card3.IP, 2, 1, 3, 0)
	ed.EditCallback(card3MaskEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card3.Mask, 2, 1, 3, 1)
	ed.EditCallback(card3GatewayEdit, "If IP is 'DCHP', this field can be ignored.", config.NetworkSettings.NetworkCards.Card3.Gateway, 2, 1, 3, 2)

	// Disk Settings Callbacks
	ed.DisplayCallback(filesystemDisplay, 3, 0)
	ed.DisplayCallback(maxfdDisplay, 3, 1)
	ed.DisplayCallback(disksizeDisplay, 3, 2)
	ed.EditCallback(filesystemEdit, "Options include: ext2", config.DiskSettings.FileSystem, 3, 0)
	ed.EditCallback(maxfdEdit, "Requires integer.", strconv.Itoa(config.DiskSettings.MaxFD), 3, 1)
	ed.EditCallback(disksizeEdit, "Disk size in MB", strconv.Itoa(config.DiskSettings.DiskSize), 3, 2)

	ed.DisplayCallback(tsHostDisplay, 4, 0)
	ed.DisplayCallback(tsS0Display, 4, 1)
	ed.DisplayCallback(tsS1Display, 4, 2)
	ed.DisplayCallback(tsS2Display, 4, 3)
	ed.DisplayCallback(tsS3Display, 4, 4)
	ed.EditCallback(tsHostEdit, "", config.TimeServers.Hostname, 4, 0)
	ed.EditCallback(tsS0Edit, "", config.TimeServers.Server0, 4, 1)
	ed.EditCallback(tsS1Edit, "", config.TimeServers.Server1, 4, 2)
	ed.EditCallback(tsS2Edit, "", config.TimeServers.Server2, 4, 3)
	ed.EditCallback(tsS3Edit, "", config.TimeServers.Server3, 4, 4)

	ed.DisplayCallback(fr00Display, 5, 0)
	ed.DisplayCallback(fr01Display, 5, 1)
	ed.DisplayCallback(fr02Display, 5, 2)
	ed.DisplayCallback(fr03Display, 5, 3)
	ed.DisplayCallback(fr00srcDisplay, 5, 0, 0)
	ed.DisplayCallback(fr00destDisplay, 5, 0, 1)
	ed.DisplayCallback(fr00proDisplay, 5, 0, 2)
	ed.DisplayCallback(fr01srcDisplay, 5, 1, 0)
	ed.DisplayCallback(fr01destDisplay, 5, 1, 1)
	ed.DisplayCallback(fr01proDisplay, 5, 1, 2)
	ed.DisplayCallback(fr02srcDisplay, 5, 2, 0)
	ed.DisplayCallback(fr02destDisplay, 5, 2, 1)
	ed.DisplayCallback(fr02proDisplay, 5, 2, 2)
	ed.DisplayCallback(fr03srcDisplay, 5, 3, 0)
	ed.DisplayCallback(fr03destDisplay, 5, 3, 1)
	ed.DisplayCallback(fr03proDisplay, 5, 3, 2)
	ed.EditCallback(fr00srcEdit, "", config.FileForwards.Rule0.Src, 5, 0, 0)
	ed.EditCallback(fr00destEdit, "", config.FileForwards.Rule0.Dest, 5, 0, 1)
	ed.EditCallback(fr00proEdit, "", config.FileForwards.Rule0.Protocol, 5, 0, 2)
	ed.EditCallback(fr01srcEdit, "", config.FileForwards.Rule0.Src, 5, 1, 0)
	ed.EditCallback(fr01destEdit, "", config.FileForwards.Rule0.Dest, 5, 1, 1)
	ed.EditCallback(fr01proEdit, "", config.FileForwards.Rule0.Protocol, 5, 1, 2)
	ed.EditCallback(fr02srcEdit, "", config.FileForwards.Rule0.Src, 5, 2, 0)
	ed.EditCallback(fr02destEdit, "", config.FileForwards.Rule0.Dest, 5, 2, 1)
	ed.EditCallback(fr02proEdit, "", config.FileForwards.Rule0.Protocol, 5, 2, 2)
	ed.EditCallback(fr03srcEdit, "", config.FileForwards.Rule0.Src, 5, 3, 0)
	ed.EditCallback(fr03destEdit, "", config.FileForwards.Rule0.Dest, 5, 3, 1)
	ed.EditCallback(fr03proEdit, "", config.FileForwards.Rule0.Protocol, 5, 3, 2)

	err = ed.Run()
	// ed.Cleanup()
	ed.Log("Name is: %v\n", file.Name)
	ed.Log("DONE RUNNING")
	if err == nil {
		saveEdit(file, isNew)
		return nil
	} else {
		return ErrQ
	}

}

func metadataDisplay() string {
	out := fmt.Sprintf("%s (%s)\n%s\n\n%s\n\n%s", config.Metadata.Name, config.Metadata.Version, config.Metadata.Author, config.Metadata.Description, config.Metadata.AppURL)
	// out := "Name: " + config.Metadata.Name + "\nDescription: " + config.Metadata.Description + "\nAuthor: " + config.Metadata.Author + "\nVersion: " + config.Metadata.Version + "\nURL: " + config.Metadata.AppURL
	return out
}

func saveDisplay() string {
	out := "Press 'enter' to export these settings to a config file."
	return out
}

func saveEdit(file *shared.BuildConfig, isNew bool) {

	cardIPs := make([]string, 0)
	cardIPs = append(cardIPs, config.NetworkSettings.NetworkCards.Card0.IP, config.NetworkSettings.NetworkCards.Card1.IP, config.NetworkSettings.NetworkCards.Card2.IP, config.NetworkSettings.NetworkCards.Card3.IP)
	cardMasks := make([]string, 0)
	cardMasks = append(cardMasks, config.NetworkSettings.NetworkCards.Card0.Mask, config.NetworkSettings.NetworkCards.Card1.Mask, config.NetworkSettings.NetworkCards.Card2.Mask, config.NetworkSettings.NetworkCards.Card3.Mask)
	cardGateways := make([]string, 0)
	cardGateways = append(cardGateways, config.NetworkSettings.NetworkCards.Card0.Gateway, config.NetworkSettings.NetworkCards.Card1.Gateway, config.NetworkSettings.NetworkCards.Card2.Gateway, config.NetworkSettings.NetworkCards.Card3.Gateway)

	// binArgs := make([]string, 0)
	// binArgs = append(binArgs, config.AppSettings.BinaryArgs)

	binArgs := config.AppSettings.BinaryArgs
	sysEnvs := config.AppSettings.SystemEnvs

	// sysEnvs := make([]string, 0)
	// sysEnvs = append(sysEnvs, config.AppSettings.SystemEnvs)

	dns := make([]string, 0)
	if config.NetworkSettings.DNS.DNS0 != "" {
		dns = append(dns, config.NetworkSettings.DNS.DNS0)
	}
	if config.NetworkSettings.DNS.DNS1 != "" {
		dns = append(dns, config.NetworkSettings.DNS.DNS1)
	}
	if config.NetworkSettings.DNS.DNS2 != "" {
		dns = append(dns, config.NetworkSettings.DNS.DNS2)
	}
	if config.NetworkSettings.DNS.DNS3 != "" {
		dns = append(dns, config.NetworkSettings.DNS.DNS3)
	}

	file.Name = config.Metadata.Name
	file.Description = config.Metadata.Description
	file.Author = config.Metadata.Author
	if isNew {
		file.ReleaseDate = time.Now()
	}
	file.Version = config.Metadata.Version
	file.AppURL = config.Metadata.AppURL

	// pmap := strings.Split(cmd.pmap, ",")
	// if len(pmap) > 0 {
	// 	for _, p := range pmap {
	// 		pmapIndividual := strings.Split(p, ":")
	// 		if len(pmapIndividual) != 3 {
	// 			return errors.New("Error: KVM, QEMU, and VIRTUALBOX hypervisors require alternative argument.\nArgument '--port-map' formatted incorrectly. Correct format should be 'x:y:z'")
	// 		}
	// 	}
	// }

	file.App.BinaryArgs = make([]string, 0)
	bargs := strings.Split(binArgs, " ")
	if len(bargs) > 0 {
		for _, a := range bargs {
			if a != "" {
				file.App.BinaryArgs = append(file.App.BinaryArgs, a)
			}
		}
	}

	file.App.SystemEnvs = make([]string, 0)
	envs := strings.Split(sysEnvs, " ")
	if len(bargs) > 0 {
		for _, a := range envs {

			aa := strings.Split(a, "=")
			if len(aa) == 2 {
				file.App.SystemEnvs = append(file.App.SystemEnvs, a)
			}

		}
	}

	// file.App.BinaryArgs = binArgs
	// file.App.SystemEnvs = sysEnvs

	file.Network.NetworkCards = make([]shared.NetworkCardConfig, 0)

	if config.NetworkSettings.NetworkCards.Card0.IP != "" {
		if strings.TrimSpace(strings.ToLower(config.NetworkSettings.NetworkCards.Card0.IP)) == "dhcp" {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP: config.NetworkSettings.NetworkCards.Card0.IP,
			})
		} else {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP:      config.NetworkSettings.NetworkCards.Card0.IP,
				Mask:    config.NetworkSettings.NetworkCards.Card0.Mask,
				Gateway: config.NetworkSettings.NetworkCards.Card0.Gateway,
			})
		}
	}
	if config.NetworkSettings.NetworkCards.Card1.IP != "" {
		if strings.TrimSpace(strings.ToLower(config.NetworkSettings.NetworkCards.Card1.IP)) != "dhcp" {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP:      config.NetworkSettings.NetworkCards.Card1.IP,
				Mask:    config.NetworkSettings.NetworkCards.Card1.Mask,
				Gateway: config.NetworkSettings.NetworkCards.Card1.Gateway,
			})
		} else {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP: config.NetworkSettings.NetworkCards.Card1.IP,
			})
		}
	}
	if config.NetworkSettings.NetworkCards.Card2.IP != "" {
		if strings.TrimSpace(strings.ToLower(config.NetworkSettings.NetworkCards.Card2.IP)) == "dhcp" {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP: config.NetworkSettings.NetworkCards.Card2.IP,
			})
		} else {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP:      config.NetworkSettings.NetworkCards.Card2.IP,
				Mask:    config.NetworkSettings.NetworkCards.Card2.Mask,
				Gateway: config.NetworkSettings.NetworkCards.Card2.Gateway,
			})
		}
	}
	if config.NetworkSettings.NetworkCards.Card3.IP != "" {
		if strings.TrimSpace(strings.ToLower(config.NetworkSettings.NetworkCards.Card3.IP)) == "dhcp" {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP: config.NetworkSettings.NetworkCards.Card3.IP,
			})
		} else {
			file.Network.NetworkCards = append(file.Network.NetworkCards, shared.NetworkCardConfig{
				IP:      config.NetworkSettings.NetworkCards.Card3.IP,
				Mask:    config.NetworkSettings.NetworkCards.Card3.Mask,
				Gateway: config.NetworkSettings.NetworkCards.Card3.Gateway,
			})
		}
	}
	file.Disk.DiskSize = config.DiskSettings.DiskSize
	file.Disk.FileSystem = config.DiskSettings.FileSystem
	file.Disk.MaxFD = config.DiskSettings.MaxFD

	file.Network.DNS = dns

	file.NTP.Hostname = config.TimeServers.Hostname
	file.NTP.Servers = make([]string, 0)
	if config.TimeServers.Server0 != "" {
		file.NTP.Servers = append(file.NTP.Servers, config.TimeServers.Server0)
	}
	if config.TimeServers.Server1 != "" {
		file.NTP.Servers = append(file.NTP.Servers, config.TimeServers.Server1)
	}
	if config.TimeServers.Server2 != "" {
		file.NTP.Servers = append(file.NTP.Servers, config.TimeServers.Server2)
	}
	if config.TimeServers.Server3 != "" {
		file.NTP.Servers = append(file.NTP.Servers, config.TimeServers.Server3)
	}

	file.Redirects.Rules = make([]shared.Redirect, 0)

	if config.FileForwards.Rule0.Src != "" {
		file.Redirects.Rules = append(file.Redirects.Rules, config.FileForwards.Rule0)
	}
	if config.FileForwards.Rule1.Src != "" {
		file.Redirects.Rules = append(file.Redirects.Rules, config.FileForwards.Rule1)
	}
	if config.FileForwards.Rule2.Src != "" {
		file.Redirects.Rules = append(file.Redirects.Rules, config.FileForwards.Rule2)
	}
	if config.FileForwards.Rule3.Src != "" {
		file.Redirects.Rules = append(file.Redirects.Rules, config.FileForwards.Rule3)
	}
}

func appSettingsDisplay() string {
	out := "Binary Args: " + config.AppSettings.BinaryArgs + "\nSystem Envs: " + config.AppSettings.SystemEnvs
	return out
}

func netSettingsDisplay() string {
	out := "DNS:\n    " + config.NetworkSettings.DNS.DNS0 + "\n    " + config.NetworkSettings.DNS.DNS1 + "\n    " + config.NetworkSettings.DNS.DNS2 + "\n    " + config.NetworkSettings.DNS.DNS3
	out = strings.TrimSpace(out)
	out += "\n\n"
	out += showCard(0, config.NetworkSettings.NetworkCards.Card0)
	out += showCard(1, config.NetworkSettings.NetworkCards.Card1)
	out += showCard(2, config.NetworkSettings.NetworkCards.Card2)
	out += showCard(3, config.NetworkSettings.NetworkCards.Card3)

	return out
}

func timeserversDisplay() string {
	out := fmt.Sprintf("Hostname: %s\n\nServer 0: %s\nServer 1: %s\nServer 2: %s\nServer 3: %s\n", config.TimeServers.Hostname, config.TimeServers.Server0, config.TimeServers.Server1, config.TimeServers.Server2, config.TimeServers.Server3)
	return out
}

func ffDisplay() string {
	out := fmt.Sprintf("File Forwarding:\n\nSource: %s\nDestination: %s\nProtocol: %s\n\nSource: %s\nDestination: %s\nProtocol: %s\n\nSource: %s\nDestination: %s\nProtocol: %s\n\nSource: %s\nDestination: %s\nProtocol: %s\n\n", config.FileForwards.Rule0.Src, config.FileForwards.Rule0.Dest, config.FileForwards.Rule0.Protocol, config.FileForwards.Rule1.Src, config.FileForwards.Rule1.Dest, config.FileForwards.Rule1.Protocol, config.FileForwards.Rule2.Src, config.FileForwards.Rule2.Dest, config.FileForwards.Rule2.Protocol, config.FileForwards.Rule3.Src, config.FileForwards.Rule3.Dest, config.FileForwards.Rule3.Protocol)
	return out
}

func tsHostDisplay() string {
	out := config.TimeServers.Hostname
	return out
}

func tsHostEdit(str string) error {
	config.TimeServers.Hostname = str
	return nil
}

func tsS0Display() string {
	out := config.TimeServers.Server0
	return out
}

func tsS0Edit(str string) error {
	config.TimeServers.Server0 = str
	return nil
}

func tsS1Display() string {
	out := config.TimeServers.Server1
	return out
}

func tsS1Edit(str string) error {
	config.TimeServers.Server1 = str
	return nil
}

func tsS2Display() string {
	out := config.TimeServers.Server2
	return out
}

func tsS2Edit(str string) error {
	config.TimeServers.Server2 = str
	return nil
}

func tsS3Display() string {
	out := config.TimeServers.Server3
	return out
}

func tsS3Edit(str string) error {
	config.TimeServers.Server3 = str
	return nil
}

func showCard(x int, card Card) string {

	if card.IP != "" {
		if strings.ToLower(card.IP) == "dhcp" {
			return fmt.Sprintf("Network Card %v\n    %s\n\n", x, card.IP)
		} else {
			return fmt.Sprintf("Network Card %v\n    %s\n    mask: %s, gateway: %s\n\n", x, card.IP, card.Mask, card.Gateway)
		}
	}

	return ""

}

func miniShowCard(card Card) string {
	if card.IP != "" {
		if strings.ToLower(card.IP) == "dhcp" {
			return fmt.Sprintf("%s\n\n", card.IP)
		} else {
			return fmt.Sprintf("%s\nmask: %s, gateway: %s\n\n", card.IP, card.Mask, card.Gateway)
		}
	}
	return ""
}

func diskSettingsDisplay() string {
	out := "Filesystem: " + config.DiskSettings.FileSystem + "\nMaxFD: " + strconv.Itoa(config.DiskSettings.MaxFD) + "\nDisk Size: " + strconv.Itoa(config.DiskSettings.DiskSize)
	return out
}

func metadataEdit(str string) error {
	return nil
}

func nameDisplay() string {
	out := config.Metadata.Name
	return out
}

func nameEdit(str string) error {
	config.Metadata.Name = str
	return nil
}

func descDisplay() string {
	out := config.Metadata.Description
	return out
}

func descEdit(str string) error {
	config.Metadata.Description = str
	return nil
}

func authorDisplay() string {
	out := config.Metadata.Author
	return out
}

func authorEdit(str string) error {
	config.Metadata.Author = str
	return nil
}

func versionDisplay() string {
	out := config.Metadata.Version
	return out
}

func versionEdit(str string) error {
	config.Metadata.Version = str
	return nil
}

func appurlDisplay() string {
	out := config.Metadata.AppURL
	return out
}

func urlEdit(str string) error {
	config.Metadata.AppURL = str
	return nil
}

func bArgsDisplay() string {
	out := config.AppSettings.BinaryArgs
	return out
}

func bArgsEdit(str string) error {
	config.AppSettings.BinaryArgs = str
	return nil
}

func sysEnvsDisplay() string {
	out := config.AppSettings.SystemEnvs
	return out
}

func sysEnvsEdit(str string) error {
	config.AppSettings.SystemEnvs = str
	return nil
}

func dnsDisplay() string {
	out := config.NetworkSettings.DNS.DNS0 + "\n" + config.NetworkSettings.DNS.DNS1 + "\n" + config.NetworkSettings.DNS.DNS2 + "\n" + config.NetworkSettings.DNS.DNS3
	out = strings.TrimSpace(out)
	out += "\n"
	return out
}

func netCardDisplay() string {
	out := showCard(0, config.NetworkSettings.NetworkCards.Card0)
	out += showCard(1, config.NetworkSettings.NetworkCards.Card1)
	out += showCard(2, config.NetworkSettings.NetworkCards.Card2)
	out += showCard(3, config.NetworkSettings.NetworkCards.Card3)
	return out
}

func dns0display() string {
	out := config.NetworkSettings.DNS.DNS0
	return out
}

func dns0Edit(str string) error {
	config.NetworkSettings.DNS.DNS0 = str
	return nil
}

func dns1display() string {
	out := config.NetworkSettings.DNS.DNS1
	return out
}

func dns1Edit(str string) error {
	config.NetworkSettings.DNS.DNS1 = str
	return nil
}

func dns2display() string {
	out := config.NetworkSettings.DNS.DNS2
	return out
}

func dns2Edit(str string) error {
	config.NetworkSettings.DNS.DNS2 = str
	return nil
}

func dns3display() string {
	out := config.NetworkSettings.DNS.DNS3
	return out
}

func dns3Edit(str string) error {
	config.NetworkSettings.DNS.DNS3 = str
	return nil
}

func card0display() string {
	out := miniShowCard(config.NetworkSettings.NetworkCards.Card0)
	return out
}

func card1display() string {
	out := miniShowCard(config.NetworkSettings.NetworkCards.Card1)
	return out
}

func card2display() string {
	out := miniShowCard(config.NetworkSettings.NetworkCards.Card2)
	return out
}

func card3display() string {
	out := miniShowCard(config.NetworkSettings.NetworkCards.Card3)
	return out
}

func card0IPdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card0.IP
	return out
}

func card0IPEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card0.IP = strings.ToLower(str)
	return nil
}

func card0Maskdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card0.Mask
	return out
}

func card0MaskEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card0.Mask = str
	return nil
}

func card0Gatewaydisplay() string {
	out := config.NetworkSettings.NetworkCards.Card0.Gateway
	return out
}

func card0GatewayEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card0.Gateway = str
	return nil
}

func card1IPdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card1.IP
	return out
}

func card1IPEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card1.IP = strings.ToLower(str)
	return nil
}

func card1Maskdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card1.Mask
	return out
}

func card1MaskEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card1.Mask = str
	return nil
}

func card1Gatewaydisplay() string {
	out := config.NetworkSettings.NetworkCards.Card1.Gateway
	return out
}

func card1GatewayEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card1.Gateway = str
	return nil
}

func card2IPdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card2.IP
	return out
}

func card2IPEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card2.IP = strings.ToLower(str)
	return nil
}

func card2Maskdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card2.Mask
	return out
}

func card2MaskEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card2.Mask = str
	return nil
}

func card2Gatewaydisplay() string {
	out := config.NetworkSettings.NetworkCards.Card2.Gateway
	return out
}

func card2GatewayEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card2.Gateway = str
	return nil
}

func card3IPdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card3.IP
	return out
}

func card3IPEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card3.IP = strings.ToLower(str)
	return nil
}

func card3Maskdisplay() string {
	out := config.NetworkSettings.NetworkCards.Card3.Mask
	return out
}

func card3MaskEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card3.Mask = str
	return nil
}

func card3Gatewaydisplay() string {
	out := config.NetworkSettings.NetworkCards.Card3.Gateway
	return out
}

func card3GatewayEdit(str string) error {
	config.NetworkSettings.NetworkCards.Card3.Gateway = str
	return nil
}

func filesystemDisplay() string {
	out := config.DiskSettings.FileSystem
	return out
}

func filesystemEdit(str string) error {
	config.DiskSettings.FileSystem = str
	return nil
}

func maxfdDisplay() string {
	out := strconv.Itoa(config.DiskSettings.MaxFD)
	return out
}

func maxfdEdit(str string) error {
	n, err := strconv.Atoi(str)
	if err == nil {
		config.DiskSettings.MaxFD = n
	}
	return err
}

func disksizeDisplay() string {
	out := strconv.Itoa(config.DiskSettings.DiskSize)
	return out
}

func disksizeEdit(str string) error {
	n, err := strconv.Atoi(str)
	if err == nil {
		config.DiskSettings.DiskSize = n
	}
	return err
}

func fr00Display() string {
	out := fmt.Sprintf("Source: %v\nDestination: %v\nProtocol: %v\n", config.FileForwards.Rule0.Src, config.FileForwards.Rule0.Dest, config.FileForwards.Rule0.Protocol)
	return out
}
func fr01Display() string {
	out := fmt.Sprintf("Source: %v\nDestination: %v\nProtocol: %v\n", config.FileForwards.Rule1.Src, config.FileForwards.Rule1.Dest, config.FileForwards.Rule1.Protocol)
	return out
}
func fr02Display() string {
	out := fmt.Sprintf("Source: %v\nDestination: %v\nProtocol: %v\n", config.FileForwards.Rule2.Src, config.FileForwards.Rule2.Dest, config.FileForwards.Rule2.Protocol)
	return out
}
func fr03Display() string {
	out := fmt.Sprintf("Source: %v\nDestination: %v\nProtocol: %v\n", config.FileForwards.Rule3.Src, config.FileForwards.Rule3.Dest, config.FileForwards.Rule3.Protocol)
	return out
}

func fr00srcDisplay() string {
	out := config.FileForwards.Rule0.Src
	return out
}
func fr00destDisplay() string {
	out := config.FileForwards.Rule0.Dest
	return out
}
func fr00proDisplay() string {
	out := config.FileForwards.Rule0.Protocol
	return out
}

func fr00srcEdit(str string) error {
	config.FileForwards.Rule0.Src = str
	return nil
}
func fr00destEdit(str string) error {
	config.FileForwards.Rule0.Dest = str
	return nil
}
func fr00proEdit(str string) error {
	config.FileForwards.Rule0.Protocol = str
	return nil
}

func fr01srcDisplay() string {
	out := config.FileForwards.Rule1.Src
	return out
}
func fr01destDisplay() string {
	out := config.FileForwards.Rule1.Dest
	return out
}
func fr01proDisplay() string {
	out := config.FileForwards.Rule1.Protocol
	return out
}

func fr01srcEdit(str string) error {
	config.FileForwards.Rule1.Src = str
	return nil
}
func fr01destEdit(str string) error {
	config.FileForwards.Rule1.Dest = str
	return nil
}
func fr01proEdit(str string) error {
	config.FileForwards.Rule1.Protocol = str
	return nil
}

func fr02srcDisplay() string {
	out := config.FileForwards.Rule2.Src
	return out
}
func fr02destDisplay() string {
	out := config.FileForwards.Rule2.Dest
	return out
}
func fr02proDisplay() string {
	out := config.FileForwards.Rule2.Protocol
	return out
}

func fr02srcEdit(str string) error {
	config.FileForwards.Rule2.Src = str
	return nil
}
func fr02destEdit(str string) error {
	config.FileForwards.Rule2.Dest = str
	return nil
}
func fr02proEdit(str string) error {
	config.FileForwards.Rule2.Protocol = str
	return nil
}

func fr03srcDisplay() string {
	out := config.FileForwards.Rule3.Src
	return out
}
func fr03destDisplay() string {
	out := config.FileForwards.Rule3.Dest
	return out
}
func fr03proDisplay() string {
	out := config.FileForwards.Rule3.Protocol
	return out
}

func fr03srcEdit(str string) error {
	config.FileForwards.Rule3.Src = str
	return nil
}
func fr03destEdit(str string) error {
	config.FileForwards.Rule3.Dest = str
	return nil
}
func fr03proEdit(str string) error {
	config.FileForwards.Rule3.Protocol = str
	return nil
}
