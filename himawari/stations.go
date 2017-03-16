package himawari

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BroadcastStation struct {
	StationID         string `json:"station_id"` // チャンネルと関連付け
	TransportStreamID int    `json:"transport_stream_id"`
	OriginalNetworkID int    `json:"original_network_id"`
	ServiceID         int    `json:"service_id"`
	Name              string `json:"name"`
}

func GetStation(id string) (*BroadcastStation, error) {

	r, err := http.Get("http://172.31.125.100:8080/v1/broadcaststations/" + id)
	if err != nil {
		fmt.Printf("%#v\r\n", r)
		return nil, err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("%#v\r\n", r)
		return nil, err
	}

	var station BroadcastStation
	json.Unmarshal(body, &station)

	return &station, nil
}

func CreateStation(b *BroadcastStation) error {
	sb, err := json.Marshal(b)
	if err != nil {
		fmt.Printf("%#v\r\n", sb)
		return err
	}

	fmt.Println(string(sb))

	r, _ := http.Post("http://172.31.125.100:8080/v1/broadcaststations/", "application/json", bytes.NewReader(sb))
	fmt.Printf("%#v\r\n", r)
	return nil
}

func ConvertEpgJson(large, middle string) int {
	return 0
}
