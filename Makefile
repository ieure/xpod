GOPATH := $(PWD)/go

deps:
	go get github.com/gorilla/feeds
	go get github.com/pkg/errors
	go get github.com/PuerkitoBio/goquery
