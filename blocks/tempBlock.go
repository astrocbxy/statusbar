package blocks

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/astrocbxy/statusbar"
)

type TempBlock struct {
	block      *statusbar.I3Block
	sensorPath string
	highTemp   int
	label      string
}

func (this *TempBlock) Init(block *statusbar.I3Block, resp *statusbar.Responder) bool {
	this.block = block

	var preferredSensors = [...]string{"thinkpad", "coretemp"}

	// Try to figure out the correct sensor
	if _, err := os.Stat("/sys/class/hwmon"); os.IsNotExist(err) {
		return false // No hwmon
	}
	sensors, err := ioutil.ReadDir("/sys/class/hwmon")
	if err != nil {
		return false // What?
	}
SensorSearch:
	for _, sensor := range sensors {
		// Check if a we have a temperature sensor
		if _, err := os.Stat("/sys/class/hwmon/" + sensor.Name() + "/temp1_input"); os.IsNotExist(err) {
			continue // No temperature sensor here
		}

		// Use this sensor
		this.sensorPath = "/sys/class/hwmon/" + sensor.Name()

		// Handle name (preferred sensors)
		rawName, err := ioutil.ReadFile(this.sensorPath + "/name")
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(string(rawName), "\n")
		for _, cur := range preferredSensors {
			if name == cur {

				// Try to get label name
				rawName, err := ioutil.ReadFile(this.sensorPath + "/temp1_label")
				if err != nil {
					this.label = ""
				} else {
					this.label = strings.TrimSuffix(string(rawName), "\n")
				}

				// Try to find high temperature
				if _, err := os.Stat(this.sensorPath + "/temp1_max"); os.IsNotExist(err) {
					rawTemp, err := ioutil.ReadFile(this.sensorPath + "/temp1_max")
					if err != nil {
						this.highTemp = 0
					} else {
						this.highTemp, _ = strconv.Atoi(string(rawTemp))
					}
				} else {
					this.highTemp = 0
				}

				break SensorSearch
			}
		}
	}

	return true
}

func (this TempBlock) Tick() {
	rawTemp, err := ioutil.ReadFile(this.sensorPath + "/temp1_input")
	if err != nil {
		this.block.FullText = "ERROR"
		this.block.Color = "#ff0202"
		return
	}
	temp, err := strconv.Atoi(strings.TrimSpace(string(rawTemp)))
	if err != nil {
		this.block.FullText = "ERROR"
		this.block.Color = "#ff0202"
		return
	}
	if this.highTemp > 0 && temp >= this.highTemp {
		this.block.Color = "#ff0202"
	} else {
		this.block.Color = ""
	}

	this.block.FullText = this.label + ": " + strconv.Itoa(temp/1000) + "°C"
}

func (this TempBlock) Click(data statusbar.I3Click) {
	cmd := exec.Command("dunstify", "test")
	cmd.Run()
}

func (this TempBlock) Block() *statusbar.I3Block {
	return this.block
}
