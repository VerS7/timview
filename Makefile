GOC=go
NAME=timview
PLATFORMS=linux windows

all: $(PLATFORMS)

linux: 
	$(GOC) build -o $(NAME) cmd/timview/main.go

windows: 
	GOOS=windows $(GOC) build -o $(NAME).exe cmd/timview/main.go

clean:
	rm $(NAME)
	rm $(NAME).exe