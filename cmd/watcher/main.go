package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/radovskyb/watcher"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	w := watcher.New()

	rules := []string{}
	directories := []string{}

	// set regular expression rules
	for _, rule := range rules {
		r := regexp.MustCompile(rule)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	// set watched directories
	for _, dir := range directories {
		err := w.AddRecursive(dir)
		check(err)
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event)

			case err := <-w.Error:
				fmt.Println(err)

			case <-w.Closed:
				fmt.Println("Watcher closed")
				return
			}
		}
	}()

	err := w.Start(time.Second)
	check(err)
}