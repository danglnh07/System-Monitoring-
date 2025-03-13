package hardware

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

type CpuInfo struct {
	Model         string    //Model name of the CPU
	Family        string    //Model family of the CPU
	MHz           float64   //CPU running frequency
	CacheSize     uint64    //Cache size
	TotalUsage    float64   //Total CPU usage
	UsagePerCores []float64 //Each core usage
	Load1         float64   //Average load (short-term load)
	Load5         float64   //Average load (mid-term load)
	Load15        float64   //Average load (long-term load)
}

func NewCpuInfo() *CpuInfo {
	return &CpuInfo{}
}

func (cpuInfo *CpuInfo) String() string {
	str := "\t\t---CPU Information---\n"

	str += fmt.Sprintf("CPU model name: %s\n", cpuInfo.Model)
	str += fmt.Sprintf("CPU model family: %s\n", cpuInfo.Family)
	str += fmt.Sprintf("Run at: %.2f MHz\n", cpuInfo.MHz)
	str += fmt.Sprintf("Cache size: %s\n", ConvertByte(cpuInfo.CacheSize))
	str += fmt.Sprintf("Total CPU usage: %.2f%%\n", cpuInfo.TotalUsage)
	str += "Usage per cores: \n"
	for core, usage := range cpuInfo.UsagePerCores {
		str += fmt.Sprintf("\tCore %d: %.2f%%\n", core, usage)
	}
	str += fmt.Sprintf("Load Avg: %.2f %.2f %.2f\n", cpuInfo.Load1, cpuInfo.Load5, cpuInfo.Load15)
	return str
}

func (cpuInfo *CpuInfo) ToHtml(tmplPath string) (string, error) {
	//Func map
	funcMap := template.FuncMap{
		"ConvertByte": ConvertByte,
	}

	//Get the template
	tmpl, err := template.New("cpuTmpl.html").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	//Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, cpuInfo)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (cpuInfo *CpuInfo) GetCPUInfo(interval time.Duration) error {
	//Get the CPU status
	cpuStat, err := cpu.Info()
	if err != nil {
		return err
	}

	/*
	 * The cpu.Info() method will return all cores (physical + multi threading/hyper threading)
	 * We only care about physical cores, which would be in the first element of the slice
	 */

	cpuInfo.Model = cpuStat[0].ModelName
	cpuInfo.Family = cpuStat[0].Family
	cpuInfo.MHz = cpuStat[0].Mhz
	cpuInfo.CacheSize = uint64(cpuStat[0].CacheSize)

	/*
	 * Get CPU usage:
	 * 1. The first paramter is the interval for which the package used to compare to get the usage
	 * If interval is 0, it will compare the value to the last call. If provide interval value, it will
	 * wait for an interval before calculating usage
	 * 2. The second parameter indicate if we want to get total CPU usage (false), or usage per core (true)
	 */

	//Get total cpu usage
	totalUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	cpuInfo.TotalUsage = totalUsage[0]

	//Get each core's usage
	coresUsage, err := cpu.Percent(interval, true)
	if err != nil {
		return err
	}
	cpuInfo.UsagePerCores = nil //Clear all remaining data before appending
	cpuInfo.UsagePerCores = append(cpuInfo.UsagePerCores, coresUsage...)

	/*
	 * Get the average load: Average load (or load average) is a measure of system activity over a period of time.
	 * It represents the average number of processes waiting for CPU time (or disk I/O) in a given time frame.
	 * On Linux, you typically see three load averages, which represent system load over: 1, 5 or 15 mins
	 * The load value is relative to the number of CPU cores:
	 * 1. If the load average is equal to the number of CPU cores, the system is fully utilized.
	 * 2. If it's below the number of CPU cores, the system is not overloaded.
	 * 3. If it's above the number of CPU cores, the system is overloaded.
	 */
	avg, err := load.Avg()
	if err != nil {
		return err
	}
	cpuInfo.Load1 = avg.Load1
	cpuInfo.Load5 = avg.Load5
	cpuInfo.Load15 = avg.Load15

	return nil
}
