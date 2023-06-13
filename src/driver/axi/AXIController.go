package axi

import (
	"os"
	"syscall"
	"unsafe"
)

const defaultMaxFileSizeInt64 = 2
const defaultMaxFileSizeInt32 = 0x3F
const defaultMemMapSize = 16

type AXIBusController struct {
	Data8B  *[defaultMaxFileSizeInt64]uint64
	data4B  *[defaultMaxFileSizeInt32]uint32
	dataRef []byte
}

func (axictl *AXIBusController) Open(offset int64) {
	f, _ := os.OpenFile("/dev/mem", os.O_SYNC|os.O_RDWR, 0644)
	b, _ := syscall.Mmap(int(f.Fd()), offset, defaultMemMapSize, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	axictl.dataRef = b
	axictl.Data8B = (*[defaultMaxFileSizeInt64]uint64)(unsafe.Pointer(&b[0]))
	axictl.data4B = (*[defaultMaxFileSizeInt32]uint32)(unsafe.Pointer(&b[0]))
}

func (axictl *AXIBusController) Read(readOffset uint32) uint32 {
	return axictl.data4B[readOffset>>2]
}
func (axictl *AXIBusController) Write(writeOffset uint32, data uint32) {
	axictl.data4B[writeOffset>>2] = data
}

func (axictl *AXIBusController) Close() {
	axictl.Data8B = nil
	axictl.data4B = nil
	axictl.dataRef = nil
}

func NewAXIController(AXIAddress int64) *AXIBusController {
	AXICTL := &AXIBusController{}
	AXICTL.Open(AXIAddress)

	return AXICTL
}
