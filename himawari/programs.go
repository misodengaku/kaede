package himawari

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetProgram(eventID int, stationID string) (*Program, error) {
	r, err := http.Get(fmt.Sprintf("http://172.31.125.100:8080/v1/programs/?event_id=%d&station=%s", eventID, stationID))
	if err != nil {
		fmt.Printf("%#v\r\n", r)
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP Error: %d", r.StatusCode)
	}

	b, _ := ioutil.ReadAll(r.Body)
	var p Program
	json.Unmarshal(b, &p)
	return &p, nil
}

func (p *Program) UploadProgram() error {

	// pg, _ := GetProgram(p.EventID, p.Station)
	sb, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("%#v\r\n", sb)
		return err
	}

	// if pg == nil {
	r, _ := http.Post("http://172.31.125.100:8080/v1/programs/", "application/json", bytes.NewReader(sb))
	if r.StatusCode > 299 {
		fmt.Println(r.StatusCode, string(sb))
		defer r.Body.Close()
		rb, _ := ioutil.ReadAll(r.Body)
		fmt.Printf("%#v\r\n", string(rb))
		log.Warn(fmt.Sprint("HTTP Error", r.StatusCode))
	}
	// } else {
	// 	req, _ := http.NewRequest("PUT", "http://172.31.125.100:8080/v1/programs/", bytes.NewReader(sb))
	// 	req.Header.Add("Content-type", "application/json")
	// 	client := new(http.Client)
	// 	r, _ := client.Do(req)
	// 	if r.StatusCode > 299 {
	// 		fmt.Println(r.StatusCode, string(sb))
	// 		defer r.Body.Close()
	// 		rb, _ := ioutil.ReadAll(r.Body)
	// 		fmt.Printf("%#v\r\n", string(rb))
	// 		panic(fmt.Sprint("HTTP Error", r.StatusCode))
	// 	}

	// }
	return nil
}
