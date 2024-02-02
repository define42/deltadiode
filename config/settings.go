package config

import (
	"github.com/olekukonko/tablewriter"
	"os"
)

type SettingsType struct {
	m map[string]SettingType
}

type SettingType struct {
	Description string
	Value       string
}

var Settings = &SettingsType{m: make(map[string]SettingType)}

func (s *SettingsType) Get(id string) string {
	return s.m[id].Value
}

func (s *SettingsType) Has(id string) bool {
	return len(s.m[id].Value) > 0
}

func (s *SettingsType) Set(id string, description string, defaultValue string) {
	if value, ok := os.LookupEnv(id); ok {
		s.m[id] = SettingType{Description: description, Value: value}
	}else{
		s.m[id] = SettingType{Description: description, Value: defaultValue}
	}
}


const (
	ACME_SERVER   = "ACME_SERVER"
	SERVER_DOMAIN = "SERVER_DOMAIN"
	DISK_USAGE_ALLOWED = "DISK_USAGE_ALLOWED"
	DATA_FOLDER = "DATA_FOLDER"
	SERVER_PORT = "SERVER_PORT"
	DMZ_TX = "DMZ_TX"
	DMZ_RX = "DMZ_RX"
	IN_TX = "IN_TX"
	IN_RX = "IN_RX"
	OUT_TX = "OUT_TX"
	OUT_RX = "OUT_RX"
)

func (s *SettingsType) Init() {
	s.Set(ACME_SERVER, "ACME server url", "")
	s.Set(SERVER_DOMAIN, "server domain name","")
	s.Set(DISK_USAGE_ALLOWED, "Allowed disk usage in percentage","75")
	s.Set(DATA_FOLDER, "data folder","./data/")
	s.Set(SERVER_PORT, "server tcp port","8000")
	s.Set(DMZ_TX, "dmx tx port","lo")
	s.Set(DMZ_RX, "dmx rx port","lo")
	s.Set(IN_TX, "in tx port","lo")
	s.Set(IN_RX, "in rx port","lo")
	s.Set(OUT_TX, "out tx port","lo")
	s.Set(OUT_RX, "out rx port","lo")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoFormatHeaders(true)
	table.SetHeader([]string{"KEY", "Description", "value"})
	for key, setting := range s.m {
		table.Append([]string{key, setting.Description, setting.Value})
	}
	table.Render()

}
