/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variablegroup

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/audit"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateAction create a variable group object.
type CreateAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateVariableGroupReq
	resp *pb.CreateVariableGroupResp

	newVarGroupID string
}

// NewCreateAction creates new CreateAction
func NewCreateAction(kit kit.Kit, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateVariableGroupReq, resp *pb.CreateVariableGroupResp) *CreateAction {
	action := &CreateAction{kit: kit, viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}

	return nil
}

func (act *CreateAction) genVariableGroupID() error {
	id, err := common.GenVariableGroupID()
	if err != nil {
		return err
	}
	act.newVarGroupID = id
	return nil
}

func (act *CreateAction) queryBusinessSharding() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryShardingReq{
		Seq: act.kit.Rid,
		Key: act.req.BizId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateVariableGroup[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QuerySharding(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QuerySharding, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateAction) initDefaultShardingDB() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.InitShardingDBReq{Seq: act.kit.Rid}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateVariableGroup[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.InitShardingDB(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager InitShardingDB, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateAction) createBusinessSharding() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateShardingReq{
		Seq:    act.kit.Rid,
		Key:    act.req.BizId,
		DbId:   database.BSCPDEFAULTSHARDINGDBID,
		DbName: database.BSCPDEFAULTSHARDINGDB,
		Memo:   "system default sharding",
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateVariableGroup[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateSharding(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateSharding, %+v", err)
	}
	if resp.Code == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.Code, resp.Message
}

func (act *CreateAction) initBusiness() (pbcommon.ErrCode, string) {
	errCode, errMsg := act.queryBusinessSharding()
	if errCode == pbcommon.ErrCode_E_OK {
		// business sharding stuff already inited.
		return pbcommon.ErrCode_E_OK, ""
	}

	if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
		// other errors.
		return errCode, errMsg
	}

	// business sharding stuff have not inited.
	if errCode, errMsg := act.initDefaultShardingDB(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// create sharding for the business.
	if errCode, errMsg := act.createBusinessSharding(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createVariableGroup() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.CreateVariableGroupReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		VarGroupId: act.newVarGroupID,
		Name:       act.req.Name,
		Memo:       act.req.Memo,
		Creator:    act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateVariableGroup[%s]| request to DataManager, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.CreateVariableGroup(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager CreateVariableGroup, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK && resp.Code != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return resp.Code, resp.Message
	}
	act.resp.Data = &pb.CreateVariableGroupResp_RespData{VarGroupId: resp.Data.VarGroupId}
	act.newVarGroupID = resp.Data.VarGroupId

	if resp.Code == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_TPL_ALREADY_EXISTS, resp.Message
	}

	// audit here on new variable group created.
	audit.Audit(int32(pbcommon.SourceType_ST_VAR_GROUP), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.BizId, act.newVarGroupID, act.kit.User, act.req.Memo)

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genVariableGroupID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, err.Error())
	}

	// init business sharding stuff in this low QPS action.
	if errCode, errMsg := act.initBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create variable group.
	if errCode, errMsg := act.createVariableGroup(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
