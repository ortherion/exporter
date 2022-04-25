package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"unicode"
)

const (
	Namespace = "BaseStation"
)


var (
	TcpAddress			  = "localhost:8080" //Tcp Address to Listen Node
	ListenAddress         = "localhost:9091" 			 //Address to listen on for web interface and telemetry
	MetricPath            = `/metrics`		 //Path under which to expose metrics
	T Telemetry
)

type Telemetry struct {
	Id string `json:"id"`
	Timestamp string `json:"timestamp"`
	Uptime string `json:"uptime"`
	Uplink string `json:"uplink"`
	Downlink string `json:"downlink"`
	LimitConnection string `json:"limit_connection"`
	MainPower string `json:"main_power"`
	Battery string `json:"battery"`
}

type Exporter struct {
	Up           prometheus.Gauge
	TotalScrapes prometheus.Counter
}

func (e *Exporter) Describe(ch chan <- *prometheus.Desc)  {
	ch <- e.Up.Desc()
	ch <- e.TotalScrapes.Desc()
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric)  {
	e.Up.Set(1)
	e.TotalScrapes.Inc()
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "id"),
			"Current id", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Id))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "timestamp"),
			"Timestamp", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Timestamp))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "uptime"),
			"uptime", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Uptime))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "uplink"),
			"uplink", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Uplink))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "downlink"),
			"downlink", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Downlink))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "battery"),
			"battery level", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.Battery))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "limit_connection"),
			"Connection is limited", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.LimitConnection))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(
			Namespace, "base_station", "main_power"),
			"Battery or 220V", nil, nil),
		prometheus.GaugeValue, ConvertToString(T.MainPower))
}

func (e Exporter) Collect(ch chan <- prometheus.Metric)  {
	e.scrape(ch)
	ch <- e.Up
	ch <- e.TotalScrapes
}

func NewExporter() *Exporter {
	return &Exporter{
		Up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "exporter",
			Name: "up",
			Help: "whether exporter is up",
		}),
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "exporter",
			Name: "scrapes_total",
			Help: "Total number of scrapes",
		}),
	}
}

func ConvertToString(str string) float64 {
	var res float64
	rstr := []rune(str)
	for _, valr := range rstr {
		if !unicode.IsDigit(valr) {
			val, err := strconv.ParseInt(str, 16, 64)
			if err != nil {
				log.Println("Convert Error")
			}
			res = float64(val)
			break

		} else {
			val, err := strconv.ParseFloat(str, 64)
			if err != nil {
				log.Println("Convert Error")
			}
			res = val
			break

		}
	}
	return res
}

func tcpServer(addr string)  {
	log.Println("Starting TCP Server on:", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Can't receive tcp packet", err)
	}
	for {
		connection, err := listener.Accept()
		log.Println("Accept Connection")
		if err != nil {
			log.Println("Connection err",err)
		}
		go handleTCPConnection(connection, &T)
	}
}

func handleTCPConnection(conn net.Conn, telemetry *Telemetry)  {
	defer conn.Close()
	data := json.NewDecoder(conn)
	if err := data.Decode(&telemetry); err != nil {
		log.Println(err)
	}
	log.Println(T)
}


func main() {
	param := os.Args
	log.Println(param[1:])
	switch len(param) {
	case 2: TcpAddress = param[1]
	case 3: TcpAddress = param[1]
			ListenAddress = param[2]
	case 4:
		TcpAddress = param[1]
		ListenAddress = param[2]
		MetricPath = param[3]
	default:
	}

	go tcpServer(TcpAddress)

	exporter := NewExporter()
	prometheus.MustRegister(exporter)
	log.Printf("Listening on: %s", ListenAddress)

	http.Handle(MetricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head>
				<title>Base station exporter</title>
			</head>
			<body>
				<h1>Prometheus exporter for sensor metrics from Base station</h1>
				<p><a href='` + MetricPath + `'>Metrics</a></p>
			</body>
			</html>
		`))

	})
	log.Fatal(http.ListenAndServe(ListenAddress, nil))

}