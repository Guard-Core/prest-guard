.PHONY: fmt vet test bench so clean

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test -race ./...

bench:
	go test -run '^$$' -bench . -benchmem .

so:
	CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags "-s -w" \
		-o lib/middlewares/guard.so .

clean:
	rm -f lib/middlewares/guard.so
