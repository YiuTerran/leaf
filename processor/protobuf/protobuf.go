package protobuf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/log"
	"github.com/golang/protobuf/proto"
)

// 对leaf中已有的Protobuf协议格式进行扩展，允许自定义的id
// -------------------------
// | id | protobuf message |
// -------------------------
type processor struct {
	littleEndian bool
	msgInfo      map[uint16]*msgInfoST
	msgID        map[reflect.Type]uint16
}

type msgInfoST struct {
	msgType       reflect.Type
	msgRouter     *chanrpc.Server
	msgHandler    msgHandlerST
	msgRawHandler msgHandlerST
}

type msgHandlerST func([]interface{})

type msgRawST struct {
	msgID      uint16
	msgRawData []byte
}

func NewProcessor() *processor {
	p := new(processor)
	p.littleEndian = false
	p.msgID = make(map[reflect.Type]uint16)
	p.msgInfo = make(map[uint16]*msgInfoST)
	return p
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *processor) Register(msg proto.Message, eventType uint16) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("protobuf message pointer required")
	}
	if _, ok := p.msgID[msgType]; ok {
		log.Fatal("message %s is already registered", msgType)
	}
	if len(p.msgInfo) >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	i := new(msgInfoST)
	i.msgType = msgType
	p.msgInfo[eventType] = i
	p.msgID[msgType] = eventType
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *processor) SetRouter(msg proto.Message, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Fatal("message %s not registered", msgType)
	}

	p.msgInfo[id].msgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *processor) SetHandler(msg proto.Message, msgHandler msgHandlerST) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Fatal("message %s not registered", msgType)
	}

	p.msgInfo[id].msgHandler = msgHandler
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *processor) SetRawHandler(id uint16, msgRawHandler msgHandlerST) {
	if id >= uint16(len(p.msgInfo)) {
		log.Fatal("message id %v not registered", id)
	}

	p.msgInfo[id].msgRawHandler = msgRawHandler
}

// goroutine safe
func (p *processor) Route(msg interface{}, userData interface{}) error {
	// raw
	if msgRaw, ok := msg.(msgRawST); ok {
		if msgRaw.msgID >= uint16(len(p.msgInfo)) {
			return fmt.Errorf("message id %v not registered", msgRaw.msgID)
		}
		i := p.msgInfo[msgRaw.msgID]
		if i.msgRawHandler != nil {
			i.msgRawHandler([]interface{}{msgRaw.msgID, msgRaw.msgRawData, userData})
		}
		return nil
	}

	// protobuf
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return fmt.Errorf("message %s not registered", msgType)
	}
	i := p.msgInfo[id]
	if i.msgHandler != nil {
		i.msgHandler([]interface{}{msg, userData})
	}
	if i.msgRouter != nil {
		i.msgRouter.Go(msgType, msg, userData)
	}
	return nil
}

// goroutine safe
func (p *processor) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}
	i, ok := p.msgInfo[id]
	if !ok {
		return nil, fmt.Errorf("message id %v not registered", id)
	}
	// msg
	if i.msgRawHandler != nil {
		return msgRawST{id, data[2:]}, nil
	} else {
		msg := reflect.New(i.msgType.Elem()).Interface()
		return msg, proto.UnmarshalMerge(data[2:], msg.(proto.Message))
	}
}

// goroutine safe
func (p *processor) Marshal(msg interface{}) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)

	// id
	_id, ok := p.msgID[msgType]
	if !ok {
		err := fmt.Errorf("message %s not registered", msgType)
		return nil, err
	}

	id := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}

	// data
	data, err := proto.Marshal(msg.(proto.Message))
	return [][]byte{id, data}, err
}

// goroutine safe
func (p *processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(id, i.msgType)
	}
}
