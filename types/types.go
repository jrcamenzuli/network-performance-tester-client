package types

type Configuration struct {
	Client struct {
		PID                  uint   `yaml:"pid"`
		LogfilePostfix       string `yaml:"log_file_postfix"`
		ServerHost           string `yaml:"server_host"`
		ServerPingPort       uint   `yaml:"server_ping_port"`
		ServerUDP_Port       uint   `yaml:"server_udp_port"`
		ServerUDP_DNS_Port   uint   `yaml:"server_udp_dns_port"`
		ServerTCP_HTTP_Port  uint   `yaml:"server_tcp_http_port"`
		ServerTCP_HTTPS_Port uint   `yaml:"server_tcp_https_port"`
		ServerTCP_DNS_Port   uint   `yaml:"server_tcp_dns_port"`
		Tests                struct {
			IdleStateOfDevice struct {
				Enable bool `yaml:"enable"`
			} `yaml:"idle_state_of_device"`
			IdleStateOfProcess struct {
				Enable bool `yaml:"enable"`
			} `yaml:"idle_state_of_process"`
			HTTP_Burst struct {
				Enable bool `yaml:"enable"`
			} `yaml:"http_burst"`
			HTTPS_Burst struct {
				Enable bool `yaml:"enable"`
			} `yaml:"https_burst"`
			HTTP_Rate struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"http_rate"`
			HTTPS_Rate struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"https_rate"`
			DNS_UDP_Burst struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"dns_udp_burst"`
			DNS_TCP_Burst struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"dns_tcp_burst"`
			DNS_UDP_Rate struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"dns_udp_rate"`
			DNS_TCP_Rate struct {
				Enable   bool `yaml:"enable"`
				Duration uint `yaml:"duration"`
			} `yaml:"dns_tcp_rate"`
			HTTP_Throughput struct {
				Enable bool `yaml:"enable"`
			} `yaml:"http_throughput"`
			HTTPS_Throughput struct {
				Enable bool `yaml:"enable"`
			} `yaml:"https_throughput"`
			Ping struct {
				Enable       bool `yaml:"enable"`
				CountSamples uint `yaml:"countSamples"`
			} `yaml:"ping"`
			Jitter struct {
				Enable           bool `yaml:"enable"`
				CountDifferences uint `yaml:"countDifferences"`
			} `yaml:"jitter"`
		} `yaml:"tests"`
	} `yaml:"client"`
}

type ProgramArgs struct {
	ConfigFile string
}
