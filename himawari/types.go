package himawari

import "time"

type Client struct {
	ServerHost string
	ServerPort int
}

type Category struct {
	ID             int    `json:"id"`
	LargeCategory  string `json:"large_category, omitempty"`
	MiddleCategory string `json:"middle_category, omitempty"`
}

type BroadcastStation struct {
	StationID         string `json:"station_id"` // チャンネルと関連付け
	TransportStreamID int    `json:"transport_stream_id"`
	OriginalNetworkID int    `json:"original_network_id"`
	ServiceID         int    `json:"service_id"`
	Name              string `json:"name"`
}

type Program struct {
	Station    string     `json:"station"` // ID
	Title      string     `json:"title"`
	Detail     string     `json:"detail"`
	Start      time.Time  `json:"start_time"`
	End        time.Time  `json:"end_time"`
	EventID    int        `json:"event_id"`
	Categories []Category `json:"categories"`
}
