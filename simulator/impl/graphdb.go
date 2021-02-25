/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/yxuco/tgdb"
	"github.com/yxuco/tgdb/factory"
)

var graph *GraphManager

// GetTGConnection returns a new connection of Graph DB
func GetTGConnection() (*GraphManager, error) {
	if graph != nil {
		return graph, nil
	}

	cf := factory.GetConnectionFactory()
	conn, err := cf.CreateAdminConnection(GraphDBConfig.URL, "admin", "admin", nil)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	gof, err := conn.GetGraphObjectFactory()
	if err != nil {
		return nil, err
	}
	gmd, err := conn.GetGraphMetadata(true)
	if err != nil {
		return nil, err
	}

	graph = &GraphManager{
		conn: conn,
		gof:  gof,
		gmd:  gmd,
	}
	return graph, nil
}

// GraphManager encapsulates standard graph DB operations
type GraphManager struct {
	conn tgdb.TGConnection
	gof  tgdb.TGGraphObjectFactory
	gmd  tgdb.TGGraphMetadata
}

// CreateNode creates an empty node in default graph
func (g *GraphManager) CreateNode(typeName string) (tgdb.TGNode, tgdb.TGError) {
	nodeType, err := g.gmd.GetNodeType(typeName)
	if err != nil {
		return nil, err
	}
	return g.gof.CreateNodeInGraph(nodeType)
}

// CreateEdge creates an empty edge in default graph
func (g *GraphManager) CreateEdge(typeName string, from, to tgdb.TGNode) (tgdb.TGEdge, tgdb.TGError) {
	edgeType, err := g.gmd.GetEdgeType(typeName)
	if err != nil {
		return nil, err
	}
	return g.gof.CreateEdgeWithEdgeType(from, to, edgeType)
}

// InsertEntity inserts a node or edge into graph
func (g *GraphManager) InsertEntity(entity tgdb.TGEntity) tgdb.TGError {
	return g.conn.InsertEntity(entity)
}

// UpdateEntity marks a node or edge for update
func (g *GraphManager) UpdateEntity(entity tgdb.TGEntity) tgdb.TGError {
	return g.conn.UpdateEntity(entity)
}

// Query executes a Gremlin query
func (g *GraphManager) Query(grem string) ([]interface{}, error) {
	rset, err := g.conn.ExecuteQuery(grem, nil)
	if err != nil {
		return nil, err
	}
	if rset == nil {
		return nil, nil
	}
	return rset.ToCollection(), nil
}

// GetNodeByKey returns a node of specified type and primary key-values
func (g *GraphManager) GetNodeByKey(nodeType string, keyValues map[string]interface{}) (tgdb.TGNode, tgdb.TGError) {
	key, err := g.gof.CreateCompositeKey(nodeType)
	for k, v := range keyValues {
		key.SetOrCreateAttribute(k, v)
	}

	node, err := g.conn.GetEntity(key, nil)
	if err != nil {
		return nil, err
	}
	if node != nil {
		if result, ok := node.(tgdb.TGNode); ok {
			return result, nil
		}
	}
	return nil, nil
}

// Commit commits the current transaction
func (g *GraphManager) Commit() (tgdb.TGResultSet, tgdb.TGError) {
	return g.conn.Commit()
}

// Disconnect disconnects from TGDB server
func (g *GraphManager) Disconnect() tgdb.TGError {
	graph = nil
	return g.conn.Disconnect()
}

func createThreshold(graph *GraphManager, threshold *Threshold) (tgdb.TGNode, error) {
	node, err := graph.CreateNode("Threshold")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("name", threshold.Name)
	node.SetOrCreateAttribute("type", threshold.ItemType)
	node.SetOrCreateAttribute("minValue", threshold.MinValue)
	node.SetOrCreateAttribute("maxValue", threshold.MaxValue)
	node.SetOrCreateAttribute("uom", threshold.UOM)

	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	return node, nil
}

func createCarrier(graph *GraphManager, carrier *Carrier) (tgdb.TGNode, error) {
	node, err := graph.CreateNode("Carrier")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("name", carrier.Name)
	node.SetOrCreateAttribute("description", carrier.Description)
	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	return node, nil
}

func createOffice(graph *GraphManager, office *Office) (tgdb.TGNode, error) {
	node, err := graph.CreateNode("Office")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("iata", office.Iata)
	node.SetOrCreateAttribute("carrier", office.Carrier)
	node.SetOrCreateAttribute("description", office.Description)
	node.SetOrCreateAttribute("gmtOffset", office.GMTOffset)
	node.SetOrCreateAttribute("latitude", office.Latitude)
	node.SetOrCreateAttribute("longitude", office.Longitude)
	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	return node, nil
}

func createRoute(graph *GraphManager, route *Route) (tgdb.TGNode, error) {
	fmt.Println("create route", route.RouteNbr)
	node, err := graph.CreateNode("Route")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("routeNbr", route.RouteNbr)
	node.SetOrCreateAttribute("type", route.RouteType)
	node.SetOrCreateAttribute("fromIata", route.From.Iata)
	node.SetOrCreateAttribute("toIata", route.To.Iata)
	node.SetOrCreateAttribute("schdDepartTime", route.SchdDepartTime)
	node.SetOrCreateAttribute("schdArrivalTime", route.SchdArrivalTime)
	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	fmt.Println("inserted route", node.GetAttribute("routeNbr").GetValue())
	return node, nil
}

func createContainer(graph *GraphManager, cons *Container) (tgdb.TGNode, error) {
	fmt.Println("create container", cons.UID)
	node, err := graph.CreateNode("Container")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("uid", cons.UID)
	node.SetOrCreateAttribute("type", cons.ConsType)
	node.SetOrCreateAttribute("monitor", cons.Product)
	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	fmt.Println("inserted container", node.GetAttribute("uid").GetValue())
	return node, nil
}

func createPackage(graph *GraphManager, pkg *Package) (tgdb.TGNode, error) {
	fmt.Println("create package", pkg.UID)
	node, err := graph.CreateNode("Package")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("uid", pkg.UID)
	node.SetOrCreateAttribute("qrCode", pkg.QRCode)
	node.SetOrCreateAttribute("handlingCd", pkg.HandlingCd)
	node.SetOrCreateAttribute("product", pkg.Product)
	node.SetOrCreateAttribute("height", pkg.Height)
	node.SetOrCreateAttribute("width", pkg.Width)
	node.SetOrCreateAttribute("depth", pkg.Depth)
	node.SetOrCreateAttribute("weight", pkg.Weight)
	node.SetOrCreateAttribute("dryIceWeight", pkg.DryIceWeight)
	node.SetOrCreateAttribute("carrier", pkg.Carrier)
	if tm, err := time.Parse(time.RFC3339, pkg.CreatedTime); err == nil {
		node.SetOrCreateAttribute("createdTime", tm.Unix())
	}
	if tm, err := time.Parse(time.RFC3339, pkg.EstPickupTime); err == nil {
		node.SetOrCreateAttribute("estPickupTime", tm.Unix())
	}
	if tm, err := time.Parse(time.RFC3339, pkg.EstDeliveryTime); err == nil {
		node.SetOrCreateAttribute("estDeliveryTime", tm.Unix())
	}

	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	fmt.Println("inserted package", node.GetAttribute("uid").GetValue())
	return node, nil
}

func createContent(graph *GraphManager, cont *Content) (tgdb.TGNode, error) {
	node, err := graph.CreateNode("Content")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("uid", cont.UID)
	node.SetOrCreateAttribute("product", cont.Product)
	node.SetOrCreateAttribute("description", cont.Description)
	node.SetOrCreateAttribute("producer", cont.Producer)
	node.SetOrCreateAttribute("itemCount", cont.ItemCount)
	node.SetOrCreateAttribute("startLotNumber", cont.StartLotNumber)
	node.SetOrCreateAttribute("endLotNumber", cont.EndLotNumber)

	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	return node, nil
}

func createAddress(graph *GraphManager, addr *Address) (tgdb.TGNode, error) {
	node, err := graph.CreateNode("Address")
	if err != nil {
		return nil, err
	}
	node.SetOrCreateAttribute("uid", addr.UID)
	node.SetOrCreateAttribute("street", addr.Street)
	node.SetOrCreateAttribute("city", addr.City)
	node.SetOrCreateAttribute("stateProvince", addr.StateProvince)
	node.SetOrCreateAttribute("postalCd", addr.PostalCd)
	node.SetOrCreateAttribute("country", addr.Country)
	node.SetOrCreateAttribute("latitude", addr.Latitude)
	node.SetOrCreateAttribute("longitude", addr.Longitude)

	if err = graph.InsertEntity(node); err != nil {
		return nil, err
	}
	return node, nil
}

func createEdgeOperates(graph *GraphManager, carrier, office tgdb.TGNode) error {
	operates, err := graph.CreateEdge("operates", carrier, office)
	if err != nil {
		return err
	}
	if err := graph.InsertEntity(operates); err != nil {
		return err
	}
	_, err = graph.Commit()

	return err
}

func createEdgeSchedules(graph *GraphManager, carrier, route tgdb.TGNode) error {
	schedules, err := graph.CreateEdge("schedules", carrier, route)
	if err != nil {
		return err
	}
	if err := graph.InsertEntity(schedules); err != nil {
		return err
	}
	_, err = graph.Commit()

	return err
}

func createEdgeDeparts(graph *GraphManager, route, office tgdb.TGNode, after time.Time) (time.Time, error) {

	// calculate random depart time according to route schedule
	tm := randomTimestamp(getAttributeAsString(route, "schdDepartTime"), getAttributeAsString(office, "gmtOffset"), 5)
	departTime := time.Unix(tm, 0)
	if departTime.Before(after) {
		departTime = correctTimeByDays(departTime, after)
		tm = departTime.Unix()
	}

	departs, err := graph.CreateEdge("departs", route, office)
	if err != nil {
		return time.Time{}, err
	}
	departs.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(departs); err != nil {
		return departTime, err
	}

	_, err = graph.Commit()
	return departTime, err
}

func createEdgeArrives(graph *GraphManager, route, office tgdb.TGNode, after time.Time) (time.Time, error) {

	// calculate random arrival time according to route schedule
	tm := randomTimestamp(getAttributeAsString(route, "schdArrivalTime"), getAttributeAsString(office, "gmtOffset"), 5)
	arrivalTime := time.Unix(tm, 0)
	if arrivalTime.Before(after) {
		arrivalTime = correctTimeByDays(arrivalTime, after)
		tm = arrivalTime.Unix()
	}

	arrives, err := graph.CreateEdge("arrives", route, office)
	if err != nil {
		return time.Time{}, err
	}
	arrives.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(arrives); err != nil {
		return arrivalTime, err
	}

	_, err = graph.Commit()
	return arrivalTime, err
}

func createEdgeBuilds(graph *GraphManager, office, container tgdb.TGNode, eventTime int64) error {
	builds, err := graph.CreateEdge("builds", office, container)
	if err != nil {
		return err
	}
	builds.SetOrCreateAttribute("eventTimestamp", eventTime)
	if err := graph.InsertEntity(builds); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeAssigned(graph *GraphManager, container, route tgdb.TGNode, eventTime int64) error {
	assigned, err := graph.CreateEdge("assigned", container, route)
	if err != nil {
		return err
	}
	assigned.SetOrCreateAttribute("eventTimestamp", eventTime)
	if err := graph.InsertEntity(assigned); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeContains(graph *GraphManager, parent, child tgdb.TGNode, inTime, outTime int64, childType string) error {
	contains, err := graph.CreateEdge("contains", parent, child)
	if err != nil {
		return err
	}
	contains.SetOrCreateAttribute("eventTimestamp", inTime)
	if childType != "C" {
		// for demo, outTime is inifinte for containers
		contains.SetOrCreateAttribute("outTimestamp", outTime)
	}
	contains.SetOrCreateAttribute("childType", childType)
	if err := graph.InsertEntity(contains); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeSender(graph *GraphManager, pkg, addr tgdb.TGNode, sender string) error {
	send, err := graph.CreateEdge("sender", pkg, addr)
	if err != nil {
		return err
	}
	send.SetOrCreateAttribute("name", sender)
	if err := graph.InsertEntity(send); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeRecipient(graph *GraphManager, pkg, addr tgdb.TGNode, recipient string) error {
	receive, err := graph.CreateEdge("recipient", pkg, addr)
	if err != nil {
		return err
	}
	receive.SetOrCreateAttribute("name", recipient)
	if err := graph.InsertEntity(receive); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeContainsContent(graph *GraphManager, pkg, cont tgdb.TGNode) error {
	contains, err := graph.CreateEdge("contains", pkg, cont)
	if err != nil {
		return err
	}
	if err := graph.InsertEntity(contains); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgePickup(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, tracking string, lat, lon float64) error {
	pickup, err := graph.CreateEdge("pickup", office, pkg)
	if err != nil {
		return err
	}
	tevent := &transferEvent{
		Carrier:   getAttributeAsString(office, "carrier"),
		Direction: "from",
		Latitude:  lat,
		Longitude: lon,
	}
	pickup.SetOrCreateAttribute("eventTimestamp", eventTime)
	pickup.SetOrCreateAttribute("trackingID", tracking)
	pickup.SetOrCreateAttribute("employeeID", createFnvHash(tevent))
	pickup.SetOrCreateAttribute("longitude", lon)
	pickup.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(pickup); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeDelivery(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, lat, lon float64) error {
	delivery, err := graph.CreateEdge("delivery", office, pkg)
	if err != nil {
		return err
	}
	tevent := &transferEvent{
		Carrier:   getAttributeAsString(office, "carrier"),
		Direction: "to",
		Latitude:  lat,
		Longitude: lon,
	}
	delivery.SetOrCreateAttribute("eventTimestamp", eventTime)
	delivery.SetOrCreateAttribute("employeeID", createFnvHash(tevent))
	delivery.SetOrCreateAttribute("longitude", lon)
	delivery.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(delivery); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

func createEdgeMeasures(graph *GraphManager, cons, thr tgdb.TGNode, measurement *Measurement) error {
	measures, err := graph.CreateEdge("measures", cons, thr)
	if err != nil {
		return err
	}

	measures.SetOrCreateAttribute("startTimestamp", measurement.PeriodStart.Unix())
	measures.SetOrCreateAttribute("eventTimestamp", measurement.PeriodEnd.Unix())
	measures.SetOrCreateAttribute("minValue", measurement.MinValue)
	measures.SetOrCreateAttribute("maxValue", measurement.MaxValue)
	measures.SetOrCreateAttribute("uom", "C")
	measures.SetOrCreateAttribute("violated", measurement.InViolation)
	if err := graph.InsertEntity(measures); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

type transferEvent struct {
	Carrier   string  `json:"carrier"`
	Direction string  `json:"direction"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func createEdgeTransfers(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, tracking string, lat, lon float64, direction string) error {
	transfers, err := graph.CreateEdge("transfers", office, pkg)
	if err != nil {
		return err
	}

	tevent := &transferEvent{
		Carrier:   getAttributeAsString(office, "carrier"),
		Direction: direction,
		Latitude:  lat,
		Longitude: lon,
	}
	transfers.SetOrCreateAttribute("eventTimestamp", eventTime)
	transfers.SetOrCreateAttribute("direction", direction)
	transfers.SetOrCreateAttribute("trackingID", tracking)
	transfers.SetOrCreateAttribute("employeeID", createFnvHash(tevent))
	transfers.SetOrCreateAttribute("longitude", lon)
	transfers.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(transfers); err != nil {
		return err
	}

	_, err = graph.Commit()
	return err
}

var carrierNodes map[string]tgdb.TGNode
var officeNodes map[string]tgdb.TGNode
var routeNodes map[string]tgdb.TGNode

// InitializeGraph inserts carrier nodes and edges into TGDB
func InitializeGraph(graph *GraphManager) error {
	carrierNodes = make(map[string]tgdb.TGNode)
	officeNodes = make(map[string]tgdb.TGNode)

	// create thresholds
	for _, th := range Thresholds {
		if _, err := createThreshold(graph, th); err != nil {
			return err
		}
	}

	// create carrier and offices
	for _, c := range Carriers {
		carrier, err := createCarrier(graph, c)
		if err != nil {
			return err
		}
		// cache carrier node for further processing
		carrierNodes[c.Name] = carrier
		for _, v := range c.Offices {
			office, err := createOffice(graph, v)
			if err != nil {
				return err
			}
			if err := createEdgeOperates(graph, carrier, office); err != nil {
				return err
			}
			// cache office node for further processing
			officeNodes[v.Carrier+":"+v.Iata] = office
		}
	}
	fmt.Println("created offices", len(officeNodes))

	// create routes
	routeNodes = make(map[string]tgdb.TGNode)
	for _, c := range Carriers {
		for _, v := range c.Offices {
			fmt.Println("init routes for", c.Name, v.Iata)
			if err := initializeRoutes(graph, v); err != nil {
				return err
			}
		}
	}

	// create containers
	for _, c := range Carriers {
		for _, v := range c.Offices {
			for _, r := range v.Routes {
				// create containers for hub inbound routes
				if !v.IsHub || r.RouteType == "G" {
					fmt.Println("init container for route ", c.Name, v.Iata, r.RouteNbr, r.To.Iata)
					if err := initializeContainers(graph, r); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// create routes and containers for a specified office
func initializeRoutes(graph *GraphManager, office *Office) error {
	for _, r := range office.Routes {
		fmt.Println("init route", r.RouteNbr)
		route, err := createRoute(graph, r)
		if err != nil {
			return err
		}
		// create departs for today
		from := officeNodes[office.Carrier+":"+r.From.Iata]
		if _, err := createEdgeDeparts(graph, route, from, time.Time{}); err != nil {
			return err
		}

		// create arrival for today
		to := officeNodes[office.Carrier+":"+r.To.Iata]
		if _, err := createEdgeArrives(graph, route, to, time.Time{}); err != nil {
			return err
		}
		// create shedules rel from carrier to route
		carrier := carrierNodes[office.Carrier]
		if err := createEdgeSchedules(graph, carrier, route); err != nil {
			return err
		}
		// cache route node for further processing
		routeNodes[office.Carrier+":"+r.From.Iata+":"+r.To.Iata] = route
	}
	return nil
}

// context for building embedded containers
type containerContext struct {
	inTime     int64
	outTime    int64
	hubInTime  int64
	hubOutTime int64
}

// create containers on a specified route, return vessel container
func initializeContainers(graph *GraphManager, route *Route) error {
	v := route.Vehicle
	vessel, err := createContainer(graph, v)
	if err != nil {
		return err
	}
	// set build and route assignment time to 1 hour before departure
	tm := randomTimestamp(route.SchdDepartTime, route.From.GMTOffset, 10) - 3600
	office := officeNodes[route.From.Carrier+":"+route.From.Iata]
	if err := createEdgeBuilds(graph, office, vessel, tm); err != nil {
		return err
	}
	toHub := routeNodes[route.From.Carrier+":"+route.From.Iata+":"+route.To.Iata]
	if err := createEdgeAssigned(graph, vessel, toHub, tm); err != nil {
		return err
	}
	context := &containerContext{
		inTime:  tm,
		outTime: randomTimestamp(route.SchdArrivalTime, route.To.GMTOffset, 5),
	}
	if route.RouteType == "A" {
		// assign same vessel to returning flight as well
		htm := randomTimestamp("00:00", route.To.GMTOffset, 10) - 3600
		hub := officeNodes[route.From.Carrier+":"+route.To.Iata]
		if err := createEdgeBuilds(graph, hub, vessel, htm); err != nil {
			return err
		}
		fromHub := routeNodes[route.From.Carrier+":"+route.To.Iata+":"+route.From.Iata]
		if err := createEdgeAssigned(graph, vessel, fromHub, htm); err != nil {
			return err
		}
		context.hubInTime = htm
		context.hubOutTime = htm - context.inTime + context.outTime
	}

	// set embedded containers
	return initializeEmbeddedContainers(graph, vessel, v.Embedded, context)
}

// create embedded containers and relationships from a parent node
func initializeEmbeddedContainers(graph *GraphManager, parent tgdb.TGNode, embedded map[string]*Container, context *containerContext) error {
	for _, c := range embedded {
		child, err := createContainer(graph, c)
		if err != nil {
			return err
		}
		if err := createEdgeContains(graph, parent, child, context.inTime, context.outTime, "C"); err != nil {
			return err
		}
		// for demo, containers are bound forever, so no need to set edge for separate in and out
		//if context.hubInTime > 0 {
		//	if err := createEdgeContains(graph, parent, child, context.hubInTime, context.hubOutTime, "C"); err != nil {
		//		return err
		//	}
		//}
		if len(c.Embedded) > 0 {
			if err := initializeEmbeddedContainers(graph, child, c.Embedded, context); err != nil {
				return err
			}
		}
	}
	return nil
}

// insert a package into TGDB
func upsertPackage(graph *GraphManager, pkg *Package) (tgdb.TGNode, error) {
	key := map[string]interface{}{
		"uid": pkg.UID,
	}
	if node, err := graph.GetNodeByKey("Package", key); err == nil && node != nil {
		fmt.Println("package exist:", pkg.UID, pkg.CreatedTime)
		return node, err
	}
	node, err := createPackage(graph, pkg)
	if err != nil {
		return node, err
	}

	from, err := upsertAddress(graph, pkg.From)
	if err != nil || from == nil {
		fmt.Println("failed to create sender address:", pkg.From.UID, pkg.From.Street)
		return nil, err
	}
	if err := createEdgeSender(graph, node, from, pkg.Sender); err != nil {
		return nil, err
	}

	to, err := upsertAddress(graph, pkg.To)
	if err != nil || to == nil {
		fmt.Println("failed to create recipient address:", pkg.To.UID, pkg.To.Street)
		return nil, err
	}
	if err := createEdgeRecipient(graph, node, to, pkg.Recipient); err != nil {
		return nil, err
	}

	return node, nil
}

// add content info of a package
func addPackageContent(graph *GraphManager, pkg tgdb.TGNode, cont *Content) error {
	key := map[string]interface{}{
		"uid": cont.UID,
	}
	if node, err := graph.GetNodeByKey("Content", key); err == nil && node != nil {
		fmt.Println("content exist:", cont.UID, cont.Product)
		return nil
	}
	node, err := createContent(graph, cont)
	if err != nil || node == nil {
		return err
	}
	return createEdgeContainsContent(graph, pkg, node)
}

func upsertAddress(graph *GraphManager, addr *Address) (tgdb.TGNode, error) {
	key := map[string]interface{}{
		"uid": addr.UID,
	}
	if node, err := graph.GetNodeByKey("Address", key); err == nil && node != nil {
		fmt.Println("address exist:", addr.UID, addr.Street)
		return node, err
	}
	return createAddress(graph, addr)
}

// AddressInfo contains key data for shipping
type AddressInfo struct {
	StateProvince string
	Longitude     float64
	Latitude      float64
}

// PackageInfo contains key data for shipping
type PackageInfo struct {
	UID           string
	HandlingCd    string
	Product       string
	Carrier       string
	EstPickupTime time.Time
	From          *AddressInfo
	To            *AddressInfo
}

func getAttributeAsString(entity tgdb.TGEntity, name string) string {
	attr := entity.GetAttribute(name)
	var result interface{}
	if attr != nil {
		result = attr.GetValue()
	}
	if result == nil {
		return ""
	}
	if v, ok := result.(string); ok {
		return v
	}
	return fmt.Sprintf("%v", result)
}

func getAttributeAsUTCTime(entity tgdb.TGEntity, name string) string {
	attr := entity.GetAttribute(name)
	var result interface{}
	if attr != nil {
		result = attr.GetValue()
	}
	if result == nil {
		return ""
	}
	if v, ok := result.(time.Time); ok {
		utc := time.FixedZone("UTC", 0)
		return v.In(utc).Format(time.RFC3339)
	}
	return fmt.Sprintf("%v", result)
}

func getAttributeAsBool(entity tgdb.TGEntity, name string) bool {
	attr := entity.GetAttribute(name)
	var result interface{}
	if attr != nil {
		result = attr.GetValue()
	}
	if result == nil {
		return false
	}
	if v, ok := result.(bool); ok {
		return v
	}
	return false
}

func getAttributeAsDouble(entity tgdb.TGEntity, name string) float64 {
	attr := entity.GetAttribute(name)
	var result interface{}
	if attr != nil {
		result = attr.GetValue()
	}
	switch v := result.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	default:
		fmt.Printf("getAttributeAsDouble ignore %v of type %T\n", v, v)
		return float64(0)
	}
}

// query package and pickup/delivery address info of a specified package-ID
func queryPackageInfo(graph *GraphManager, packageID string) (*PackageInfo, error) {
	key := map[string]interface{}{
		"uid": packageID,
	}
	node, err := graph.GetNodeByKey("Package", key)
	if err != nil || node == nil {
		fmt.Println("failed to find package", packageID, err)
		return nil, err
	}
	result := &PackageInfo{
		UID:           packageID,
		HandlingCd:    getAttributeAsString(node, "handlingCd"),
		Product:       getAttributeAsString(node, "product"),
		Carrier:       getAttributeAsString(node, "carrier"),
		EstPickupTime: node.GetAttribute("estPickupTime").GetValue().(time.Time),
	}
	if addr, err := queryAddressInfo(graph, packageID, "sender"); err == nil {
		result.From = addr
	}
	if addr, err := queryAddressInfo(graph, packageID, "recipient"); err == nil {
		result.To = addr
	}
	return result, nil
}

// query sender/recipient address of a specified package
func queryAddressInfo(graph *GraphManager, packageID, addressType string) (*AddressInfo, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').outE('%s').inV();", packageID, addressType)
	nodes, err := graph.Query(query)
	if err != nil {
		return nil, err
	}
	if len(nodes) < 1 {
		return nil, fmt.Errorf("%s address not found", addressType)
	}
	node, ok := nodes[0].(tgdb.TGNode)
	if !ok {
		return nil, fmt.Errorf("query result %T is not a tgdb.TGNode", nodes[0])
	}

	return &AddressInfo{
		StateProvince: getAttributeAsString(node, "stateProvince"),
		Longitude:     getAttributeAsDouble(node, "longitude"),
		Latitude:      getAttributeAsDouble(node, "latitude"),
	}, nil
}

// query office node of a specified carrier and iata
func queryOffice(graph *GraphManager, carrier, iata string) (tgdb.TGNode, error) {
	key := map[string]interface{}{
		"iata":    iata,
		"carrier": carrier,
	}
	return graph.GetNodeByKey("Office", key)
}

// query package detail of a specified package-ID
func queryPackageDetail(graph *GraphManager, packageID string) (*PackageRequest, error) {
	key := map[string]interface{}{
		"uid": packageID,
	}
	node, err := graph.GetNodeByKey("Package", key)
	if err != nil || node == nil {
		fmt.Println("failed to find package", packageID, err)
		return nil, err
	}
	result := &PackageRequest{
		UID:        packageID,
		HandlingCd: getAttributeAsString(node, "handlingCd"),
		Height:     getAttributeAsDouble(node, "height"),
		Width:      getAttributeAsDouble(node, "width"),
		Depth:      getAttributeAsDouble(node, "depth"),
		Weight:     getAttributeAsDouble(node, "weight"),
	}

	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').outE('sender').values('name');", packageID)
	if nodes, err := graph.Query(query); err == nil && len(nodes) > 0 {
		result.Sender = nodes[0].(string)
	}

	query = fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').outE('recipient').values('name');", packageID)
	if nodes, err := graph.Query(query); err == nil && len(nodes) > 0 {
		result.Recipient = nodes[0].(string)
	}

	if addr, err := queryAddress(graph, packageID, "sender"); err == nil {
		result.From = addr
	}
	if addr, err := queryAddress(graph, packageID, "recipient"); err == nil {
		result.To = addr
	}
	if cont, err := queryContent(graph, packageID); err == nil && cont != nil {
		result.Content = cont
	}
	return result, nil
}

// query sender/recipient address of a specified package
func queryAddress(graph *GraphManager, packageID, addressType string) (*Address, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').outE('%s').inV();", packageID, addressType)
	nodes, err := graph.Query(query)
	if err != nil {
		return nil, err
	}
	if len(nodes) < 1 {
		return nil, fmt.Errorf("%s address not found", addressType)
	}
	node, ok := nodes[0].(tgdb.TGNode)
	if !ok {
		return nil, fmt.Errorf("query result %T is not a tgdb.TGNode", nodes[0])
	}

	return &Address{
		Street:        getAttributeAsString(node, "street"),
		City:          getAttributeAsString(node, "city"),
		StateProvince: getAttributeAsString(node, "stateProvince"),
		PostalCd:      getAttributeAsString(node, "postalCd"),
		Country:       getAttributeAsString(node, "country"),
		Longitude:     getAttributeAsDouble(node, "longitude"),
		Latitude:      getAttributeAsDouble(node, "latitude"),
	}, nil
}

// query content of a specified package
func queryContent(graph *GraphManager, packageID string) (*Content, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').outE('contains').inV();", packageID)
	nodes, err := graph.Query(query)
	if err != nil || len(nodes) < 1 {
		return nil, err
	}
	node, ok := nodes[0].(tgdb.TGNode)
	if !ok {
		return nil, fmt.Errorf("query result %T is not a tgdb.TGNode", nodes[0])
	}
	return &Content{
		Product:        getAttributeAsString(node, "product"),
		Description:    getAttributeAsString(node, "description"),
		Producer:       getAttributeAsString(node, "producer"),
		ItemCount:      int(getAttributeAsDouble(node, "itemCount")),
		StartLotNumber: getAttributeAsString(node, "startLotNumber"),
		EndLotNumber:   getAttributeAsString(node, "endLotNumber"),
	}, nil
}

// update graph for package pickup at specified office and send to its hub office, return the time when plane arrives at the hub
func handlePickup(graph *GraphManager, pkg *PackageInfo, office *Office) (time.Time, error) {
	var err error
	key := map[string]interface{}{
		"iata":    office.Iata,
		"carrier": office.Carrier,
	}
	origin, err := graph.GetNodeByKey("Office", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("office node is not found for %s %s", office.Carrier, office.Iata)
	}
	key = map[string]interface{}{
		"uid": pkg.UID,
	}
	node, err := graph.GetNodeByKey("Package", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("package node is not found for %s", pkg.UID)
	}

	// calculate local pickup time based on its distance from the origin office
	pickupDelay := localDelayHours(pkg.From.Latitude, pkg.From.Longitude, office)
	pickupTime, arrivalTime, err := localPickup(graph, pickupDelay, origin, node)
	if err != nil {
		return time.Time{}, err
	}
	if err := createEdgePickup(graph, origin, node, pickupTime.Unix(), pkg.UID, pkg.From.Latitude, pkg.From.Longitude); err != nil {
		return time.Time{}, err
	}
	if pkg.HandlingCd == "P" && IsMonitored(pkg.Product) {
		// record it on blockchain
		if req, err := queryPackageDetail(graph, pkg.UID); err == nil {
			err := sendPackagePickup(office.Carrier, pkg.UID, pickupTime, req)
			if err != nil {
				fmt.Println("Failed to send blockchain request for pickup", err)
			}
		}
	}
	return originRoute(graph, arrivalTime, origin, node)
}

// update local truck pickup and return pickup time and the time for truck to arrive at the origin office
func localPickup(graph *GraphManager, pickupDelay float64, origin, pkg tgdb.TGNode) (time.Time, time.Time, error) {

	// get the local route
	iata := getAttributeAsString(origin, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','G').values('routeNbr');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("no local route found at %s", iata)
	}
	routeNbr := data[0].(string)
	route, err := graph.GetNodeByKey("Route", map[string]interface{}{"routeNbr": routeNbr})
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// get last depart time of the local route
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('departs').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("pickup route depart time not found for %s", routeNbr)
	}
	departTime := data[0].(time.Time)

	// get last arrival time of the local route
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('arrives').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("pickup route arrival time not found for %s", routeNbr)
	}
	arrivalTime := data[0].(time.Time)

	if departTime.Before(time.Now()) {
		// last route time is old, so create new pickup route depart and arrival for a new day
		departTime, err = createEdgeDeparts(graph, route, origin, time.Now())
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if arrivalTime, err = createEdgeArrives(graph, route, origin, departTime); err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().values('uid');", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().has('Container','monitor','%s').values('uid');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, arrivalTime, fmt.Errorf("no container found for %s and %s", routeNbr, product)
	}
	uid := data[0].(string)
	cons, err := graph.GetNodeByKey("Container", map[string]interface{}{"uid": uid})
	if err != nil {
		return time.Time{}, arrivalTime, err
	}

	// add simulated temperature measurement
	if handling == "P" && IsMonitored(product) {
		createMonitorMeasurements(graph, cons,
			getAttributeAsString(route, "schdDepartTime"),
			getAttributeAsString(route, "schdArrivalTime"),
			getAttributeAsString(origin, "gmtOffset"),
			getAttributeAsString(origin, "gmtOffset"))
	}

	// add package to the parent container
	pickupTime := departTime.Add(time.Minute * time.Duration(int(pickupDelay*60)))
	err = createEdgeContains(graph, cons, pkg, pickupTime.Unix(), arrivalTime.Unix(), "P")
	return pickupTime, arrivalTime, err
}

// update origin route to hub and return the time for plane to arrive at the hub
func originRoute(graph *GraphManager, arrivalTime time.Time, origin, pkg tgdb.TGNode) (time.Time, error) {

	// get the origin route
	iata := getAttributeAsString(origin, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','A').values('routeNbr');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no origin route found at %s", iata)
	}
	routeNbr := data[0].(string)
	route, err := graph.GetNodeByKey("Route", map[string]interface{}{"routeNbr": routeNbr})
	if err != nil {
		return time.Time{}, err
	}

	// get last origin route depart time
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('departs').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("origin route depart time not found for %s", routeNbr)
	}
	departTime := data[0].(time.Time)

	// get last origin route arrival time at hub
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('arrives').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("origin route arrival time not found for %s", routeNbr)
	}
	hubTime := data[0].(time.Time)

	if departTime.Before(arrivalTime) {
		// last route time is old, so create origin route depart for a new day
		if departTime, err = createEdgeDeparts(graph, route, origin, arrivalTime); err != nil {
			return time.Time{}, err
		}
	}

	// retrieve hub office
	carrier := getAttributeAsString(origin, "carrier")
	toIata := getAttributeAsString(route, "toIata")
	var hub tgdb.TGNode
	key := map[string]interface{}{
		"iata":    toIata,
		"carrier": carrier,
	}
	hub, err = graph.GetNodeByKey("Office", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("hub office node is not found for %s %s", carrier, toIata)
	}

	if hubTime.Before(departTime) {
		// last route time is old, so create origin route arrival for a new day
		if hubTime, err = createEdgeArrives(graph, route, hub, departTime); err != nil {
			return time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")

	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().values('uid');", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().outE('contains').inV().has('Container','monitor','%s').values('uid');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return hubTime, fmt.Errorf("no container found for %s and %s", routeNbr, product)
	}
	uid := data[0].(string)
	cons, err := graph.GetNodeByKey("Container", map[string]interface{}{"uid": uid})
	if err != nil {
		return hubTime, err
	}

	// add simulated temperature measurement
	if handling == "P" && IsMonitored(product) {
		createMonitorMeasurements(graph, cons,
			getAttributeAsString(route, "schdDepartTime"),
			getAttributeAsString(route, "schdArrivalTime"),
			getAttributeAsString(origin, "gmtOffset"),
			getAttributeAsString(hub, "gmtOffset"))
	}

	// add package to the parent container
	err = createEdgeContains(graph, cons, pkg, departTime.Unix(), hubTime.Unix(), "P")
	return hubTime, err
}

// transfer a package between 2 hub offices of different carriers
func handleTransfer(graph *GraphManager, pkg *PackageInfo, originHub, destHub *Office, hubTime time.Time) error {
	var err error
	key := map[string]interface{}{
		"iata":    originHub.Iata,
		"carrier": originHub.Carrier,
	}
	origin, err := graph.GetNodeByKey("Office", key)
	if err != nil {
		return fmt.Errorf("office node is not found for %s %s", originHub.Carrier, originHub.Iata)
	}
	key = map[string]interface{}{
		"uid": pkg.UID,
	}
	node, err := graph.GetNodeByKey("Package", key)
	if err != nil {
		return fmt.Errorf("package node is not found for %s", pkg.UID)
	}
	if err := createEdgeTransfers(graph, origin, node, hubTime.Unix(), pkg.UID, originHub.Latitude, originHub.Longitude, "from"); err != nil {
		return err
	}

	key = map[string]interface{}{
		"iata":    destHub.Iata,
		"carrier": destHub.Carrier,
	}
	dest, err := graph.GetNodeByKey("Office", key)
	if err != nil {
		return fmt.Errorf("office node is not found for %s %s", destHub.Carrier, destHub.Iata)
	}
	ackTime := hubTime.Add(time.Second * time.Duration(30))
	if err := createEdgeTransfers(graph, dest, node, ackTime.Unix(), pkg.UID, destHub.Latitude, destHub.Longitude, "to"); err != nil {
		return err
	}

	if pkg.HandlingCd == "P" && IsMonitored(pkg.Product) {
		// record it on blockchain
		if err := sendPackageTransfer(originHub.Carrier, destHub.Carrier, pkg.UID, hubTime, originHub.Latitude, originHub.Longitude); err != nil {
			fmt.Println("Failed to send blockchain request for transfer", err)
		}
		if err := sendPackageTransferAck(originHub.Carrier, destHub.Carrier, pkg.UID, ackTime, destHub.Latitude, destHub.Longitude); err != nil {
			fmt.Println("Failed to send blockchain request for transfer ack", err)
		}
	}
	return nil
}

// update graph for package delivery from hub to the specified destination office
func handleDelivery(graph *GraphManager, pkg *PackageInfo, office *Office, hubTime time.Time) (time.Time, error) {
	var err error
	key := map[string]interface{}{
		"iata":    office.Iata,
		"carrier": office.Carrier,
	}
	dest, err := graph.GetNodeByKey("Office", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("office node is not found for %s %s", office.Carrier, office.Iata)
	}
	key = map[string]interface{}{
		"uid": pkg.UID,
	}
	node, err := graph.GetNodeByKey("Package", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("package node is not found for %s", pkg.UID)
	}

	arrivalTime, err := deliveryRoute(graph, hubTime, dest, node)

	// calculate local delivery time based on its distance from the destination office
	deliveryDelay := localDelayHours(pkg.To.Latitude, pkg.To.Longitude, office)
	deliveryTime, err := localDelivery(graph, arrivalTime, deliveryDelay, dest, node)
	if err != nil {
		return deliveryTime, err
	}
	if err := createEdgeDelivery(graph, dest, node, deliveryTime.Unix(), pkg.To.Latitude, pkg.To.Longitude); err != nil {
		return deliveryTime, err
	}
	if pkg.HandlingCd == "P" && IsMonitored(pkg.Product) {
		// record it on blockchain
		err := sendPackageDelivery(office.Carrier, pkg.UID, deliveryTime, pkg.To.Latitude, pkg.To.Longitude)
		if err != nil {
			fmt.Println("Failed to send blockchain request for delivery", err)
		}
	}
	return deliveryTime, nil
}

// update delivery route from hub and return the time for plane to arrive at the dest office
func deliveryRoute(graph *GraphManager, hubTime time.Time, dest, pkg tgdb.TGNode) (time.Time, error) {

	// get the destination route
	iata := getAttributeAsString(dest, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','toIata','%s').has('Route','type','A').values('routeNbr');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no destination route found at %s", iata)
	}
	routeNbr := data[0].(string)
	route, err := graph.GetNodeByKey("Route", map[string]interface{}{"routeNbr": routeNbr})
	if err != nil {
		return time.Time{}, err
	}

	// get last destination route depart time from hub
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('departs').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("destination route depart time not found for %s", routeNbr)
	}
	departTime := data[0].(time.Time)

	// get last destination route arrival time at destination office
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('arrives').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("destination route arrival time not found for %s", routeNbr)
	}
	arrivalTime := data[0].(time.Time)

	// retrieve hub office
	carrier := getAttributeAsString(dest, "carrier")
	fromIata := getAttributeAsString(route, "fromIata")
	var hub tgdb.TGNode
	key := map[string]interface{}{
		"iata":    fromIata,
		"carrier": carrier,
	}
	hub, err = graph.GetNodeByKey("Office", key)
	if err != nil {
		return time.Time{}, fmt.Errorf("hub office node is not found for %s %s", carrier, fromIata)
	}

	if departTime.Before(hubTime) {
		// last route time is old, so create destination route depart for a new day
		if departTime, err = createEdgeDeparts(graph, route, hub, hubTime); err != nil {
			return time.Time{}, err
		}
	}

	if arrivalTime.Before(departTime) {
		// last route time is old, so create destination route arrival for a new day
		if arrivalTime, err = createEdgeArrives(graph, route, dest, departTime); err != nil {
			return time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")

	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().values('uid');", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().outE('contains').inV().has('Container','monitor','%s').values('uid');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return arrivalTime, fmt.Errorf("no container found for %s and %s", routeNbr, product)
	}
	uid := data[0].(string)
	cons, err := graph.GetNodeByKey("Container", map[string]interface{}{"uid": uid})
	if err != nil {
		return arrivalTime, err
	}

	// add simulated temperature measurement
	if handling == "P" && IsMonitored(product) {
		createMonitorMeasurements(graph, cons,
			getAttributeAsString(route, "schdDepartTime"),
			getAttributeAsString(route, "schdArrivalTime"),
			getAttributeAsString(hub, "gmtOffset"),
			getAttributeAsString(dest, "gmtOffset"))
	}

	// add package to the parent container
	err = createEdgeContains(graph, cons, pkg, departTime.Unix(), arrivalTime.Unix(), "P")
	return arrivalTime, err
}

// update local truck delivery and return the package delivery time
func localDelivery(graph *GraphManager, arrivalTime time.Time, deliveryDelay float64, dest, pkg tgdb.TGNode) (time.Time, error) {

	// get the local route
	iata := getAttributeAsString(dest, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','G').values('routeNbr');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no local route found at %s", iata)
	}
	routeNbr := data[0].(string)
	route, err := graph.GetNodeByKey("Route", map[string]interface{}{"routeNbr": routeNbr})
	if err != nil {
		return time.Time{}, err
	}

	// get last depart time of the local route
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('departs').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("delivery route depart time not found for %s", routeNbr)
	}
	departTime := data[0].(time.Time)

	if departTime.Before(arrivalTime) {
		// last route time is old, so create new delivery route depart and arrival for a new day
		departTime, err = createEdgeDeparts(graph, route, dest, arrivalTime)
		if err != nil {
			return time.Time{}, err
		}
		if _, err = createEdgeArrives(graph, route, dest, departTime); err != nil {
			return time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().values('uid');", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().has('Container','monitor','%s').values('uid');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return arrivalTime, fmt.Errorf("no container found for %s and %s", routeNbr, product)
	}
	uid := data[0].(string)
	cons, err := graph.GetNodeByKey("Container", map[string]interface{}{"uid": uid})
	if err != nil {
		return arrivalTime, err
	}

	// add simulated temperature measurement
	if handling == "P" && IsMonitored(product) {
		gmtOffset := getAttributeAsString(dest, "gmtOffset")
		createMonitorMeasurements(graph, cons,
			getAttributeAsString(route, "schdDepartTime"),
			getAttributeAsString(route, "schdArrivalTime"),
			gmtOffset, gmtOffset)
	}

	// add package to the parent container
	deliveryTime := departTime.Add(time.Minute * time.Duration(int(deliveryDelay*60)))
	err = createEdgeContains(graph, cons, pkg, departTime.Unix(), deliveryTime.Unix(), "P")
	return deliveryTime, err
}

// generate monitoring events if a container is monitored by a specified threshold
func createMonitorMeasurements(graph *GraphManager, cons tgdb.TGNode, schdDepart, schdArrival, departGmtOffset, arrivalGmtOffset string) error {
	monitor := getAttributeAsString(cons, "monitor")
	if len(monitor) == 0 {
		// ignore if container is not monitored
		return nil
	}
	threshold, err := graph.GetNodeByKey("Threshold", map[string]interface{}{"name": monitor})
	if err != nil || threshold == nil {
		// ignore if no threshold is found
		return nil
	}
	minValue := getAttributeAsDouble(threshold, "minValue")
	maxValue := getAttributeAsDouble(threshold, "maxValue")

	// monitor the periods of the next 3 days
	for d := 0; d < 3; d++ {
		monitorStart, monitorEnd := measurementPeriod(schdDepart, schdArrival, departGmtOffset, arrivalGmtOffset, d)

		if containerIsMonitored(getAttributeAsString(cons, "uid"), monitorEnd) {
			// skip if the container measurement already exist in TGDB
			continue
		}
		measures := randomThresholdViolation(monitorStart, monitorEnd, minValue, maxValue, FabricConfig.ViolationRate)
		for _, m := range measures {
			// create edge measures from cons to threshold
			err := createEdgeMeasures(graph, cons, threshold, m)
			if err != nil {
				fmt.Println("failed to create measurement", err)
			}
		}
	}

	return nil
}

func containerIsMonitored(consUID string, monitorEnd time.Time) bool {
	// query last monitor end time
	query := fmt.Sprintf("gremlin://g.V().has('Container','uid','%s').outE('measures').order().by('eventTimestamp',desc).limit(1).values('eventTimestamp');", consUID)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return false
	}
	lastMonitorTime := data[0].(time.Time)

	// consider monitored if monitor date is ealier than last monitored date
	if lastMonitorTime.YearDay() > monitorEnd.YearDay() {
		return true
	}
	// consider them equivalent if time difference is within 1 hour
	if math.Abs(monitorEnd.Sub(lastMonitorTime).Hours()) < 1 {
		return true
	}
	return false
}

// For a specified package UID, return map of containerUID -> violationMeasurement,
// assuming that at most one violation period for each embedding container
func queryThresholdViolation(graph *GraphManager, uid string) (map[string]*Measurement, error) {
	// query time periods when package is on route
	periods, err := queryOnRoutePeriods(graph, uid)
	if err != nil {
		return nil, err
	}
	// query monitored container
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').inE('contains').outV();", uid)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	result := make(map[string]*Measurement)
	for i, node := range data {
		cons := node.(tgdb.TGNode)
		consUID := getAttributeAsString(cons, "uid")
		m, err := queryContainerViolation(graph, consUID, periods[i].PeriodStart, periods[i].PeriodEnd)
		if err == nil && m != nil {
			result[consUID] = m
		}
	}
	return result, nil
}

type timePeriod struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func queryOnRoutePeriods(graph *GraphManager, uid string) ([]*timePeriod, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').inE('contains');", uid)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}
	var result []*timePeriod
	for _, edge := range data {
		contains := edge.(tgdb.TGEdge)
		period := &timePeriod{
			PeriodStart: contains.GetAttribute("eventTimestamp").GetValue().(time.Time),
			PeriodEnd:   contains.GetAttribute("outTimestamp").GetValue().(time.Time),
		}
		result = append(result, period)
	}
	return result, nil
}

// return threshold violation period of a container within the specified time range
func queryContainerViolation(graph *GraphManager, consUID string, periodStart, periodEnd time.Time) (*Measurement, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Container','uid','%s').outE('measures').has('violated',1);", consUID)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	for _, edge := range data {
		measures := edge.(tgdb.TGEdge)
		violationStart := measures.GetAttribute("startTimestamp").GetValue().(time.Time)
		violationEnd := measures.GetAttribute("eventTimestamp").GetValue().(time.Time)
		if violationStart.Before(periodStart) {
			violationStart = periodStart
		}
		if periodEnd.Before(violationEnd) {
			violationEnd = periodEnd
		}
		if violationStart.Before(violationEnd) {
			// found a violation measurement
			return &Measurement{
				PeriodStart: violationStart,
				PeriodEnd:   violationEnd,
				MinValue:    getAttributeAsDouble(measures, "minValue"),
				MaxValue:    getAttributeAsDouble(measures, "maxValue"),
				InViolation: true,
			}, nil
		}
	}
	return nil, nil
}

// return measurements of a container within the specified time range
func queryContainerMeasurements(graph *GraphManager, consUID string, periodStart, periodEnd time.Time) (bool, []*monitorData, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Container','uid','%s').outE('measures').order().by('eventTimestamp');", consUID)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return false, nil, err
	}

	var result []*monitorData
	violated := false
	for _, edge := range data {
		measures := edge.(tgdb.TGEdge)
		measureStart := measures.GetAttribute("startTimestamp").GetValue().(time.Time)
		measureEnd := measures.GetAttribute("eventTimestamp").GetValue().(time.Time)
		if measureStart.After(periodEnd) || measureEnd.Before(periodStart) {
			continue
		}
		if measureStart.Before(periodStart) {
			measureStart = periodStart
		}
		if periodEnd.Before(measureEnd) {
			measureEnd = periodEnd
		}
		if measureStart.Before(measureEnd) {
			// collect the measurement
			utc := time.FixedZone("UTC", 0)
			m := &monitorData{
				PeriodStart: measureStart.In(utc).Format(time.RFC3339),
				PeriodEnd:   measureEnd.In(utc).Format(time.RFC3339),
				MinValue:    getAttributeAsDouble(measures, "minValue"),
				MaxValue:    getAttributeAsDouble(measures, "maxValue"),
				InViolation: getAttributeAsBool(measures, "violated"),
			}
			if m.InViolation {
				violated = true
			}
			result = append(result, m)
		}
	}
	return violated, result, nil
}

type packageTransit struct {
	UID      string          `json:"uid"`
	Timeline []*transitEvent `json:"timeline"`
	Routes   []*routeDetail  `json:"routes"`
}

type transitEvent struct {
	EventTimestamp string  `json:"eventTime"`
	EventType      string  `json:"eventType"`
	Location       string  `json:"location"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	RouteRef       string  `json:"route,omitempty"`
}

type routeDetail struct {
	RouteNbr      string         `json:"routeNbr"`
	RouteType     string         `json:"-"`
	DepartureTime string         `json:"departureTime"`
	From          string         `json:"from"`
	FromLatitude  float64        `json:"-"`
	FromLongitude float64        `json:"-"`
	ArrivalTime   string         `json:"arrivalTime"`
	To            string         `json:"to"`
	ToLatitude    float64        `json:"-"`
	ToLongitude   float64        `json:"-"`
	ContainerPath string         `json:"containers"`
	Violated      bool           `json:"violated"`
	Measurements  []*monitorData `json:"measurements"`
}

type monitorData struct {
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
	MinValue    float64 `json:"minValue"`
	MaxValue    float64 `json:"maxValue"`
	InViolation bool    `json:"violated"`
}

// return package transit timeline
func queryPackageTransit(graph *GraphManager, uid string) (*packageTransit, error) {
	relatedNodes, err := queryRelatedNodes(graph, uid)
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').inE().order().by('eventTimestamp');", uid)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	var timeline []*transitEvent
	var routes []*routeDetail
	for _, edge := range data {
		event := edge.(tgdb.TGEdge)
		switch event.GetEntityType().GetName() {
		case "pickup":
			// do nothing, will be added by contains
		case "contains":
			eventTime := getAttributeAsUTCTime(event, "eventTimestamp")
			key := fmt.Sprintf("contains-%s", eventTime)
			cons := relatedNodes[key]
			periodStart := event.GetAttribute("eventTimestamp").GetValue().(time.Time)
			periodEnd := event.GetAttribute("outTimestamp").GetValue().(time.Time)
			if rd, err := queryRouteDetail(graph, cons, periodStart, periodEnd); err == nil {
				routes = append(routes, rd)
				if rd.RouteType == "G" && eventTime > rd.DepartureTime {
					// add pickup
					pickup := &transitEvent{
						EventTimestamp: eventTime,
						EventType:      "pickup",
						RouteRef:       rd.RouteNbr,
					}
					if addr, err := queryAddress(graph, uid, "sender"); err == nil {
						pickup.Latitude = addr.Latitude
						pickup.Longitude = addr.Longitude
						pickup.Location = fmt.Sprintf("%s, %s, %s", addr.Street, addr.City, addr.StateProvince)
					}
					timeline = append(timeline, pickup)
				} else {
					timeline = append(timeline, &transitEvent{
						EventTimestamp: eventTime,
						EventType:      "depart",
						Location:       rd.From,
						Latitude:       rd.FromLatitude,
						Longitude:      rd.FromLongitude,
						RouteRef:       rd.RouteNbr,
					})
				}
				outTime := getAttributeAsUTCTime(event, "outTimestamp")
				if rd.RouteType == "G" && outTime < rd.ArrivalTime {
					// add delivery
					delivery := &transitEvent{
						EventTimestamp: outTime,
						EventType:      "deliver",
						RouteRef:       rd.RouteNbr,
					}
					if addr, err := queryAddress(graph, uid, "recipient"); err == nil {
						delivery.Latitude = addr.Latitude
						delivery.Longitude = addr.Longitude
						delivery.Location = fmt.Sprintf("%s, %s, %s", addr.Street, addr.City, addr.StateProvince)
					}
					timeline = append(timeline, delivery)
				} else {
					timeline = append(timeline, &transitEvent{
						EventTimestamp: outTime,
						EventType:      "arrive",
						Location:       rd.To,
						Latitude:       rd.ToLatitude,
						Longitude:      rd.ToLongitude,
						RouteRef:       rd.RouteNbr,
					})
				}
			}
		case "transfers":
			eventTime := getAttributeAsUTCTime(event, "eventTimestamp")
			key := fmt.Sprintf("transfers-%s", eventTime)
			office := relatedNodes[key]
			eventType := "transfer"
			if getAttributeAsString(event, "direction") == "to" {
				eventType = "transferAck"
			}
			loc := fmt.Sprintf("%s: %s, %s", getAttributeAsString(office, "carrier"), getAttributeAsString(office, "iata"), getAttributeAsString(office, "description"))
			timeline = append(timeline, &transitEvent{
				EventTimestamp: eventTime,
				EventType:      eventType,
				Location:       loc,
				Latitude:       getAttributeAsDouble(event, "latitude"),
				Longitude:      getAttributeAsDouble(event, "longitude"),
			})
		case "delivery":
			// do nothing, added by contains
		default:
			fmt.Println("ignore package relationship", event.GetEntityType().GetName())
		}
	}
	return &packageTransit{
		UID:      uid,
		Timeline: timeline,
		Routes:   routes,
	}, nil
}

func queryRelatedNodes(graph *GraphManager, uid string) (map[string]tgdb.TGNode, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').inE().outV().simplePath().path();", uid)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}
	result := make(map[string]tgdb.TGNode)
	for _, path := range data {
		entities, ok := path.([]interface{})
		if !ok || len(entities) < 3 {
			return nil, errors.New("query did not return path with 3 entities")
		}
		edge := entities[1].(tgdb.TGEdge)
		node := entities[2].(tgdb.TGNode)
		key := fmt.Sprintf("%s-%s", edge.GetEntityType().GetName(), getAttributeAsUTCTime(edge, "eventTimestamp"))
		result[key] = node
	}
	return result, nil
}

// retrieve details of a route corresponding to a package's parent container at a specified on-route start and end time
func queryRouteDetail(graph *GraphManager, cons tgdb.TGNode, periodStart, periodEnd time.Time) (*routeDetail, error) {
	result := &routeDetail{}
	// get measurements of the base container if type is 'F'
	if getAttributeAsString(cons, "type") == "F" {
		if violated, measurements, err := queryContainerMeasurements(graph, getAttributeAsString(cons, "uid"), periodStart, periodEnd); err == nil {
			result.Violated = violated
			result.Measurements = measurements
		}
	}

	// navigate to root vessel container
	path := getAttributeAsString(cons, "uid")
	var err error
	for getAttributeAsString(cons, "type") != "V" {
		cons, path, err = queryParentContainer(graph, cons, path)
		if err != nil || cons == nil {
			fmt.Println("failed to query parent", cons, err)
			return nil, err
		}
	}
	result.ContainerPath = path

	// get assigned route
	query := fmt.Sprintf("gremlin://g.V().has('Container','uid','%s').outE('assigned').inV();", getAttributeAsString(cons, "uid"))
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		fmt.Println("failed to query route", query, err)
		return nil, err
	}
	route := data[0].(tgdb.TGNode)
	result.RouteType = getAttributeAsString(route, "type")

	// get departure event detailss
	departEvt, err := queryRouteEvent(graph, route, "departs", result.RouteType, periodStart)
	if err != nil {
		if len(data) > 1 {
			// evaluate next route to find a match on depart time
			route = data[1].(tgdb.TGNode)
			result.RouteType = getAttributeAsString(route, "type")
			departEvt, err = queryRouteEvent(graph, route, "departs", result.RouteType, periodStart)
		}
		if err != nil {
			fmt.Println("failed to query route departure", route, err)
			return nil, err
		}
	}
	result.RouteNbr = departEvt.RouteNbr
	result.DepartureTime = departEvt.EventTime
	result.From = departEvt.Location
	result.FromLatitude = departEvt.Latitude
	result.FromLongitude = departEvt.Longitude

	// get arrival event details
	arriveEvt, err := queryRouteEvent(graph, route, "arrives", result.RouteType, periodEnd)
	if err != nil {
		fmt.Println("failed to query route arrival", route, err)
		return nil, err
	}
	result.ArrivalTime = arriveEvt.EventTime
	result.To = arriveEvt.Location
	result.ToLatitude = arriveEvt.Latitude
	result.ToLongitude = arriveEvt.Longitude

	return result, nil
}

type routeEvent struct {
	RouteNbr  string
	EventTime string
	Location  string
	Latitude  float64
	Longitude float64
}

// route event info for eventType 'departs' or 'arrives'
func queryRouteEvent(graph *GraphManager, route tgdb.TGNode, eventType, routeType string, refTime time.Time) (*routeEvent, error) {
	query := fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('%s').inV().simplePath().path();", getAttributeAsString(route, "routeNbr"), eventType)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	// pick route with the following conditions
	//   1. routeType='A': eventTime is within 30 minute of the refTime
	//   2. routeType='G' and eventType='arrives': eventTime > refTime and within 8 hours (i.e, deliveryTime before route end)
	//   3. routeType='G' and eventType='departs': eventTime < refTime and within 8 hours (i.e., pickupTime after route start)
	for _, path := range data {
		entities, ok := path.([]interface{})
		if !ok || len(entities) < 3 {
			return nil, errors.New("query did not return path with 3 entities")
		}
		edge := entities[1].(tgdb.TGEdge)
		node := entities[2].(tgdb.TGNode)
		eventTime := edge.GetAttribute("eventTimestamp").GetValue().(time.Time)
		var related bool
		if routeType == "A" {
			related = math.Abs(refTime.Sub(eventTime).Minutes()) < 30
		} else {
			if eventType == "departs" {
				related = eventTime.Before(refTime.Add(time.Duration(30)*time.Minute)) && math.Abs(refTime.Sub(eventTime).Hours()) <= 8
			} else {
				related = refTime.Before(eventTime.Add(time.Duration(30)*time.Minute)) && math.Abs(eventTime.Sub(refTime).Hours()) <= 8
			}
		}
		if related {
			loc := fmt.Sprintf("%s: %s, %s", getAttributeAsString(node, "carrier"), getAttributeAsString(node, "iata"), getAttributeAsString(node, "description"))
			return &routeEvent{
				RouteNbr:  getAttributeAsString(route, "routeNbr"),
				EventTime: getAttributeAsUTCTime(edge, "eventTimestamp"),
				Location:  loc,
				Latitude:  getAttributeAsDouble(node, "latitude"),
				Longitude: getAttributeAsDouble(node, "longitude"),
			}, nil
		}
	}
	return nil, fmt.Errorf("faild to retrieve route event %s", eventType)
}

func queryParentContainer(graph *GraphManager, cons tgdb.TGNode, path string) (tgdb.TGNode, string, error) {
	uid := getAttributeAsString(cons, "uid")
	query := fmt.Sprintf("gremlin://g.V().has('Container','uid','%s').inE('contains').outV();", uid)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		fmt.Println("queryParentContainer", uid, query, err)
		return nil, path, err
	}

	node := data[0].(tgdb.TGNode)
	return node, fmt.Sprintf("%s.%s", getAttributeAsString(node, "uid"), path), nil
}
