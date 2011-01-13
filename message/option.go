package dhcp

import "bytes"
import "io"
import "os"
import "net"

type Option interface {
  OptionType()(byte)
  Bytes()([]byte)
}

// Write an option to an io.Writer, including tag  & length
// (if length is appropriate to the tag type).
// Utilizes the MarshalOption as the underlying serializer.
func WriteOption(w io.Writer, a Option)(err os.Error){
  out, err := MarshalOption(a)
  if err == nil {
    _, err = w.Write(out)
  }
  return
}


type option struct {
  tag byte
  data []byte
}

// A more json.Marshal like version of WriteOption.
func MarshalOption(o Option)(out []byte, err os.Error){
  switch o.OptionType() {
    case OPT_PAD, OPT_END:
      out = []byte{o.OptionType()}
    default:
      dlen := len(o.Bytes())
      if dlen > 253 {
        err = os.NewError("Data too long to marshal")
      } else {
        out = make([]byte, dlen + 2)
        out[0], out[1] = o.OptionType(), byte(dlen)
        copy(out[2:], o.Bytes())
      }
  }
  return
}

func (self option)Bytes()([]byte){ return self.data }
func (self option)OptionType()(byte){ return self.tag }

func NewOption(tag byte, data []byte)(Option){
  return &option{tag:tag, data: data}
}

// NB: We don't validate that you have /any/ IP's in the option here,
// simply that if you do that they're valid. Most DHCP options are only
// valid with 1(+|) values
func IP4sOption(tag byte, ips []net.IP)(opt Option, err os.Error){
  var out []byte = make([]byte, 4* len(ips))
  for i := range(ips){
    ip := ips[i].To4()
    if ip == nil {
      err = os.NewError("ip is not a valid IPv4 address")
    } else {
      copy(out[i*4:], []byte(ip))
    }
    if err != nil { break }
  }
  opt = NewOption(tag, out)
  return
}

// NB: We don't validate that you have /any/ IP's in the option here,
// simply that if you do that they're valid. Most DHCP options are only
// valid with 1(+|) values
func IP4Option(tag byte, ips net.IP)(opt Option, err os.Error){
  ips = ips.To4()
  if ips == nil {
    err = os.NewError("ip is not a valid IPv4 address")
    return
  }
  opt = NewOption(tag, []byte(ips))
  return
}

// NB: I'm not checking tag : min length here!
func StringOption(tag byte, s string)(opt Option, err os.Error){
  opt = &option{tag: tag, data: bytes.NewBufferString(s).Bytes()}
  return
}


func ParseOptions(in []byte)(opts []Option, err os.Error){
  pos := 0
  for pos < len(in) && err == nil {
    var tag = in[pos]
    pos++
    switch tag {
      case OPT_PAD, OPT_END:
        opts = append(opts, NewOption(tag, []byte{}))
      default:
        if len(in) - pos >= 1 {
          _len := in[pos]
          pos++
          opts = append(opts, NewOption(tag, in[pos:pos+int(_len)]))
          pos += int(_len)
        }
    }
    //log.Printf("ParseLoop: (%v) %d/%d", err, pos, len(in))
  }
  return
}


