package hcxtools

/*
#cgo LDFLAGS: -L. -lhcxpcapngtool
#include <stdlib.h>

// Declare the convert_pcap function from the shared library
int convert_pcap(int argc, char *argv[]);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func ConvertPCAPToHashcatFormat(inputFile, outputFile string) error {
	// Prepare arguments for the convert_pcap function
	args := []string{"", inputFile, "-o", outputFile}
	argc := C.int(len(args))
	argv := make([]*C.char, len(args))

	// Convert Go string slices to C strings
	for i, arg := range args {
		argv[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(argv[i])) // Free memory once done
	}

	// Call the convert_pcap function from the shared library
	ret := C.convert_pcap(argc, &argv[0])
	if ret != 0 {
		return fmt.Errorf("hcxpcapngtool conversion failed with code %d", ret)
	}

	return nil
}
