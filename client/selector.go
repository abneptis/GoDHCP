package dhcp

import "os"

var NoOfferSelected = os.NewError("No offer was selected")
var NoOfferFound = os.NewError("No offer was found")

type OfferSelector interface {
  SelectOffer([]Message)(Message, os.Error)
}

type fifoSelector struct {}
func (self fifoSelector)SelectOffer(in []Message)(msg Message, err os.Error){
  for i := range(in){
    t, err := in[i].DHCPMessageType()
    if err == nil && t == DHCPOFFER {
      return in[i], nil
    }
  }
  err = NoOfferFound
  return
}
