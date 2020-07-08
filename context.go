package dyndns

const maxBufferSize = 1000

// Context is shared channel space to synchronize various go routines.
type Context struct {
	cfg           *Config
	exit          chan bool
	exitRoutine   chan bool
	reload        chan bool
	reloadRoutine chan bool
	error         chan error
	fatalError    error
}

func (s *Server) initContext() {
	if s.ctx != nil {
		return
	}
	ctx := &Context{
		cfg: s.cfg,
	}
	ctx.exit = make(chan bool)
	ctx.exitRoutine = make(chan bool, maxBufferSize)
	ctx.reload = make(chan bool)
	ctx.reloadRoutine = make(chan bool)
	ctx.error = make(chan error, maxBufferSize)
	s.ctx = ctx
	return
}
