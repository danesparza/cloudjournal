package journal

/*
	JSON data example:

	{
		"__CURSOR": "s=7fe895b45f18448daa12dfe9ec1d2993;i=230;b=6b9d0f62f43c4b0bb0f61848b4da3b15;m=3b5c2b9;t=5cff2c0f5338d;x=274b5b63cb69c9f7",
		"__REALTIME_TIMESTAMP": "1636016409883533",
		"__MONOTONIC_TIMESTAMP": "62243513",
		"_BOOT_ID": "6b9d0f62f43c4b0bb0f61848b4da3b15",
		"PRIORITY": "6",
		"_MACHINE_ID": "b8043161058e4f26a87fb9d0978451b6",
		"_HOSTNAME": "dashboard",
		"SYSLOG_FACILITY": "3",
		"_UID": "0",
		"_GID": "0",
		"_SYSTEMD_SLICE": "system.slice",
		"_TRANSPORT": "stdout",
		"_CAP_EFFECTIVE": "3fffffffff",
		"SYSLOG_IDENTIFIER": "daydash",
		"_PID": "282",
		"_COMM": "daydash",
		"_EXE": "/usr/bin/daydash",
		"_CMDLINE": "/usr/bin/daydash start --config=/etc/daydash/config.yaml",
		"_SYSTEMD_CGROUP": "/system.slice/daydash.service",
		"_SYSTEMD_UNIT": "daydash.service",
		"_SYSTEMD_INVOCATION_ID": "294437e7d25d4b3eafae9737cf1f1577",
		"MESSAGE": "{\"historyttl\":2592000000000000,\"level\":\"info\",\"machineid\":\"23d5aa419eca8a1f24afaa8f9b581ffd4d13d428b0680c2234de8df1956bc360\",\"msg\":\"System started\",\"time\":\"2021-11-04T05:00:09-04:00\"}"
	}
*/

type Entry struct {
	Cursor                  string `json:"__CURSOR"`
	RealtimeTimestamp       string `json:"__REALTIME_TIMESTAMP"`
	MonotonicTimestamp      string `json:"__MONOTONIC_TIMESTAMP"`
	BootID                  string `json:"_BOOT_ID"`
	Priority                string `json:"PRIORITY"`
	MachineID               string `json:"_MACHINE_ID"`
	Hostname                string `json:"_HOSTNAME"`
	SyslogFacility          string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier        string `json:"SYSLOG_IDENTIFIER"`
	UID                     string `json:"_UID"`
	GID                     string `json:"_GID"`
	Transport               string `json:"_TRANSPORT"`
	Codefile                string `json:"CODE_FILE"`
	Codeline                string `json:"CODE_LINE"`
	Codefunction            string `json:"CODE_FUNCTION"`
	MessageID               string `json:"MESSAGE_ID"`
	Result                  string `json:"RESULT"`
	PID                     string `json:"_PID"`
	Comm                    string `json:"_COMM"`
	EXE                     string `json:"_EXE"`
	CmdLine                 string `json:"_CMDLINE"`
	CapEffective            string `json:"_CAP_EFFECTIVE"`
	SystemDCGroup           string `json:"_SYSTEMD_CGROUP"`
	SystemDUnit             string `json:"_SYSTEMD_UNIT"`
	SystemDSlice            string `json:"_SYSTEMD_SLICE"`
	Unit                    string `json:"UNIT"`
	Message                 string `json:"MESSAGE"`
	SourceRealtimeTimestamp string `json:"_SOURCE_REALTIME_TIMESTAMP"`
	SystemDInvocationID     string `json:"_SYSTEMD_INVOCATION_ID"`
}