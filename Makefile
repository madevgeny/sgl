default: build

package = sgl

.PHONY: default build

file_list = sgl.go \
		version.go

build:format
		go build -p 4 -o ${package} ${file_list}

run:format
		go run -p 4 -race ${file_list}

format:
		go fmt ${file_list}
