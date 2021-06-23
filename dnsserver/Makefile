wgEndpoint = localhost:51820
piSSHAddress = pi@192.168.0.126
version ?= 0.0.0

define build_arch_os
	GOARCH=$(1) GOOS=$(2) go build -o "./bin/dnsserver-$(1)-$(2)-$(version)"
	chmod a+x "./bin/wgapi-$(1)-$(2)"
endef

build: build_linux build_darwin build_arm

build_arm:
	$(call build_arch_os,arm,linux)
build_linux:
	$(call build_arch_os,amd64,linux)
build_darwin:
	$(call build_arch_os,amd64,darwin)
test_arm: build_arm
	scp -r bin/wgapi-arm-linux ${piSSHAddress}:./Downloads/
	ssh ${piSSHAddress} sudo Downloads/wgapi-arm-linux --wg-endpoint ${wgEndpoint}
test_darwin: build_darwin
	bin/wgapi-amd64-darwin --wg-endpoint ${wgEndpoint}
release: build