/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestInitializePackage(t *testing.T) {
	fmt.Println("TestInitializePackage")

	// parse sample request
	sample, err := ioutil.ReadFile("../package.json")
	assert.NoError(t, err, "read sample packcage requet should not throw error")
	req := &PackageRequest{}
	err = json.Unmarshal(sample, req)
	assert.NoError(t, err, "unmarshal sample request should not throw error")

	// initialize sample package
	pkg, err := initializePackage(req)
	assert.NoError(t, err, "initialize sample package should not throw error")

	// verify generated timestamps
	createTime, err := time.Parse(time.RFC3339, pkg.CreatedTime)
	assert.NoError(t, err, "created time should be valid")
	pickupTime, err := time.Parse(time.RFC3339, pkg.EstPickupTime)
	assert.NoError(t, err, "estimated pickup time should be valid")
	deliveryTime, err := time.Parse(time.RFC3339, pkg.EstDeliveryTime)
	assert.NoError(t, err, "estimated delivery time should be valid")
	assert.True(t, createTime.Before(pickupTime), "created time should be before pickup time")
	assert.True(t, pickupTime.Before(deliveryTime), "pickup time should be before delivery time")
	fmt.Println(pkg.CreatedTime, pkg.EstPickupTime, pkg.EstDeliveryTime)
	fmt.Println(time.Unix(createTime.Unix(), 0), time.Unix(pickupTime.Unix(), 0), time.Unix(deliveryTime.Unix(), 0))

	// verify hash IDs
	assert.Equal(t, "PfizerVaccine", pkg.Product, "product should be 'PfizerVaccine'")
	assert.Greater(t, len(pkg.UID), 0, "package UID should not be blank")
	assert.Greater(t, len(pkg.From.UID), 0, "origin address UID should not be blank")
	assert.Greater(t, len(pkg.To.UID), 0, "destination address UID should not be blank")

	// verify QR code & print out QR Code image file
	data, err := readQRCode(pkg.QRCode)
	assert.NoError(t, err, "QR code should be a readable image")
	err = ioutil.WriteFile("package.png", pkg.QRCode, 0644)
	assert.NoError(t, err, "write QR code to png file should not throw error")
	var qrdata map[string]interface{}
	err = json.Unmarshal([]byte(data), &qrdata)
	assert.NoError(t, err, "QR code should contain a valid JSON object")
	assert.Equal(t, pkg.UID, qrdata["uid"].(string), "QR Code uid should match package ID")
}

func TestRandomThresholdViolation(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour * time.Duration(2))
	measures := randomThresholdViolation(start, end, -80, -50, FabricConfig.ViolationRate)
	if len(measures) > 1 {
		// assert violation period
		violation := measures[1]
		assert.True(t, violation.InViolation, "second period should be in violation")
		assert.Less(t, float64(-50), violation.MaxValue, "violation value should be greater than -50")
	}
	// assert first period
	assert.False(t, measures[0].InViolation, "first period should not be in violation")
	assert.GreaterOrEqual(t, float64(-50), measures[0].MaxValue, "normal value should be less than -50")
}
