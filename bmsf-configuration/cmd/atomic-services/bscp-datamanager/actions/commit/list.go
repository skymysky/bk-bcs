/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package commit

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// ListAction is commit list action object.
type ListAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHistoryCommitsReq
	resp *pb.QueryHistoryCommitsResp

	sd *dbsharding.ShardingDB

	totalCount int64
	commits    []database.Commit
}

// NewListAction creates new ListAction.
func NewListAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHistoryCommitsReq, resp *pb.QueryHistoryCommitsResp) *ListAction {
	action := &ListAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	commits := []*pbcommon.Commit{}
	for _, st := range act.commits {
		commit := &pbcommon.Commit{
			BizId:         st.BizID,
			CommitId:      st.CommitID,
			AppId:         st.AppID,
			CfgId:         st.CfgID,
			Operator:      st.Operator,
			ReleaseId:     st.ReleaseID,
			MultiCommitId: st.MultiCommitID,
			Memo:          st.Memo,
			State:         st.State,
			CommitMode:    st.CommitMode,
			CreatedAt:     st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		commits = append(commits, commit)
	}
	act.resp.Data = &pb.QueryHistoryCommitsResp_RespData{TotalCount: uint32(act.totalCount), Info: commits}
	return nil
}

func (act *ListAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.AppId) == 0 && len(act.req.CfgId) == 0 && len(act.req.Operator) == 0 {
		return errors.New("invalid input data, app_id or cfg_id or operator is required")
	}

	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	if act.req.Page == nil {
		return errors.New("invalid input data, page is required")
	}
	if err = common.ValidateInt32("page.start", act.req.Page.Start,
		database.BSCPEMPTY, math.MaxInt32); err != nil {
		return err
	}
	if err = common.ValidateInt32("page.limit", act.req.Page.Limit,
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *ListAction) queryHistoryCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, ""
	}

	// query type, 0:All(default)  1:Init  2:Confirmed  3:Canceled
	whereState := fmt.Sprintf("Fstate IN (%d, %d, %d)",
		pbcommon.CommitState_CS_INIT, pbcommon.CommitState_CS_CONFIRMED, pbcommon.CommitState_CS_CANCELED)

	if act.req.QueryType == 1 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_INIT)
	} else if act.req.QueryType == 2 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CONFIRMED)
	} else if act.req.QueryType == 3 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CANCELED)
	}

	err := act.sd.DB().
		Model(&database.Commit{}).
		Where(&database.Commit{
			BizID:    act.req.BizId,
			AppID:    act.req.AppId,
			CfgID:    act.req.CfgId,
			Operator: act.req.Operator,
		}).
		Where(whereState).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ListAction) queryHistoryCommits() (pbcommon.ErrCode, string) {
	// query type, 0:All(default)  1:Init  2:Confirmed  3:Canceled
	whereState := fmt.Sprintf("Fstate IN (%d, %d, %d)",
		pbcommon.CommitState_CS_INIT, pbcommon.CommitState_CS_CONFIRMED, pbcommon.CommitState_CS_CANCELED)

	if act.req.QueryType == 1 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_INIT)
	} else if act.req.QueryType == 2 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CONFIRMED)
	} else if act.req.QueryType == 3 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CANCELED)
	}

	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Commit{
			BizID:    act.req.BizId,
			AppID:    act.req.AppId,
			CfgID:    act.req.CfgId,
			Operator: act.req.Operator,
		}).
		Where(whereState).
		Find(&act.commits).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query commits count.
	if errCode, errMsg := act.queryHistoryCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query commits list.
	if errCode, errMsg := act.queryHistoryCommits(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
