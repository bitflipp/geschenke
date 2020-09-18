all: run

clean:
	rm -rf dist

run: clean
	go run *.go

dist: clean
	CGO_ENABLED=0 go build -o dist/geschenke
	upx dist/geschenke
	cp -r templates static dist
