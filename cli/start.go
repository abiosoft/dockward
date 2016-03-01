package cli

func Start(){
	conf := parseCli()

	if conf.Host {
		forwardToHost(conf)
		return
	}

	exitIfErr(setupDocker())
	forwardToDocker(conf)

	<-trapInterrupts(nil)
}
