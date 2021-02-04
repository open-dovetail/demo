/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var configFile = "./config.json"

func setup() error {
	return readConfig(configFile)
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
	assert.Equal(t, 4, len(Carriers["SLS"].Offices), "SLS should have 4 offices")
	assert.Equal(t, "SLS", Carriers["SLS"].Name, "SLS should have a name 'SLS'")
	assert.Equal(t, 4, len(Carriers["NLS"].Offices), "NLS should have 4 offices")
	assert.Equal(t, "DEN", Carriers["SLS"].Offices["DEN"].Iata, "Denver IATA should be 'DEN'")
	assert.Equal(t, "-07:00", Carriers["SLS"].Offices["DEN"].GMTOffset, "Denver GMT offset should be '-07:00'")
	assert.Equal(t, -104.9903, Carriers["SLS"].Offices["DEN"].Longitude, "DEN's longitude should be -104.9903")
	assert.Equal(t, "CO", Carriers["SLS"].Offices["DEN"].State, "DEN's state should be 'CO'")
	assert.Equal(t, "DEN", Hubs["NLS"].Iata, "NLS hub should be 'DEN'")
}

func TestRandomAddress(t *testing.T) {
	fmt.Println("TestRandomAddress")
	addr := &Address{
		StateProvince: "CO",
	}
	office := findOfficeByState(addr.StateProvince)
	assert.NotNil(t, office, "office in CO should not be nil")
	assert.Equal(t, "DEN", office.Iata, "office IATA should be 'DEN'")
	addr.Latitude, addr.Longitude = randomGPSLocation(office)
	// fmt.Printf("office %v address %v\n", office, addr)
	delay := localDelayHours(addr.Latitude, addr.Longitude, office)
	// fmt.Printf("time delay %f\n", delay)
	assert.Less(t, delay, 7.0, "local time delay should be less than 7 hours")
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
