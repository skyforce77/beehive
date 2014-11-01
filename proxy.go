package beehive

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kandoo/beehive/Godeps/_workspace/src/github.com/coreos/etcd/raft/raftpb"
	"github.com/kandoo/beehive/Godeps/_workspace/src/github.com/golang/glog"
	"github.com/kandoo/beehive/connpool"
)

func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&connpool.Dialer{
				MaxConnPerAddr: 64,
			}).Dial,
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConnsPerHost: 64,
		},
	}
}

type proxyBee struct {
	localBee
	proxy proxy
}

type proxy struct {
	client *http.Client

	to       string
	stateURL string
	msgURL   string
	cmdURL   string
	raftURL  string

	errors uint64
}

func newProxyWithAddr(client *http.Client, addr string) proxy {
	// TODO(soheil): add scheme.
	p := proxy{
		client:   client,
		to:       addr,
		stateURL: buildURL("http", addr, serverV1StatePath),
		msgURL:   buildURL("http", addr, serverV1MsgPath),
		cmdURL:   buildURL("http", addr, serverV1CmdPath),
		raftURL:  buildURL("http", addr, serverV1RaftPath),
	}
	return p
}

func (h *hive) newProxy(to uint64) (proxy, error) {
	a, err := h.hiveAddr(to)
	if err != nil {
		return proxy{}, err
	}

	return newProxyWithAddr(h.client, a), nil
}

func (p proxy) sendMsg(m *msg) error {
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(m); err != nil {
		return err
	}

	res, err := p.client.Post(p.msgURL, "application/x-gob", &data)
	if err != nil {
		p.errors++
		return err
	}

	p.errors = 0
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		return errors.New(string(b.Bytes()))
	}
	return nil
}

func (p proxy) sendCmd(c *cmd) (interface{}, error) {
	// TODO(soheil): We need to add a retry strategy here.
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(c); err != nil {
		return nil, err
	}

	glog.V(2).Infof("Proxy to %v sends command %v", p.to, c)
	pRes, err := p.client.Post(p.cmdURL, "application/x-gob", &data)
	if err != nil {
		p.errors++
		return nil, err
	}
	p.errors = 0
	glog.V(2).Infof("Proxy to %v receives the result for command %v", p.to, c)

	defer pRes.Body.Close()
	if pRes.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(pRes.Body)
		return nil, errors.New(string(b.Bytes()))
	}
	cRes := cmdResult{}
	if err := gob.NewDecoder(pRes.Body).Decode(&cRes); err != nil {
		return nil, err
	}
	return cRes.get()
}

// TODO(soheil): We should batch here.
func (b *proxyBee) handleMsg(mh msgAndHandler) {
	mh.msg.MsgTo = b.ID()

	glog.V(2).Infof("Proxy %v sends msg %v", b, mh.msg)
	if err := b.proxy.sendMsg(mh.msg); err != nil {
		glog.Errorf("cannot send message %v to %v: %v", mh.msg, b, err)
	}
}

func (b *proxyBee) handleCmd(cc cmdAndChannel) {
	switch cc.cmd.Data.(type) {
	case cmdStop, cmdStart:
		b.localBee.handleCmd(cc)
	default:
		d, err := b.proxy.sendCmd(&cc.cmd)
		if cc.ch != nil {
			cc.ch <- cmdResult{Data: d, Err: err}
		}
	}
}

func (p proxy) doSendRaft(url string, m raftpb.Message) error {
	d, err := m.Marshal()
	if err != nil {
		glog.Fatalf("cannot marshal raft message")
	}

	r, err := p.client.Post(url, "application/x-protobuf", bytes.NewBuffer(d))
	if err != nil {
		p.errors++
		return err
	}
	p.errors = 0

	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(r.Body)
		return errors.New(string(b.Bytes()))
	}
	return nil
}

func (p proxy) sendRaft(m raftpb.Message) error {
	glog.V(2).Infof("proxy to %v sends raft message %v", p.to, m)
	return p.doSendRaft(p.raftURL, m)
}

func (p proxy) sendBeeRaft(app string, b uint64, m raftpb.Message) error {
	url := buildURL("http", p.to,
		fmt.Sprintf("%s/%s/%v", serverV1RaftPath, app, b))
	glog.V(2).Infof("proxy to %v sends bee raft message %v", url, m)
	return p.doSendRaft(url, m)
}

func (p proxy) state() (hiveState, error) {
	s := hiveState{}
	r, err := p.client.Get(p.stateURL)
	if err != nil {
		p.errors++
		return s, err
	}

	p.errors = 0

	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(r.Body)
		return s, errors.New(string(b.Bytes()))
	}

	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&s)
	return s, err
}

// TODO(soheil): Maybe start should return an error.
func (b *proxyBee) start() {
	b.status = beeStatusStarted
	glog.V(2).Infof("%v started", b)

	for b.status == beeStatusStarted {
		select {
		case d, ok := <-b.dataCh:
			if !ok {
				return
			}
			b.handleMsg(d)

		case c, ok := <-b.ctrlCh:
			if !ok {
				return
			}
			b.handleCmd(c)
		}
	}
}

func (b *proxyBee) String() string {
	return "proxy " + b.localBee.String()
}
