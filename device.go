package onvif

import (
	"strings"
)

var deviceXMLNs = []string{
	// `xmlns:tds="http://www.onvif.org/ver10/device/wsdl"`,
	// `xmlns:tt="http://www.onvif.org/ver10/schema"`,
	`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
	`xmlns:xsd="http://www.w3.org/2001/XMLSchema"`,
}

// GetInformation fetch information of ONVIF camera
func (device Device) GetInformation() (DeviceInformation, error) {
	// Create SOAP
	soap := SOAP{
		Body:     "<GetDeviceInformation  xmlns=\"http://www.onvif.org/ver10/media/wsdl\"/>",
		XMLNs:    deviceXMLNs,
		User:     device.User,
		Password: device.Password,
		URI:      "/onvif/device_service",
		Method:   "POST",
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return DeviceInformation{}, err
	}

	// Parse response to interface
	deviceInfo, err := response.ValueForPath("Envelope.Body.GetDeviceInformationResponse")
	if err != nil {
		return DeviceInformation{}, err
	}
	//fmt.Println("deviceInfo:", deviceInfo)
	// Parse interface to struct
	result := DeviceInformation{}
	if mapInfo, ok := deviceInfo.(map[string]interface{}); ok {
		result.Manufacturer = interfaceToString(mapInfo["Manufacturer"])
		result.Model = interfaceToString(mapInfo["Model"])
		result.FirmwareVersion = interfaceToString(mapInfo["FirmwareVersion"])
		result.SerialNumber = interfaceToString(mapInfo["SerialNumber"])
		result.HardwareID = interfaceToString(mapInfo["HardwareId"])
	}

	return result, nil
}

func (device Device) GetServices() (DeviceServices, error) {

	soap := SOAP{
		XMLNs: deviceXMLNs,
		Body: `<GetServices xmlns="http://www.onvif.org/ver10/device/wsdl">
			<IncludeCapability>false</IncludeCapability>
		</GetServices>`,
		User:     device.User,
		Password: device.Password,
		URI:      "/onvif/device_service",
		Method:   "POST",
	}
	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return DeviceServices{}, err
	}

	// Parse response to interface
	deviceServices, err := response.ValueForPath("Envelope.Body.GetServicesResponse")
	if err != nil {
		return DeviceServices{}, err
	}
	//fmt.Println(deviceServices)
	// Parse interface to struct
	result := DeviceServices{}
	// if mapInfo, ok := deviceServices.(map[string]interface{}); ok {
	// 	Service := mapInfo["Service"]
	// 	fmt.Println("Service1111:", Service)
	//
	// }

	for _, v := range deviceServices.(map[string]interface{}) {
		// fmt.Print(k)
		// fmt.Print("-----------1")
		// fmt.Println(v)

		for _, v2 := range v.([]interface{}) {
			// fmt.Print(k2)
			// fmt.Print("-----------2")
			// fmt.Println(v2)
			// fmt.Println(v2.(map[string]interface{})["XAddr"])
			switch v2.(map[string]interface{})["Namespace"] {
			case "http://www.onvif.org/ver10/device/wsdl":
				result.Devices_service = "http://10.5.0.241/onvif/device_service"
			case "http://www.onvif.org/ver10/media/wsdl":
				result.Media = "http://10.5.0.241/onvif/Media"
			case "http://www.onvif.org/ver10/events/wsdl":
				result.Events = "http://10.5.0.241/onvif/Events"
			case "http://www.onvif.org/ver20/ptz/wsdl":
				result.PTZ = "http://10.5.0.241/onvif/PTZ"
			case "http://www.onvif.org/ver20/imaging/wsdl":
				result.Imageing = "http://10.5.0.241/onvif/Imaging"
			case "http://www.onvif.org/ver10/deviceIO/wsdl":
				result.DeviceIO = "http://10.5.0.241/onvif/DeviceIO"
			case "http://www.onvif.org/ver20/analytics/wsdl":
				result.Analytics = "http://10.5.0.241/onvif/Analytics"
			case "http://www.onvif.org/ver10/recording/wsdl":
				result.Recording = "http://10.5.0.241/onvif/Recording"
			case "http://www.onvif.org/ver10/search/wsdl":
				result.SearchRecording = "http://10.5.0.241/onvif/SearchRecording"
			case "http://www.onvif.org/ver10/replay/wsdl":
				result.Replay = "http://10.5.0.241/onvif/Replay"
			}

			// for k3, v3 := range v2.(map[string]interface{}) {
			// 	fmt.Print(k3)
			// 	fmt.Print("-----------3")
			// 	fmt.Println(v3)
			//
			// 	switch Namespace {
			// 	case condition:
			//
			// 	}
			//
			// }
		}

	}

	return result, nil
}

// GetCapabilities fetch info of ONVIF camera's capabilities
func (device Device) GetCapabilities() (DeviceCapabilities, error) {
	// Create SOAP
	soap := SOAP{
		XMLNs: deviceXMLNs,
		Body: `<tds:GetCapabilities>
			<tds:Category>All</tds:Category>
		</tds:GetCapabilities>`,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return DeviceCapabilities{}, err
	}

	// Get network capabilities
	envelopeBodyPath := "Envelope.Body.GetCapabilitiesResponse.Capabilities"
	ifaceNetCap, err := response.ValueForPath(envelopeBodyPath + ".Device.Network")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	netCap := NetworkCapabilities{}
	if mapNetCap, ok := ifaceNetCap.(map[string]interface{}); ok {
		netCap.DynDNS = interfaceToBool(mapNetCap["DynDNS"])
		netCap.IPFilter = interfaceToBool(mapNetCap["IPFilter"])
		netCap.IPVersion6 = interfaceToBool(mapNetCap["IPVersion6"])
		netCap.ZeroConfig = interfaceToBool(mapNetCap["ZeroConfiguration"])
	}

	// Get events capabilities
	ifaceEventsCap, err := response.ValueForPath(envelopeBodyPath + ".Events")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	eventsCap := make(map[string]bool)
	if mapEventsCap, ok := ifaceEventsCap.(map[string]interface{}); ok {
		for key, value := range mapEventsCap {
			if strings.ToLower(key) == "xaddr" {
				continue
			}

			key = strings.Replace(key, "WS", "", 1)
			eventsCap[key] = interfaceToBool(value)
		}
	}

	// Get streaming capabilities
	ifaceStreamingCap, err := response.ValueForPath(envelopeBodyPath + ".Media.StreamingCapabilities")
	if err != nil {
		return DeviceCapabilities{}, err
	}

	streamingCap := make(map[string]bool)
	if mapStreamingCap, ok := ifaceStreamingCap.(map[string]interface{}); ok {
		for key, value := range mapStreamingCap {
			key = strings.Replace(key, "_", " ", -1)
			streamingCap[key] = interfaceToBool(value)
		}
	}

	// Create final result
	deviceCapabilities := DeviceCapabilities{
		Network:   netCap,
		Events:    eventsCap,
		Streaming: streamingCap,
	}

	return deviceCapabilities, nil
}

// GetDiscoveryMode fetch network discovery mode of an ONVIF camera
func (device Device) GetDiscoveryMode() (string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetDiscoveryMode/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return "", err
	}

	// Parse response
	discoveryMode, _ := response.ValueForPathString("Envelope.Body.GetDiscoveryModeResponse.DiscoveryMode")
	return discoveryMode, nil
}

// GetScopes fetch scopes of an ONVIF camera
func (device Device) GetScopes() ([]string, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetScopes/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return nil, err
	}

	// Parse response to interface
	ifaceScopes, err := response.ValuesForPath("Envelope.Body.GetScopesResponse.Scopes")
	if err != nil {
		return nil, err
	}

	// Convert interface to array of scope
	scopes := []string{}
	for _, ifaceScope := range ifaceScopes {
		if mapScope, ok := ifaceScope.(map[string]interface{}); ok {
			scope := interfaceToString(mapScope["ScopeItem"])
			scopes = append(scopes, scope)
		}
	}

	return scopes, nil
}

// GetHostname fetch hostname of an ONVIF camera
func (device Device) GetHostname() (HostnameInformation, error) {
	// Create SOAP
	soap := SOAP{
		Body:  "<tds:GetHostname/>",
		XMLNs: deviceXMLNs,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return HostnameInformation{}, err
	}

	// Parse response to interface
	ifaceHostInfo, err := response.ValueForPath("Envelope.Body.GetHostnameResponse.HostnameInformation")
	if err != nil {
		return HostnameInformation{}, err
	}

	// Parse interface to struct
	hostnameInfo := HostnameInformation{}
	if mapHostInfo, ok := ifaceHostInfo.(map[string]interface{}); ok {
		hostnameInfo.Name = interfaceToString(mapHostInfo["Name"])
		hostnameInfo.FromDHCP = interfaceToBool(mapHostInfo["FromDHCP"])
	}

	return hostnameInfo, nil
}
