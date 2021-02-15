/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"fmt"
	"strconv"
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
	fmt.Println("create content", cont.UID)
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
	fmt.Println("inserted content", node.GetAttribute("uid").GetValue())
	return node, nil
}

func createAddress(graph *GraphManager, addr *Address) (tgdb.TGNode, error) {
	fmt.Println("create address", addr.UID)
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
	fmt.Println("inserted address", node.GetAttribute("uid").GetValue())
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
	fmt.Println("committed operates", err)
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
	fmt.Println("committed schedules", err)
	return err
}

func createEdgeDeparts(graph *GraphManager, route, office tgdb.TGNode, after time.Time) (time.Time, error) {
	fmt.Println("create departs", getAttributeAsString(route, "routeNbr"), getAttributeAsString(office, "iata"))
	// calculate random depart time according to route schedule
	tm := randomTimestamp(getAttributeAsString(route, "schdDepartTime"), getAttributeAsString(office, "gmtOffset"), 5)
	departTime := time.Unix(tm, 0)
	if departTime.Before(after) {
		departTime = correctTimeByDays(departTime, after)
		tm = departTime.Unix()
	}
	fmt.Println("depart time", tm, departTime)

	departs, err := graph.CreateEdge("departs", route, office)
	if err != nil {
		return time.Time{}, err
	}
	departs.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(departs); err != nil {
		return departTime, err
	}

	_, err = graph.Commit()
	fmt.Println("committed departs", err)
	return departTime, err
}

func createEdgeArrives(graph *GraphManager, route, office tgdb.TGNode, after time.Time) (time.Time, error) {
	fmt.Println("create arrives", getAttributeAsString(route, "routeNbr"), getAttributeAsString(office, "iata"))
	// calculate random depart time according to route schedule
	tm := randomTimestamp(getAttributeAsString(route, "schdArrivalTime"), getAttributeAsString(office, "gmtOffset"), 5)
	arrivalTime := time.Unix(tm, 0)
	if arrivalTime.Before(after) {
		arrivalTime = correctTimeByDays(arrivalTime, after)
		tm = arrivalTime.Unix()
	}
	fmt.Println("arrival time", tm, arrivalTime)

	arrives, err := graph.CreateEdge("arrives", route, office)
	if err != nil {
		return time.Time{}, err
	}
	arrives.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(arrives); err != nil {
		return arrivalTime, err
	}

	_, err = graph.Commit()
	fmt.Println("committed arrives", err)
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
	fmt.Println("committed builds", err)
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
	fmt.Println("committed assigned", err)
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
	fmt.Println("committed contains", err)
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
	fmt.Println("committed sender", err)
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
	fmt.Println("committed recipient", err)
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
	fmt.Println("committed contains content", err)
	return err
}

func createEdgePickup(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, tracking string, lat, lon float64) error {
	pickup, err := graph.CreateEdge("pickup", office, pkg)
	if err != nil {
		return err
	}

	pickup.SetOrCreateAttribute("eventTimestamp", eventTime)
	pickup.SetOrCreateAttribute("trackingID", tracking)
	pickup.SetOrCreateAttribute("employeeID", strconv.FormatInt(time.Now().Unix(), 10))
	pickup.SetOrCreateAttribute("longitude", lon)
	pickup.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(pickup); err != nil {
		return err
	}

	_, err = graph.Commit()
	fmt.Println("committed pickup", err)
	return err
}

func createEdgeDelivery(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, lat, lon float64) error {
	delivery, err := graph.CreateEdge("delivery", office, pkg)
	if err != nil {
		return err
	}

	delivery.SetOrCreateAttribute("eventTimestamp", eventTime)
	delivery.SetOrCreateAttribute("employeeID", strconv.FormatInt(time.Now().Unix(), 10))
	delivery.SetOrCreateAttribute("longitude", lon)
	delivery.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(delivery); err != nil {
		return err
	}

	_, err = graph.Commit()
	fmt.Println("committed delivery", err)
	return err
}

func createEdgeTransfers(graph *GraphManager, office, pkg tgdb.TGNode, eventTime int64, tracking string, lat, lon float64, direction string) error {
	transfers, err := graph.CreateEdge("transfers", office, pkg)
	if err != nil {
		return err
	}

	transfers.SetOrCreateAttribute("eventTimestamp", eventTime)
	transfers.SetOrCreateAttribute("direction", direction)
	transfers.SetOrCreateAttribute("trackingID", tracking)
	transfers.SetOrCreateAttribute("employeeID", strconv.FormatInt(time.Now().Unix(), 10))
	transfers.SetOrCreateAttribute("longitude", lon)
	transfers.SetOrCreateAttribute("latitude", lat)
	if err := graph.InsertEntity(transfers); err != nil {
		return err
	}

	_, err = graph.Commit()
	fmt.Println("committed transfers", err)
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

func getAttributeAsString(node tgdb.TGNode, name string) string {
	attr := node.GetAttribute(name)
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

func getAttributeAsDouble(node tgdb.TGNode, name string) float64 {
	attr := node.GetAttribute(name)
	var result interface{}
	if attr != nil {
		result = attr.GetValue()
	}
	switch v := result.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
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
	pickupTime := estimatePUDTime(office.GMTOffset, pickupDelay)
	err = createEdgePickup(graph, origin, node, pickupTime.Unix(), pkg.UID, pkg.From.Latitude, pkg.From.Longitude)
	if err != nil {
		return time.Time{}, err
	}
	arrivalTime, err := localPickup(graph, pickupTime, origin, node)
	if err != nil {
		return arrivalTime, err
	}
	return originRoute(graph, arrivalTime, origin, node)
}

// update local truck pickup and return the time for truck to arrive at the origin office
func localPickup(graph *GraphManager, pickupTime time.Time, origin, pkg tgdb.TGNode) (time.Time, error) {

	// get the local route
	iata := getAttributeAsString(origin, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','G');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no local route found at %s", iata)
	}
	route := data[0].(tgdb.TGNode)
	routeNbr := getAttributeAsString(route, "routeNbr")

	// get last arrival time of the local route
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('arrives').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("pickup route arrival time not found for %s", routeNbr)
	}
	arrivalTime := data[0].(time.Time)

	if arrivalTime.Before(pickupTime) {
		// last route time is old, so create new pickup route depart and arrival for a new day
		departTime, err := createEdgeDeparts(graph, route, origin, time.Now())
		if err != nil {
			return time.Time{}, err
		}
		if arrivalTime, err = createEdgeArrives(graph, route, origin, departTime); err != nil {
			return time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV();", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().has('Container','monitor','%s');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return arrivalTime, fmt.Errorf("no container found for %s", routeNbr)
	}
	cons := data[0].(tgdb.TGNode)

	// add package to the parent container
	err = createEdgeContains(graph, cons, pkg, pickupTime.Unix(), arrivalTime.Unix(), "P")
	return arrivalTime, err
}

// update origin route to hub and return the time for plane to arrive at the hub
func originRoute(graph *GraphManager, arrivalTime time.Time, origin, pkg tgdb.TGNode) (time.Time, error) {

	// get the origin route
	iata := getAttributeAsString(origin, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','A');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no origin route found at %s", iata)
	}
	route := data[0].(tgdb.TGNode)
	routeNbr := getAttributeAsString(route, "routeNbr")

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

	if hubTime.Before(departTime) {
		// last route time is old, so create origin route arrival for a new day
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
		if hubTime, err = createEdgeArrives(graph, route, hub, departTime); err != nil {
			return time.Time{}, err
		}
	}

	// find container to add package
	handling := getAttributeAsString(pkg, "handlingCd")
	product := getAttributeAsString(pkg, "product")

	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV();", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().outE('contains').inV().has('Container','monitor','%s');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return hubTime, fmt.Errorf("no container found for %s", routeNbr)
	}
	cons := data[0].(tgdb.TGNode)

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
	err = createEdgeTransfers(graph, origin, node, hubTime.Unix(), pkg.UID, originHub.Latitude, originHub.Longitude, "from")
	if err != nil {
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
	return createEdgeTransfers(graph, dest, node, hubTime.Unix(), pkg.UID, destHub.Latitude, destHub.Longitude, "to")
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
	deliveryTime := estimatePUDTime(office.GMTOffset, deliveryDelay)
	if deliveryTime.Before(arrivalTime) {
		// correct delivery time by adding days
		deliveryTime = correctTimeByDays(deliveryTime, arrivalTime)
	}
	err = createEdgeDelivery(graph, dest, node, deliveryTime.Unix(), pkg.To.Latitude, pkg.To.Longitude)
	if err != nil {
		return time.Time{}, err
	}
	return localDelivery(graph, arrivalTime, deliveryTime, dest, node)
}

// update delivery route from hub and return the time for plane to arrive at the dest office
func deliveryRoute(graph *GraphManager, hubTime time.Time, dest, pkg tgdb.TGNode) (time.Time, error) {

	// get the destination route
	iata := getAttributeAsString(dest, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','toIata','%s').has('Route','type','A');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no destination route found at %s", iata)
	}
	route := data[0].(tgdb.TGNode)
	routeNbr := getAttributeAsString(route, "routeNbr")

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

	if departTime.Before(hubTime) {
		// last route time is old, so create destination route depart for a new day
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

	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV();", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().outE('contains').inV().has('Container','monitor','%s');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return arrivalTime, fmt.Errorf("no container found for %s", routeNbr)
	}
	cons := data[0].(tgdb.TGNode)

	// add package to the parent container
	err = createEdgeContains(graph, cons, pkg, departTime.Unix(), arrivalTime.Unix(), "P")
	return arrivalTime, err
}

// update local truck delivery and return the package delivery time
func localDelivery(graph *GraphManager, arrivalTime, deliveryTime time.Time, dest, pkg tgdb.TGNode) (time.Time, error) {

	// get the local route
	iata := getAttributeAsString(dest, "iata")
	query := fmt.Sprintf("gremlin://g.V().has('Route','fromIata','%s').has('Route','type','G');", iata)
	data, err := graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("no local route found at %s", iata)
	}
	route := data[0].(tgdb.TGNode)
	routeNbr := getAttributeAsString(route, "routeNbr")

	// get last depart time of the local route
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').outE('departs').order().by('eventTimestamp', desc).values('eventTimestamp').limit(1);", routeNbr)
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return time.Time{}, fmt.Errorf("delivery route depart time not found for %s", routeNbr)
	}
	departTime := data[0].(time.Time)

	if departTime.Before(arrivalTime) {
		// last route time is old, so create new delivery route depart and arrival for a new day
		departTime, err := createEdgeDeparts(graph, route, dest, arrivalTime)
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
	query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV();", routeNbr)
	if handling == "P" && IsMonitored(product) {
		query = fmt.Sprintf("gremlin://g.V().has('Route','routeNbr','%s').inE('assigned').outV().outE('contains').inV().has('Container','monitor','%s');", routeNbr, product)
	}
	data, err = graph.Query(query)
	if err != nil || len(data) == 0 {
		return arrivalTime, fmt.Errorf("no container found for %s", routeNbr)
	}
	cons := data[0].(tgdb.TGNode)

	// add package to the parent container
	err = createEdgeContains(graph, cons, pkg, departTime.Unix(), deliveryTime.Unix(), "P")
	return deliveryTime, err
}

func edgeQuery(conn tgdb.TGConnection) {
	memberName := "Napoleon Bonaparte"
	fmt.Printf("\n*** edgeQuery %s\n", memberName)
	query := fmt.Sprintf("gremlin://g.V().has('houseMemberType', 'memberName', '%s').bothE();", memberName)
	rset, err := conn.ExecuteQuery(query, nil)
	if err != nil {
		fmt.Printf("query error: %v\n", err)
	}
	for rset.HasNext() {
		edge := rset.Next().(tgdb.TGEdge)
		fmt.Println("Edge")
		attrs, _ := edge.GetAttributes()
		for _, a := range attrs {
			fmt.Printf("\tattribute %s -> %v\n", a.GetName(), a.GetValue())
		}
		n := edge.GetVertices()
		for i, v := range n {
			fmt.Printf("\tnode %d: %d\n", i, v.GetVirtualId())
		}
	}
}
