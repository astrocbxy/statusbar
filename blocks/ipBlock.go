package blocks

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/astrocbxy/statusbar"
)

type IpBlock struct {
	block  *statusbar.I3Block
	failed bool
}

func (this *IpBlock) Init(block *statusbar.I3Block, resp *statusbar.Responder) bool {
	this.block = block
	return true
}

func (this IpBlock) Tick() {
	if this.failed {
		return
	}

	this.block.FullText = ""
	this.block.Color = ""

	f, err := os.Open("/proc/net/route")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		this.failed = true
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		// Only consider default routes (to 0.0.0.0)
		if parts[1] == "00000000" {
			iface := parts[0]
			ipAddress := this.GetInterfaceIp(iface)

			// get SSID if wifi network
			ssid := ""
			if this.CommandExists("nmcli") {
				ssid = this.GetNmSsid(iface)
			} else if this.CommandExists("wpa_cli") {
				ssid = this.GetWpaSupplSsid()
			}
			if ssid != "" {
				ssid = " - " + ssid
			}

			// Write to Block
			if this.block.FullText == "" {
				this.block.FullText = iface + ssid + " - " + ipAddress
			} else {
				this.block.FullText += " " + iface + ssid + " - " + ipAddress
			}
		}
	}
	if this.block.FullText == "" {
		this.block.FullText = "No Link"
		this.block.Color = "#ff0202"
	}
}

func (this IpBlock) GetInterfaceIp(ifaceName string) string {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		this.failed = true
	}
	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		this.failed = true
	}
	for _, address := range addrs {
		// check the address type and that it's not a loopback
		if ipnet, err := address.(*net.IPNet); err && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "No IP"
}

func (this IpBlock) GetWpaSupplSsid() string {
	out, err := exec.Command("bash", "-c", "wpa_cli status | grep 'ssid=' | cut -d'=' -f2").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		this.failed = true
		return ""
	}
	outString := string(out)
	if len([]rune(outString)) > 0 {
		return outString
	}
	return ""
}

func (this IpBlock) GetNmSsid(iface string) string {
	out, err := exec.Command("bash", "-c", "nmcli connection show | grep "+iface+" | grep wifi").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		this.failed = true
		return ""
	}
	outString := string(out)
	if strings.Contains(outString, iface) {
		return strings.Fields(outString)[0]
	}
	return ""
}

func (this IpBlock) CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (this IpBlock) Click(data statusbar.I3Click) {
}

func (this IpBlock) Block() *statusbar.I3Block {
	return this.block
}
