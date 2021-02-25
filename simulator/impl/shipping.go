/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

// Address for sender and recipient
type Address struct {
	UID           string  `json:"-"`
	Street        string  `json:"street"`
	City          string  `json:"city"`
	StateProvince string  `json:"state-province"`
	PostalCd      string  `json:"postal-code"`
	Country       string  `json:"country"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
}

// Package describes attributes of a package; json attributes will be stored in QR code
type Package struct {
	UID             string   `json:"uid"`
	QRCode          []byte   `json:"-"`
	HandlingCd      string   `json:"handling"`
	Product         string   `json:"-"`
	Height          float64  `json:"-"`
	Width           float64  `json:"-"`
	Depth           float64  `json:"-"`
	Weight          float64  `json:"-"`
	DryIceWeight    float64  `json:"-"`
	Carrier         string   `json:"carrier"`
	CreatedTime     string   `json:"created"`
	EstPickupTime   string   `json:"-"`
	EstDeliveryTime string   `json:"-"`
	Sender          string   `json:"sender"`
	From            *Address `json:"from"`
	Recipient       string   `json:"recipient"`
	To              *Address `json:"to"`
}

// Content contained in a package
type Content struct {
	UID            string `json:"-"`
	Product        string `json:"product"`
	Description    string `json:"description"`
	Producer       string `json:"producer"`
	ItemCount      int    `json:"count"`
	StartLotNumber string `json:"start-lot-number"`
	EndLotNumber   string `json:"end-lot-number"`
}

// PackageRequest defines JSON string for a shipment request
type PackageRequest struct {
	UID          string   `json:"uid,omitempty"`
	HandlingCd   string   `json:"handling"`
	Height       float64  `json:"height"`
	Width        float64  `json:"width"`
	Depth        float64  `json:"depth"`
	Weight       float64  `json:"weight"`
	DryIceWeight float64  `json:"dry-ice-weight,omitempty"`
	Sender       string   `json:"sender"`
	From         *Address `json:"from"`
	Recipient    string   `json:"recipient"`
	To           *Address `json:"to"`
	Content      *Content `json:"content"`
}

// PackageResponse returns data of newly created shipping label
type PackageResponse struct {
	UID             string   `json:"uid"`
	HandlingCd      string   `json:"handling"`
	Product         string   `json:"product"`
	Carrier         string   `json:"carrier"`
	CreatedTime     string   `json:"created"`
	EstPickupTime   string   `json:"estimated-pickup"`
	EstDeliveryTime string   `json:"estimated-delivery"`
	Sender          string   `json:"sender"`
	From            *Address `json:"from"`
	Recipient       string   `json:"recipient"`
	To              *Address `json:"to"`
}

// PrintShippingLabel processes a PackageConfig JSON request
func PrintShippingLabel(request string) ([]byte, error) {
	req := &PackageRequest{}
	err := json.Unmarshal([]byte(request), req)
	if err != nil {
		return nil, err
	}
	pkg, err := initializePackage(req)
	if err != nil {
		return nil, err
	}
	req.Content.UID = pkg.UID + "-1"

	graph, err := GetTGConnection()
	if err != nil {
		return nil, err
	}
	node, err := upsertPackage(graph, pkg)
	if err != nil {
		return nil, err
	}
	err = addPackageContent(graph, node, req.Content)

	resp := &PackageResponse{
		UID:             pkg.UID,
		HandlingCd:      pkg.HandlingCd,
		Product:         pkg.Product,
		Carrier:         pkg.Carrier,
		CreatedTime:     pkg.CreatedTime,
		EstPickupTime:   pkg.EstPickupTime,
		EstDeliveryTime: pkg.EstDeliveryTime,
		Sender:          pkg.Sender,
		From:            pkg.From,
		Recipient:       pkg.Recipient,
		To:              pkg.To,
	}
	return json.Marshal(resp)
}

func initializePackage(req *PackageRequest) (*Package, error) {
	pkg := &Package{
		HandlingCd:   req.HandlingCd,
		Height:       req.Height,
		Width:        req.Width,
		Depth:        req.Depth,
		Weight:       req.Weight,
		DryIceWeight: req.DryIceWeight,
		Sender:       req.Sender,
		From:         req.From,
		Recipient:    req.Recipient,
		To:           req.To,
	}

	// select pickup office
	origin := findOfficeByState(pkg.From.StateProvince)
	if origin == nil {
		return nil, fmt.Errorf("sender state '%s' is not serviced by any carrier", pkg.From.StateProvince)
	}
	if pkg.From.Latitude*pkg.From.Longitude <= 0 {
		lat, lon := randomGPSLocation(origin)
		pkg.From.Latitude = lat
		pkg.From.Longitude = lon
		pkg.From.UID = createFnvHash(pkg.From)
	}
	pickupDelay := localDelayHours(pkg.From.Latitude, pkg.From.Longitude, origin)

	// select destination office
	dest := findOfficeByState(pkg.To.StateProvince)
	if dest == nil {
		return nil, fmt.Errorf("recipient state '%s' is not serviced by any carrier", pkg.To.StateProvince)
	}
	if pkg.To.Latitude*pkg.To.Longitude <= 0 {
		lat, lon := randomGPSLocation(dest)
		pkg.To.Latitude = lat
		pkg.To.Longitude = lon
	}
	pkg.To.UID = createFnvHash(pkg.To)
	deliveryDelay := localDelayHours(pkg.To.Latitude, pkg.To.Longitude, dest)

	// set package attributes
	pkg.Product = req.Content.Product
	pkg.Carrier = origin.Carrier
	pkg.CreatedTime = time.Now().Format(time.RFC3339)
	pkg.UID = createFnvHash(pkg)
	pickupTime := estimatePUDTime(origin.GMTOffset, pickupDelay)
	deliveryTime := estimatePUDTime(dest.GMTOffset, deliveryDelay)
	dd := pickupTime.YearDay() - deliveryTime.YearDay() + 1
	if dd > 0 {
		deliveryTime = deliveryTime.Add(time.Hour * time.Duration(dd*24))
	}
	pkg.EstPickupTime = pickupTime.Format(time.RFC3339)
	pkg.EstDeliveryTime = deliveryTime.Format(time.RFC3339)

	// generate QR code containing package json doc
	qrbytes, err := json.Marshal(pkg)
	if err != nil {
		fmt.Println("Failed to marshal package data", err)
		return nil, err
	}
	qrcode, err := createQRCode(string(qrbytes))
	if err != nil {
		fmt.Println("Failed to create QR code", err)
		return nil, err
	}
	pkg.QRCode = qrcode

	return pkg, nil
}

// estimate pickup and delivery time assuming start at 8:00 am local time, with local delay in hours
func estimatePUDTime(gmtOffset string, delay float64) time.Time {
	// construct time at specified event HH:mm and GMT offset
	c := time.Now()
	d := c.Format("2006-01-02")
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT%s:00%s", d, "08:00", gmtOffset))
	if err != nil {
		t = time.Now()
	}

	// add local delay if pickup already started for today
	if t.Before(c) {
		t = t.Add(time.Hour * time.Duration(24))
	}
	t = t.Add(time.Minute * time.Duration(int(delay*60)))

	return t
}

// correct an estimated time to be after a reference time by adding days
func correctTimeByDays(estimated, after time.Time) time.Time {
	if estimated.After(after) {
		return estimated
	}
	dd := after.YearDay() - estimated.YearDay()
	t := estimated
	if dd > 0 {
		t = estimated.Add(time.Hour * time.Duration(dd*24))
	}
	if t.Before(after) {
		t = t.Add(time.Hour * time.Duration(24))
	}
	return t
}

// returns random GPS (latitude, longitude) within the 0.2 degree distance from the office location
func randomGPSLocation(office *Office) (float64, float64) {
	dlat := -0.2 + rand.Float64()*0.4
	dlon := -0.2 + rand.Float64()*0.4
	return math.Round((office.Latitude+dlat)*10000) / 10000, math.Round((office.Longitude+dlon)*10000) / 10000
}

// calculate local pickup/delivery delay in hours based on distance from office
func localDelayHours(latitude, longitude float64, office *Office) float64 {
	dlat := math.Abs(latitude - office.Latitude)
	dlon := math.Abs(longitude - office.Longitude)
	return 7.0 * (dlat + dlon) / 0.4
}

// return FNV-1a hash of an object using JSON encoder
func createFnvHash(data interface{}) string {
	h := fnv.New64a()
	json.NewEncoder(h).Encode(data)
	return fmt.Sprintf("%x", h.Sum64())
}

// create png image for QR code containing specified data, return content of resulting png image
func createQRCode(data string) ([]byte, error) {
	qrWriter := qrcode.NewQRCodeWriter()
	hints := map[gozxing.EncodeHintType]interface{}{
		gozxing.EncodeHintType_ERROR_CORRECTION: decoder.ErrorCorrectionLevel_M,
	}
	matrix, err := qrWriter.Encode(data, gozxing.BarcodeFormat_QR_CODE, 250, 250, hints)
	if err != nil {
		return nil, err
	}

	// create PNG file
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.White
			if matrix.Get(x, y) {
				c = color.Black
			}
			img.Set(x, y, c)
		}
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decode png image to get text from QR code
func readQRCode(png []byte) (string, error) {
	bytes.NewReader(png)
	img, _, err := image.Decode(bytes.NewReader(png))
	if err != nil {
		return "", err
	}

	// prepare BinaryBitmap
	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return "", err
	}
	return result.GetText(), nil
}

// PickupPackage simulates pickup of a package of specified uid
func PickupPackage(packageID string) error {

	graph, err := GetTGConnection()
	if err != nil {
		return err
	}
	pkg, err := queryPackageInfo(graph, packageID)
	if err != nil {
		return err
	}

	originOffice := findOfficeByState(pkg.From.StateProvince)
	if originOffice == nil {
		return fmt.Errorf("No office serves sender state %s", pkg.From.StateProvince)
	}

	hubTime, err := handlePickup(graph, pkg, originOffice)
	if err != nil {
		return err
	}

	destOffice := findOfficeByState(pkg.To.StateProvince)
	if destOffice == nil {
		return fmt.Errorf("No office serves recipient state %s", pkg.To.StateProvince)
	}
	if destOffice.Carrier != originOffice.Carrier {
		originHub, ok := Hubs[originOffice.Carrier]
		if !ok {
			return fmt.Errorf("No hub office defined for carrier %s", originOffice.Carrier)
		}
		destHub, ok := Hubs[destOffice.Carrier]
		if !ok {
			return fmt.Errorf("No hub office defined for carrier %s", destOffice.Carrier)
		}
		err = handleTransfer(graph, pkg, originHub, destHub, hubTime)
		if err != nil {
			return err
		}
	}
	if _, err = handleDelivery(graph, pkg, destOffice, hubTime); err != nil {
		return err
	}

	// notify blockchain if there are threshold violations
	if mms, err := queryThresholdViolation(graph, packageID); err == nil && len(mms) > 0 {
		for c, m := range mms {
			if err := sendTemperatureUpdate(packageID, c, m); err != nil {
				fmt.Println("failed to send temperature violation to blockchain", err)
			}
		}
	}
	return err
}

// QueryPackageTimeline return transit timeline of a package of specified uid
func QueryPackageTimeline(packageID string) ([]byte, error) {

	graph, err := GetTGConnection()
	if err != nil {
		return nil, err
	}

	transit, err := queryPackageTransit(graph, packageID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(transit)
}

// Measurement is randomly generated measurement against a threshold
type Measurement struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	MinValue    float64
	MaxValue    float64
	InViolation bool
}

// randomly generate a period of threshold violation covering 1% of the total period. violationRate is the rate for including a violation period.
func randomThresholdViolation(periodStart, periodEnd time.Time, minValue, maxValue float64, violationRate float64) []*Measurement {
	startSecond := periodStart.Unix()
	endSecond := periodEnd.Unix()

	var violation *Measurement
	violationPeriod := int64(rand.Float64() * float64(endSecond-startSecond) / 10.0)
	if rand.Float64() < violationRate && violationPeriod > 0 {
		violationStart := startSecond + int64(rand.Float64()*float64(endSecond-startSecond))
		violationEnd := violationStart + violationPeriod
		if violationEnd > endSecond {
			violationEnd = endSecond
		}
		violation = &Measurement{
			PeriodStart: time.Unix(violationStart, 0),
			PeriodEnd:   time.Unix(violationEnd, 0),
			InViolation: true,
		}
		violation.MinValue, violation.MaxValue = randomMeasurementRange(maxValue, 2*maxValue-minValue)
	}

	var result []*Measurement
	if violation == nil {
		// do not generate violation period
		m := &Measurement{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			InViolation: false,
		}
		m.MinValue, m.MaxValue = randomMeasurementRange(minValue, maxValue)
		result = append(result, m)
	} else {
		if periodStart.Before(violation.PeriodStart) {
			m := &Measurement{
				PeriodStart: periodStart,
				PeriodEnd:   violation.PeriodStart,
				InViolation: false,
			}
			m.MinValue, m.MaxValue = randomMeasurementRange(minValue, maxValue)
			result = append(result, m)
		}
		result = append(result, violation)
		if violation.PeriodEnd.Before(periodEnd) {
			m := &Measurement{
				PeriodStart: violation.PeriodEnd,
				PeriodEnd:   periodEnd,
				InViolation: false,
			}
			m.MinValue, m.MaxValue = randomMeasurementRange(minValue, maxValue)
			result = append(result, m)
		}
	}
	return result
}

func randomMeasurementRange(minValue, maxValue float64) (float64, float64) {
	nv1 := minValue + rand.Float64()*(maxValue-minValue)
	nv1 = math.Round(nv1*100) / 100
	nv2 := minValue + rand.Float64()*(maxValue-minValue)
	nv2 = math.Round(nv2*100) / 100
	if nv1 < nv2 {
		return nv1, nv2
	}
	return nv2, nv1
}

// returns measurement start and end time for simulation of a container on the day after today
func measurementPeriod(schdDepart, schdArrival, departGmtOffset, arrivalGmtOffset string, delayOfDay int) (time.Time, time.Time) {
	depart := scheduledTimeOfDay(schdDepart, departGmtOffset, delayOfDay)
	depart = depart.Add(time.Hour * time.Duration(-1))
	arrival := scheduledTimeOfDay(schdArrival, arrivalGmtOffset, delayOfDay)
	arrival = arrival.Add(time.Hour * time.Duration(1))

	return depart, arrival
}

// construct time at specified schedule HH:mm and GMT offset +/-HH:mm
func scheduledTimeOfDay(schedule, gmtOffset string, delayOfDay int) time.Time {
	// construct time at specified event HH:mm and GMT offset
	c := time.Now()
	d := c.Format("2006-01-02")
	t, _ := time.Parse(time.RFC3339, fmt.Sprintf("%sT%s:00%s", d, schedule, gmtOffset))
	if delayOfDay > 0 {
		t = t.Add(time.Hour * time.Duration(delayOfDay*24))
	}
	return t
}

// TemperatureUpdate contains data sent to blockchain to update package temperature
type TemperatureUpdate struct {
	UID         string  `json:"uid"`
	ContainerID string  `json:"containerID"`
	PeriodStart string  `json:"periodStart"`
	EventTime   string  `json:"eventTime"`
	MinValue    float64 `json:"minValue"`
	MaxValue    float64 `json:"maxValue"`
	InViolation bool    `json:"inViolation"`
}

// send temperature update event to blockchain
func sendTemperatureUpdate(uid, consUID string, measurement *Measurement) error {
	utc := time.FixedZone("UTC", 0)
	msg := &TemperatureUpdate{
		UID:         uid,
		ContainerID: consUID,
		PeriodStart: measurement.PeriodStart.In(utc).Format(time.RFC3339),
		EventTime:   measurement.PeriodEnd.In(utc).Format(time.RFC3339),
		MinValue:    measurement.MinValue,
		MaxValue:    measurement.MaxValue,
		InViolation: measurement.InViolation,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, status, err := postToBlockchain(FabricConfig.BlockchainUser, FabricConfig.UpdateTemperature, data, 0)
	fmt.Println("blockchain update temperature", status, string(resp))
	return err
}

// PackageTransaction contains data sent to blockchain for key package transactions
// where PackageDetail is json serialized from PackageRequest
type PackageTransaction struct {
	UID           string  `json:"uid"`
	EventTime     string  `json:"eventTime"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Carrier       string  `json:"carrier,omitempty"`
	ToCarrier     string  `json:"toCarrier,omitempty"`
	PackageDetail string  `json:"packageDetail,omitempty"`
}

// send pickup event to blockchain
func sendPackagePickup(carrier, uid string, pickupTime time.Time, request *PackageRequest) error {
	detail, err := json.Marshal(request)
	if err != nil {
		return nil
	}
	utc := time.FixedZone("UTC", 0)
	trans := &PackageTransaction{
		UID:           uid,
		EventTime:     pickupTime.In(utc).Format(time.RFC3339),
		Latitude:      request.From.Latitude,
		Longitude:     request.From.Longitude,
		PackageDetail: string(detail),
	}
	data, err := json.Marshal(trans)
	if err != nil {
		return err
	}
	user := Carriers[carrier].BlockchainUser
	resp, status, err := postToBlockchain(user, FabricConfig.Pickup, data, 0)
	fmt.Println("pickup package", status, string(resp))
	return err
}

// send delivery event to blockchain
func sendPackageDelivery(carrier, uid string, deliveryTime time.Time, lat, lon float64) error {

	utc := time.FixedZone("UTC", 0)
	trans := &PackageTransaction{
		UID:       uid,
		EventTime: deliveryTime.In(utc).Format(time.RFC3339),
		Latitude:  lat,
		Longitude: lon,
	}
	data, err := json.Marshal(trans)
	if err != nil {
		return err
	}
	user := Carriers[carrier].BlockchainUser
	resp, status, err := postToBlockchain(user, FabricConfig.Delivery, data, 0)
	fmt.Println("deliver package", status, string(resp))
	return err
}

// send transfer event to blockchain
func sendPackageTransfer(carrier, toCarrier, uid string, transferTime time.Time, lat, lon float64) error {

	utc := time.FixedZone("UTC", 0)
	trans := &PackageTransaction{
		UID:       uid,
		EventTime: transferTime.In(utc).Format(time.RFC3339),
		ToCarrier: toCarrier,
		Latitude:  lat,
		Longitude: lon,
	}
	data, err := json.Marshal(trans)
	if err != nil {
		return err
	}
	user := Carriers[carrier].BlockchainUser
	resp, status, err := postToBlockchain(user, FabricConfig.Transfer, data, 0)
	fmt.Println("transfer package", status, string(resp))
	return err
}

// send transfer ack event to blockchain
func sendPackageTransferAck(carrier, toCarrier, uid string, ackTime time.Time, lat, lon float64) error {

	utc := time.FixedZone("UTC", 0)
	trans := &PackageTransaction{
		UID:       uid,
		EventTime: ackTime.In(utc).Format(time.RFC3339),
		Carrier:   carrier,
		Latitude:  lat,
		Longitude: lon,
	}
	data, err := json.Marshal(trans)
	if err != nil {
		return err
	}
	user := Carriers[toCarrier].BlockchainUser
	resp, status, err := postToBlockchain(user, FabricConfig.TransferAck, data, 0)
	fmt.Println("transfer package ack", status, string(resp))
	return err
}

// send POST request to blockchain service; service URL and types are defined in monitor config
func postToBlockchain(user, service string, content []byte, timeout int) ([]byte, string, error) {
	if !FabricConfig.Enabled {
		// do not send blockchain request if monitoring is disabled
		fmt.Println("Blockchain request", user, service, string(content))
		return nil, "Monitoring disabled", nil
	}

	if timeout <= 0 {
		// default time out to 5 second
		timeout = 5
	}
	client := http.Client{
		Timeout: time.Duration(timeout * int(time.Second)),
	}
	url := fmt.Sprintf("%s/%s", FabricConfig.BlockchainService, service)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	if err != nil {
		return nil, "Bad request", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(user, "")

	response, err := client.Do(request)
	if err != nil {
		if response == nil {
			return nil, "Error response", err
		}
		return nil, response.Status, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		if response == nil {
			return nil, "Error response", err
		}
		return nil, response.Status, err
	}
	return data, response.Status, nil
}
