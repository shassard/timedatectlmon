// +build linux

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const (
	TimeDateCtlExe = "/usr/bin/timedatectl"
	TimeDateFormat = "Mon 2006-01-02 15:04:05 MST"
)

type TimeDateCtl struct {
	LocalTime      time.Time      `json:"local_time"`
	RTCTime        time.Time      `json:"rtc_time"`
	RTCInLocalTime bool           `json:"rtc_in_localtime"`
	locationPtr    *time.Location `json:"-"`
	Location       string         `json:"location"`
	Synchronized   bool           `json:"ntp_synchronized"`
	NTPEnabled     bool           `json:"ntp_enabled"`
}

// IsYes converts *yes* to true, otherwise returns false.
func IsYes(answer string) bool {
	if answer == "yes" {
		return true
	}
	return false
}

// ParseTimeDateCtl parses the output from *timedatectl show*.
func ParseTimeDateCtl() (*TimeDateCtl, error) {
	out, err := exec.Command(TimeDateCtlExe, "show").Output()
	if err != nil {
		return nil, err
	}

	var td TimeDateCtl

	for _, line := range bytes.Split(out, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) != 2 {
			continue
		}

		k := string(parts[0])
		v := string(parts[1])

		switch k {
		case "Timezone":
			td.locationPtr, err = time.LoadLocation(v)
			if err != nil {
				return nil, err
			}
		case "LocalRTC":
			td.RTCInLocalTime = IsYes(v)
		case "NTP":
			td.NTPEnabled = IsYes(v)
		case "NTPSynchronized":
			td.Synchronized = IsYes(v)
		case "TimeUSec":
			td.LocalTime, err = time.ParseInLocation(TimeDateFormat, v, td.locationPtr)
			if err != nil {
				return nil, err
			}
		case "RTCTimeUSec":
			td.RTCTime, err = time.ParseInLocation(TimeDateFormat, v, td.locationPtr)
			if err != nil {
				return nil, err
			}
		}
	}

	td.Location = td.locationPtr.String()

	return &td, nil
}

func main() {
	d, err := ParseTimeDateCtl()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	b, err := json.Marshal(&d)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Printf("%s", b)
}
