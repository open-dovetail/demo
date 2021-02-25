/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var configFile = "../config.json"

func setup() error {

	err := Initialize(configFile)
	if err != nil {
		return nil
	}
	FabricConfig.Enabled = false
	return setupDemoGraph()
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("FAILED %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Setup successful")
	os.Exit(m.Run())
}

func TestConfig(t *testing.T) {
	fmt.Println("TestConfig")

	// verify carriers
	assert.Equal(t, 4, len(Carriers["SLS"].Offices), "SLS should have 4 offices")
	assert.Equal(t, "SLS", Carriers["SLS"].Name, "SLS should have a name 'SLS'")
	assert.Equal(t, 4, len(Carriers["NLS"].Offices), "NLS should have 4 offices")
	assert.Equal(t, "DEN", Carriers["SLS"].Offices["DEN"].Iata, "Denver IATA should be 'DEN'")
	assert.Equal(t, "-07:00", Carriers["SLS"].Offices["DEN"].GMTOffset, "Denver GMT offset should be '-07:00'")
	assert.Equal(t, -104.9903, Carriers["SLS"].Offices["DEN"].Longitude, "DEN's longitude should be -104.9903")
	assert.Equal(t, "CO", Carriers["SLS"].Offices["DEN"].State, "DEN's state should be 'CO'")
	assert.Equal(t, "DEN", Hubs["NLS"].Iata, "NLS hub should be 'DEN'")

	// verify threshods
	assert.Equal(t, 4, len(Thresholds), "config should have specified 4 products")
	thr := Thresholds["PfizerVaccine"]
	assert.Equal(t, "PfizerVaccine", thr.Name, "Threshold name should match the product name")
	assert.Equal(t, "P", thr.ItemType, "PfizerVaccine should be considered perishable")
	assert.Equal(t, float64(-80), thr.MinValue, "PfizerVaccine should be kept above -80 C")
	assert.Equal(t, float64(-60), thr.MaxValue, "PfizerVaccine should be kept below -60 C")

	// verify graph DB connection parameters
	assert.Equal(t, "tcp://127.0.0.1:8222/{dbName=shipdb}", GraphDBConfig.URL, "GraphDB should be configured to shipdb")
	assert.Equal(t, "scott", GraphDBConfig.User, "graphdb user should be configured as 'scott'")
	assert.Equal(t, "scott", GraphDBConfig.Passwd, "graphdb password should be configured as 'scott'")

	// verify fabric config parameters
	assert.Equal(t, "http://127.0.0.1:7979", FabricConfig.BlockchainService, "blockchain service url should be configured")
	assert.Equal(t, float64(0.5), FabricConfig.ViolationRate, "threshold violation rate should be configured")
	assert.Equal(t, "iot@org1", FabricConfig.BlockchainUser, "Blockchain user for monitoring should be configured")
	assert.Equal(t, "shipping/pickuppackage", FabricConfig.Pickup, "pickup request should be configured")
	assert.Equal(t, "shipping/transferpackage", FabricConfig.Transfer, "transfer request should be configured")
	assert.Equal(t, "shipping/transferpackageack", FabricConfig.TransferAck, "transfer ack request should be configured")
	assert.Equal(t, "shipping/deliverpackage", FabricConfig.Delivery, "delivery request should be configured")
	assert.Equal(t, "shipping/updatetemperature", FabricConfig.UpdateTemperature, "temperature update request should be configured")
}

func TestRandomTimestamp(t *testing.T) {
	// get location of the same timezone
	ref := time.Now().Format("2006-01-02T15:04:05") + "-05:00"
	nyt, err := time.Parse(time.RFC3339, ref)
	assert.NoError(t, err)

	// generate random timestamp
	tm := randomTimestamp("16:30", "-05:00", 5)
	v := time.Unix(tm, 0)
	v = v.In(nyt.Location())
	diff := math.Abs(float64(v.Minute() - 30))
	assert.LessOrEqual(t, diff, float64(5), "random timestamp should be less than 5 minutes")
}

func TestArrivalTime(t *testing.T) {
	fmt.Println("TestArrivalTime")
	to := &Office{
		GMTOffset: "-05:00",
		Longitude: -74.0060,
		Latitude:  40.7128,
	}
	from := &Office{
		GMTOffset: "-07:00",
		Longitude: -104.9903,
		Latitude:  39.7392,
	}
	arrival := arrivalTime("16:00", from, to)
	assert.Equal(t, "22:07", arrival, "local arrival time should be 22:07")
}

func TestCreateRoutes(t *testing.T) {
	fmt.Println("TestCreateRoutes")
	carrier := Carriers["SLS"]
	hub := carrier.Offices["DEN"]
	assert.Equal(t, 4, len(hub.Routes), "Hub should have 4 routes")
	// find an airplane route
	var route *Route
	for _, r := range hub.Routes {
		if r.RouteType == "A" {
			route = r
			break
		}
	}
	assert.Equal(t, "V", route.Vehicle.ConsType, "vehicle container type should be 'V'")
	assert.Equal(t, len(Thresholds), len(route.Vehicle.Embedded), "plane's ULD count should match number of thresholds")
	for _, uld := range route.Vehicle.Embedded {
		assert.Equal(t, "U", uld.ConsType, "airplain should contain ULDs")
		assert.Equal(t, 1, len(uld.Embedded), "ULD should contain 1 freezer")
		for _, c := range uld.Embedded {
			assert.Equal(t, "F", c.ConsType, "ULD should contain freezer")
		}
	}
}
