//    Copyright 2018 Tharanga Nilupul Thennakoon
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"os"
	"quebic-faas/common"
	"quebic-faas/types"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	setupMgrCompCmds()
	setupMgrCompFlags()
}

var mgrCompCmd = &cobra.Command{
	Use:   "components",
	Short: "Components commonds",
	Long:  `Components commonds`,
}

func setupMgrCompCmds() {

	mgrCompCmd.AddCommand(mgrCompGetAllCmd)

}

func setupMgrCompFlags() {
}

var mgrCompGetAllCmd = &cobra.Command{
	Use:   "ls",
	Short: "components : get-all",
	Long:  `components : get-all`,
	Run: func(cmd *cobra.Command, args []string) {
		mgrCompGetALL(cmd, args)
	},
}

func mgrCompGetALL(cmd *cobra.Command, args []string) {

	mgrService := appContainer.GetMgrService()
	comps, err := mgrService.ManagerComponentGetALL()
	if err != nil {
		prepareErrorResponse(cmd, err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Host", "Post"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(prepareMgrCompTable(comps))
	table.Render()

}

func prepareMgrCompTable(data []types.ManagerComponent) [][]string {

	var rows [][]string

	for _, val := range data {

		id := val.ID
		host := val.Deployment.Host
		port := val.Deployment.Port

		rows = append(rows, []string{id, host, common.IntToStr(port)})

	}

	return rows

}
