package cli

// Start  starts the cli application.
func Start() {
	conf := parseCli()

	if conf.Host {
		forwardToHost(conf)
		return
	}

	exitIfErr(setupDocker())
	forwardToDocker(conf)

	<-trapInterrupts(nil)
}
