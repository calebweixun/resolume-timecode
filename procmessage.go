package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chabad360/go-osc/osc"
)

var (
	clipName         = ""
	directionForward = true

	timeLeft string

	clipLength float32
	posPrev    float32

	activeClipPath string
)

func procMsg(data *osc.Message) {
	if !strings.HasPrefix(data.Address, clipPath) {
		return
	}

	if strings.HasSuffix(data.Address, "/connected") {
		procConnected(data)
		return
	}

	if !strings.HasPrefix(data.Address, monitoredClipPath()) {
		return
	}

	switch {
	case strings.HasSuffix(data.Address, "/position"):
		procPos(data)
	case strings.HasSuffix(data.Address, "/direction"):
		procDirection(data)
	case strings.HasSuffix(data.Address, "/name"):
		procName(data)
	case strings.HasSuffix(data.Address, "/duration"):
		procDuration(data)
	case strings.HasSuffix(data.Address, "/connect"):
		reset()
	case strings.Contains(data.Address, "/select"):
		reset()
	}
}

func procDirection(data *osc.Message) {
	value, ok := firstNumber(data)
	if !ok {
		return
	}
	directionForward = value != 0
	if !directionForward {
		posPrev = 1 - posPrev
	}
}

func procName(data *osc.Message) {
	if len(data.Arguments) == 0 {
		return
	}
	name, ok := data.Arguments[0].(string)
	if !ok {
		return
	}
	clipName = name
	clipNameBinding.Set("Clip Name: " + clipName)
	broadcast.Publish(osc.NewMessage("/name", clipName))
}

func procDuration(data *osc.Message) {
	value, ok := firstNumber(data)
	if !ok {
		return
	}
	clipLength = (value * 604800) + 0.001
	clipLengthBinding.Set(fmt.Sprintf("Clip Length: %.3fs", clipLength))
	broadcast.Publish(osc.NewMessage("/duration", clipLength))
}

func reset() {
	activeClipPath = ""
	requestClipMetadata()
	requestLayerClipDiscovery()

	posPrev = 0
	broadcast.Publish(osc.NewMessage("/reset"))
	broadcast.Send()
}

// requestClipMetadata keeps the selected clip name and duration synchronized
// even when Resolume does not proactively publish OSC state changes.
func requestClipMetadata() {
	path := monitoredClipPath()
	name := osc.NewMessage(path+"/name", "?")
	duration := osc.NewMessage(path+"/transport/position/behaviour/duration", "?")
	if _, err := oscServer.WriteTo(osc.NewBundle(name, duration), OSCAddr+":"+OSCPort); err != nil {
		fmt.Println(err)
	}
}

// requestPosition actively polls Resolume so the clock also works when its OSC
// output preset is not configured to continuously publish transport position.
func requestPosition() {
	position := osc.NewMessage(monitoredClipPath()+"/transport/position", "?")
	if _, err := oscServer.WriteTo(position, OSCAddr+":"+OSCPort); err != nil {
		fmt.Println(err)
	}
}

// monitoredClipPath maps a layer path to its currently playing clip. Once a
// connected clip is discovered, its absolute path is preferred.
func monitoredClipPath() string {
	if activeClipPath != "" {
		return activeClipPath
	}
	if layerPath, ok := monitoredLayerPath(clipPath); ok {
		return layerPath + "/clips/playing"
	}
	return clipPath
}

func monitoredLayerPath(path string) (string, bool) {
	path = strings.TrimSuffix(path, "/")
	if path == "/composition/selectedlayer" {
		return path, true
	}
	const prefix = "/composition/layers/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	layer := strings.TrimPrefix(path, prefix)
	if strings.Contains(layer, "/") {
		return "", false
	}
	n, err := strconv.Atoi(layer)
	return path, err == nil && n > 0
}

func requestLayerClipDiscovery() {
	layerPath, ok := monitoredLayerPath(clipPath)
	if !ok || layerPath == "/composition/selectedlayer" {
		return
	}
	connected := osc.NewMessage(layerPath+"/clips/*/connected", "?")
	if _, err := oscServer.WriteTo(connected, OSCAddr+":"+OSCPort); err != nil {
		fmt.Println(err)
	}
}

func procConnected(data *osc.Message) {
	value, ok := firstNumber(data)
	if !ok {
		return
	}
	path := strings.TrimSuffix(data.Address, "/connected")
	if value == 0 {
		if activeClipPath == path {
			activeClipPath = ""
			posPrev = 0
		}
		return
	}
	// A transition can report multiple connected clips in one wildcard query.
	// Keep the current target until Resolume explicitly disconnects it instead
	// of oscillating between candidates on every discovery pass.
	if activeClipPath != "" {
		return
	}
	activeClipPath = path
	posPrev = 0
	requestClipMetadata()
}

func firstNumber(data *osc.Message) (float32, bool) {
	if len(data.Arguments) == 0 {
		return 0, false
	}
	switch value := data.Arguments[0].(type) {
	case float32:
		return value, true
	case float64:
		return float32(value), true
	case int32:
		return float32(value), true
	case int64:
		return float32(value), true
	case int:
		return float32(value), true
	default:
		return 0, false
	}
}

func procPos(data *osc.Message) {
	pos, ok := firstNumber(data)
	if !ok {
		return
	}

	if !directionForward {
		pos = 1 - pos
	}

	// Position is absolute, so every valid reply can update the clock. The old
	// interval filter permanently stalled after loops, seeks, and clip changes.
	posPrev = pos

	if clipInvert {
		pos = 1 - pos
	}

	t := (clipLength * 1000) * (1 - pos)

	timeActual := time.UnixMilli(int64(t)).UTC()

	timeLeft = fmt.Sprintf("-%02d:%02d:%02d.%03d", timeActual.Hour(), timeActual.Minute(), timeActual.Second(), timeActual.Nanosecond()/1000000)
	broadcast.Publish(osc.NewMessage("/time", timeLeft, fmt.Sprintf("%.3fs", clipLength)))
	broadcast.Send()

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}
