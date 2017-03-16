package himawari

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Program struct {
	Station    string     `json:"station"` // ID
	Title      string     `json:"title"`
	Detail     string     `json:"detail"`
	Start      time.Time  `json:"start_time"`
	End        time.Time  `json:"end_time"`
	EventID    int        `json:"event_id"`
	Categories []Category `json:"categories"`
}

func (p *Program) UploadProgram() error {
	sb, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("%#v\r\n", sb)
		return err
	}

	// fmt.Println(string(sb))

	http.Post("http://172.31.125.100:8080/v1/programs/", "application/json", bytes.NewReader(sb))
	// fmt.Printf("%#v\r\n", r)
	return nil
}
