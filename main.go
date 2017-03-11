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
)

const (
	curX    = "curX"
	curY    = "curY"
	scrW    = "scrW"
	scrH    = "scrH"
	boxW    = "boxW"
	boxH    = "boxH"
	deltaX  = "deltaX"
	deltaY  = "deltaY"
	evtType = "evtType"
)

var intParams = []string{curX, curY, scrW, scrH, boxW, boxH, deltaX, deltaY}

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

func main() {
	var hostport string
	var displayWidth, displayHeight, xadjust, yadjust int

	flag.StringVar(&hostport, "http", ":9831", "which host port to start mouser server in")
	flag.IntVar(&displayWidth, "width", 1366, "width of screen")
	flag.IntVar(&displayHeight, "height", 768, "height of screen")
	flag.IntVar(&xadjust, "xadjust", 0, "ladjust")
	flag.IntVar(&yadjust, "yadjust", 0, "radjust")
	flag.Parse()

	http.HandleFunc("/mousedata", func(w http.ResponseWriter, r *http.Request) {
		params, err := parseNums(r, intParams)
		if err != nil {
			log.Printf("Error: parsing int params in handleMouse: %s", err)
			return
		}

		log.Printf("%s: %v", r.FormValue(evtType), params)

		switch r.FormValue(evtType) {
		case "pan":
			boxLeft := (params[scrW] - params[boxW]) / 2
			boxRight := params[scrW] - boxLeft
			boxTop := (params[scrH] - params[boxH]) / 2
			boxBtm := params[scrH] - boxTop

			// if any inputs outside the bounding box, ignore
			if params[curX] < boxLeft || params[curX] > boxRight || params[curY] < boxTop || params[curY] > boxBtm {
				break
			}

			if params[deltaX] < params[boxW]/10 {
				params[deltaX] = 0
			}

			if params[deltaY] < params[boxH]/10 {
				params[deltaY] = 0
			}

			xPer := float32(params[deltaX]) / float32(params[boxW])
			yPer := float32(params[deltaY]) / float32(params[boxH])

			x := int(xPer * float32(displayWidth) * 0.2)
			y := int(yPer * float32(displayHeight) * 0.2)

			log.Printf("xPer:%.2f yPer:%.2f x:%d y:%d", xPer, yPer, x, y)

			exec.Command("xdotool", "mousemove_relative", "--", strconv.Itoa(x), strconv.Itoa(y)).Run()
		case "tap":
			exec.Command("xdotool", "click", "1").Run()
		default:
			log.Printf("Error: unsupported event type: %s", evtType)
		}
	})
	http.HandleFunc("/", serveAsset)
	http.ListenAndServe(hostport, nil)

}
