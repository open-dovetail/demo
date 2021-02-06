/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import (
	"fmt"
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
