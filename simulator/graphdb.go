/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"fmt"

	"github.com/yxuco/tgdb"
	"github.com/yxuco/tgdb/factory"
)

// GetTGConnection returns a new connection of Graph DB
func GetTGConnection() (*GraphManager, error) {
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

	return &GraphManager{
		conn: conn,
		gof:  gof,
		gmd:  gmd,
	}, nil
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

// Commit commits the current transaction
func (g *GraphManager) Commit() (tgdb.TGResultSet, tgdb.TGError) {
	return g.conn.Commit()
}

// Disconnect disconnects from TGDB server
func (g *GraphManager) Disconnect() tgdb.TGError {
	return g.conn.Disconnect()
}

func createThreshold(graph *GraphManager, threshold *Threshold) (tgdb.TGNode, tgdb.TGError) {
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

func createCarrier(graph *GraphManager, carrier *Carrier) (tgdb.TGNode, tgdb.TGError) {
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

func createOffice(graph *GraphManager, office *Office) (tgdb.TGNode, tgdb.TGError) {
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

func createRoute(graph *GraphManager, route *Route) (tgdb.TGNode, tgdb.TGError) {
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

func createContainer(graph *GraphManager, cons *Container) (tgdb.TGNode, tgdb.TGError) {
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

func createEdgeDeparts(graph *GraphManager, route, office tgdb.TGNode) error {
	fmt.Println("create departs", route.GetAttribute("routeNbr").GetValue(), office.GetAttribute("iata").GetValue())
	departs, err := graph.CreateEdge("departs", route, office)
	if err != nil {
		return err
	}
	eventTime := route.GetAttribute("schdDepartTime").GetValue().(string)
	gmtOffset := office.GetAttribute("gmtOffset").GetValue().(string)
	tm := randomTimestamp(eventTime, gmtOffset, 5)
	departs.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(departs); err != nil {
		return err
	}

	_, err = graph.Commit()
	fmt.Println("committed departs", err)
	return err
}

func createEdgeArrives(graph *GraphManager, route, office tgdb.TGNode) error {
	arrives, err := graph.CreateEdge("arrives", route, office)
	if err != nil {
		return err
	}
	eventTime := route.GetAttribute("schdArrivalTime").GetValue().(string)
	gmtOffset := office.GetAttribute("gmtOffset").GetValue().(string)
	tm := randomTimestamp(eventTime, gmtOffset, 5)
	arrives.SetOrCreateAttribute("eventTimestamp", tm)
	if err := graph.InsertEntity(arrives); err != nil {
		return err
	}

	_, err = graph.Commit()
	fmt.Println("committed arrives", err)
	return err
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
		if err := createEdgeDeparts(graph, route, from); err != nil {
			return err
		}
		// create arrival for today
		to := officeNodes[office.Carrier+":"+r.To.Iata]
		if err := createEdgeArrives(graph, route, to); err != nil {
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
		if context.hubInTime > 0 {
			if err := createEdgeContains(graph, parent, child, context.hubInTime, context.hubOutTime, "C"); err != nil {
				return err
			}
		}
		if len(c.Embedded) > 0 {
			if err := initializeEmbeddedContainers(graph, child, c.Embedded, context); err != nil {
				return err
			}
		}
	}
	return nil
}

func execQuery(conn tgdb.TGConnection) {
	startYear := 1800
	endYear := 1900
	fmt.Printf("\n*** execQuery born between (%d, %d)\n", startYear, endYear)

	query := fmt.Sprintf("gremlin://g.V().has('houseMemberType', 'yearBorn', between(%d, %d));", startYear, endYear)
	rset, err := conn.ExecuteQuery(query, nil)
	if err != nil {
		fmt.Printf("query error: %v\n", err)
	}
	for rset.HasNext() {
		if member, ok := rset.Next().(tgdb.TGNode); ok {
			fmt.Printf("Found member %v\n", member.GetAttribute("memberName").GetValue())
			if attrs, err := member.GetAttributes(); err == nil {
				for _, v := range attrs {
					fmt.Printf("\tattribute %s => %v\n", v.GetName(), v.GetValue())
				}
			}
		}
	}
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
