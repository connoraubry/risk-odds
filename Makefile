build:
	mkdir -p bin/
	go build -o bin/main .
test:
	go test 
clean:
	rm bin/*
	rm c.out
