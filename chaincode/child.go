// sycoin
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	//"time"

	shim "github.com/hyperledger/fabric/core/chaincode/shim"
)

//=================================================================================================================================
//	 Structure Definitions
//=================================================================================================================================
type Chaincode struct {
}

/* 儿童信息账本
1  身份ID（身份证号）
2  名字
3  性别
4  年龄
5  父亲名
6  母亲名
7  紧急电话
8  备注
*/
type ChildInfo struct {
	CustID  string `json:"CustID"`
	CustName string `json:"CustName"`
	Sex string `json:"Sex"`
	Age string `json:"Age"`
	Father string `json:"Father"`
	Mother string `json:"Mother"`
	EmeTel string `json:"EmeTel"`
	Memo string `json:"Memo"`
}

/* 路线轨迹账本
1  轨迹号
2  经度
3  纬度
*/
type RouteTrack struct {
	TrackNumber  string `json:"TrackNumber"`
	Date string  `json:"Date"`
	Latitude int64 `json:"Latitude"`
	Longitude int64 `json:"Longitude"`
}

/*路线轨迹查询返回结构
1  []路线轨迹账本
2  错误码
*/
type ListToTr struct {
	ListTrade []RouteTrack `json:"ListTrade"`
	ErrCode     string     `json:"ErrCode"`
}

//常量定义
const (
	// 定义LOG文件名
	CHAINCODE_LOG_FILE   = "child.log"
	CHAINCODE_LOG_PREFIX = ""
	// 儿童信息账本KEY名
	CHILD_INFO = "Child_Info"
	// 路线轨迹账本KEY名
	ROUTE_TRACK = "Route_Track"
	// 儿童信息登陆
	FUNC_CUST_REG = "CustReg"
	// 儿童信息查询
	FUNC_CUST_QUERY = "CustQuery"
	// 路线轨迹登录
	FUNC_ROUTE_TRACK_REG = "RouteTrackReg"
	// 路线轨迹查询
	FUNC_ROUTE_TRACK_QUERY = "RouteTrackQuery"
)

//============================================================================================================
//	 Function GetCustomerAccount 儿童信息查询
//============================================================================================================
func (t *Chaincode) CustQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		logger.Debugf("[CustQuery]参数错误。预期1个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("CustQuery function expects 1 argument")
	}
	custID := args[0]
	key := CHILD_INFO + custID
	errMsg := ""

	logger.Debugf("[QueryExchangeRate]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}

	logger.Debugf("[QueryExchangeRate]Query Result for KEY:%s,VALUE=%s\n", key, string(getState_Byte))

	if getState_Byte == nil {
		errMsg = fmt.Sprintf("Query Result is Nil For Id:%s", custID)
		return nil, errors.New(errMsg)
	}

	return getState_Byte, nil
}

//============================================================================================================
//	 Function Init
//============================================================================================================
func (t *Chaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	return nil, nil
}

//============================================================================================================
//	 Function Invoke Invoke路由函数
//============================================================================================================
func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	// 儿童信息登陆
	case FUNC_CUST_REG:
		return t.CustReg(stub, function, args)
	// 路线轨迹登录
	case FUNC_ROUTE_TRACK_REG:
		return t.RouteTrackReg(stub, function, args)
	}
	return nil, errors.New("Invalid Function Call:" + function)
}

//============================================================================================================
//	 Function CustReg 儿童信息登陆
//============================================================================================================
func (t *Chaincode) CustReg(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 8 {
		logger.Debugf("[CustReg]参数错误。预期8个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("AccountCReg function expects 8 argument")
	}
	custID := args[0]
	custName := args[1]
	sex := args[2]
	age := args[3]
	father := args[4]
	mother := args[5]
	emeTel := args[6]
	memo := args[7]

	childInfo := ChildInfo{
		CustID:   custID,
		CustName:   custName,
		Sex:   sex,
		Age:   age,
		Father:   father,
		Mother: mother,
		EmeTel:  emeTel,
		Memo: memo}

	key := CHILD_INFO + custID
	putState_Byte, _ := json.Marshal(&childInfo)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err := stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return putState_Byte, nil
}

//============================================================================================================
//	 Function RouteTrackReg 路线轨迹登录
//============================================================================================================
func (t *Chaincode) RouteTrackReg(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 经度
	var latitude int64
	// 纬度
	var longitude int64
	// 错误信息
	var errMsg string
	if len(args) != 4 {
		logger.Debugf("[RouteTrackReg]参数错误。预期4个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("RouteTrackReg function expects 4 argument")
	}
	custID := args[0]
	date := args[1]
	trackNumber := custID + date

	latitude, err := strconv.ParseInt(args[2], 10, 64) //strconv.Atoi(args[1])
	if err != nil {
		logger.Debugf("[RouteTrackReg]参数错误。经度应为数字,AMOUNT=%s。参数列表=%v", args[2], args)
		errMsg = fmt.Sprintf("Invalid argument,Latitude should be number. args[2]:%s", args[2])
		return nil, errors.New(errMsg)
	}
	longitude, err = strconv.ParseInt(args[3], 10, 64) //strconv.Atoi(args[1])
	if err != nil {
		logger.Debugf("[RouteTrackReg]参数错误。纬度应为数字,AMOUNT=%s。参数列表=%v", args[3], args)
		errMsg = fmt.Sprintf("Invalid argument,longitude should be number. args[3]:%s", args[3])
		return nil, errors.New(errMsg)
	}
	routeTrack := RouteTrack{
		TrackNumber:     trackNumber,
		Date:		date,
		Latitude: 	latitude,
		Longitude:	longitude}

	key := ROUTE_TRACK + trackNumber
	putState_Byte, _ := json.Marshal(&routeTrack)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return putState_Byte, nil
}

//============================================================================================================
//	 Function RouteTrackQuery 路线轨迹查询
//============================================================================================================
func (t *Chaincode) RouteTrackQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 身份ID（身份证号）
	var custID string
	// 开始时间
	var startDate string
	// 结束时间
	var endDate string
	// 错误信息
	var err error
	var errMsg string

	// 路线轨迹账本
	var routeTrack RouteTrack
	// 返回账本结构
	var listToTr ListToTr
	// 账本数组
	var transArray []RouteTrack

	if len(args) != 3 {
		logger.Debugf("[RouteTrackQuery]参数错误。预期3个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("RouteTrackQuery function expects 3 argument")
	}
	custID = args[0]
	startDate = args[1]
	endDate = args[2]
	errMsg = ""

	startKey := ROUTE_TRACK + custID + startDate
	endKey := ROUTE_TRACK + custID + endDate

	iter, err := stub.RangeQueryState(startKey, endKey)
	if err != nil {
		return nil, err
	}
	for iter.HasNext() {
		_, value, err := iter.Next()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(value, &routeTrack)
		if err != nil {
			return nil, err
		}
		transArray = append(transArray, routeTrack)
	}
	listToTr.ListTrade = transArray
	listToTr.ErrCode = errMsg

	bytes, _ := json.Marshal(&listToTr)

	return bytes, nil
}

//============================================================================================================
//	 Function Query Query路由函数
//============================================================================================================
func (t *Chaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	// 儿童信息查询
	case FUNC_CUST_QUERY:
		return t.CustQuery(stub, function, args)
	// 路线轨迹查询
	case FUNC_ROUTE_TRACK_QUERY:
		return t.RouteTrackQuery(stub, function, args)
	}
	return nil, errors.New("Invalid Function Call:" + function)
}

// Log定义
var logger *shim.ChaincodeLogger = shim.NewLogger("CHILD")

//============================================================================================================
//	 Function main函数
//============================================================================================================
func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		panic(err)
	}
}
