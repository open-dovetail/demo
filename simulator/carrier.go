/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Carriers is configurations from config file
var Carriers map[string]*Carrier

// Hubs caches carrier's hub offices
var Hubs map[string]*Office

// Carrier defines a carrier and its office locations
type Carrier struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Offices     map[string]*Office `json:"offices"`
}

// Office defines an office location of a carrier
type Office struct {
	Iata        string  `json:"iata"`
	IsHub       bool    `json:"hub"`
	Description string  `json:"description"`
	GMTOffset   string  `json:"gmtOffset"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	State       string  `json:"state"`
	Routes      map[string]*Route
}

// Address for sender and recipient
type Address struct {
	UID           string
	Street        string
	City          string
	StateProvince string
	PostalCd      string
	Country       string
	Longitude     float64
	Latitude      float64
}

// Route generated for carriers
type Route struct {
	RouteNbr        string
	RouteType       string
	SchdDepartTime  string
	SchdArrivalTime string
	From            *Office
	To              *Office
}

// read configure file to populate Carriers for test
func readConfig(configFile string) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	rand.Seed(time.Now().UnixNano())
	Carriers = make(map[string]*Carrier)
	err = json.Unmarshal(data, &Carriers)
	if err != nil {
		return err
	}
	Hubs = make(map[string]*Office)
	for n, c := range Carriers {
		c.Name = n
		for i, v := range c.Offices {
			v.Iata = i
			if len(v.Description) > 0 {
				tokens := strings.Split(v.Description, ",")
				if len(tokens) > 1 {
					v.State = strings.TrimSpace(tokens[1])
				}
			}
			if len(v.GMTOffset) > 0 {
				ch := v.GMTOffset[0:1]
				if ch != "+" && ch != "-" {
					v.GMTOffset = "+" + v.GMTOffset
				}
			}
			if v.IsHub {
				Hubs[n] = v
			}
		}
	}
	return nil
}

// iterate over Carrier's offices to find the first office in a state
func findOfficeByState(state string) *Office {
	for _, c := range Carriers {
		for _, v := range c.Offices {
			if v.State == state {
				return v
			}
		}
	}
	return nil
}

// returns random GPS (latitude, longitude) within the 0.2 degree distance from the office location
func randomGPSLocation(office *Office) (float64, float64) {
	dlat := -0.2 + rand.Float64()*0.4
	dlon := -0.2 + rand.Float64()*0.4
	return office.Latitude + dlat, office.Longitude + dlon
}

// calculate local pickup/delivery delay in hours based on distance from office
func localDelayHours(latitude, longitude float64, office *Office) float64 {
	dlat := math.Abs(latitude - office.Latitude)
	dlon := math.Abs(longitude - office.Longitude)
	return 7.0 * (dlat + dlon) / 0.4
}

func flightTime(from, to *Office) float64 {
	dlat := from.Latitude - to.Latitude
	dlon := from.Longitude - to.Longitude
	dist := math.Sqrt(dlat*dlat + dlon*dlon)
	return dist * 4.0 / 30.0
}

// return (hour, minute) for a GMT offset of format +HH:mm
func parseGMTOffset(offset string) (int, int) {
	tokens := strings.Split(offset, ":")
	h, err := strconv.Atoi(tokens[0])
	if err != nil {
		return 0, 0
	}
	var m int
	if len(tokens) > 1 {
		m, _ = strconv.Atoi(tokens[1])
	}
	return h, m
}

// estimate local arrival time at destination office in format HH:mm
func arrivalTime(depart string, from, to *Office) string {
	// get destination timezone location
	toTime := fmt.Sprintf("2000-01-01T%s:00%s", depart, to.GMTOffset)
	t, err := time.Parse(time.RFC3339, toTime)
	if err != nil {
		return ""
	}
	loc := t.Location()

	// convert depart time to destination timezone
	fromTime := fmt.Sprintf("2000-01-01T%s:00%s", depart, from.GMTOffset)
	t, _ = time.Parse(time.RFC3339, fromTime)
	t = t.In(loc)

	// add flight time
	ft := int(flightTime(from, to) * 60)
	t = t.Add(time.Minute * time.Duration(ft))
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

func createRoutes(carrier *Carrier) {
	hub := Hubs[carrier.Name]
	seq := 0
	for _, v := range carrier.Offices {
		v.Routes = make(map[string]*Route)
		if !v.IsHub {
			// outbound flight from hub
			seq++
			rn := fmt.Sprintf("%s%03d", carrier.Name, seq)
			r := &Route{
				RouteNbr:        rn,
				RouteType:       "A",
				SchdDepartTime:  "16:00",
				SchdArrivalTime: arrivalTime("16:00", hub, v),
				From:            hub,
				To:              v,
			}
			v.Routes[rn] = r

			// inbound flight to hub
			seq++
			rn = fmt.Sprintf("%s%03d", carrier.Name, seq)
			r = &Route{
				RouteNbr:        rn,
				RouteType:       "A",
				SchdDepartTime:  "00:00",
				SchdArrivalTime: arrivalTime("16:00", v, hub),
				From:            v,
				To:              hub,
			}
			v.Routes[rn] = r
		}

		// local ground truck route
		seq++
		rn := fmt.Sprintf("%s%03d", carrier.Name, seq)
		r := &Route{
			RouteNbr:        rn,
			RouteType:       "G",
			SchdDepartTime:  "08:00",
			SchdArrivalTime: "15:00",
			From:            v,
			To:              v,
		}
		v.Routes[rn] = r
	}
}
