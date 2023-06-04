package main

type Counter struct {
	Name      string  `json:"data"`
	Occurence int     `json:"occurence"`
	Percent   float64 `json:"percent"`
}

type Record struct {
	LogID    string      `json:"syslog_identifier"`
	Time     uint64      `json:"_source_realtime_timestamp,string"`
	Command  string      `json:"_cmdline"`
	Binary   string      `json:"_exe"`
	Unit     string      `json:"_systemd_unit"`
	Priority int64       `json:"priority,string"`
	Message  interface{} `json:"message"`
}
