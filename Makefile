include $(GOROOT)/src/Make.inc

TARG=dhcp


GOFILES=\
	message/message.go\
	message/consts.go\
	message/option.go\
        message/dhcp_message_types.go\
        message/dhcp_option_types.go\
	socket/socket.go\
	agent/finder.go\

include $(GOROOT)/src/Make.pkg


