default: spectacle

spectacle:
	go build -tags pkgconfig -o $@ ./

clean:
	rm -rf spectacle

.PHONY: spectacle clean
