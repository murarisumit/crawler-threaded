APP=webcrawler

build:
	go build

clean:
	rm $(APP)

get-deps:
	go get github.com/puerkitobio/goquery
	go get github.com/romana/rlog

log-info:
	export RLOG_LOG_LEVEL=INFO

log-debug:
	export RLOG_LOG_LEVEL=DEBUG

log-notime:
	export RLOG_LOG_NOTIME=yes

run: | build
	./webcrawler -u https://monzo.com -d 2

# run: | build
# 	./webcrawler -u https://sumit.murari.me/ -d 3
