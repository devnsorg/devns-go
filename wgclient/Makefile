apiEndpoint = http://192.168.0.126:9999/wg
piSSHAddress = pi@192.168.0.126

define build_arch_os
	GOARCH=$(1) GOOS=$(2) go build -o "./bin/wgclient-$(1)-$(2)"
	chmod a+x "./bin/wgclient-$(1)-$(2)"
endef

build: build_linux build_darwin build_arm

build_arm:
	$(call build_arch_os,arm,linux)
build_linux:
	$(call build_arch_os,amd64,linux)
build_darwin:
	$(call build_arch_os,amd64,darwin)
deploy_and_test_arm: build_arm
	scp -r bin/wgclient-arm-linux ${piSSHAddress}:./Downloads/
	ssh ${piSSHAddress} sudo Downloads/wgclient-arm-linux --api-endpoint ${apiEndpoint}
test_darwin: build_darwin
	bin/wgclient-amd64-darwin --api-endpoint ${apiEndpoint}
all: build