package main

import (
	"fmt"
	"github.com/fayizk1/syslog"
	"github.com/robertkowalski/graylog-golang"
	"code.google.com/p/gcfg"
	"os"
	"os/signal"
	"syscall"
	"strings"
)

type handler struct {
	*syslog.BaseHandler
}

type Server struct {
	Uri string
	Grayloghostname string
	Graylogport int
}

var server Server 

type Filter struct {
	Tag []string
	Message []string
}
var filter Filter

type ConfigReader struct {
        Server struct {
                Uri string
		Grayloghostname string
		Graylogport int
        }
	Filter struct {
		Tag []string
		Message []string
	}
}
var cfgrd ConfigReader

func init() {
	err := gcfg.ReadFileInto(&cfgrd, "server.gcfg")
	if err == nil {
		server.Uri = cfgrd.Server.Uri
		server.Grayloghostname = cfgrd.Server.Grayloghostname
		server.Graylogport = cfgrd.Server.Graylogport
		filter.Tag = cfgrd.Filter.Tag
		filter.Message = cfgrd.Filter.Message
	} else {
		server = Server{"0.0.0.0:514", "127.0.0.1", 12201}
//		filter =  Filter{[], []}
		fmt.Println("Config Error:", err)
	}
	fmt.Println(filter.Message, filter.Tag)
}

func filterfn(m *syslog.Message) bool {
	for i := range filter.Tag {
		//fmt.Println(filter.Tag[i])
		if m.Tag == filter.Tag[i] {
			return false
		}
	}
	for i:= range filter.Message {
		if strings.Contains(m.Content, filter.Message[i]) {
			return false
		}
	}
	return true
}

func newHandler() *handler {
	h := handler{syslog.NewBaseHandler(512, filterfn, false)}
	go h.mainLoop() 
	return &h
}

func (h *handler) mainLoop() {
	g := gelf.New(gelf.Config{
		GraylogPort:     server.Graylogport,
		GraylogHostname: server.Grayloghostname,
		Connection:      "wan",
	})
	for {
		m := h.Get()
		if m == nil {
			break
		}
		g.Log(m.Gelf())
	//	fmt.Println(m)
		//fmt.Println(m.Gelf())
	}
	fmt.Println("Exit handler")
	h.End()
}

func main() {
	s := syslog.NewServer()
	s.AddHandler(newHandler())
	s.Listen(server.Uri)

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	fmt.Println("Shutdown the server...")
	s.Shutdown()
	fmt.Println("Server is down")
}
