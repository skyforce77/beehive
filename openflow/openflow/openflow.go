// Automatically generated by Packet Go code generator.
package openflow

import (
  "encoding/binary"
  "errors"
  "fmt"
  "net"

  "github.com/packet/packet/src/go/packet"

)

type OpenflowVersions int

const (
  OPENFLOW_1_0 OpenflowVersions = 1
  OPENFLOW_1_1 OpenflowVersions = 2
  OPENFLOW_1_2 OpenflowVersions = 3
  OPENFLOW_1_3 OpenflowVersions = 4
)

type OpenflowConstants int

const (
  OFP_ETH_ALEN OpenflowConstants = 6
  OFP_MAX_PORT_NAME_LEN OpenflowConstants = 16
  OFP_MAX_TABLE_NAME_LEN OpenflowConstants = 32
)

type OpenflowType int

const (
  OFPT_HELLO OpenflowType = 0
  OFPT_ERROR OpenflowType = 1
  OFPT_ECHO_REQUEST OpenflowType = 2
  OFPT_ECHO_REPLY OpenflowType = 3
  OFPT_VENDOR OpenflowType = 4
  OFPT_FEATURES_REQUEST OpenflowType = 5
  OFPT_FEATURES_REPLY OpenflowType = 6
  OFPT_GET_CONFIG_REQUEST OpenflowType = 7
  OFPT_GET_CONFIG_REPLY OpenflowType = 8
  OFPT_SET_CONFIG OpenflowType = 9
)

func NewOpenflowHeaderWithBuf(b []byte) OpenflowHeader {
  return OpenflowHeader{packet.Packet{Buf: b}}
}

func NewOpenflowHeader() OpenflowHeader {
  s := 8
  b := make([]byte, s)
  p := OpenflowHeader{packet.Packet{Buf: b}}
  p.Init()
  return p
}

type OpenflowHeader struct {
  packet.Packet
}

func (this OpenflowHeader) minSize() int {
  return 8
}

type OpenflowHeaderConn struct {
  net.Conn
  buf []byte
  offset int
}

func NewOpenflowHeaderConn(c net.Conn) OpenflowHeaderConn {
  return OpenflowHeaderConn {
    Conn: c,
    buf: make([]byte, packet.DefaultBufSize),
  }
}

func (c *OpenflowHeaderConn) Write(pkts []OpenflowHeader) error {
  for _, p := range pkts {
    s := p.Size()
    b := p.Buffer()[:s]
    n := 0
    for s > 0 {
      var err error
      if n, err = c.Conn.Write(b); err != nil {
        return fmt.Errorf("Error in write: %v", err)
      }
      s -= n
    }
  }
  return nil
}

func (c *OpenflowHeaderConn) Read(pkts []OpenflowHeader) (int, error) {
  if len(c.buf) == c.offset {
    newSize := packet.DefaultBufSize
    if newSize < len(c.buf) {
      newSize = 2 * len(c.buf)
    }

    buf := make([]byte, newSize)
    copy(buf, c.buf[:c.offset])
    c.buf = buf
  }

  r, err := c.Conn.Read(c.buf[c.offset:])
  if err != nil {
    return 0, err
  }

  r += c.offset

  s := 0
  n := 0
  for i := range pkts {
    p := NewOpenflowHeaderWithBuf(c.buf)

    pSize := p.Size()
    if r < s+pSize {
      break
    }

    pkts[i] = p
    s += pSize
    n++
  }

  c.offset = r - s
  if c.offset < 0 {
    panic("Invalid value for offset")
  }

  c.buf = c.buf[s:]
  if c.offset < len(c.buf) {
    return n, nil
  }

  buf := make([]byte, packet.DefaultBufSize)
  copy(buf, c.buf[:c.offset])
  c.buf = buf
  return n, nil
}

func (this *OpenflowHeader) Init() {
  this.SetLength(uint16(this.minSize()))
  // Invariants.
}

func (this OpenflowHeader) Size() int {
  if len(this.Buf) < this.minSize() {
    return 0
  }

  size := int(this.Length())
  return size
}

func ConvertToOpenflowHeader(p packet.Packet) (OpenflowHeader, error) {
  if !IsOpenflowHeader(p) {
    return NewOpenflowHeaderWithBuf(nil), errors.New("Cannot convert to openflow.OpenflowHeader")
  }

  return NewOpenflowHeaderWithBuf(p.Buf), nil
}

func IsOpenflowHeader(p packet.Packet) bool {
  return true
}

func (this *OpenflowHeader) Version() uint8 {
  offset := this.VersionOffset()
  res := uint8(this.Buf[offset])
  return res
}

func (this *OpenflowHeader) SetVersion(v uint8) {
  offset := this.VersionOffset()
  this.Buf[offset] = byte(v); offset++
}

func (this *OpenflowHeader) VersionOffset() int {
  offset := 0
  return offset
}


func (this *OpenflowHeader) Type() uint8 {
  offset := this.TypeOffset()
  res := uint8(this.Buf[offset])
  return res
}

func (this *OpenflowHeader) SetType(t uint8) {
  offset := this.TypeOffset()
  this.Buf[offset] = byte(t); offset++
}

func (this *OpenflowHeader) TypeOffset() int {
  offset := 1
  return offset
}


func (this *OpenflowHeader) Length() uint16 {
  offset := this.LengthOffset()
  res := binary.BigEndian.Uint16(this.Buf[offset:])
  return res
}

func (this *OpenflowHeader) SetLength(l uint16) {
  offset := this.LengthOffset()
  binary.BigEndian.PutUint16(this.Buf[offset:], l); offset += 2
}

func (this *OpenflowHeader) LengthOffset() int {
  offset := 2
  return offset
}


func (this *OpenflowHeader) Xid() uint32 {
  offset := this.XidOffset()
  res := binary.BigEndian.Uint32(this.Buf[offset:])
  return res
}

func (this *OpenflowHeader) SetXid(x uint32) {
  offset := this.XidOffset()
  binary.BigEndian.PutUint32(this.Buf[offset:], x); offset += 4
}

func (this *OpenflowHeader) XidOffset() int {
  offset := 4
  return offset
}


func NewOpenflowHelloWithBuf(b []byte) OpenflowHello {
  return OpenflowHello{OpenflowHeader{packet.Packet{Buf: b}}}
}

func NewOpenflowHello() OpenflowHello {
  s := 8
  b := make([]byte, s)
  p := OpenflowHello{OpenflowHeader{packet.Packet{Buf: b}}}
  p.Init()
  return p
}

type OpenflowHello struct {
  OpenflowHeader
}

func (this OpenflowHello) minSize() int {
  return 8
}

type OpenflowHelloConn struct {
  net.Conn
  buf []byte
  offset int
}

func NewOpenflowHelloConn(c net.Conn) OpenflowHelloConn {
  return OpenflowHelloConn {
    Conn: c,
    buf: make([]byte, packet.DefaultBufSize),
  }
}

func (c *OpenflowHelloConn) Write(pkts []OpenflowHello) error {
  for _, p := range pkts {
    s := p.Size()
    b := p.Buffer()[:s]
    n := 0
    for s > 0 {
      var err error
      if n, err = c.Conn.Write(b); err != nil {
        return fmt.Errorf("Error in write: %v", err)
      }
      s -= n
    }
  }
  return nil
}

func (c *OpenflowHelloConn) Read(pkts []OpenflowHello) (int, error) {
  if len(c.buf) == c.offset {
    newSize := packet.DefaultBufSize
    if newSize < len(c.buf) {
      newSize = 2 * len(c.buf)
    }

    buf := make([]byte, newSize)
    copy(buf, c.buf[:c.offset])
    c.buf = buf
  }

  r, err := c.Conn.Read(c.buf[c.offset:])
  if err != nil {
    return 0, err
  }

  r += c.offset

  s := 0
  n := 0
  for i := range pkts {
    p := NewOpenflowHelloWithBuf(c.buf)

    pSize := p.Size()
    if r < s+pSize {
      break
    }

    pkts[i] = p
    s += pSize
    n++
  }

  c.offset = r - s
  if c.offset < 0 {
    panic("Invalid value for offset")
  }

  c.buf = c.buf[s:]
  if c.offset < len(c.buf) {
    return n, nil
  }

  buf := make([]byte, packet.DefaultBufSize)
  copy(buf, c.buf[:c.offset])
  c.buf = buf
  return n, nil
}

func (this *OpenflowHello) Init() {
  this.OpenflowHeader.Init()
  this.SetLength(uint16(this.minSize()))
  // Invariants.
  this.SetType(uint8(0)) // type
}

func (this OpenflowHello) Size() int {
  if len(this.Buf) < this.minSize() {
    return 0
  }

  size := int(this.Length())
  return size
}

func ConvertToOpenflowHello(p OpenflowHeader) (OpenflowHello, error) {
  if !IsOpenflowHello(p) {
    return NewOpenflowHelloWithBuf(nil), errors.New("Cannot convert to openflow.OpenflowHello")
  }

  return NewOpenflowHelloWithBuf(p.Buf), nil
}

func IsOpenflowHello(p OpenflowHeader) bool {
  return p.Type() == 0 && true
}