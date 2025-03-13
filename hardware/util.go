package hardware

import "fmt"

const (
	GB float64 = 1024 * 1024 * 1024
	MB float64 = 1024 * 1024
	KB float64 = 1024
)

func ConvertByte(value uint64) string {
	//We always expect to receive value in Byte, so we just have to convert it into 1 level higher unit
	switch {
	case value > (1024 * 1024 * 1024): //If the memory value is more than 1GB, convert to GB
		return fmt.Sprintf("%.2f GB", float64(value)/GB)
	case value > (1024 * 1024): //If the memory value is more than 1MB, convert to MB
		return fmt.Sprintf("%.2f MB", float64(value)/MB)
	case value > 1024: //If the memory value is more than 1KB, convert to KB
		return fmt.Sprintf("%.2f KB", float64(value)/KB)
	default: //If this not exceed a KB, then just keep the Byte value
		return fmt.Sprintf("%d B", uint32(value))
	}
}
