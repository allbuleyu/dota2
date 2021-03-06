package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/allbuleyu/dota2/config"
	"github.com/mongodb/mongo-go-driver/bson"
	"net/http"
	"strconv"
	"time"
)

type TeamInfo struct {
	Team_id int64 `json:"-"`

	// The team's unique id.
	Name string
	Tag string

	// Unix timestamp of when the team was created.
	Time_created time.Duration
	Calibration_games_remaining int64

	// The UGC id for the team logo. You can resolve this with the GetUGCFileDetails method.
	// GetUGCFileDetails = https://wiki.teamfortress.com/wiki/WebAPI/GetUGCFileDetails
	Logo uint64
	Logo_sponsor uint64



	// The team's ISO 3166-1 country-code.
	Country_code string

	// The URL the team provided upon creation.
	Url string

	// Amount of matches played with the current roster
	Games_played int64

	// 32-bit account ID. Where N is incremental from 0 for every player on the team.
	Player_0_account_id uint64
	Player_1_account_id uint64
	Player_2_account_id uint64
	Player_3_account_id uint64
	Player_4_account_id uint64
	Player_5_account_id uint64

	// 32-bit account ID of the team's admin.
	Admin_account_id uint64
}

// startTeamId The. team id to start returning results from.
// teamsRequested. The amount of teams to return.
// 通过比赛的战队ID查找战队信息才不会有战队ID出错的情况
func GetTeamsInfo()  {
	client, _ := config.NewMongoClient("")
	ctx, _ := context.WithTimeout(context.Background(), config.CtxTimeOutDuration)

	filter := bson.M{}
	lastOne := client.Database("d2log").Collection("data_teams").FindOne(ctx, filter)
	teamInfo := TeamInfo{}
	lastOne.Decode(&teamInfo)

	var startTeamId, teamsRequested int64
	startTeamId = teamInfo.Team_id + 1
	addr := "https://api.steampowered.com/IDOTA2Match_570/GetTeamInfoByTeamID/v1/"
	query := map[string]string{
		"start_at_team_id": strconv.FormatInt(startTeamId, 10),
		"teams_requested": strconv.FormatInt(teamsRequested, 10),
	}
	addr = config.Addr(addr, query)

	request, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		fmt.Println("request err", err)
		return
	}

	resp, err := http.DefaultClient.Do(request)

	// decoder
	decoder := json.NewDecoder(resp.Body)

	result := struct {
		Result struct{
			Teams []TeamInfo
			Status int64
		}
	}{}

	err = decoder.Decode(&result)
	if err != nil {
		fmt.Println("decode err :", err)
		return
	}

	teams := result.Result.Teams
	if len(teams) == 0 {
		fmt.Println("teams is empty by idstart", startTeamId)
		return
	}


	insertData := make([]interface{}, 0)
	for i := range teams {
		teams[i].Team_id = startTeamId + int64(i)
		insertData = append(insertData, teams[i])
	}


	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("client connect err", err)
	}

	dbRes, err := client.Database("d2log").Collection("data_teams").InsertMany(ctx, insertData)
	fmt.Println(dbRes.InsertedIDs)
}