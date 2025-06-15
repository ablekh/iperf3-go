package protocol

// TestConfig represents the test configuration sent by the client
type TestConfig struct {
	Protocol        string `json:"protocol,omitempty"`
	Time            int    `json:"time,omitempty"`
	Parallel        int    `json:"parallel,omitempty"`
	Reverse         bool   `json:"reverse,omitempty"`
	Window          int    `json:"window,omitempty"`
	Length          int    `json:"len,omitempty"`
	Bandwidth       int64  `json:"bandwidth,omitempty"`
	Fqrate          int64  `json:"fqrate,omitempty"`
	Pacing          int    `json:"pacing_timer,omitempty"`
	Burst           int    `json:"burst,omitempty"`
	Bidir           bool   `json:"bidirectional,omitempty"`
	TOS             int    `json:"tos,omitempty"`
	FlowLabel       int    `json:"flowlabel,omitempty"`
	Title           string `json:"title,omitempty"`
	ExtraData       string `json:"extra_data,omitempty"`
	GetServerOutput bool   `json:"get_server_output,omitempty"`
	UDPCountersMode bool   `json:"udp_counters_64bit,omitempty"`
	ZeroCopy        bool   `json:"zerocopy,omitempty"`
	OmitSec         int    `json:"omit,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	Blockcount      int64  `json:"blockcount,omitempty"`
}

// TestResults represents the complete test results
type TestResults struct {
	Start TestStart `json:"start"`
	End   TestEnd   `json:"end"`
}

// TestStart represents the test start information
type TestStart struct {
	Connected     []Connection `json:"connected"`
	Version       string       `json:"version"`
	SystemInfo    string       `json:"system_info"`
	Timestamp     Timestamp    `json:"timestamp"`
	ConnectingTo  ConnectingTo `json:"connecting_to"`
	Cookie        string       `json:"cookie"`
	TCPMSSDefault int          `json:"tcp_mss_default,omitempty"`
	SockBufsize   int          `json:"sock_bufsize,omitempty"`
	SNDBufActual  int          `json:"sndbuf_actual,omitempty"`
	RCVBufActual  int          `json:"rcvbuf_actual,omitempty"`
}

// Connection represents a connection info
type Connection struct {
	Socket     int    `json:"socket"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
}

// Timestamp represents a timestamp
type Timestamp struct {
	Time     int64 `json:"time"`
	Timesecs int64 `json:"timesecs"`
}

// ConnectingTo represents connection target info
type ConnectingTo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TestEnd represents the test end results
type TestEnd struct {
	Streams               []StreamResult `json:"streams"`
	SumSent               StreamResult   `json:"sum_sent"`
	SumReceived           StreamResult   `json:"sum_received"`
	CPUUtilizationPercent CPUUtilization `json:"cpu_utilization_percent"`
	SenderTCPCongestion   string         `json:"sender_tcp_congestion,omitempty"`
	ReceiverTCPCongestion string         `json:"receiver_tcp_congestion,omitempty"`
}

// StreamResult represents results for a single stream
type StreamResult struct {
	Socket        int     `json:"socket,omitempty"`
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	Seconds       float64 `json:"seconds"`
	Bytes         int64   `json:"bytes"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Retransmits   int     `json:"retransmits,omitempty"`
	SndCwnd       int     `json:"snd_cwnd,omitempty"`
	RTT           int     `json:"rtt,omitempty"`
	RTTVar        int     `json:"rttvar,omitempty"`
	PMTU          int     `json:"pmtu,omitempty"`
	Omitted       bool    `json:"omitted,omitempty"`
	Sender        bool    `json:"sender,omitempty"`
}

// CPUUtilization represents CPU utilization statistics
type CPUUtilization struct {
	HostTotal    float64 `json:"host_total"`
	HostUser     float64 `json:"host_user"`
	HostSystem   float64 `json:"host_system"`
	RemoteTotal  float64 `json:"remote_total"`
	RemoteUser   float64 `json:"remote_user"`
	RemoteSystem float64 `json:"remote_system"`
}

// Interval represents an interval measurement
type Interval struct {
	Socket        int     `json:"socket"`
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	Seconds       float64 `json:"seconds"`
	Bytes         int64   `json:"bytes"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Retransmits   int     `json:"retransmits,omitempty"`
	SndCwnd       int     `json:"snd_cwnd,omitempty"`
	RTT           int     `json:"rtt,omitempty"`
	RTTVar        int     `json:"rttvar,omitempty"`
	PMTU          int     `json:"pmtu,omitempty"`
	Omitted       bool    `json:"omitted"`
}
