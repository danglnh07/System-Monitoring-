package hardware

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"

	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	PID                int32   //Process ID
	Name               string  //Process name
	NumberOfThreadUsed int32   //Number of threads that process currently used
	CpuUsagePercent    float64 //The CPU usage of that process
	MemoryUsed         uint64  //The amount of memory the current process is holding in RAM (not including swap)
}

func NewProcessInfo() *ProcessInfo {
	return &ProcessInfo{}
}

func (procInfo *ProcessInfo) String() string {
	str := fmt.Sprintf("PID: %d\n", procInfo.PID)
	str += fmt.Sprintf("Process name: %s\n", procInfo.Name)
	str += fmt.Sprintf("Number of thread used: %d\n", procInfo.NumberOfThreadUsed)
	str += fmt.Sprintf("CPU Usage: %.2f%%\n", procInfo.CpuUsagePercent)
	str += fmt.Sprintf("Memory used: %s", ConvertByte(procInfo.MemoryUsed))

	return str
}

func (procInfo *ProcessInfo) GetProcessInfo(runningProc *process.Process) error {
	var err error

	//Get process PID
	procInfo.PID = runningProc.Pid

	//Get process's name
	procInfo.Name, err = runningProc.Name()
	if err != nil {
		return err
	}

	//Get the number of threads used by a process
	procInfo.NumberOfThreadUsed, err = runningProc.NumThreads()
	if err != nil {
		return err
	}

	//Get the cpu usage (%) of that process
	procInfo.CpuUsagePercent, err = runningProc.CPUPercent()
	if err != nil {
		return err
	}

	//Get the memory information of that process
	memoryStat, err := runningProc.MemoryInfo()
	if err != nil {
		return err
	}
	/*
	 * Resident Set Size (RSS), which is the portion of a process's memory that is held in RAM
	 * (i.e., not swapped to disk).
	 * It represents the actual physical memory being used by the process.
	 * VMS: Virtual Memory Size (total allocated memory, including swapped out).
	 */
	procInfo.MemoryUsed = memoryStat.RSS //Get the current memory that process is taken in RAM

	return err
}

type Processes []ProcessInfo

func NewProcesses() *Processes {
	return &Processes{}
}

func (processes *Processes) String() string {
	str := "\t\t---Process Infomation---\n"
	for _, processInfo := range *processes {
		str += fmt.Sprintf("%s\n---\n", processInfo.String())
	}
	return str
}

func (processes Processes) Len() int {
	return len(processes)
}

func (processes Processes) Less(i, j int) bool {
	/*
	 * A process is evaluated by its threads used, cpu usage and memory usage
	 * Evaluation: 20% * Threads used + 40% * CPU Usage + 40% * Memory Usage
	 * The sort is descending
	 */
	proc1, proc2 := processes[i], processes[j]
	stat1 := float64(proc1.NumberOfThreadUsed)*0.2 + proc1.CpuUsagePercent*0.4 + float64(proc1.MemoryUsed)*0.4
	stat2 := float64(proc2.NumberOfThreadUsed)*0.2 + proc2.CpuUsagePercent*0.4 + float64(proc2.MemoryUsed)*0.4

	return stat1 > stat2
}

func (processes Processes) Swap(i, j int) {
	processes[i], processes[j] = processes[j], processes[i]
}

func (processes *Processes) ToHtml(tmplPath string) (string, error) {
	//Func map
	funcMap := template.FuncMap{
		"ConvertByte": ConvertByte,
	}

	//Get the template
	tmpl, err := template.New("processTmpl.html").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	//Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, processes)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (processes *Processes) GetAllProcessInfo() error {
	//Clean the processes to avoid duplicate
	*processes = (*processes)[:0]

	//Get all running processes
	runningProcesses, err := process.Processes()
	if err != nil {
		return err
	}

	procInfo := NewProcessInfo()

	for _, runningProc := range runningProcesses {
		//If we find some process with PID = 0, ignore them (PID = 0 usually idle, which is not what we want to track)
		if runningProc.Pid > 0 {
			err = procInfo.GetProcessInfo(runningProc)
			//If we get some error getting information about 1 process, then we just ignore it and continue
			if err != nil {
				continue
			}
			*processes = append(*processes, *procInfo)
		}
	}

	//Sort the processes based on threads used, CPU usage and memory usage
	sort.Sort(processes)

	return err
}
