package cli

import "fmt"

var hostsModuleColName = []string{"address", "group", "user", "password", "date"}

func DisplayHostsModuleInfo() {
	for _, colname := range hostsModuleColName {
		fmt.Println(colname)
	}
}
