package main

//go:generate go-bindata assets/

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	curX    = "curX"
	curY    = "curY"
	scrW    = "scrW"
	scrH    = "scrH"
	boxW    = "boxW"
	boxH    = "boxH"
	evtType = "evtType"
)

var intParams = []string{curX, curY, scrW, scrH, boxW, boxH}

func serveAsset(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.String()
	log.Println(urlpath)
	if urlpath == "/" {
		urlpath = "/index.html"
	}

	name := filepath.Join("assets", urlpath)
	log.Println(name)

	info, err := AssetInfo(name)
	if err != nil {
		log.Printf("Error statting %s: %s", name, err)
		http.Error(w, fmt.Sprintf("internal server error statting: %s", name), http.StatusInternalServerError)
		return
	}

	data, err := Asset(name)
	if err != nil {
		log.Printf("Error reading %s: %s", name, err)
		http.Error(w, fmt.Sprintf("internal server error reading: %s", name), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, filepath.Base(name), info.ModTime(), bytes.NewReader(data))
}

func parseNums(r *http.Request, labels []string) (map[string]int, error) {
	ret := map[string]int{}
	for _, label := range labels {
		i, err := strconv.Atoi(r.FormValue(label))
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %s", label, err)
		}
		ret[label] = i
	}
	return ret, nil
}

type SyncXY struct {
	sync.Mutex
	x, y int
}

func NewSyncXY(x, y int) *SyncXY {
	return &SyncXY{
		x: x,
		y: y,
	}
}

func (l *SyncXY) SetPos(x, y int) {
	l.Lock()
	l.x = x
	l.y = y
	l.Unlock()
}

func (l *SyncXY) GetPos() (int, int) {
	var x, y int
	l.Lock()
	x = l.x
	y = l.y
	l.Unlock()
	return x, y
}

func main() {
	var hostport string
	var displayWidth, displayHeight int

	lastPos := NewSyncXY(-1, -1)

	flag.StringVar(&hostport, "http", ":9831", "which host port to start mouser server in")
	flag.IntVar(&displayWidth, "width", 1366, "width of screen")
	flag.IntVar(&displayHeight, "height", 768, "width of screen")

	http.HandleFunc("/mousedata", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseNums(r, intParams)
		if err != nil {
			log.Printf("Error: parsing int params in handleMouse: %s", err)
			return
		}

		log.Printf("%s: %v", r.FormValue(evtType), params)

		switch r.FormValue(evtType) {
		case "panstart":
			lastPos.SetPos(params[curX], params[curY])
		case "panend":
			lastPos.SetPos(-1, -1)

		case "panmove":
			/*
				boxLeft := (params[scrW] - params[boxW]) / 2
				boxRight := params[scrW] - boxLeft
				boxTop := (params[scrH] - params[boxH]) / 2
				boxBtm := params[scrH] - boxTop

				if params[curX] < boxLeft {
					params[curX] = 0
				}

				if params[curX] > boxRight {
					params[curX] = boxRight
				}

				if params[curY] < boxTop {
					params[curY] = boxTop
				}

				if params[curY] > boxBtm {
					params[curY] = boxBtm
				}

				lastX, lastY := lastPos.GetPos()

				xPer := float32(params[curX]-boxLeft-lastX) / float32(params[boxW])
				yPer := float32(params[curY]-boxTop-lastY) / float32(params[boxH])

				lastPos.SetPos(params[curX], params[curY])
				x := strconv.Itoa(int(xPer * float32(displayWidth)))
				y := strconv.Itoa(int(yPer * float32(displayHeight)))
			*/

			lastX, lastY := lastPos.GetPos()
			if lastX == -1 || lastY == -1 {
				break
			}

			boxLeft := (params[scrW] - params[boxW]) / 2
			boxRight := params[scrW] - boxLeft
			boxTop := (params[scrH] - params[boxH]) / 2
			boxBtm := params[scrH] - boxTop

			if params[curX] < boxLeft || params[curX] > boxRight {
				break
			}

			if params[curY] < boxTop || params[curY] > boxBtm {
				break
			}

			deltaX := params[curX] - lastX
			deltaY := params[curY] - lastY
			lastPos.SetPos(params[curX], params[curY])

			amplifyX := float32(displayWidth) / float32(params[boxW])
			amplifyY := float32(displayHeight) / float32(params[boxH])

			scrDeltaX := int(amplifyX * float32(deltaX))
			scrDeltaY := int(amplifyY * float32(deltaY))

			log.Println(scrDeltaX, scrDeltaY)
			exec.Command("xdotool", "mousemove_relative", "--", strconv.Itoa(scrDeltaX), strconv.Itoa(scrDeltaY)).Run()
		case "tap":
			exec.Command("xdotool", "click", "1").Run()
		default:
			log.Printf("Error: unsupported event type: %s", evtType)
		}
	})
	http.HandleFunc("/", serveAsset)
	http.ListenAndServe(hostport, nil)

}
