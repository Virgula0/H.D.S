Clients communicate with the server using **`gRPC`**. This enables clients to identify whether they are the intended receiver for a specific cracking task.

The communication channel also supports a **bidirectional stream**, allowing the server and client to exchange messages in real time during `hashcat` execution.

---

## **What Does the Client Do?**

A client performs the following tasks:

1. **Waits for tasks** from the server.
2. Upon receiving a task, it **acknowledges the server**.
3. The server then **removes the task from the `pending` queue** and updates its status. Meanwhile, the client saves the **base64-encoded hash file** into a temporary directory.
4. Once saved as a **`.PCAP` file**, the client converts it into a **hash format compatible with `hashcat`**.
5. The client uses **`hcxtools`** for the conversion. This library supports multiple operations on `.PCAP` files and beyond.
6. After conversion, **`hashcat` begins execution**, applying user-defined or default options.
7. **Logs and status updates** generated by `hashcat` are sent asynchronously to the server.
8. If `hashcat` successfully cracks the password, the **result is sent back to the server**.
9. The client then resets itself and **waits for the next task**.

---

## **Gocat**

The client uses the **`gocat`** dependency to execute `hashcat` from within Go. Since `hashcat` is written in **C**, a **porting layer** was required to bridge the two environments.

---

## **Hcxtools**

For our use case, we rely specifically on **`hcxpcapngtool`** from the `hcxtools` suite.

This tool doesn’t natively support building as a shared library. To work around this limitation and enable its integration with Go, we **modified its entry point** using `sed`:

```bash
sed -i 's/int main(int argc, char \*argv\[\])/int convert_pcap(int argc, char *argv\[\])/' hcxpcapngtool.c
```

This command replaces the standard `main` function signature with `convert_pcap`. We then compile it into a shared library:

```bash
cc -fPIC -shared -o /app/client/libhcxpcapngtool.so /app/hcxtools/hcxpcapngtool.c -lz -lssl -lcrypto -DVERSION_TAG=\"6.3.5\" -DVERSION_YEAR=\"2024\"
```

This shared library can now be directly imported and used in Go:

> [!NOTE]  
> **File:** `client/internal/hcxtools/hcxpcapngtool.go`

```go
/*
#cgo LDFLAGS: -L../../ -lhcxpcapngtool
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
		defer C.free(unsafe.Pointer(argv[i]))
	}

	// Call the convert_pcap function from the shared library
	ret := C.convert_pcap(argc, &argv[0])
	if ret != 0 {
		return fmt.Errorf("hcxpcapngtool conversion failed with code %d", ret)
	}

	return nil
}
```

While this solution works for our current requirements, future improvements could include **porting the library fully to Go**. However, this is considered **out of scope** for the current project.

---

## **Compile and Run**

> [!IMPORTANT]  
> The following dependencies needs to be installed before proceeding, even if you're using compiled binaries from releases

```bash
sudo apt update -y && \
sudo apt install -y --no-install-recommends \
    protobuf-compiler \
    libminizip-dev \
    ocl-icd-libopencl1 \
    opencl-headers \
    git \
    git-lfs \ 
    pocl-opencl-icd \
    build-essential \
    ca-certificates \
    libz-dev \
    libssl-dev \
    dbus \
    libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev
```

> [!IMPORTANT]  
> The file `/etc/machine-id` must exist on your machine.

Follow these steps to compile and run the client, run it from project root dir

```bash
git submodule update --init --remote --recursive && \
git lfs install && \  
git lfs pull
```

1. **You need to install `hashcat` 6.1.1. This step is necesary only for the first time.**

```bash
cd externals/gocat
sudo make install
sudo make set-user-permissions USER=${USER}
cd ../../
```

2. **Install protobuf**

> [!NOTE]
> This was tested out using go `1.23.4`. Other version may have problems.

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 &&
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```

3. **Build client**

```bash
cd client
make build
```

Produces the following files tree in `build`

```
├── client
├── hashcat.hctune -> /usr/local/share/hashcat/hashcat.hctune
├── hashcat.hcstat2 -> /usr/local/share/hashcat/hashcat.hcstat2
├── libhcxpcapngtool.so
├── modules -> /usr/local/share/hashcat/modules
└── OpenCL -> /usr/local/share/hashcat/OpenCL
```

4. **Run with**

```bash
make run-compiled
```

but remember to set these env variables first

```bash
export GRPC_URL=localhost:7777 # change with gRPC address
export GRPC_TIMEOUT=10s #leave this timeout by default
export LD_LIBRARY_PATH=.:$LD_LIBRARY_PATH 
```