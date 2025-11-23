GOC=go
GOARCH=386
NAME=timview
PLATFORMS=linux windows

all: $(PLATFORMS)

linux: 
	GOARCH=$(GOARCH) $(GOC) build -o $(NAME) cmd/timview/main.go

windows: 
	GOARCH=$(GOARCH) GOOS=windows $(GOC) build -o $(NAME).exe cmd/timview/main.go

clean:
	rm $(NAME)
	rm $(NAME).exe