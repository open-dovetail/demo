/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yxuco/tgdb"
)

func TestInitializeGraph(t *testing.T) {
	fmt.Println("TestInitializeGraph")

	err := Initialize(configFile)
	assert.NoError(t, err, "initialize config should not throw error")
	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	query := fmt.Sprintf("gremlin://g.V().has('Carrier', 'name', '%s');", "SLS")
	result, err := graph.Query(query)
	assert.NoError(t, err, "Gremlin query should not return error")
	if len(result) == 0 {
		// initalize graph only if carriers have not been created yet
		err = InitializeGraph(graph)
		assert.NoError(t, err, "initialize GraphDB should not throw error")
	} else {
		node, ok := result[0].(tgdb.TGNode)
		assert.True(t, ok, "query result should be a TGNode")
		attrs, err := node.GetAttributes()
		assert.NoError(t, err, "get attributes should not throw error")
		fmt.Println("carrier SLS already exists")
		for _, v := range attrs {
			fmt.Printf("\t%s => %v\n", v.GetName(), v.GetValue())
		}
	}
}

func TestPrintShippingLabel(t *testing.T) {
	fmt.Println("TestPrintShippingLabel")

	// make sure graph DB is initialized
	err := Initialize(configFile)
	assert.NoError(t, err, "initialize config should not throw error")
	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	query := fmt.Sprintf("gremlin://g.V().has('Carrier', 'name', '%s');", "SLS")
	result, err := graph.Query(query)
	assert.NoError(t, err, "Gremlin query should not return error")
	if len(result) == 0 {
		// initalize graph only if carriers have not been created yet
		err = InitializeGraph(graph)
		assert.NoError(t, err, "initialize GraphDB should not throw error")
	}

	// parse sample request
	sample, err := ioutil.ReadFile("./package.json")
	assert.NoError(t, err, "read sample packcage requet should not throw error")

	_, err = PrintShippingLabel(string(sample))
	assert.NoError(t, err, "print shipping label should not throw error")

	// verify package
	result, err = graph.Query("gremlin://g.V().has('Package','product','PfizerVaccine').values('uid');")
	assert.NoError(t, err, "package query should not throw error")
	assert.Greater(t, len(result), 0, "one or more packages should exist in TGDB")

	// verify package out node count
	query = fmt.Sprintf("gremlin://g.V().has('Package', 'uid', '%s').out();", result[0].(string))
	result, err = graph.Query(query)
	assert.NoError(t, err, "package out-nodes query should not throw error")
	assert.Equal(t, 3, len(result), "package should have 3 out nodes")
}

func TestQueryPackage(t *testing.T) {
	fmt.Println("TestQueryPackage")

	// connect to TGDB
	err := Initialize(configFile)
	assert.NoError(t, err, "initialize config should not throw error")
	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	result, err := graph.Query("gremlin://g.V().has('Package','product','PfizerVaccine').values('estPickupTime');")
	assert.NoError(t, err, "package timestamp query should not throw error")
	for _, r := range result {
		fmt.Printf("%T: %v\n", r, r)
	}

	result, err = graph.Query("gremlin://g.V().has('Package','product','PfizerVaccine').values('uid');")
	assert.NoError(t, err, "package query should not throw error")
	assert.Less(t, 0, len(result), "one or more packages should exist in TGDB")
	fmt.Printf("uid %T, %v\n", result[0], result[0])
	query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').out();", result[0].(string))
	result, err = graph.Query(query)
	assert.NoError(t, err, "package out-node query should not throw error")
	assert.Equal(t, 3, len(result), "package should have 3 out nodes")

	// assert.Fail(t, "test")
}
