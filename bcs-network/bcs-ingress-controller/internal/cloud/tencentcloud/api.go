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

package tencentcloud

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/qcloud"
)

const (
	// APIRequestLimitExceededCode err code for api request exceeded limit
	APIRequestLimitExceededCode = 4400
	// APIWrongStatusCode err code for api request status code
	APIWrongStatusCode = 4000
)

// APIWrapper wrap 2017 clb api
type APIWrapper struct {
	// api for tencent cloud clb v2 api
	apiCli qcloud.APIInterface

	// domain clb domain
	domain string

	// secretID for tencent cloud
	secretID string

	// secretKey for tencent cloud
	secretKey string

	throttler throttle.RateLimiter
}

// NewAPIWrapper create APIWrapper
func NewAPIWrapper() (*APIWrapper, error) {
	a := &APIWrapper{}
	err := a.loadEnv()
	if err != nil {
		return nil, err
	}

	clbClient := qcloud.NewClient(fmt.Sprintf("https://%s/v2/index.php", a.domain), a.secretKey)
	// here we don't use cvm client, so leave it nil
	clbAPI := qcloud.NewAPI(clbClient, nil)
	a.apiCli = clbAPI
	a.throttler = throttle.NewTokenBucket(int64(throttleQPS), int64(bucketSize))

	return a, nil
}

func (a *APIWrapper) checkErrCode(errCode int) {
	if errCode == APIRequestLimitExceededCode {
		blog.Warnf("request exceed limit, have a rest for %d second", waitPeriodLBDealing)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	} else if errCode == APIWrongStatusCode {
		blog.Warnf("clb is dealing another action, have a rest for %d second", waitPeriodLBDealing)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	}
}

func (a *APIWrapper) loadEnv() error {
	secretID := os.Getenv(EnvNameTencentCloudAccessKeyID)
	secretKey := os.Getenv(EnvNameTencentCloudAccessKey)
	a.domain = DefaultTencentCloudClbV2Domain
	a.secretID = secretID
	a.secretKey = secretKey
	return nil
}

func (a *APIWrapper) tryThrottle() {
	now := time.Now()
	a.throttler.Accept()
	if latency := time.Since(now); latency > maxLatency {
		pc, _, _, _ := runtime.Caller(2)
		callerName := runtime.FuncForPC(pc).Name()
		blog.Infof("Throttling request took %d ms, function: %s", latency, callerName)
	}
}

func (a *APIWrapper) waitTaskDone(region string, taskID int) error {
	for counter := 0; counter <= maxRetry; counter++ {
		resp, err := a.describeLoadBalancersTaskResult(region, taskID)
		if err != nil {
			blog.Errorf("describe task %d result failed, err %s", taskID, err.Error())
			return fmt.Errorf("describe task %d result failed, err %s", taskID, err.Error())
		}
		if resp.Data.Status == TaskStatusDealing {
			blog.Warn("clb is dealing")
			time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
			continue
		} else if resp.Data.Status == TaskStatusFailed {
			blog.Errorf("task %s is failed", taskID)
			return fmt.Errorf("task %d is failed", taskID)
		} else if resp.Data.Status == TaskStatusSucceed {
			blog.Infof("task %d is done", taskID)
			return nil
		}
		return fmt.Errorf("error status of task %d", resp.Data.Status)
	}
	blog.Errorf("wait for task %d result timeout", taskID)
	return fmt.Errorf("wait for task %d result timeout", taskID)
}

func (a *APIWrapper) describeLoadBalancersTaskResult(region string, requestID int) (
	*qcloud.DescribeLoadBalancersTaskResultOutput, error) {
	req := new(qcloud.DescribeLoadBalancersTaskResultInput)
	req.Action = "DescribeLoadBalancersTaskResult"
	req.Region = region
	req.SecretID = a.secretID
	req.RequestID = requestID

	var err error
	var resp *qcloud.DescribeLoadBalancersTaskResultOutput
	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.DescribeLoadBalanceTaskResult(req)
		if err != nil {
			blog.Errorf("DescribeLoadBalanceTaskResult failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeLoadBalanceTaskResult failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			blog.Errorf("DescribeLoadBalanceTaskResult falied, errcode %d", resp.Code)
			return nil, fmt.Errorf("DescribeLoadBalanceTaskResult falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		blog.Errorf("DescribeLoadBalanceTaskResult out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeLoadBalanceTaskResult out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// Create4LayerListener create 4 layer listener
func (a *APIWrapper) Create4LayerListener(region string, req *qcloud.CreateForwardLBFourthLayerListenersInput) (
	string, error) {
	req.Action = "CreateForwardLBFourthLayerListeners"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.CreateForwardLBFourthLayerListenersOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.Create4LayerListener(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("CreateForwardLBFourthLayerListeners failed, err %s", err.Error())
			return "", fmt.Errorf("CreateForwardLBFourthLayerListeners failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("CreateForwardLBFourthLayerListeners falied, errcode %d", resp.Code)
			return "", fmt.Errorf("CreateForwardLBFourthLayerListeners falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("CreateForwardLBFourthLayerListeners out of maxRetry %d", maxRetry)
		return "", fmt.Errorf("CreateForwardLBFourthLayerListeners out of maxRetry %d", maxRetry)
	}
	err = a.waitTaskDone(region, resp.RequestID)
	if err != nil {
		result = metrics.LibCallStatusErr
		return "", err
	}
	return resp.ListenerIds[0], nil
}

// DescribeForwardLBListeners describe forward lb listeners
func (a *APIWrapper) DescribeForwardLBListeners(region string, req *qcloud.DescribeForwardLBListenersInput) (
	*qcloud.DescribeForwardLBListenersOutput, error) {

	req.Action = "DescribeForwardLBListeners"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.DescribeForwardLBListenersOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.DescribeForwardLBListeners(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("DescribeForwardLBListeners failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeForwardLBListeners failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("DescribeForwardLBListeners falied, errcode %d", resp.Code)
			return nil, fmt.Errorf("DescribeForwardLBListeners falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("DescribeForwardLBListeners out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeForwardLBListeners out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// DescribeForwardLBBackends wrap DescribeForwardLBBackends
func (a *APIWrapper) DescribeForwardLBBackends(region string, req *qcloud.DescribeForwardLBBackendsInput) (
	*qcloud.DescribeForwardLBBackendsOutput, error) {

	req.Action = "DescribeForwardLBBackends"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.DescribeForwardLBBackendsOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.DescribeForwardLBBackends(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("DescribeForwardLBBackends failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeForwardLBBackends failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("DescribeForwardLBBackends falied, errcode %d", resp.Code)
			return nil, fmt.Errorf("DescribeForwardLBBackends falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("DescribeForwardLBBackends out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeForwardLBBackends out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// RegInstancesWith4LayerListener register instance with 4 layer listener
func (a *APIWrapper) RegInstancesWith4LayerListener(region string,
	req *qcloud.RegisterInstancesWithForwardLBFourthListenerInput) error {

	req.Action = "RegisterInstancesWithForwardLBFourthListener"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.RegisterInstancesWithForwardLBFourthListenerOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.RegInstancesWith4LayerListener(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("RegisterInstancesWithForwardLBFourthListener failed, err %s", err.Error())
			return fmt.Errorf("RegisterInstancesWithForwardLBFourthListener failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("RegisterInstancesWithForwardLBFourthListener falied, errcode %d", resp.Code)
			return fmt.Errorf("RegisterInstancesWithForwardLBFourthListener falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("RegisterInstancesWithForwardLBFourthListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("RegisterInstancesWithForwardLBFourthListener out of maxRetry %d", maxRetry)
	}
	err = a.waitTaskDone(region, resp.RequestID)
	if err != nil {
		result = metrics.LibCallStatusErr
		return err
	}
	return nil
}

// DeRegInstancesWith4LayerListener deregister instance with 4 layer listener
func (a *APIWrapper) DeRegInstancesWith4LayerListener(region string,
	req *qcloud.DeregisterInstancesFromForwardLBFourthListenerInput) error {

	req.Action = "DeregisterInstancesFromForwardLBFourthListener"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.DeregisterInstancesFromForwardLBFourthListenerOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.DeRegInstancesWith4LayerListener(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("DeregisterInstancesFromForwardLBFourthListener failed, err %s", err.Error())
			return fmt.Errorf("DeregisterInstancesFromForwardLBFourthListener failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("DeregisterInstancesFromForwardLBFourthListener falied, errcode %d", resp.Code)
			return fmt.Errorf("DeregisterInstancesFromForwardLBFourthListener falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("DeregisterInstancesFromForwardLBFourthListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeregisterInstancesFromForwardLBFourthListener out of maxRetry %d", maxRetry)
	}
	err = a.waitTaskDone(region, resp.RequestID)
	if err != nil {
		result = metrics.LibCallStatusErr
		return err
	}
	return nil
}

// DeleteListener wrapper DeleteListener
func (a *APIWrapper) DeleteListener(region string, req *qcloud.DeleteForwardLBListenerInput) error {
	req.Action = "DeleteForwardLBListener"
	req.Nonce = uint(rand.Uint32())
	req.Region = region
	req.SecretID = a.secretID
	req.Timestamp = uint(time.Now().Unix())

	var err error
	var resp *qcloud.DeleteForwardLBListenerOutput

	startTime := time.Now()
	result := metrics.LibCallStatusOK
	defer metrics.ReportLibRequestMetric(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudAPI,
		req.Action, result, startTime)

	counter := 1
	for ; counter <= maxRetry; counter++ {
		a.tryThrottle()
		req.Nonce = uint(rand.Uint32())
		req.Timestamp = uint(time.Now().Unix())
		resp, err = a.apiCli.DeleteListener(req)
		if err != nil {
			result = metrics.LibCallStatusErr
			blog.Errorf("DeleteForwardLBListener failed, err %s", err.Error())
			return fmt.Errorf("DeleteForwardLBListener failed, err %s", err.Error())
		}
		a.checkErrCode(resp.Code)
		if resp.Code == APIRequestLimitExceededCode || resp.Code == APIWrongStatusCode {
			continue
		}
		if resp.Code != 0 {
			result = metrics.LibCallStatusErr
			blog.Errorf("DeleteForwardLBListener falied, errcode %d", resp.Code)
			return fmt.Errorf("DeleteForwardLBListener falied, errcode %d", resp.Code)
		}
		break
	}
	if counter > maxRetry {
		result = metrics.LibCallStatusErr
		blog.Errorf("DeleteForwardLBListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteForwardLBListener out of maxRetry %d", maxRetry)
	}
	err = a.waitTaskDone(region, resp.RequestID)
	if err != nil {
		result = metrics.LibCallStatusErr
		return err
	}
	return nil
}
