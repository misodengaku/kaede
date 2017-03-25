package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"./himawari"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
)

type EpgDatetime int64

func (e EpgDatetime) GetTime() time.Time {
	return time.Unix(int64(e)/10000, 0)
}

type EpgCategory struct {
	Large struct {
		JaJP string `json:"ja_JP"`
		En   string `json:"en"`
	} `json:"large"`
	Middle struct {
		JaJP string `json:"ja_JP"`
		En   string `json:"en"`
	} `json:"middle"`
}

type EpgProgram struct {
	Channel  string        `json:"channel"` // ID
	Title    string        `json:"title"`
	Detail   string        `json:"detail"`
	Start    EpgDatetime   `json:"start"`
	End      EpgDatetime   `json:"end"`
	Duration int           `json:"duration"`
	Category []EpgCategory `json:"category"`
	EventID  int           `json:"event_id"`
	FreeCA   bool          `json:"freeCA"`
	Video    struct {
		Resolution string `json:"resolution"`
		Aspect     string `json:"aspect"`
	} `json:"video"`
	Audio []struct {
		Type     string `json:"type"`
		Langcode string `json:"langcode"`
		Extdesc  string `json:"extdesc"`
	} `json:"audio"`
	// 以下とりあえず無視
	Extdetail  []interface{} `json:"extdetail"`
	Attachinfo []interface{} `json:"attachinfo"`
}

type EpgdumpJSON []struct {
	himawari.BroadcastStation
	ID       string       `json:"id"`
	Programs []EpgProgram `json:"programs"`
}

type Channel struct {
	PhysicalChannel int
	LogicalChannels []himawari.BroadcastStation
}

type Device struct {
	Path   string
	IsOpen bool
}

type Devices []Device

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	db, _ := sql.Open("sqlite3", "./foo.db")
	defer db.Close()
	// app.Name = "kaede"
	// app.Usage = "make an explosive entrance"
	app.Commands = []cli.Command{
		{
			Name:    "chscan",
			Aliases: []string{"c"},
			Usage:   "scan channel",
			Action: func(c *cli.Context) error {
				channelScan()
				return nil
			},
		},
		{
			Name:    "tschscan",
			Aliases: []string{"t"},
			Usage:   "scan channel from ts files",
			Action: func(c *cli.Context) error {
				channelScanFromTSFiles()
				return nil
			},
		},
		{
			Name:    "sync",
			Aliases: []string{"s"},
			Usage:   "sync",
			Action: func(c *cli.Context) error {
				syncFromEpgDump()
				return nil
			},
		},
		{
			Name:    "server",
			Aliases: []string{"f"},
			Usage:   "server",
			Action: func(c *cli.Context) error {
				syncFromEpgDump()
				return nil
			},
		},
		{
			Name:    "daemon",
			Aliases: []string{"f"},
			Usage:   "daemon",
			Action: func(c *cli.Context) error {
				syncFromEpgDump()
				return nil
			},
		},
		{
			Name:    "execrecord",
			Aliases: []string{"r"},
			Usage:   "execrecord TASKID",
			Action: func(c *cli.Context) error {
				if len(c.Args()) < 1 {
					fmt.Println("schedule id is needed")
					return nil
				}
				fmt.Printf("Hello %q\r\n", c.Args().Get(0))
				// create table schedule(id integer primary key, schedule_id text not null, start_date text not null, duration integer);
				// insert into schedule (schedule_id, start_date, duration) values ("test", "2017-03-19T22:24:30+09:00", 1800);
				sql := fmt.Sprintf(`SELECT * FROM schedule WHERE schedule_id="%s" LIMIT 1;`, c.Args().Get(0))
				fmt.Println(sql)
				rows, err := db.Query(sql)
				defer rows.Close()
				for rows.Next() {
					id, scheduleID, dateStr, duration := 0, "", "", 0
					err = rows.Scan(&id, &scheduleID, &dateStr, &duration)
					if err != nil {
						return err
					}

					fmt.Println(id, scheduleID, dateStr, duration)
				}
				fmt.Println("ik")
				// syncFromEpgDump()
				return nil
			},
		},
	}
	app.Run(os.Args)
	fmt.Println("app exit")
}

func deviceMapInit(devices ...string) Devices {
	dl := make([]Device, len(devices))
	for i, d := range devices {
		dl[i] = Device{Path: d, IsOpen: false}
	}
	return dl
}

func scanStationName(device Device, ch, scantime int) []himawari.BroadcastStation {
	tsname := fmt.Sprintf("./kaede_scan_gr%d.ts", ch)
	scantimeStr := strconv.Itoa(scantime)
	err := exec.Command("recpt1", "--b25", "--strip", strconv.Itoa(ch), scantimeStr, tsname, "--device", device.Path).Run()
	if err != nil {
		fmt.Println("[ERROR] recpt1 execute failed!!!")
		return nil
	}
	epg := dumpTS(tsname)
	if len(epg) > 0 {
		cnl := make([]himawari.BroadcastStation, len(epg))
		for i, cn := range epg {
			cnl[i] = cn.BroadcastStation
			cnl[i].StationID = fmt.Sprintf("GR1_%d", cnl[i].ServiceID)
		}
		return cnl
	}
	return nil
}

func (devices Devices) GetDevice() chan Device {
	device := make(chan Device)
	go func() {
		for {
			for i, d := range devices {
				if !d.IsOpen {
					devices[i].IsOpen = true
					device <- d
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return device
}

func (devices Devices) FreeDevice(path string) {
	for i, d := range devices {
		if d.Path == path {
			devices[i].IsOpen = false
		}
	}
}

func channelScanFromTSFiles() {
	fis, err := ioutil.ReadDir("./")

	if err != nil {
		panic(err)
	}

	channelList := []Channel{}
	r := regexp.MustCompile(`kaede_scan_gr(\d+).ts`)
	for _, fi := range fis {
		f := r.FindAllStringSubmatch(fi.Name(), -1)
		if len(f) != 1 {
			continue
		}

		fname := fi.Name()
		ch, _ := strconv.Atoi(f[0][1])
		fmt.Println(fi.Name())

		epg := dumpTS(fname)
		channel := Channel{PhysicalChannel: ch}
		if len(epg) > 0 {
			cnl := make([]himawari.BroadcastStation, len(epg))
			for i, cn := range epg {
				cnl[i] = cn.BroadcastStation
				cnl[i].StationID = fmt.Sprintf("GR1_%d", cnl[i].ServiceID)
			}
			channel.LogicalChannels = cnl
		}
		channelList = append(channelList, channel)
	}
	chb, _ := json.Marshal(channelList)
	ioutil.WriteFile("channelList.json", chb, 666)
}

func channelScan() {
	// UHF channel scan
	wait := 5 * time.Second
	scantime := 30
	channelList := []Channel{}
	devices := deviceMapInit("/dev/pt1video2", "/dev/pt1video3")
	fmt.Println("channel scan started.")

	for channel := 16; channel < 63; channel++ {
		d := <-devices.GetDevice()
		fmt.Println("scanning:", channel)
		// fmt.Println("locked", d.Path)
		// fmt.Println("goroutine", runtime.NumGoroutine())

		go func(ch int, device Device) {
			time.Sleep(wait)
			defer func() {
				devices.FreeDevice(d.Path)
				// fmt.Println("unlocked", device.Path)
			}()
			// fmt.Println("start scan", ch, device.Path)
			cmd := exec.Command("recpt1", "--b25", "--strip", strconv.Itoa(ch), "3", "/dev/null", "--device", device.Path)
			errPipe, _ := cmd.StderrPipe()
			scanner := bufio.NewScanner(errPipe)
			isTuned := make(chan bool)

			go func() {
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "C/N") {
						// fmt.Println("\tchannel tuned.", ch, device.Path)
						isTuned <- true
						return
					}
				}

				// fmt.Println("\tchscan closed", device.Path)
				cmd.Wait()
				// fmt.Println("process exited", device.Path)
				isTuned <- false
			}()

			cmd.Start()

			if !<-isTuned {
				// チューニング失敗
				// fmt.Println("tunning fail", device.Path)
				return
			}

			cmd.Process.Kill()
			time.Sleep(wait)
			fmt.Println(ch, "name scanning...", device.Path)

			channel := Channel{PhysicalChannel: ch}
			for i := 0; i < 3; i++ {
				result := scanStationName(device, ch, scantime)
				if result != nil {
					channel.LogicalChannels = result
					fmt.Println("detect: ", channel.PhysicalChannel, channel.LogicalChannels, device.Path)
					channelList = append(channelList, channel)
					break
				} else {
					fmt.Println("\tstation scan fail: ", ch, device.Path, "retrying...")
					time.Sleep(wait * 2)
				}
			}
			// fmt.Println("scan complete", device.Path)
		}(channel, d)
	}

	chb, _ := json.MarshalIndent(channelList, "", "\t")
	ioutil.WriteFile("channelList.json", chb, 666)

}

func syncFromEpgDump() {
	fis, err := ioutil.ReadDir("./")

	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile(`kaede_scan_gr(\d+).ts`)
	for _, fi := range fis {
		f := r.FindAllStringSubmatch(fi.Name(), -1)
		if len(f) != 1 {
			continue
		}

		fname := fi.Name()
		fmt.Println(fi.Name())

		epg := dumpTS(fname)
		workQueue := make(chan himawari.Program, 100)
		go func() {
			for {
				ps := <-workQueue
				ps.UploadProgram()
			}
		}()

		// return
		for _, v := range epg {
			v.BroadcastStation.StationID = v.ID
			_, err := himawari.GetStation(v.ID)
			if err != nil {
				if err.Error() == "HTTP Error: 404" {
					// 局が存在しない場合作成
					himawari.CreateStation(&v.BroadcastStation)
				}
			}

			pc := make(chan struct{}, len(v.Programs))
			for _, p := range v.Programs {
				ps := himawari.Program{}
				ps.Station = p.Channel
				ps.EventID = p.EventID
				ps.Start = p.Start.GetTime()
				ps.End = p.End.GetTime()
				ps.Title = p.Title
				ps.Detail = p.Detail
				// fmt.Println(ps.Title)
				ps.Categories = []himawari.Category{}
				for _, c := range p.Category {
					ch, _ := himawari.SearchCategories(c.Large.JaJP, c.Middle.JaJP)
					// fmt.Printf("%#v", ch)
					if len(ch) < 1 {
						fmt.Printf("error \"%s\" \"%s\"\r\n", c.Large.JaJP, c.Middle.JaJP)
					}
					ps.Categories = append(ps.Categories, ch[0])
				}
				// time.Sleep(100 * time.Mill	isecond)
				workQueue <- ps
				pc <- struct{}{}
			}
			for i := 0; i < len(v.Programs); i++ {
				<-pc
			}
			fmt.Println(len(v.Programs), "registered")

		}
	}
}

func dumpTS(filename string) EpgdumpJSON {
	exec.Command("epgdump", "json", filename, "./dump.json").Run()
	var epg EpgdumpJSON
	data, _ := ioutil.ReadFile("dump.json")
	json.Unmarshal(data, &epg)
	return epg
}
