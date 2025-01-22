build-client:
	go mod verify
	go mod tidy
	go build -o build/client -buildvcs=false
	mv libhcxpcapngtool.so build/libhcxpcapngtool.so
.PHONY: build-client

build-hcxtools:
	sed -i 's/int main(int argc, char \*argv\[\])/int convert_pcap(int argc, char *argv\[\])/' ../externals/hcxtools/hcxpcapngtool.c
	cc -fPIC -shared -o libhcxpcapngtool.so ../externals/hcxtools/hcxpcapngtool.c -lz -lssl -lcrypto -DVERSION_TAG=\"6.3.5\" -DVERSION_YEAR=\"2024\"
.PHONY: build-hcxtoolshcxtools

links-build:
	mkdir -p build
	@{ \
		if [ ! -e "build/OpenCL" ]; then \
			ln -s /usr/local/share/hashcat/OpenCL build/OpenCL; \
			echo "Created symbolic link for OpenCL"; \
		fi; \
		if [ ! -e "build/hashcat.hctune" ]; then \
			ln -s /usr/local/share/hashcat/hashcat.hctune build/hashcat.hctune; \
		fi; \
		if [ ! -e "build/modules" ]; then \
			ln -s /usr/local/share/hashcat/modules build/modules; \
		fi; \
	}
.PHONY: links-build

links:
	@{ \
		if [ ! -e "OpenCL" ]; then \
			ln -s /usr/local/share/hashcat/OpenCL OpenCL; \
			echo "Created symbolic link for OpenCL"; \
		fi; \
		if [ ! -e "hashcat.hctune" ]; then \
			ln -s /usr/local/share/hashcat/hashcat.hctune hashcat.hctune; \
		fi; \
		if [ ! -e "modules" ]; then \
			ln -s /usr/local/share/hashcat/modules modules; \
		fi; \
	}
.PHONY: links

clean:
	rm -rf OpenCL
	rm -rf modules
	rm -rf kernels
	rm -f *.hctune 
	rm -f *.dictstat*
	rm -f main
.PHONY: clean

install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2
.PHONY: install-linter

run-compiled:
	cd build && \
	LD_LIBRARY_PATH=.:$LD_LIBRARY_PATH ./client && \
	cd ..
.PHONY: run-compiled

go-vuln:
	govulncheck ./...
.PHONY: go-vuln

lint:	
	@golangci-lint run
.PHONY: lint

proto: ## Builds proto files
	@../proto.sh
.PHONY: proto

test: ## Starts tests locally
	go test ./... --count=1 -p 1 # -p 1 do not run tests parallel. Avoid race conditions when starting/tearing down tests
.PHONY: test

build: proto links-build build-hcxtools build-client
run: build run-compiled