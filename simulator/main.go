/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package simulator

import "fmt"

func main() {
	fmt.Println("hello")
	readConfig("./config.json")
	fmt.Printf("read config %v\n", Carriers)
}
