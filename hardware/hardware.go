package hardware

import (
	"bytes"
	"html/template"
)

const (
	SYSTEM_TMPL  = "./templates/systemTmpl.html"
	DISK_TMPL    = "./templates/diskTmpl.html"
	CPU_TMPL     = "./templates/cpuTmpl.html"
	PROCESS_TMPL = "./templates/processTmpl.html"
	NET_TMPL     = "./templates/netTmpl.html"
	TMPL         = "./templates/tmpl.html"
)

type Hardware struct {
	SysInfo     *SystemInfo
	DiskInfo    *DiskInfo
	CpuInfo     *CpuInfo
	ProcessInfo *Processes
	NetInfo     *Connections
}

func NewHardware() *Hardware {
	return &Hardware{
		SysInfo:     NewSystemInfo(),
		DiskInfo:    NewDiskInfo(),
		CpuInfo:     NewCpuInfo(),
		ProcessInfo: NewProcesses(),
		NetInfo:     NewConnections(),
	}
}

func (hardware *Hardware) String() string {
	str := "\t\t\t---Hardware Information---\n"

	str += hardware.SysInfo.String() + "\n"
	str += hardware.DiskInfo.String() + "\n"
	str += hardware.CpuInfo.String() + "\n"
	str += hardware.ProcessInfo.String() + "\n"
	str += hardware.NetInfo.String()

	return str
}

func (hardware *Hardware) ToHtml(tmplPath string) (string, error) {
	//Get the template
	tmpl, err := template.New("tmpl.html").ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	//Get all the template string
	sysTmpl, err := hardware.SysInfo.ToHtml(SYSTEM_TMPL)
	if err != nil {
		return "", err
	}

	diskTmpl, err := hardware.DiskInfo.ToHtml(DISK_TMPL)
	if err != nil {
		return "", err
	}

	cpuTmpl, err := hardware.CpuInfo.ToHtml(CPU_TMPL)
	if err != nil {
		return "", err
	}

	processesTmpl, err := hardware.ProcessInfo.ToHtml(PROCESS_TMPL)
	if err != nil {
		return "", err
	}

	netTmpl, err := hardware.NetInfo.ToHtml(NET_TMPL)
	if err != nil {
		return "", err
	}

	// Use template.HTML instead of string to prevent HTML escaping
	data := struct {
		SysTmpl       template.HTML
		DiskTmpl      template.HTML
		CpuTmpl       template.HTML
		ProcessesTmpl template.HTML
		NetTmpl       template.HTML
	}{
		SysTmpl:       template.HTML(sysTmpl),
		DiskTmpl:      template.HTML(diskTmpl),
		CpuTmpl:       template.HTML(cpuTmpl),
		ProcessesTmpl: template.HTML(processesTmpl),
		NetTmpl:       template.HTML(netTmpl),
	}

	//Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (hardware *Hardware) CollectData() error {
	var err error
	err = hardware.SysInfo.GetSystemInfo()
	if err != nil {
		return err
	}

	err = hardware.DiskInfo.GetDiskInfo()
	if err != nil {
		return err
	}

	err = hardware.CpuInfo.GetCPUInfo(0)
	if err != nil {
		return err
	}

	err = hardware.ProcessInfo.GetAllProcessInfo()
	if err != nil {
		return err
	}

	err = hardware.NetInfo.GetAllConnection()
	if err != nil {
		return err
	}

	return nil
}
