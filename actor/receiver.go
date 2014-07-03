package actor

type RcvrId struct {
	StageId   StageId
	ActorName ActorName
	Id        uint32
}

func (r *RcvrId) isNil() bool {
	return len(r.StageId) == 0 && len(r.ActorName) == 0 && r.Id == 0
}

type receiver interface {
	id() RcvrId
	enque(mh msgAndHandler)
	start()
}

type localRcvr struct {
	asyncRoutine
	ctx recvContext
	rId RcvrId
}

func (rcvr *localRcvr) id() RcvrId {
	return rcvr.rId
}

func (rcvr *localRcvr) start() {
	for {
		select {
		case d, ok := <-rcvr.dataCh:
			if !ok {
				return
			}
			rcvr.handleMsg(d)

		case c, ok := <-rcvr.ctrlCh:
			if !ok {
				return
			}
			rcvr.handleCmd(c)
		}
	}
}

func (rcvr *localRcvr) handleMsg(mh msgAndHandler) {
	mh.handler.Recv(mh.msg, &rcvr.ctx)
}

func (rcvr *localRcvr) handleCmd(cmd routineCmd) {
	switch {
	case cmd.cmdType == stopRoutine:
		close(rcvr.dataCh)
		close(rcvr.ctrlCh)
		close(rcvr.waitCh)
	}
}

func (rcvr *localRcvr) enque(mh msgAndHandler) {
	rcvr.dataCh <- mh
}