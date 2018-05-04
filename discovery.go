package onvif

import (
	"errors"
//	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/satori/go.uuid"
)

var errWrongDiscoveryResponse = errors.New("Response is not related to discovery request")

// StartDiscovery send a WS-Discovery message and wait for all matching device to respond
func StartDiscovery(duration time.Duration) ([]Device, error) {
	// Get list of interface address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []Device{}, err
	}

	// Fetch IPv4 address
	ipAddrs := []string{}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if ok && !ipAddr.IP.IsLoopback() && ipAddr.IP.To4() != nil {
			ipAddrs = append(ipAddrs, ipAddr.IP.String())
		}
	}

	// Create initial discovery results
	discoveryResults := []Device{}

	// Discover device on each interface's network
	for _, ipAddr := range ipAddrs {
		devices, err := discoverDevices(ipAddr, duration)
		if err != nil {
			return []Device{}, err
		}

		discoveryResults = append(discoveryResults, devices...)
	}

	return discoveryResults, nil
}

func discoverDevices(ipAddr string, duration time.Duration) ([]Device, error) {
	// Create WS-Discovery request
	//fmt.Println("discoverDevices:", ipAddr)

	id, err := uuid.NewV4()
	if err != nil {
		return []Device{}, err
	}
	requestID := "uuid:" + id.String()

	//fmt.Println("requestID:", requestID)

	// Create UDP address for local and multicast address
	localAddress, err := net.ResolveUDPAddr("udp4", ipAddr+":0")
	if err != nil {
		return []Device{}, err
	}

	multicastAddress, err := net.ResolveUDPAddr("udp4", "239.255.255.250:3702")
	if err != nil {
		return []Device{}, err
	}

	// Create UDP connection to listen for respond from matching device
	conn, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		return []Device{}, err
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(duration))
	if err != nil {
		return []Device{}, err
	}

	err1 := discoverMessageV1_1(requestID,conn,multicastAddress)

	err2 := discoverMessageV1_2(requestID,conn,multicastAddress)

	if err1 != nil && err2 != nil {
		return []Device{}, err1
	}

	// Create initial discovery results
	discoveryResults := []Device{}

	// Keep reading UDP message until timeout
	for {
		// Create buffer and receive UDP response
		buffer := make([]byte, 10*1024)
		_, _, err = conn.ReadFromUDP(buffer)

		// Check if connection timeout
		if err != nil {
			if udpErr, ok := err.(net.Error); ok && udpErr.Timeout() {
				break
			} else {
				return discoveryResults, err
			}
		}

		// Read and parse WS-Discovery response
		device, err := readDiscoveryResponse(requestID, buffer)
		if err != nil && err != errWrongDiscoveryResponse {
			return discoveryResults, err
		}
		//fmt.Println("readDiscoveryResponse:", device.XAddr)
		// Push device to results
		discoveryResults = append(discoveryResults, device)
	}
	return discoveryResults, nil

}

func discoverMessageV1_1(requestID string,conn *net.UDPConn,multicastAddress *net.UDPAddr) (error) {

	request := `
		<?xml version="1.0" encoding="UTF-8"?>
		<Envelope
				xmlns:tds="http://www.onvif.org/ver10/device/wsdl"
				xmlns="http://www.w3.org/2003/05/soap-envelope">
				<Header>
						<wsa:MessageID xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">` + requestID + `</wsa:MessageID>
						<wsa:To xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
						<wsa:Action xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
				</Header>
				<Body>
						<Probe
								xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
								xmlns:xsd="http://www.w3.org/2001/XMLSchema"
								xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
								<Types>tds:Device</Types>
								<Scopes />
						</Probe>
				</Body>
			</Envelope>`

	// Clean WS-Discovery message
	request = regexp.MustCompile(`\>\s+\<`).ReplaceAllString(request, "><")
	request = regexp.MustCompile(`\s+`).ReplaceAllString(request, " ")


	// Send WS-Discovery request to multicast address
	_, err := conn.WriteToUDP([]byte(request), multicastAddress)

	return err

}
func discoverMessageV1_2(requestID string,conn *net.UDPConn,multicastAddress *net.UDPAddr) (error) {

	request := `
		<?xml version="1.0" encoding="utf-8"?>
		<Envelope
				xmlns:tds="http://www.onvif.org/ver10/network/wsdl"
				xmlns="http://www.w3.org/2003/05/soap-envelope">
				<Header>
						<wsa:MessageID xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">` + requestID + `</wsa:MessageID>
						<wsa:To xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
						<wsa:Action xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
				</Header>
				<Body>
						<Probe
								xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
								xmlns:xsd="http://www.w3.org/2001/XMLSchema"
								xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
								<Types>dn:NetworkVideoTransmitter</Types>
								<Scopes />
						</Probe>
				</Body>
			</Envelope>`


	// Clean WS-Discovery message
	request = regexp.MustCompile(`\>\s+\<`).ReplaceAllString(request, "><")
	request = regexp.MustCompile(`\s+`).ReplaceAllString(request, " ")

	_, err := conn.WriteToUDP([]byte(request), multicastAddress)

	return err
}
// readDiscoveryResponse reads and parses WS-Discovery response
func readDiscoveryResponse(messageID string, buffer []byte) (Device, error) {
	// Inital result
	result := Device{}

	// Parse XML to map
	mapXML, err := mxj.NewMapXml(buffer)
	if err != nil {
		return result, err
	}

	// Check if this response is for our request
	responseMessageID, _ := mapXML.ValueForPathString("Envelope.Header.RelatesTo")
	if responseMessageID != messageID {
		return result, errWrongDiscoveryResponse
	}

	// Get device's ID and clean it
	deviceID, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.EndpointReference.Address")
	deviceID = strings.Replace(deviceID, "urn:uuid:", "", 1)

	// Get device's name
	deviceName := ""
	scopes, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.Scopes")
	for _, scope := range strings.Split(scopes, " ") {
		if strings.HasPrefix(scope, "onvif://www.onvif.org/name/") {
			deviceName = strings.Replace(scope, "onvif://www.onvif.org/name/", "", 1)
			deviceName = strings.Replace(deviceName, "_", " ", -1)
			break
		}
	}

	// Get device's xAddrs
	xAddrs, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.XAddrs")
	listXAddr := strings.Split(xAddrs, " ")
	if len(listXAddr) == 0 {
		return result, errors.New("Device does not have any xAddr")
	}

	// Finalize result
	result.ID = deviceID
	result.Name = deviceName
	result.XAddr = listXAddr[0]

	return result, nil
}
