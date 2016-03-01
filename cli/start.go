package cli

import "fmt"

// Start  starts the cli application.
func Start() {
	defer func() {
		if err := recover(); err != nil {
			exit(fmt.Errorf("%v", err))
		}
	}()

	conf := parseCli()

	if conf.Host {
		forwardToHost(conf)
		return
	}

	exitIfErr(setupDocker())
	forwardToDocker(conf)

	<-trapInterrupts(nil)
}
