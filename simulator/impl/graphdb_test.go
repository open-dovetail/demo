/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package impl

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yxuco/tgdb"
)

func setupDemoGraph() error {
	graph, err := GetTGConnection()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("gremlin://g.V().has('Carrier', 'name', '%s');", "SLS")
	result, err := graph.Query(query)
	if err != nil {
		return err
	}
	if len(result) == 0 {
		// initalize graph only if carriers have not been created yet
		return InitializeGraph(graph)
	}
	return nil
}

func TestInitializeGraph(t *testing.T) {
	fmt.Println("TestInitializeGraph")

	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	query := fmt.Sprintf("gremlin://g.V().has('Carrier', 'name', '%s');", "SLS")
	result, err := graph.Query(query)
	assert.NoError(t, err, "Gremlin query should not return error")
	assert.Equal(t, 1, len(result), "carrier query should return 1 node")

	node, ok := result[0].(tgdb.TGNode)
	assert.True(t, ok, "query result should be a TGNode")
	attrs, err := node.GetAttributes()
	assert.NoError(t, err, "get attributes should not throw error")
	for _, v := range attrs {
		name := v.GetName()
		value := v.GetValue().(string)
		if name == "name" {
			assert.Equal(t, "SLS", value, "carrier name should be 'SLS'")
		} else {
			assert.Equal(t, "South Logistics Services", value, "carrier description should be 'South Logistics Services'")
		}
	}
}

func TestPrintShippingLabel(t *testing.T) {
	fmt.Println("TestPrintShippingLabel")

	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	// parse sample request
	sample, err := ioutil.ReadFile("../package.json")
	assert.NoError(t, err, "read sample packcage requet should not throw error")

	_, err = PrintShippingLabel(string(sample))
	assert.NoError(t, err, "print shipping label should not throw error")

	// verify package
	result, err := graph.Query("gremlin://g.V().has('Package','product','PfizerVaccine').values('uid');")
	assert.NoError(t, err, "package query should not throw error")
	assert.Greater(t, len(result), 0, "one or more packages should exist in TGDB")

	// verify package out node count
	query := fmt.Sprintf("gremlin://g.V().has('Package', 'uid', '%s').out();", result[0].(string))
	result, err = graph.Query(query)
	assert.NoError(t, err, "package out-nodes query should not throw error")
	assert.Equal(t, 3, len(result), "package should have 3 out nodes")
}

func TestPickupPackage(t *testing.T) {
	fmt.Println("TestPickupPackage")

	graph, err := GetTGConnection()
	assert.NoError(t, err, "connect to TGDB should not throw error")

	result, err := graph.Query("gremlin://g.V().has('Package','handlingCd','P').values('uid');")
	assert.NoError(t, err, "package uid query should not throw error")

	// simulate pickup/delivery of a newly created package
	for _, attr := range result {
		uid := attr.(string)
		query := fmt.Sprintf("gremlin://g.V().has('Package','uid','%s');", uid)
		result, err = graph.Query(query)
		assert.NoError(t, err, "package query should not throw error")
		assert.Equal(t, 1, len(result), "query should return 1 package")

		// check if it has already been picked up
		query = fmt.Sprintf("gremlin://g.V().has('Package','uid','%s').inE('pickup');", uid)
		result, err = graph.Query(query)
		assert.NoError(t, err, "query pickup event should not throw error")
		if len(result) > 0 {
			// already picked up, so skip it
			continue
		}

		// simulate package pickup
		err = PickupPackage(uid)
		assert.NoError(t, err, "pickup package should not throw exception")
		break
	}
}
