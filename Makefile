.PHONY: test profile

test:
	go test -bench=. -benchmem ./...

profile:
	go test -bench=. -benchmem ./fhirpath -cpuprofile=cpu.prof -memprofile=mem.prof
