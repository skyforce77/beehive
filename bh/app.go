package bh

import (
	"errors"

	"github.com/golang/glog"
)

type AppName string

// Apps simply process and exchange messages. App methods are not
// thread-safe and we assume that neither are its map and receive functions.
type App interface {
	// Handles a specific message type using the handler. If msgType is an
	// name of msgType's reflection type.
	// instnace of MsgType, we use it as the type. Otherwise, we use the qualified
	Handle(msgType interface{}, h Handler) error
	// Hanldes a specific message type using the map and receive functions. If
	// msgType is an instnace of MsgType, we use it as the type. Otherwise, we use
	// the qualified name of msgType's reflection type.
	HandleFunc(msgType interface{}, m Map, r Recv) error

	// Regsiters the app's detached handler.
	Detached(h DetachedHandler) error
	// Registers the detached handler using functions.
	DetachedFunc(start Start, stop Stop, r Recv) error

	// Returns the state of this app that is shared among all instances and the
	// map function. This state is NOT thread-safe and apps must synchronize for
	// themselves.
	State() State
	// Returns the app name.
	Name() AppName

	// Whether the app is sticky.
	Sticky() bool
	// Sets whether the app is sticky, i.e., should not be migrated.
	SetSticky(sticky bool)
}

// An applications map function that maps a specific message to the set of keys
// in state dictionaries. This method is assumed not to be thread-safe and is
// called sequentially.
type Map func(m Msg, c Context) MapSet

// An application recv function that handles a message. This method is called in
// parallel for different map-sets and sequentially within a map-set.
type Recv func(m Msg, c RecvContext)

// The interface msg handlers should implement.
type Handler interface {
	Map(m Msg, c Context) MapSet
	Recv(m Msg, c RecvContext)
}

// Detached handlers, in contrast to normal Handlers with Map and Recv, start in
// their own go-routine and emit messages. They do not listen on a particular
// message and only recv replys in their receive functions.
// Note that each app can have only one detached handler.
type DetachedHandler interface {
	// Starts the handler. Note that this will run in a separate goroutine, and
	// you can block.
	Start(ctx RecvContext)
	// Stops the handler. This should notify the start method perhaps using a
	// channel.
	Stop(ctx RecvContext)
	// Receives replies to messages emitted in this handler.
	Recv(m Msg, ctx RecvContext)
}

// Start function of a detached handler.
type Start func(ctx RecvContext)

// Stop function of a detached handler.
type Stop func(ctx RecvContext)

type funcHandler struct {
	mapFunc  Map
	recvFunc Recv
}

func (h *funcHandler) Map(m Msg, c Context) MapSet {
	return h.mapFunc(m, c)
}

func (h *funcHandler) Recv(m Msg, c RecvContext) {
	h.recvFunc(m, c)
}

type funcDetached struct {
	startFunc Start
	stopFunc  Stop
	recvFunc  Recv
}

func (h *funcDetached) Start(c RecvContext) {
	h.startFunc(c)
}

func (h *funcDetached) Stop(c RecvContext) {
	h.stopFunc(c)
}

func (h *funcDetached) Recv(m Msg, c RecvContext) {
	h.recvFunc(m, c)
}

type app struct {
	name     AppName
	hive     *hive
	mapper   *mapper
	handlers map[MsgType]Handler
	sticky   bool
}

func (a *app) HandleFunc(msgType interface{}, m Map, r Recv) error {
	return a.Handle(msgType, &funcHandler{m, r})
}

func (a *app) DetachedFunc(start Start, stop Stop, rcv Recv) error {
	return a.Detached(&funcDetached{start, stop, rcv})
}

func (a *app) Handle(msg interface{}, h Handler) error {
	if a.mapper == nil {
		glog.Fatalf("App's mapper is nil!")
	}

	t := msgType(msg)
	a.hive.RegisterMsg(msg)
	return a.registerHandler(t, h)
}

func (a *app) registerHandler(t MsgType, h Handler) error {
	_, ok := a.handlers[t]
	if ok {
		return errors.New("A handler for this message type already exists.")
	}

	a.handlers[t] = h
	a.hive.registerHandler(t, a.mapper, h)
	return nil
}

func (a *app) handler(t MsgType) Handler {
	return a.handlers[t]
}

func (a *app) Detached(h DetachedHandler) error {
	return a.mapper.registerDetached(h)
}

func (a *app) State() State {
	return a.mapper.state()
}

func (a *app) Name() AppName {
	return a.name
}

func (a *app) SetSticky(sticky bool) {
	a.sticky = sticky
}

func (a *app) Sticky() bool {
	return a.sticky
}

func (a *app) initMapper() {
	// TODO(soheil): Maybe stop the previous mapper if any?
	a.mapper = &mapper{
		asyncRoutine: asyncRoutine{
			dataCh: make(chan msgAndHandler, a.hive.config.DataChBufSize),
			ctrlCh: make(chan routineCmd),
		},
		ctx: context{
			hive: a.hive,
			app:  a,
		},
		keyToRcvrs: make(map[DictionaryKey]receiver),
		idToRcvrs:  make(map[RcvrId]receiver),
	}
}
