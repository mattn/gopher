all : gopher.exe

gopher.exe : main.go embed.go Makefile gopher.syso
	go build -ldflags="-H windowsgui"

gopher.syso : gopher.rc
	windres gopher.rc gopher.syso

clean :
	-rm *.syso *.exe
