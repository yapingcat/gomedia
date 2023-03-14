package rtmp

import (
    "encoding/binary"
    "fmt"
    "math"
)

type AMF0_DATA_TYPE int

const (
    AMF0_NUMBER AMF0_DATA_TYPE = iota
    AMF0_BOOLEAN
    AMF0_STRING
    AMF0_OBJECT
    AMF0_MOVIECLIP
    AMF0_NULL
    AMF0_UNDEFINED
    AMF0_REFERENCE
    AMF0_ECMA_ARRAY
    AMF0_OBJECT_END
    AMF0_STRICT_ARRAY
    AMF0_DATE
    AMF0_LONG_STRING
    AMF0_UNSUPPORTED
    AMF0_RECORDSET
    AMF0_XML_DOCUMENT
    AMF0_TYPED_OBJECT
    AMF0_AVMPLUS_OBJECT
)

var NullItem []byte = []byte{byte(AMF0_NULL)}
var EndObj []byte = []byte{0, 0, byte(AMF0_OBJECT_END)}

type amf0Item struct {
    amfType AMF0_DATA_TYPE
    length  int
    value   interface{}
}

func (amf *amf0Item) encode() []byte {
    buf := make([]byte, amf.length+4+8)
    switch amf.amfType {
    case AMF0_NUMBER:
        buf[0] = byte(AMF0_NUMBER)
        binary.BigEndian.PutUint64(buf[1:], math.Float64bits(amf.value.(float64)))
        return buf[:9]
    case AMF0_BOOLEAN:
        buf[0] = byte(AMF0_BOOLEAN)
        v := amf.value.(bool)
        if v {
            buf[1] = 1
        } else {
            buf[1] = 0
        }
        return buf[0:2]
    case AMF0_STRING:
        buf[0] = byte(AMF0_STRING)
        buf[1] = byte(uint16(amf.length) >> 8)
        buf[2] = byte(uint16(amf.length))
        copy(buf[3:], []byte(amf.value.(string)))
        return buf[0 : 3+amf.length]
    case AMF0_MOVIECLIP:
    case AMF0_NULL:
        buf[0] = byte(AMF0_NULL)
        return buf[0:1]
    case AMF0_UNDEFINED:
    case AMF0_REFERENCE:
    case AMF0_ECMA_ARRAY:
    case AMF0_STRICT_ARRAY:
    case AMF0_DATE:
    case AMF0_LONG_STRING:
    case AMF0_UNSUPPORTED:
    case AMF0_RECORDSET:
    case AMF0_XML_DOCUMENT:
    case AMF0_TYPED_OBJECT:
    case AMF0_AVMPLUS_OBJECT:
    default:
        panic("unsupport")
    }
    return nil
}

func (amf *amf0Item) decode(data []byte) int {
    _ = data[0]
    amf.amfType = AMF0_DATA_TYPE(data[0])
    switch amf.amfType {
    case AMF0_NUMBER:
        amf.length = 8
        v := math.Float64frombits(binary.BigEndian.Uint64(data[1:]))
        amf.value = v
        return 9
    case AMF0_BOOLEAN:
        amf.length = 1
        if data[1] == 1 {
            amf.value = true
        } else {
            amf.value = false
        }
        return 2
    case AMF0_STRING:
        amf.length = int(binary.BigEndian.Uint16(data[1:]))
        str := make([]byte, amf.length)
        copy(str, data[3:3+amf.length])
        amf.value = str
        return 3 + amf.length
    case AMF0_NULL:
    case AMF0_LONG_STRING:
        amf.length = int(binary.BigEndian.Uint32(data[1:]))
        str := make([]byte, amf.length)
        copy(str, data[5:5+amf.length])
        return 5 + amf.length
    case AMF0_UNDEFINED:
    case AMF0_ECMA_ARRAY:
        return 5
    default:
        panic(fmt.Sprintf("unsupport amf type %d", amf.amfType))
    }
    return 1
}

func makeStringItem(str string) amf0Item {
    item := amf0Item{
        amfType: AMF0_STRING,
        length:  len(str),
        value:   str,
    }
    return item
}

func makeNumberItem(num float64) amf0Item {
    item := amf0Item{
        amfType: AMF0_NUMBER,
        value:   num,
    }
    return item
}

func makeBoolItem(v bool) amf0Item {
    item := amf0Item{
        amfType: AMF0_BOOLEAN,
        value:   v,
    }
    return item
}

type amfObjectItem struct {
    name  string
    value amf0Item
}

type amfObject struct {
    items []*amfObjectItem
}

func (object *amfObject) encode() []byte {
    obj := make([]byte, 1)
    obj[0] = byte(AMF0_OBJECT)
    for _, item := range object.items {
        lenbytes := make([]byte, 2)
        binary.BigEndian.PutUint16(lenbytes, uint16(len(item.name)))
        obj = append(obj, lenbytes...)
        obj = append(obj, []byte(item.name)...)
        obj = append(obj, item.value.encode()...)
    }
    obj = append(obj, EndObj...)
    return obj
}

func (object *amfObject) decode(data []byte) int {
    total := 1
    data = data[1:]
    isArray := false
    for len(data) > 0 {
        if data[0] == 0x00 && data[1] == 0x00 && data[2] == byte(AMF0_OBJECT_END) {
            total += 3
            if isArray {
                isArray = false
                continue
            } else {
                break
            }
        }
        length := binary.BigEndian.Uint16(data)
        name := string(data[2 : 2+length])
        item := amf0Item{}
        l := item.decode(data[2+length:])
        if item.amfType == AMF0_ECMA_ARRAY {
            isArray = true
        } else {
            obj := &amfObjectItem{
                name:  name,
                value: item,
            }
            object.items = append(object.items, obj)
        }
        data = data[2+int(length)+l:]
        total += 2 + int(length) + l
    }
    return total
}

func decodeAmf0(data []byte) (items []amf0Item, objs []amfObject) {
    for len(data) > 0 {
        switch AMF0_DATA_TYPE(data[0]) {
        case AMF0_ECMA_ARRAY:
            data = data[5:]
            fallthrough
        case AMF0_OBJECT:
            obj := amfObject{}
            l := obj.decode(data)
            data = data[l:]
            objs = append(objs, obj)
        default:
            item := amf0Item{}
            l := item.decode(data)
            data = data[l:]
            items = append(items, item)
        }
    }
    return
}
