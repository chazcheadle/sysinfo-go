package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// System structure to hold an instance of the system information.
type System struct {
	Host struct {
		Hostname        string `json:"hostname"`
		OS              string `json:"os"`
		Platform        string `json:"platform"`
		PlatformFamily  string `json:"platformFamily"`
		PlatformVersion string `json:"platformVersion"`
		KernelVersion   string `json:"kernelVersion"`
	} `json:"host"`
	CPU struct {
		Cores int `json:"cores"`
		Load  struct {
			Load1  float64 `json:"load1"`
			Load5  float64 `json:"load5"`
			Load15 float64 `json:"load15"`
		} `json:"load"`
	} `json:"cpu"`
	Mem struct {
		Total       uint64  `json:"total"`
		Free        uint64  `json:"free"`
		Available   uint64  `json:"available"`
		UsedPercent float64 `json:"usedPercent"`
	} `json:"memory"`
	Disks map[string]disk
	Error string `json:"error"`
}

type disk struct {
	All  uint64
	Free uint64
	Used uint64
}

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// DiskUsage of path/disk
func DiskUsage(path string) (diskstatus DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	diskstatus.All = fs.Blocks * uint64(fs.Bsize)
	diskstatus.Free = fs.Bfree * uint64(fs.Bsize)
	diskstatus.Used = diskstatus.All - diskstatus.Free
	return
}

func sysHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Retrieve Sol data.
	sys, err := getSysData()
	if err != nil {
		// Send Bad Request status code and Error message.
		http.Error(w, sys.Error, 400)
	} else {
		// Convert Sol struct to JSON for output.
		buffer, _ := json.MarshalIndent(sys, "", "    ")

		// Send proper JSON reponse to ResponseWriter.
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(buffer)
	}
}

func getSysData() (sys *System, err error) {

	// Initialize and instance of System.
	sys = &System{}

	// Explicitly list drives to be checked.
	drives := []string{"/"}
	// Create the map for sys.Disks.
	sys.Disks = make(map[string]disk)
	// Iterate over the drives.
	for _, drive := range drives {
		// Get drive usage information.
		dUsage := DiskUsage(drive)
		// Create a temporary instance of a disk to store data.
		d := &disk{}
		d.All = dUsage.All
		d.Free = dUsage.Free
		d.Used = dUsage.Used
		// create an element in the slice with the drive name and data.
		sys.Disks[drive] = *d
	}

	virtualMemory, _ := mem.VirtualMemory()
	sys.Mem.Total = virtualMemory.Total
	sys.Mem.Free = virtualMemory.Free
	sys.Mem.Available = virtualMemory.Available
	sys.Mem.UsedPercent = virtualMemory.UsedPercent

	cpu, _ := cpu.Info()
	sys.CPU.Cores = len(cpu)

	loadAvg, _ := load.Avg()
	sys.CPU.Load.Load1 = loadAvg.Load1
	sys.CPU.Load.Load5 = loadAvg.Load5
	sys.CPU.Load.Load15 = loadAvg.Load15

	hostInfo, _ := host.Info()
	sys.Host.Hostname = hostInfo.Hostname
	sys.Host.OS = hostInfo.OS
	sys.Host.Platform = hostInfo.Platform
	sys.Host.PlatformFamily = hostInfo.PlatformFamily
	sys.Host.PlatformVersion = hostInfo.PlatformVersion
	sys.Host.KernelVersion = hostInfo.KernelVersion

	buffer, _ := json.MarshalIndent(sys, "", "    ")

	fmt.Println(string(buffer))

	return sys, nil
}

func main() {
	// Instantiate a new router.
	router := httprouter.New()

	router.GET("/sys", sysHandler)
	http.ListenAndServe(":3002", router)

}
