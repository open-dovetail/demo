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

// InitializeGraph inserts carrier nodes and edges into TGDB
func InitializeGraph(graph *GraphManager) error {
	for _, c := range Carriers {
		carrier, err := createCarrier(graph, c)
		if err != nil {
			return nil
		}
		for _, v := range c.Offices {
			office, err := createOffice(graph, v)
			if err != nil {
				return nil
			}
			operates, err := graph.CreateEdge("operates", carrier, office)
			if err := graph.InsertEntity(operates); err != nil {
				return nil
			}
			// Note: must commit here. it does not work to commit after the loop
			if _, err := graph.Commit(); err != nil {
				return err
			}
		}
	}
	//	_, err := graph.Commit()
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
