// Service Discovery using Dell OpenManage Enterprise

package main

import (
	"log"
	"net"
	"net/url"
	"strconv"
)

type ScrapeConfig struct {
	Targets        []string          `json:"targets"`
	Labels         map[string]string `json:"labels"`
	targetsFetched int
}

func buildTarget(device string) (string, bool) {
	u, err := url.Parse(device)
	if err != nil {
		log.Printf("Could build target host %. Error: ", device, err.Error())
		return device, false
	}
	host, _, _ := net.SplitHostPort(u.Host)
	return "https://" + config.Hostname + "?target=" + host, true
}

func omeFindAllTargets(ome *OmeConfig, sc *ScrapeConfig) bool {
	step := 100
	path := "/api/DeviceService/Devices?$top=" + strconv.Itoa(step) + "&skip=" + strconv.Itoa(sc.targetsFetched)
	log.Print("Fetching devices from OME at " + path)
	data, ok := redfishGet(&ome.host, path)
	if !ok {
		return false
	}
	values := data["value"].(list)
	sc.targetsFetched += len(values)
	for _, v := range values {
		device := v.(dict)
		var idracUrl string
		iDRAC := false
		dms := device["DeviceManagement"].(list)
		for _, dm := range dms {
			d := dm.(dict)
			mps := d["ManagementProfile"].(list)
			for _, mp := range mps {
				m := mp.(dict)
				agentname, ok := m["AgentName"].(string)
				if ok && agentname == "iDRAC" {
					iDRAC = true
					url, ok := m["ManagementURL"].(string)
					if ok && url != "" {
						idracUrl = url
					}
				}
			}
		}

		if iDRAC {
			target, ok := buildTarget(idracUrl)
			if ok {
				sc.Targets = append(sc.Targets, target)
			} else {
				log.Print("could not build target")
				log.Print(target)
			}
		}
	}

	_, hasmore := data["@odata.nextLink"].(string)

	if hasmore {
		return omeFindAllTargets(ome, sc)
	}

	return true
}
