package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mattn/gopher"
)

var (
	exit    = flag.Bool("x", false, "exit all gophers")
	list    = flag.Bool("l", false, "list gophers")
	message = flag.String("m", "", "send message to gopher")
	link    = flag.String("u", "", "send url to gopher")
	clear   = flag.Bool("c", false, "clear message queue")
	jump    = flag.Bool("j", false, "jump gopher")
)

func main() {
	flag.Parse()

	gophers := gopher.Lookup()

	if *list {
		for _, gopher := range gophers {
			fmt.Println(gopher.Hwnd(), gopher.Name())
		}
		os.Exit(0)
	}

	if len(gophers) == 0 {
		fmt.Fprintln(os.Stderr, "Gopher not exists")
		os.Exit(1)
	}

	if *exit {
		// broadcast terminate message
		for _, gopher := range gophers {
			gopher.Terminate()
		}
		return
	}

	if *jump {
		// broadcast jump message
		for _, gopher := range gophers {
			gopher.Jump()
		}
		return
	}

	if *clear {
		// broadcast clear message
		for _, gopher := range gophers {
			gopher.Clear()
		}
		if *message == "" {
			return
		}
	}

	if *message != "" {
		// choice one of gophers to display message
		rand.Seed(time.Now().UnixNano())
		gopher := gophers[rand.Int()%len(gophers)]
		for {
			if err := gopher.Message(*message, *link); err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	} else {
		flag.Usage()
		os.Exit(1)
	}
}
