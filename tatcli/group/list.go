package group

import (
	"github.com/ovh/tat"
	"github.com/ovh/tat/tatcli/internal"
	"github.com/spf13/cobra"
)

var (
	criteria tat.GroupCriteria
)

func init() {
	cmdGroupList.Flags().StringVarP(&criteria.IDGroup, "idGroup", "", "", "Search by id of group")
	cmdGroupList.Flags().StringVarP(&criteria.Name, "name", "", "", "Search by group name")
	cmdGroupList.Flags().StringVarP(&criteria.NameRegex, "nameRegex", "", "", "Search by group name (supports regex syntax), example: 'foo' will search for the group exactly named 'foo' while '.*foo.*' will search for groups containing 'foo' in their name")
	cmdGroupList.Flags().StringVarP(&criteria.Description, "description", "", "", "Search by description of group")
	cmdGroupList.Flags().StringVarP(&criteria.DateMinCreation, "dateMinCreation", "", "", "Filter result on dateCreation, timestamp Unix Format")
	cmdGroupList.Flags().StringVarP(&criteria.DateMaxCreation, "dateMaxCreation", "", "", "Filter result on dateCreation, timestamp Unix format")
}

var cmdGroupList = &cobra.Command{
	Use:   "list",
	Short: "List all groups: tatcli group list [<skip>] [<limit>], tatcli group list -h to see all criteria",
	Run: func(cmd *cobra.Command, args []string) {
		criteria.Skip, criteria.Limit = internal.GetSkipLimit(args)
		out, err := internal.Client().GroupList(&criteria)
		internal.Check(err)
		internal.Print(out)
	},
}
