package main

import (
	"net/http"
	"os"
	"io/ioutil"
	"os/signal"
	"syscall"

	"github.com/takama/daemon"
	"log"
	"time"
	"fmt"
	"bytes"
	"encoding/json"
)



const (

	// name of the service
	name        = "PleskDynamicDNS Client"
	description = "PleskDynamicDNS Ping deamon"
)




var stdlog, errlog *log.Logger
var config Config
var ticker *time.Ticker

type Service struct {
	daemon.Daemon
}

type packet_struct struct {
	Token string
	Ip string
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: myservice install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}


	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)


	work()

	for {
		select {
		case killSignal := <-interrupt:
			stdlog.Println("Signal received:", killSignal)
			stdlog.Println("Stoping service ")
			if killSignal == os.Interrupt {
				return "Daemon was interruped by system signal", nil
			}
			return "Daemon was killed", nil

		}
	}

	ticker.Stop()


	return usage, nil
}

func work() {
	go func() {
		for range ticker.C {
			sendIP()
		}
	}()
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	config = readConfig()
	ticker = time.NewTicker(time.Second * time.Duration(config.Delay))
}

func main() {


	srv, err := daemon.New(name, description)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()

	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}


	fmt.Println(status)


}

func sendIP() {

	externalIP := getExternalIp()
	if (externalIP == "") {
		errlog.Println("Error while retreiving external IP address")
	}

	packet := packet_struct{Token:config.Token, Ip:externalIP}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(packet)

	resp, err := http.Post(config.Server + ":" + config.Port, "application/json; charset=utf-8", buffer)

	if err != nil {
		errlog.Println("Error communicating with dns server api")
		errlog.Println(err)
		return
	}

	stdlog.Println("Sending IP address to server.")
	defer resp.Body.Close()

}

func getExternalIp() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
	}
	if resp.StatusCode == 200 {
		// OK
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return ""
}
