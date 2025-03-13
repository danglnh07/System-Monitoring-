package hardware

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/shirou/gopsutil/disk"
)

type PartitionInfo struct {
	DeviceName string //Curent partition
	Total      uint64 //Total size
	Free       uint64 //Free storage remain
}

func NewPartitionInfo() *PartitionInfo {
	return &PartitionInfo{}
}

func (parInfo *PartitionInfo) String() string {
	str := fmt.Sprintf("Partition: %s\n", parInfo.DeviceName)
	str += fmt.Sprintf("Total size: %s\n", ConvertByte(parInfo.Total))
	str += fmt.Sprintf("Free size: %s\n", ConvertByte(parInfo.Free))
	return str
}

type DiskInfo []PartitionInfo

func NewDiskInfo() *DiskInfo {
	return &DiskInfo{}
}

func (diskInfo *DiskInfo) String() string {
	str := "\t\t---Disk Information---\n"
	for _, partition := range *diskInfo {
		str += partition.String() + "---\n"
	}
	return str
}

func (diskInfo *DiskInfo) ToHtml(tmplPath string) (string, error) {
	//Func map
	funcMap := template.FuncMap{
		"ConvertByte": ConvertByte,
	}

	//Get the template
	tmpl, err := template.New("diskTmpl.html").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	//Execute template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, diskInfo)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (diskInfo *DiskInfo) GetDiskInfo() error {
	//Clean the disk info before processing
	*diskInfo = (*diskInfo)[:0]

	//We get all the partition in the system (only physical devices like hard disks, CDROM,...)
	partitions, err := disk.Partitions(false) //false mean only physical devices
	if err != nil {
		return err
	}

	var (
		diskStat *disk.UsageStat
		parInfo  = NewPartitionInfo()
	)

	//For each partition, we loop through each and get their stat
	for _, partition := range partitions {
		diskStat, err = disk.Usage(partition.Mountpoint)
		if err != nil {
			return err
		}
		parInfo.DeviceName = partition.Device
		parInfo.Total = diskStat.Total
		parInfo.Free = diskStat.Free

		*diskInfo = append(*diskInfo, *parInfo)
	}

	return nil
}
