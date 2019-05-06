package main

import (
	"net/http"
	"strconv"
	"time"

	// "fmt"
	"github.com/sirupsen/logrus"
	"github.com/drael/GOnetstat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	socketTCPListening = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "socket_tcp_port_listening",
			Help: "Listening server sockets TCP",},
		[]string{"Port"},
	)
	chInitPorts = make(chan []int64)
)

func init() {
	prometheus.MustRegister(socketTCPListening)
	chInitPorts <- getPortListenning()
	go func() {
		for {
			portsListing := <-chInitPorts
			for _, port := range portsListing {
				updateInitPort(port)
			}
		}
	}()
}

func getPortListenning() (portListening []int64) {
	sockets := GOnetstat.Tcp()
	sockets = append(sockets, GOnetstat.Tcp6()...)
	for _, socket := range sockets {
		if socket.State == "LISTEN" {
			portListening = append(portListening, socket.Port)
		}
	}
	return portListening
}

func checkPortInList(port int64, listPorts []int64) bool {
	for _, value := range listPorts {
		if value == port {
			return true
		}
	}
	return false
}

func updateInitPort(port int64) {
	for _, value := range initPorts {
		if value == port {
			continue
		}
		initPorts = append(initPorts, port)
	}
}

func main() {
	go func() {
		for {
			portsListing := getPortListenning()
			// go func ()  {
			for _, port := range initPorts {
				if checkPortInList(port, portsListing) {
					socketTCPListening.WithLabelValues(strconv.FormatInt(port, 10)).Set(1)
				} else {
					metric, err := socketTCPListening.GetMetricWithLabelValues(strconv.FormatInt(port, 10))
					if err != nil {
						log.Error(err)
						return
					}
					metric.Set(0)
				}
			}
			// }()
			// for _, port := range(portsListing){
			//   updateInitPort(port)
			// }
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to server on port :8001")
	log.Fatal(http.ListenAndServe(":8001", nil))
}
