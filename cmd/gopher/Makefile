IMAGES = $(wildcard data/*)

all : gopher.exe

gopher.exe : main.go bindata.go Makefile gopher.syso
	go build -ldflags="-H windowsgui"

bindata.go : $(IMAGES) main.go
	go-bindata data

gopher.syso : gopher.rc
	windres gopher.rc gopher.syso

clean :
	-rm *.syso *.exe
