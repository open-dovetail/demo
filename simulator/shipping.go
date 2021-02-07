/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
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
	Longitude     float64 `json:"-"`
	Latitude      float64 `json:"-"`
}

// Package describes attributes of a package
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

// PackageConfig defines JSON string for a shipment request
type PackageConfig struct {
	HandlingCd   string   `json:"handling"`
	Height       float64  `json:"height"`
	Width        float64  `json:"width"`
	Depth        float64  `json:"depth"`
	Weight       float64  `json:"weight"`
	DryIceWeight float64  `json:"dry-ice-weight"`
	Sender       string   `json:"sender"`
	From         *Address `json:"from"`
	Recipient    string   `json:"recipient"`
	To           *Address `json:"to"`
	Content      *Content `json:"content"`
}

// PrintShippingLabel processes a PackageConfig JSON request
func PrintShippingLabel(request string) error {
	req := &PackageConfig{}
	err := json.Unmarshal([]byte(request), req)
	if err != nil {
		return err
	}
	pkg, err := initializePackage(req)
	if err != nil {
		return err
	}
	fmt.Println("process package", pkg)
	return nil
}

func initializePackage(req *PackageConfig) (*Package, error) {
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
	lat, lon := randomGPSLocation(origin)
	pkg.From.UID = createFnvHash(pkg.From)
	pkg.From.Latitude = lat
	pkg.From.Longitude = lon
	pickupDelay := localDelayHours(lat, lon, origin)

	// select destination office
	dest := findOfficeByState(pkg.To.StateProvince)
	if dest == nil {
		return nil, fmt.Errorf("recipient state '%s' is not serviced by any carrier", pkg.To.StateProvince)
	}
	lat, lon = randomGPSLocation(dest)
	pkg.To.UID = createFnvHash(pkg.To)
	pkg.To.Latitude = lat
	pkg.To.Longitude = lon
	deliveryDelay := localDelayHours(lat, lon, dest)

	// set package attributes
	pkg.Product = req.Content.Product
	pkg.Carrier = origin.Carrier
	pkg.CreatedTime = time.Now().Format(time.RFC3339)
	pkg.UID = createFnvHash(pkg)
	pickupTime := estimatePUDTime(origin.GMTOffset, pickupDelay)
	deliveryTime := estimatePUDTime(dest.GMTOffset, deliveryDelay)
	dd := deliveryTime.YearDay() - pickupTime.YearDay()
	if dd < 1 {
		deliveryTime = deliveryTime.Add(time.Hour * time.Duration((1-dd)*24))
	}
	pkg.EstPickupTime = pickupTime.Format(time.RFC3339)
	pkg.EstDeliveryTime = deliveryTime.Format(time.RFC3339)

	// generate QR code
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
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT%s:00%s", d, "8:00", gmtOffset))
	if err != nil {
		t = time.Now()
	}

	// add local delay
	t = t.Add(time.Minute * time.Duration(int(delay*60)))
	if t.Before(c) {
		t = t.Add(time.Hour * time.Duration(24))
	}
	return t
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

// return FNV-1a hash of an object using JSON encoder
func createFnvHash(data interface{}) string {
	h := fnv.New64a()
	json.NewEncoder(h).Encode(data)
	return fmt.Sprintf("%x", h.Sum64())
}

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
