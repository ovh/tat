package message

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"

	"github.com/ovh/tat"
	"github.com/ovh/tat/tatcli/internal"
	"github.com/spf13/cobra"
)

var (
	criteria         tat.MessageCriteria
	stream           bool
	execMsg, execErr []string
)

func init() {
	cmdMessageList.Flags().StringVarP(&criteria.TreeView, "treeView", "", "", "Tree View of messages: onetree or fulltree. Default: notree")
	cmdMessageList.Flags().StringVarP(&criteria.IDMessage, "idMessage", "", "", "Search by IDMessage")
	cmdMessageList.Flags().StringVarP(&criteria.InReplyOfID, "inReplyOfID", "", "", "Search by IDMessage InReply")
	cmdMessageList.Flags().StringVarP(&criteria.InReplyOfIDRoot, "inReplyOfIDRoot", "", "", "Search by IDMessage IdRoot")
	cmdMessageList.Flags().StringVarP(&criteria.AllIDMessage, "allIDMessage", "", "", "Search in All ID Message (idMessage, idReply, idRoot)")
	cmdMessageList.Flags().StringVarP(&criteria.Text, "text", "", "", "Search by text")
	cmdMessageList.Flags().StringVarP(&criteria.Topic, "topic", "", "", "Search by topic")
	cmdMessageList.Flags().StringVarP(&criteria.Label, "label", "", "", "Search by label: could be labelA,labelB")
	cmdMessageList.Flags().StringVarP(&criteria.StartLabel, "startLabel", "", "", "Search by a label prefix: --startLabel='mykey:,myKey2:'")
	cmdMessageList.Flags().StringVarP(&criteria.NotLabel, "notLabel", "", "", "Search by label (exclude): could be labelA,labelB")
	cmdMessageList.Flags().StringVarP(&criteria.AndLabel, "andLabel", "", "", "Search by label (and) : could be labelA,labelB")
	cmdMessageList.Flags().StringVarP(&criteria.Tag, "tag", "", "", "Search by tag : could be tagA,tagB")
	cmdMessageList.Flags().StringVarP(&criteria.StartTag, "startTag", "", "", "Search by a tag prefix: --startTag='mykey:,myKey2:'")
	cmdMessageList.Flags().StringVarP(&criteria.NotTag, "notTag", "", "", "Search by tag (exclude) : could be tagA,tagB")
	cmdMessageList.Flags().StringVarP(&criteria.AndTag, "andTag", "", "", "Search by tag (and) : could be tagA,tagB")
	cmdMessageList.Flags().StringVarP(&criteria.DateMinCreation, "dateMinCreation", "", "", "Search by dateCreation (timestamp), select messages where dateCreation >= dateMinCreation")
	cmdMessageList.Flags().StringVarP(&criteria.DateMaxCreation, "dateMaxCreation", "", "", "Search by dateCreation (timestamp), select messages where dateCreation <= dateMaxCreation")
	cmdMessageList.Flags().StringVarP(&criteria.DateMinUpdate, "dateMinUpdate", "", "", "Search by dateUpdate (timestamp), select messages where dateUpdate >= dateMinUpdate")
	cmdMessageList.Flags().StringVarP(&criteria.DateMaxUpdate, "dateMaxUpdate", "", "", "Search by dateUpdate (timestamp), select messages where dateUpdate <= dateMaxUpdate")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefCreation, "dateRefCreation", "", "", "This have to be used with dateRefDeltaMinCreation and / or dateRefDeltaMaxCreation. This could be BeginningOfMinute, BeginningOfHour, BeginningOfDay, BeginningOfWeek, BeginningOfMonth, BeginningOfQuarter, BeginningOfYear")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefDeltaMinCreation, "dateRefDeltaMinCreation", "", "", "Add seconds to dateRefCreation flag")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefDeltaMaxCreation, "dateRefDeltaMaxCreation", "", "", "Add seconds to dateRefCreation flag")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefUpdate, "dateRefUpdate", "", "", "This have to be used with dateRefDeltaMinUpdate and / or dateRefDeltaMaxUpdate. This could be BeginningOfMinute, BeginningOfHour, BeginningOfDay, BeginningOfWeek, BeginningOfMonth, BeginningOfQuarter, BeginningOfYear")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefDeltaMinUpdate, "dateRefDeltaMinUpdate", "", "", "Add seconds to dateRefUpdate flag")
	cmdMessageList.Flags().StringVarP(&criteria.DateRefDeltaMaxUpdate, "dateRefDeltaMaxUpdate", "", "", "Add seconds to dateRefUpdate flag")
	cmdMessageList.Flags().StringVarP(&criteria.LastMinCreation, "lastMinCreation", "", "", "Search by dateCreation (duration in second), select messages where dateCreation >= now - lastMinCreation")
	cmdMessageList.Flags().StringVarP(&criteria.LastMaxCreation, "lastMaxCreation", "", "", "Search by dateCreation (duration in second), select messages where dateCreation <= now - lastMaxCreation")
	cmdMessageList.Flags().StringVarP(&criteria.LastMinUpdate, "lastMinUpdate", "", "", "Search by dateUpdate (duration in second), select messages where dateUpdate >= now - lastMinCreation")
	cmdMessageList.Flags().StringVarP(&criteria.LastMaxUpdate, "lastMaxUpdate", "", "", "Search by dateUpdate (duration in second), select messages where dateUpdate <= now - lastMaxCreation")
	cmdMessageList.Flags().StringVarP(&criteria.LastHourMinCreation, "lastHourMinCreation", "", "", "Search by dateCreation, select messages where dateCreation >= Now Beginning Of Hour - (60 * lastHourMinCreation)")
	cmdMessageList.Flags().StringVarP(&criteria.LastHourMaxCreation, "lastHourMaxCreation", "", "", "Search by dateCreation, select messages where dateCreation <= Now Beginning Of Hour - (60 * lastHourMaxCreation)")
	cmdMessageList.Flags().StringVarP(&criteria.LastHourMinUpdate, "lastHourMinUpdate", "", "", "Search by dateUpdate, select messages where dateUpdate >= Now Beginning Of Hour - (60 * lastHourMinCreation)")
	cmdMessageList.Flags().StringVarP(&criteria.LastHourMaxUpdate, "lastHourMaxUpdate", "", "", "Search by dateUpdate, select messages where dateUpdate <= Now Beginning Of Hour - (60 * lastHourMaxCreation)")
	cmdMessageList.Flags().StringVarP(&criteria.Username, "username", "", "", "Search by username : could be usernameA,usernameB")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMinNbReplies, "limitMinNbReplies", "", "", "In onetree mode, filter root messages with more or equals minNbReplies")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMaxNbReplies, "limitMaxNbReplies", "", "", "In onetree mode, filter root messages with min or equals maxNbReplies")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMinNbVotesUP, "limitMinNbVotesUP", "", "", "Search by nbVotesUP")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMaxNbVotesUP, "limitMaxNbVotesUP", "", "", "Search by nbVotesUP")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMinNbVotesDown, "limitMinNbVotesDown", "", "", "Search by nbVotesDown")
	cmdMessageList.Flags().StringVarP(&criteria.LimitMaxNbVotesDown, "limitMaxNbVotesDown", "", "", "Search by nbVotesDown")
	cmdMessageList.Flags().StringVarP(&criteria.OnlyMsgRoot, "onlyMsgRoot", "", "", "--onlyMsgRoot=true: restricts to root message only (inReplyOfIDRoot empty). If treeView is used, limit search criteria to root message, replies are still given, independently of search criteria.")
	cmdMessageList.Flags().StringVarP(&criteria.OnlyMsgReply, "onlyMsgReply", "", "", "--onlyMsgReply=true: restricts to reply message only (inReplyOfIDRoot not empty). If treeView is used, limit search criteria to reply, messages root are still given, independently of search criteria.")
	cmdMessageList.Flags().StringVarP(&criteria.OnlyCount, "onlyCount", "", "", "--onlyCount=true: only count messages, without retrieve msg. limit, skip, treeview criterias are ignored.")
	cmdMessageList.Flags().StringVarP(&criteria.SortBy, "sortBy", "", "", "--sortBy=-dateCreation: sort message. Use '-' to reverse sort. Default is --sortBy=-dateCreation. You can use: text, topic, inReplyOfID, inReplyOfIDRoot, nbLikes, labels, likers, votersUP, votersDown, nbVotesUP, nbVotesDown, userMentions, urls, tags, dateCreation, dateUpdate, author, nbReplies")
	cmdMessageList.Flags().BoolVarP(&stream, "stream", "s", false, "stream messages --stream. Request tat each 10s, default sort: dateUpdate")
	cmdMessageList.Flags().StringSliceVarP(&execMsg, "exec", "", nil, `--stream required. Exec a cmd on each new message: --stream --exec 'myLights --pulse blue --duration=1000' With only --onlyMsgCount=true : --exec min:max:cmda --exec min:max:cmdb, example: --exec 0:4:'cmdA' --exec 5::'cmdb'`)
	cmdMessageList.Flags().StringSliceVarP(&execErr, "execErr", "", nil, `--stream required. Exec a cmd on each error while requesting tat: --stream --exec 'myLights --pulse blue --duration=1000' --execErr 'myLights --pulse red --duration=2000'`)
}

var cmdMessageList = &cobra.Command{
	Use:     "list",
	Short:   "List all messages on one topic: tatcli msg list <Topic> <skip> <limit>",
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			internal.Exit("Invalid argument to list message: See tatcli msg list --help\n")
		}

		if stream {
			cmdMessageListStream(args[0])
			return
		}

		criteria.Skip, criteria.Limit = internal.GetSkipLimit(args)
		c := internal.Client()
		var err error
		var out interface{}
		if criteria.OnlyCount == tat.True {
			out, err = c.MessageCount(args[0], &criteria)
		} else {
			out, err = c.MessageList(args[0], &criteria)
		}
		internal.Check(err)
		internal.Print(out)
	},
}

func cmdMessageListStream(topic string) {
	c := internal.Client()
	lastTime := float64(time.Now().Unix())
	lastID := ""
	lastCount := -1

	for {
		criteria.Skip = 0
		criteria.Limit = 10
		criteria.SortBy = "dateUpdate"
		criteria.DateMinUpdate = fmt.Sprintf("%f", lastTime)

		if criteria.OnlyCount == tat.True {
			out, err := c.MessageCount(topic, &criteria)
			if err != nil {
				processExecError(err)
				continue
			}
			processCount(out.Count)
			if out.Count != lastCount {
				time.Sleep(10 * time.Second)
			}
			continue
		}

		out, err := c.MessageList(topic, &criteria)
		if err != nil {
			processExecError(err)
			continue
		}

		for _, m := range out.Messages {
			if lastID != m.ID {
				processMsg(m)
				lastTime = m.DateUpdate
			}
		}

		// do not wait if request reach criteria.Limit
		if len(out.Messages) < criteria.Limit {
			time.Sleep(10 * time.Second)
		}
	}
}

func processExecError(err error) {
	fmt.Printf("Error:%s", err)
	for _, ex := range execErr {
		execCmd(ex)
	}
}

func processCount(count int) {
	fmt.Printf("count:%d\n", count)
	h := "Invalid --exec with --onlyMsgCount. Use --exec min:max:cmd, example: --exec 0:4:'myLights --pulse blue --duration=1000'\n"

	maxInt := int(^uint(0) >> 1)

	for _, ex := range execMsg {
		tuple := strings.Split(ex, ":")
		if len(tuple) != 3 {
			internal.Exit(h)
		}

		min, errmin := strconv.Atoi(tuple[0])
		if errmin != nil {
			internal.Exit(h)
		}

		max := maxInt
		if tuple[1] != "" {
			var errmax error
			max, errmax = strconv.Atoi(tuple[1])
			if errmax != nil {
				internal.Exit(h)
			}
		}

		if count >= min && count <= max {
			execCmd(tuple[2])
		}
	}
}

func processMsg(msg tat.Message) {
	fmt.Printf("%s %s %s\n", time.Unix(int64(msg.DateUpdate), 0).Format(time.Stamp), msg.Author.Username, msg.Text)
	for _, ex := range execMsg {
		execCmd(ex)
	}
}

func execCmd(ex string) {
	opts := strings.Split(ex, " ")
	if ex != "" {
		s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
		cmd := exec.Command(opts[0], opts[1:]...)
		s.Start()
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		s.Stop()
	}

}
