default: test build

fmt:
	go fmt github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

# Peform a development (current-platform-only) build.
dev: version fmt
	go build -o _bin/terraform-provider-ddcloud

# Perform a full (all-platforms) build.
build: version build-windows64 build-linux64 build-mac64

build-windows64:
	GOOS=windows GOARCH=amd64 go build -o _bin/windows-amd64/terraform-provider-ddcloud.exe

build-linux64:
	GOOS=linux GOARCH=amd64 go build -o _bin/linux-amd64/terraform-provider-ddcloud

build-mac64:
	GOOS=darwin GOARCH=amd64 go build -o _bin/darwin-amd64/terraform-provider-ddcloud

# Produce archives for a GitHub release.
dist: build
	zip -9 _bin/windows-amd64.zip _bin/windows-amd64/terraform-provider-ddcloud.exe
	zip -9 _bin/linux-amd64.zip _bin/linux-amd64/terraform-provider-ddcloud
	zip -9 _bin/darwin-amd64.zip _bin/darwin-amd64/terraform-provider-ddcloud

test: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

# Run acceptance tests (since they're long-running, enable retry).
testacc: fmt
	rm -f "${PWD}/AccTest.log"
	TF_ACC=1 TF_LOG=DEBUG TF_LOG_PATH="${PWD}/AccTest.log" \
	DD_COMPUTE_EXTENDED_LOGGING=1 \
	DD_COMPUTE_MAX_RETRY=6 DD_COMPUTE_RETRY_DELAY=10 \
		go test -v \
		github.com/DimensionDataResearch/dd-cloud-compute-terraform/vendor/ddcloud \
		-timeout 120m \
		-run=TestAcc${TEST}

version:
	echo "package main\n\n// ProviderVersion is the current version of the ddcloud terraform provider.\nconst ProviderVersion = \"v1.0.1 (`git rev-parse HEAD`)\"" > ./version-info.go
