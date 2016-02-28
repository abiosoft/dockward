package main
import (
	"os/signal"
	"fmt"
	"syscall"
	"os"
)

// trapInterrupts traps OS interrupt signals.
func trapInterrupts(exit chan struct{}) chan struct{} {
	done := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			fmt.Print("OS Interrupt signal received. Performing cleanup...")
			cleanUp()
			fmt.Println("Done.")
			done <- struct{}{}
		case <-exit:
			cleanUp()
			done <- struct{}{}
		}

	}()
	return done
}

// cleanUpFuncs is list of functions to call before application exits.
var cleanUpFuncs []func()

// addCleanUpFunc adds a function to cleanUpFuncs.
func addCleanUpFunc(f func()) {
	cleanUpFuncs = append(cleanUpFuncs, f)
}

// cleanUp calls all functions in cleanUpFuncs.
func cleanUp() {
	for i := range cleanUpFuncs {
		cleanUpFuncs[i]()
	}
}
