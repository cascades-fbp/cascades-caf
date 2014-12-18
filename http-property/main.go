package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	caf "github.com/cascades-fbp/cascades-caf"
	httputils "github.com/cascades-fbp/cascades-http/utils"
	"github.com/cascades-fbp/cascades/components/utils"
	"github.com/cascades-fbp/cascades/runtime"
	zmq "github.com/pebbe/zmq4"
)

var (
	intervalEndpoint = flag.String("port.int", "10s", "Component's input port endpoint")
	requestEndpoint  = flag.String("port.req", "", "Component's input port endpoint")
	templateEndpoint = flag.String("port.tmpl", "", "Component's input port endpoint")
	propertyEndpoint = flag.String("port.prop", "", "Component's output port endpoint")
	responseEndpoint = flag.String("port.resp", "", "Component's output port endpoint")
	bodyEndpoint     = flag.String("port.body", "", "Component's output port endpoint")
	errorEndpoint    = flag.String("port.err", "", "Component's error port endpoint")
	jsonFlag         = flag.Bool("json", false, "Print component documentation in JSON")
	debug            = flag.Bool("debug", false, "Enable debug mode")
)

type RequestIP struct {
	URL         string              `json:"url"`
	Method      string              `json:"method"`
	ContentType string              `json:"content-type"`
	Headers     map[string][]string `json:"headers"`
}

func assertError(err error) {
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if *jsonFlag {
		doc, _ := registryEntry.JSON()
		fmt.Println(string(doc))
		os.Exit(0)
	}

	if *requestEndpoint == "" || *templateEndpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *propertyEndpoint == "" && *responseEndpoint == "" && *bodyEndpoint == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.SetFlags(0)
	if *debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	var err error

	defer zmq.Term()

	// Input sockets
	// Interval socket
	intSock, err := utils.CreateInputPort(*intervalEndpoint)
	assertError(err)
	defer intSock.Close()

	// Request socket
	reqSock, err := utils.CreateInputPort(*requestEndpoint)
	assertError(err)
	defer reqSock.Close()

	// Property template socket
	tmplSock, err := utils.CreateInputPort(*templateEndpoint)
	assertError(err)
	defer tmplSock.Close()

	// Output sockets
	// Property socket
	var propSock *zmq.Socket
	if *propertyEndpoint != "" {
		propSock, err = utils.CreateOutputPort(*propertyEndpoint)
		assertError(err)
		defer propSock.Close()
	}

	// Response socket
	var respSock *zmq.Socket
	if *responseEndpoint != "" {
		respSock, err = utils.CreateOutputPort(*responseEndpoint)
		assertError(err)
		defer respSock.Close()
	}
	// Response body socket
	var bodySock *zmq.Socket
	if *bodyEndpoint != "" {
		bodySock, err = utils.CreateOutputPort(*bodyEndpoint)
		assertError(err)
		defer bodySock.Close()
	}
	// Error socket
	var errSock *zmq.Socket
	if *errorEndpoint != "" {
		errSock, err = utils.CreateOutputPort(*errorEndpoint)
		assertError(err)
		defer errSock.Close()
	}

	// Ctrl+C handling
	utils.HandleInterruption()

	//TODO: setup input ports monitoring to close sockets when upstreams are disconnected

	// Setup socket poll items
	poller := zmq.NewPoller()
	poller.Add(intSock, zmq.POLLIN)
	poller.Add(reqSock, zmq.POLLIN)
	poller.Add(tmplSock, zmq.POLLIN)

	// This is obviously dangerous but we need it to deal with our custom CA's
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	client.Timeout = 30 * time.Second

	var (
		interval     time.Duration
		ip           [][]byte
		request      *RequestIP
		propTemplate *caf.PropertyTemplate
		httpRequest  *http.Request
	)

	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			log.Println("Error polling ports:", err.Error())
			continue
		}
		for _, socket := range sockets {
			if socket.Socket == nil {
				log.Println("ERROR: could not find socket in polling items array")
				continue
			}
			ip, err = socket.Socket.RecvMessageBytes(0)
			if err != nil {
				log.Println("Error receiving message:", err.Error())
				continue
			}
			if !runtime.IsValidIP(ip) {
				log.Println("Invalid IP:", ip)
				continue
			}
			switch socket.Socket {
			case intSock:
				interval, err = time.ParseDuration(string(ip[1]))
				log.Println("Interval specified:", interval)
			case reqSock:
				err = json.Unmarshal(ip[1], &request)
				if err != nil {
					log.Println("ERROR: failed to unmarshal request:", err.Error())
					continue
				}
				log.Println("Request specified:", request)
			case tmplSock:
				err = json.Unmarshal(ip[1], &propTemplate)
				if err != nil {
					log.Println("ERROR: failed to unmarshal template:", err.Error())
					continue
				}
				log.Println("Template specified:", propTemplate)

			default:
				log.Println("ERROR: IP from unhandled socket received!")
				continue
			}
		}
		if interval > 0 && request != nil && propTemplate != nil {
			log.Println("Component configured. Moving on...")
			break
		}
	}

	log.Println("Started...")
	ticker := time.NewTicker(interval)
	for _ = range ticker.C {
		httpRequest, err = http.NewRequest(request.Method, request.URL, nil)
		assertError(err)

		// Set the accepted Content-Type
		if request.ContentType != "" {
			httpRequest.Header.Add("Content-Type", request.ContentType)
		}

		// Set any additional headers if provided
		for k, v := range request.Headers {
			httpRequest.Header.Add(k, v[0])
		}

		response, err := client.Do(httpRequest)
		if err != nil {
			log.Printf("ERROR performing HTTP %s %s: %s\n", request.Method, request.URL, err.Error())
			if errSock != nil {
				errSock.SendMessageDontwait(runtime.NewPacket([]byte(err.Error())))
			}
			continue
		}

		resp, err := httputils.Response2Response(response)
		if err != nil {
			log.Println("ERROR converting response to reply:", err.Error())
			if errSock != nil {
				errSock.SendMessageDontwait(runtime.NewPacket([]byte(err.Error())))
			}
			continue
		}

		// Property output socket
		if propSock != nil {
			var (
				data interface{}
				buf  bytes.Buffer
				out  []byte
			)
			ts := time.Now().Unix()
			prop := &caf.Property{
				ID:        propTemplate.ID,
				Name:      propTemplate.Name,
				Type:      propTemplate.Type,
				Group:     propTemplate.Group,
				Timestamp: &ts,
			}

			tmpl, err := template.New("value").Parse(propTemplate.Template)
			if err != nil {
				log.Println("ERROR parsing the template:", err.Error())
				continue
			}
			if strings.HasSuffix(request.ContentType, "json") {
				err = json.Unmarshal(resp.Body, &data)
				if err != nil {
					log.Println("ERROR unmarshaling the JSON response:", err.Error())
					continue
				}
			} else {
				// TODO: support other content-types
				log.Printf("WARNING processing of %s is not supported", request.ContentType)
				continue
			}

			err = tmpl.Execute(&buf, data)
			if err != nil {
				log.Println("ERROR executing the template:", err.Error())
				continue
			}

			switch propTemplate.Type {
			case "string":
				v := buf.String()
				prop.StringValue = &v

			case "float":
				v, err := strconv.ParseFloat(buf.String(), 64)
				prop.Value = &v
				if err != nil {
					log.Println("ERROR parsing float:", err.Error())
					continue
				}

			case "bool":
				v, err := strconv.ParseBool(buf.String())
				prop.BoolValue = &v
				if err != nil {
					log.Println("ERROR parsing bool:", err.Error())
					continue
				}

			case "json":
				err = json.Unmarshal(buf.Bytes(), prop.JSONValue)

				if err != nil {
					log.Println("ERROR marshaling the result in JSON:", err.Error())
					continue
				}

			default:
				log.Printf("WARNING marshaling to %s is not supported", propTemplate.Type)
				continue
			}
			out, _ = json.Marshal(prop)
			propSock.SendMessage(runtime.NewPacket(out))
		}

		// Extra output sockets (e.g., for debugging)
		if respSock != nil {
			ip, err = httputils.Response2IP(resp)
			if err != nil {
				log.Println("ERROR converting reply to IP:", err.Error())
				if errSock != nil {
					errSock.SendMessageDontwait(runtime.NewPacket([]byte(err.Error())))
				}
			} else {
				respSock.SendMessage(ip)
			}
		}
		if bodySock != nil {
			bodySock.SendMessage(runtime.NewPacket(resp.Body))
		}
	}
}
