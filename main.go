package main

func main() {
	args := parseCli()

	if args.Host {
		forwardToHost(args)
		return
	}

	exitIfErr(setupDocker())
	forwardToDocker(args)

	<-trapInterrupts(nil)
}
