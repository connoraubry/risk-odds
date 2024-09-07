build:
	mkdir -p bin/
	go build -o bin/main .
test:
	go test 
clean:
	rm bin/*
	rm c.out
	rm profile.out
bench:
	go test -bench=. -cpuprofile profile.out
profile:
	go tool pprof -http=:8080 profile.out 
