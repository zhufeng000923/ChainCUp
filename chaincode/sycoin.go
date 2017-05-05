// sycoin
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	shim "github.com/hyperledger/fabric/core/chaincode/shim"
)

//=================================================================================================================================
//	 Structure Definitions
//=================================================================================================================================
type Chaincode struct {
}

/* 兑换比率账本
1  积分数
2  苏银币
3  人民币
*/
type JSBankExchangeRate struct {
	Point  int64 `json:"Point"`
	SyCoin int64 `json:"SyCoin"`
	RMB    int64 `json:"RMB"`
}

/* 银行账本结构
1  积分总额
2  苏银币总额
*/
type JSBankSYCoin struct {
	TotalPoint  int64 `json:"TotalPoint"`
	TotalSyCoin int64 `json:"TotalSyCoin"`
}

/* 客户账本结构
1  客户ECIF号
2  苏银币余额
3  可用苏银币余额
4  冻结苏银币余额
*/
type JSBankCustomerSYCoin struct {
	ECIF   string `json:"AccountId"`
	Amount int64  `json:"Amount"`
	Avail  int64  `json:"Avail"`
	Frozen int64  `json:"Frozen"`
}

/* 商户账本结构
1  商户号
2  积分余额
3  苏银币余额
*/
type JSBankMerchantSYCoin struct {
	Id     string `json:"AccountId"`
	Point  int64  `json:"Point"`
	Amount int64  `json:"Amount"`
}

/*交易信息账本结构 KEY:JSBank_Trade_+ID+流水号
1  流水号 （YYYYMMDDHHMMSS+采番号12位+交易类型：BU/EX/GC/GM）
2  出金账户ID
3  入金账户ID
4  交易类型 （BU：购买,EX：兑换,GC：客户赠与/转让,GM:商户赠与/转让）
5  交易苏银币金额
6  出金账户余额
7  入金账户余额
8  交易日期 (YYYYMMDD)
9  商品编码
10 摘要
*/
type TransactionInfo struct {
	SerialNumber      string `json:"SerialNumber"`
	CashOutAccount    string `json:"CashOutAccount"`
	CashInAccount     string `json:"CashInAccount"`
	TransactionType   string `json:"TransactionType"`
	TransactionAmount int64  `json:"TransactionAmount"`
	CashOutBalance    int64  `json:"CashOutBalance"`
	CashInBalance     int64  `json:"CashInBalance"`
	TransactionDate   string `json:"TransactionDate"`
	ProcuctID         string `json:"ProcuctID"`
	Memo              string `json:"Memo"`
}

/*交易记录查询返回结构
1  []错易信息账本
2  错误码
*/
type ListTrans struct {
	ListTrade []TransactionInfo `json:"ListTrade"`
	Error     string            `json:"Error"`
}

//常量定义
const (
	// 定义LOG文件名
	CHAINCODE_LOG_FILE   = "sycoin.log"
	CHAINCODE_LOG_PREFIX = ""
	// 苏银币兑换比率
	JSBANK_SYCOIN_EXCHANGE_RATE = "JSBank_SYCoin_Exchange_Rate"
	// 银行全局账本KEY名
	JSBANK_SYCOIN = "JSBank_SYCoin"
	// 客户账本KEY名
	JSBANK_SYCOIN_C = "JSBank_SYCoin_C_"
	// 商户账本KEY名
	JSBANK_SYCOIN_M = "JSBank_SYCoin_M_"
	// 交易记录全局账本KEY名
	JSBANK_TRADE = "JSBank_Trade_"
	// 全局数据初期化
	FUNC_INITIALIZE = "Initialize"
	// 设置兑换比率
	FUNC_SET_EXCHANGE_RATE = "SetExchangeRate"
	// 设置兑换比率
	FUNC_QUERY_EXCHANGE_RATE = "QueryExchangeRate"
	// 客户账户注册
	FUNC_ACCOUNT_C_REG = "AccountCReg"
	// 商户账户注册
	FUNC_ACCOUNT_M_REG = "AccountMReg"
	// 赠与
	FUNC_SYCOIN_GIFT = "SyCoinGift"
	// 积分兑换账户
	FUNC_POINT_EXCHANGE = "PointExchange"
	// 积分兑换苏银币查询
	FUNC_POINT_EXCHANGE_QUERY = "PointExchangeQuery"
	// 银行全局账本查询
	FUNC_ACCOUNT_B_QUERY = "AccountBQuery"
	// 客户账户查询
	FUNC_ACCOUNT_C_QUERY = "AccountCQuery"
	// 商户账户查询
	FUNC_ACCOUNT_M_QUERY = "AccountMQuery"
	// 交易记录查询
	FUNC_TRANSACTION_QUERY = "TransationInfoQuery"
	// 交易记录查询
	FUNC_TRANSACTION_RANGE_QUERY = "TransationInfoRangeQuery"
	// 客户购买商品
	FUNC_COMMODITY_TRADE = "CommodityTrade"
	// 记录交易
	FUNC_TRANSACTION_SET = "TransationInfoSet"
	FUNC_SYCOIN_ISSUE    = "SyCoinIssue"
	// 交易类型()
	TRANS_TYPE_BU = "BU"
	TRANS_TYPE_EX = "EX"
	TRANS_TYPE_GC = "GC"
	TRANS_TYPE_GM = "GM"
	TRANS_TYPE_SI = "SI"
	// 账户类型
	ACCOUNT_TYPE_C = "C"
	ACCOUNT_TYPE_M = "M"
	// 错误信息
	ERR_CUSTOMER_ACCOUNT_NOT_EXIST   = "101"
	ERR_CUSTOMER_INSUFFICIENT_SYCOIN = "102"
	ERR_MERCHANT_ACCOUNT_NOT_EXIST   = "201"
	ERR_MERCHANT_INSUFFICIENT_SYCOIN = "202"
	ERR_SYS_ERR                      = "301"
)

//============================================================================================================
//	 Function ExchangePointCalc 计算积分兑换苏银币
//============================================================================================================
//参数列表
//====================================
//Input
//01 交易积分数:point
//====================================
//Output
//01 苏银币金额：iSyCoinRet
//02 人民币金额：iRMBRet
//03 error
//====================================
//============================================================================================================
func (t *Chaincode) ExchangePointCalc(stub shim.ChaincodeStubInterface, point int64) (int64, int64, error) {
	var errMsg string

	if point == 0 {
		return 0, 0, nil
	}

	getState_Byte, err := stub.GetState(JSBANK_SYCOIN_EXCHANGE_RATE)
	if getState_Byte == nil {
		errMsg = fmt.Sprintf("Fail to get Exchange Rate of KEY:%s", JSBANK_SYCOIN_EXCHANGE_RATE)
		return 0, 0, errors.New(errMsg)
	}
	jsBankExchangeRate := JSBankExchangeRate{}
	err = json.Unmarshal(getState_Byte, &jsBankExchangeRate)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Exchange Rate format.%s", string(getState_Byte))
		return 0, 0, errors.New(errMsg)
	}
	fExPoint := float64(point)
	fSyCoin := float64(jsBankExchangeRate.SyCoin)
	fPoint := float64(jsBankExchangeRate.Point)
	fRMB := float64(jsBankExchangeRate.RMB)
	calcSyCoinRet := fExPoint * fSyCoin / fPoint
	iSyCoinRet := int64(calcSyCoinRet)
	calcRMBRet := fExPoint * fRMB / fPoint
	iRMBRet := int64(calcRMBRet)
	logger.Debugf("[ExchangePointCalc]EXCHANGE CALC.Input Point=%d,RATE POINT=%d,RATE SYCOIN=%d,RATE RMB=%d,CALC SYCOIN RESULT=%d,CALC RMB RESULT=%d\n", point, jsBankExchangeRate.Point, jsBankExchangeRate.SyCoin, jsBankExchangeRate.RMB, calcSyCoinRet, calcRMBRet)
	return iSyCoinRet, iRMBRet, nil
}

//============================================================================================================
//	 Function GetCustomerAccount 获取客户账户Utility方法
//============================================================================================================
//参数列表
//====================================
//Input
//01 客户号（本行ECIF号）:id
//====================================
//Output
//01 客户账本：JSBankCustomerSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) GetCustomerAccount(stub shim.ChaincodeStubInterface, id string) (*JSBankCustomerSYCoin, error) {

	key := JSBANK_SYCOIN_C + id
	errMsg := ""

	logger.Debugf("[GetCustomerAccount]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}

	logger.Debugf("[GetCustomerAccount]Query Result for KEY:%s,VALUE=%s\n", key, string(getState_Byte))

	if getState_Byte == nil {
		errMsg = fmt.Sprintf("Query Result is Nil For Id:%s", id)
		return nil, errors.New(errMsg)
	}
	jsbankCustomerSyCoin := JSBankCustomerSYCoin{}
	err = json.Unmarshal(getState_Byte, &jsbankCustomerSyCoin)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Customer Account format.%s", string(getState_Byte))
		return nil, errors.New(errMsg)
	}
	return &jsbankCustomerSyCoin, nil
}

//============================================================================================================
//	 Function GetMerchantAccount 获取商户账户Utility方法
//============================================================================================================
//参数列表
//====================================
//Input
//01 商户号:id
//====================================
//Output
//01 商户账户：JSBankMerchantSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) GetMerchantAccount(stub shim.ChaincodeStubInterface, id string) (*JSBankMerchantSYCoin, error) {

	key := JSBANK_SYCOIN_M + id
	errMsg := ""

	logger.Debugf("[GetMerchantAccount]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}

	logger.Debugf("[GetMerchantAccount]Query Result for KEY:%s,VALUE=%s\n", key, string(getState_Byte))

	if getState_Byte == nil {
		errMsg = fmt.Sprintf("Query Result is Nil For Id:%s", id)
		return nil, errors.New(errMsg)
	}
	jsbankMerchantSYCoin := JSBankMerchantSYCoin{}
	err = json.Unmarshal(getState_Byte, &jsbankMerchantSYCoin)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Customer Account format.%s", string(getState_Byte))
		return nil, errors.New(errMsg)
	}
	return &jsbankMerchantSYCoin, nil
}

//============================================================================================================
//	 Function BeforZeroEdit 入出金账号补零方法
//============================================================================================================
//参数列表
//====================================
//Input
//01 账号:inputData
//02 账号总长：size
//====================================
//Output
//01 修正后账号：returnStr
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) BeforZeroEdit(stub shim.ChaincodeStubInterface, inputData string, size int) (string, error) {
	var baseStr string = "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMM"
	var baseByte []byte = []byte(baseStr)
	var returnStr string = ""
	numString := []byte(inputData)
	numlen := len(numString)
	if numlen < size {
		returnStr = string(append(baseByte[:size-numlen], []byte(numString)...))
	} else {
		returnStr = string([]byte(numString)[numlen-size:])
	}
	return returnStr, nil
}

//============================================================================================================
//	 Function Init
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	return nil, nil
}

//============================================================================================================
//	 Function Invoke Invoke路由函数
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	// 初期化账本
	case FUNC_INITIALIZE:
		return t.Initialize(stub, function, args)
	// 设置苏银币兑换比率
	case FUNC_SET_EXCHANGE_RATE:
		return t.SetExchangeRate(stub, function, args)
	// 注册客户账户
	case FUNC_ACCOUNT_C_REG:
		return t.AccountCReg(stub, function, args)
	// 注册商户账户
	case FUNC_ACCOUNT_M_REG:
		return t.AccountMReg(stub, function, args)
	// 积分兑换
	case FUNC_POINT_EXCHANGE:
		return t.ExchangePoint(stub, function, args)
	// 赠与
	case FUNC_SYCOIN_GIFT:
		return t.Gift(stub, function, args)
	// 客户购买商品
	case FUNC_COMMODITY_TRADE:
		return t.CommodityTrade(stub, function, args)
	//银行苏银币发行
	case FUNC_SYCOIN_ISSUE:
		return t.SYCoinIssue(stub, function, args)
	}
	return nil, errors.New("Invalid Function Call:" + function)
}

//============================================================================================================
//	 Function Initialize 初期化账本
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 银行总积分数
//		02 银行总苏银币数
//		03 兑换比率（积分数)
//		04 兑换比率（苏银币)
//		05 兑换比率（RMB)
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) Initialize(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	// 错误信息
	var errMsg string
	if len(args) != 5 {
		logger.Debugf("[Initialize]参数错误。预期5个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("Initialize function expects 2 arguments")
	}
	// 银行总积分数
	totalPoint, err := strconv.ParseInt(args[0], 10, 64) //strconv.Atoi(args[0])
	if err != nil {
		logger.Debugf("[Initialize]参数错误。积分总量应为数字,TotalPoint=%s。参数列表=%v", args[0], args)
		errMsg = fmt.Sprintf("Invalid argument, args[0]:%s", args[0])
		return nil, errors.New(errMsg)
	}
	// 银行总苏银币数
	totalSyCoin, err := strconv.ParseInt(args[1], 10, 64) //strconv.Atoi(args[1])
	if err != nil {
		logger.Debugf("[Initialize]参数错误。苏银币总量应为数字,TotalSyCoin=%s。参数列表=%v", args[1], args)
		errMsg = fmt.Sprintf("Invalid argument, args[1]:%s", args[1])
		return nil, errors.New(errMsg)
	}
	point, err := strconv.ParseInt(args[2], 10, 64) //strconv.Atoi(args[2])
	if err != nil {
		logger.Debugf("[Initialize]参数错误。兑换比率（积分数)应为数字,Point=%s。参数列表=%v", args[2], args)
		errMsg = fmt.Sprintf("Invalid argument, args[0]:", args[0])
		return nil, errors.New(errMsg)
	}
	sycoin, err := strconv.ParseInt(args[3], 10, 64) //strconv.Atoi(args[3])
	if err != nil {
		logger.Debugf("[Initialize]参数错误。兑换比率（苏银币)应为数字,SyCoin=%s。参数列表=%v", args[3], args)
		errMsg = fmt.Sprintf("Invalid argument, args[1]:", args[1])
		return nil, errors.New(errMsg)
	}
	rmb, err := strconv.ParseInt(args[4], 10, 64) //strconv.Atoi(args[4])
	if err != nil {
		logger.Debugf("[Initialize]参数错误。兑换比率（rmb)应为数字,RMB=%s。参数列表=%v", args[4], args)
		errMsg = fmt.Sprintf("Invalid argument, args[1]:", args[1])
		return nil, errors.New(errMsg)
	}
	jsBankSyCoin := JSBankSYCoin{
		TotalPoint:  totalPoint,
		TotalSyCoin: totalSyCoin}
	putState_Byte, _ := json.Marshal(&jsBankSyCoin)
	key := JSBANK_SYCOIN

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}

	jsBankExchangeRate := JSBankExchangeRate{
		Point:  point,
		SyCoin: sycoin,
		RMB:    rmb}
	putState_Byte, _ = json.Marshal(&jsBankExchangeRate)
	key = JSBANK_SYCOIN_EXCHANGE_RATE

	logger.Debugf("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

//============================================================================================================
//	 Function SetExchangeRate 设置苏银币兑换比率
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 兑换比率（积分数)
//		02 兑换比率（苏银币)
//		03 兑换比率（RMB)
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) SetExchangeRate(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var errMsg string
	if len(args) != 3 {
		logger.Debugf("[SetExchangeRate]参数错误。预期3个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("SetExchangeRate function expects 3 arguments")
	}
	point, err := strconv.ParseInt(args[0], 10, 64) //strconv.Atoi(args[0])
	if err != nil {
		logger.Debugf("[SetExchangeRate]参数错误。积分数应为数字,Point=%s。参数列表=%v", args[0], args)
		errMsg = fmt.Sprintf("Invalid argument, args[0]:", args[0])
		return nil, errors.New(errMsg)
	}
	sycoin, err := strconv.ParseInt(args[1], 10, 64) //strconv.Atoi(args[2])
	if err != nil {
		logger.Debugf("[SetExchangeRate]参数错误。苏银币应为数字,SyCoin=%s。参数列表=%v", args[1], args)
		errMsg = fmt.Sprintf("Invalid argument, args[1]:", args[1])
		return nil, errors.New(errMsg)
	}
	rmb, err := strconv.ParseInt(args[2], 10, 64) //strconv.Atoi(args[3])
	if err != nil {
		logger.Debugf("[SetExchangeRate]参数错误。人民币应为数字,RMB=%s。参数列表=%v", args[2], args)
		errMsg = fmt.Sprintf("Invalid argument, args[2]:%s", args[2])
		return nil, errors.New(errMsg)
	}
	jsBankExchangeRate := JSBankExchangeRate{
		Point:  point,
		SyCoin: sycoin,
		RMB:    rmb}
	putState_Byte, _ := json.Marshal(&jsBankExchangeRate)
	key := JSBANK_SYCOIN_EXCHANGE_RATE

	logger.Debugf("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return putState_Byte, nil
}

//============================================================================================================
//	 Function QueryExchangeRate 获取苏银币兑换比率
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args(无)
//====================================
//Output
//01 兑换比率账本:JSBankExchangeRate
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) QueryExchangeRate(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	key := JSBANK_SYCOIN_EXCHANGE_RATE
	bytes, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	/*rate := JSBankExchangeRate{
		Point:  10000,
		SyCoin: 15,
		RMB:    10}
	str, _ := json.Marshal(&rate)*/
	return bytes, nil
}

//============================================================================================================
//	 Function AccountCReg 注册客户账户
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 客户号（本行ECIF号）
//====================================
//Output
//01 客户账本:JSBankCustomerSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) AccountCReg(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		logger.Debugf("[AccountCReg]参数错误。预期1个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("AccountCReg function expects 1 argument")
	}
	ecif := args[0]
	jsBankCustomerSYCoin := JSBankCustomerSYCoin{
		ECIF:   ecif,
		Amount: 0,
		Avail:  0,
		Frozen: 0}

	key := JSBANK_SYCOIN_C + ecif
	putState_Byte, _ := json.Marshal(&jsBankCustomerSYCoin)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err := stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return putState_Byte, nil
}

//============================================================================================================
//	 Function AccountMReg 注册商户账户
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 商户号
//		02 商户苏银币余额
//====================================
//Output
//01 商户账本:JSBankMerchantSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) AccountMReg(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 商户苏银币数
	var amount int64
	// 错误信息
	var errMsg string
	if len(args) != 2 {
		logger.Debugf("[AccountMReg]参数错误。预期2个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("AccountCReg function expects 2 argument")
	}
	id := args[0]
	amount, err := strconv.ParseInt(args[1], 10, 64) //strconv.Atoi(args[1])
	if err != nil {
		logger.Debugf("[AccountMReg]参数错误。苏银币数额应为数字,AMOUNT=%s。参数列表=%v", args[1], args)
		errMsg = fmt.Sprintf("Invalid argument,Amount should be number. args[1]:%s", args[1])
		return nil, errors.New(errMsg)
	}
	jsBankCustomerSYCoin := JSBankMerchantSYCoin{
		Id:     id,
		Point:  0,
		Amount: amount}

	key := JSBANK_SYCOIN_M + id
	putState_Byte, _ := json.Marshal(&jsBankCustomerSYCoin)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	return putState_Byte, nil
}

//============================================================================================================
//	 Function ExchangePoint 兑换积分
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易流水号
//		02 客户ECIF号
//		03 交易积分数
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) ExchangePoint(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 交易流水号
	var serialNumber string
	// 客户ECIF号
	var ecif string
	// 交易积分数
	var point int64
	// 错误信息
	var errMsg string
	if len(args) != 3 {
		logger.Debugf("[ExchangePoint]参数错误。预期3个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("ExchangePoint function expects 3 argument")
	}
	serialNumber = args[0]
	ecif = args[1]
	point, err := strconv.ParseInt(args[2], 10, 64) //strconv.Atoi(args[2])
	if err != nil {
		logger.Debugf("[ExchangePoint]参数错误。交易积分数额应为数字,POINT=%s。参数列表=%v", args[2], args)
		errMsg = fmt.Sprintf("Invalid argument,Amount should be number. args[2]:%s", args[2])
		return nil, errors.New(errMsg)
	}
	now := time.Now()
	date := now.Format("20060102")
	// 客户账本
	jsBankCustomerSYCoin, err := t.GetCustomerAccount(stub, ecif)
	if err != nil {
		errMsg = fmt.Sprintf("Customer Account not exists. ECIF=%s", ecif)
		return nil, errors.New(errMsg)
	}

	// 银行账本
	bankKey := JSBANK_SYCOIN
	getState_Byte, err := stub.GetState(bankKey)
	if err != nil {
		errMsg = fmt.Sprintf("Cannot get Bank Account.KEY=%s", bankKey)
		return nil, errors.New(errMsg)
	}
	bankAccountStr := string(getState_Byte)
	jsBankSyCoin := JSBankSYCoin{}
	err = json.Unmarshal(getState_Byte, &jsBankSyCoin)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Bank Account Format. DATA=%s", bankAccountStr)
		return nil, errors.New(errMsg)
	}

	// 计算苏银币兑换（注意是Int型）
	syCoin, _, _ := t.ExchangePointCalc(stub, point)
	// syCoin = 1000
	// 计入客户账本
	accountKey := JSBANK_SYCOIN_C + ecif
	jsBankCustomerSYCoin.Amount += syCoin
	jsBankCustomerSYCoin.Avail += syCoin
	custAccountRenewStr, _ := json.Marshal(&jsBankCustomerSYCoin)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", accountKey, string(custAccountRenewStr))
	err = stub.PutState(accountKey, custAccountRenewStr)
	if err != nil {
		return nil, err
	}

	// 记入银行账本
	jsBankSyCoin.TotalPoint += point
	jsBankSyCoin.TotalSyCoin -= syCoin
	jsBankBytes, _ := json.Marshal(&jsBankSyCoin)
	stub.PutState(JSBANK_SYCOIN, jsBankBytes)

	// 记入交易记录账本
	var paras []string
	// 1流水号
	paras = append(paras, serialNumber)
	// 2出金账户ID
	paras = append(paras, JSBANK_SYCOIN)
	// 3入金账户ID
	paras = append(paras, ecif)
	// 4交易类型
	paras = append(paras, TRANS_TYPE_EX)
	// 5交易苏银币金额
	paras = append(paras, strconv.FormatInt(syCoin, 10)) //strconv.Itoa(syCoin))
	//6 出金账户余额
	paras = append(paras, strconv.FormatInt(jsBankSyCoin.TotalSyCoin, 10)) //strconv.Itoa(jsBankSyCoin.TotalSyCoin))
	// 7入金账户余额
	paras = append(paras, strconv.FormatInt(jsBankCustomerSYCoin.Amount, 10)) //strconv.Itoa(jsBankCustomerSYCoin.Amount))
	// 8交易日期
	paras = append(paras, date)
	// 9商品编码
	paras = append(paras, "")
	// 10摘要
	paras = append(paras, "")

	_, err = t.TransationInfoSet(stub, paras)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

//============================================================================================================
//	 Function Gift 赠送积分
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易流水号
//		02 转出客户号
//		03 转出客户类型
//		04 转入客户号
//		05 转入客户类型
//		06 交易苏银币金额
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) Gift(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 交易流水号
	var serialNumber string
	// 转出客户号，转出客户类型，转入客户号，转入客户类型，
	var fromAccountId, fromAccountType, toAccountId, toAccountType string
	// 交易苏银币数
	var sycoin int64
	// 错误信息
	var errMsg string
	if len(args) != 6 {
		logger.Debugf("[Gift]参数错误。预期6个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("ExchangePoint function expects 6 argument")
	}
	serialNumber = args[0]
	fromAccountId = args[1]
	fromAccountType = args[2]
	toAccountId = args[3]
	toAccountType = args[4]
	sycoinStr := args[5]
	if fromAccountType != ACCOUNT_TYPE_C && fromAccountType != ACCOUNT_TYPE_M {
		logger.Debugf("[Gift]参数错误。转入客户类型必须是C或者M，参数列表=%v", args)
		errMsg = fmt.Sprintf("Invalid Parameter.FromAccountType or ToAccountType must be C or M. args=%v", args)
		return nil, errors.New(errMsg)
	}
	if toAccountType != ACCOUNT_TYPE_C && toAccountType != ACCOUNT_TYPE_M {
		logger.Debugf("[Gift]参数错误。转出客户类型必须是C或者M，参数列表=%v", args)
		errMsg = fmt.Sprintf("Invalid Parameter.FromAccountType or ToAccountType must be C or M. args=%v", args)
		return nil, errors.New(errMsg)
	}
	// 转换苏银币为整形
	sycoin, err := strconv.ParseInt(sycoinStr, 10, 64) //strconv.Atoi(sycoinStr)
	if err != nil {
		logger.Debugf("[Gift]参数错误。交易苏银币数额应为数字,SYCOIN=%s。参数列表=%v", sycoinStr, args)
		errMsg = fmt.Sprintf("Invalid argument,SYCOIN should be number. args[5]:%s", sycoinStr)
		return nil, errors.New(errMsg)
	}

	var getState_Byte []byte
	var fromCustomerSYCoin *JSBankCustomerSYCoin
	var toCustomerSYCoin *JSBankCustomerSYCoin
	var fromMerchantSYCoin *JSBankMerchantSYCoin
	var toMerchantSYCoin *JSBankMerchantSYCoin
	var fromTotalSyCoin, fromAvailSyCoin, toTotalSyCoin, toAvailSyCoin int64

	fromCustomerSYCoin = nil
	toCustomerSYCoin = nil
	fromMerchantSYCoin = nil
	toMerchantSYCoin = nil

	// 获取日期
	now := time.Now()
	date := now.Format("20060102")
	// 获取账本
	if fromAccountType == ACCOUNT_TYPE_C {
		fromCustomerSYCoin, err = t.GetCustomerAccount(stub, fromAccountId)
		if err != nil {
			return nil, err
		}
		fromTotalSyCoin = fromCustomerSYCoin.Amount
		fromAvailSyCoin = fromCustomerSYCoin.Avail
	} else {
		fromMerchantSYCoin, err = t.GetMerchantAccount(stub, fromAccountId)
		if err != nil {
			return nil, err
		}
		fromTotalSyCoin = fromMerchantSYCoin.Amount
		fromAvailSyCoin = fromMerchantSYCoin.Amount
	}

	if toAccountType == ACCOUNT_TYPE_C {
		toCustomerSYCoin, err = t.GetCustomerAccount(stub, toAccountId)
		if err != nil {
			return nil, err
		}
		toTotalSyCoin = toCustomerSYCoin.Amount
		toAvailSyCoin = toCustomerSYCoin.Avail
	} else {
		toMerchantSYCoin, err = t.GetMerchantAccount(stub, toAccountId)
		if err != nil {
			return nil, err
		}
		toTotalSyCoin = toMerchantSYCoin.Amount
		toAvailSyCoin = toMerchantSYCoin.Amount
	}

	// 计算余额是否可用
	if fromAvailSyCoin < sycoin || fromTotalSyCoin < sycoin {
		errMsg = fmt.Sprintf("Insuffcient amount for from account.from id=%s,amount=%d,avail=%d,transfer sycoin=%d", fromAccountId, fromTotalSyCoin, fromAvailSyCoin, sycoin)
		return nil, errors.New(errMsg)
	}

	fromTotalSyCoin -= sycoin
	fromAvailSyCoin -= sycoin
	toTotalSyCoin += sycoin
	toAvailSyCoin += sycoin

	if fromAccountType == ACCOUNT_TYPE_C {
		fromCustomerSYCoin.Amount = fromTotalSyCoin
		fromCustomerSYCoin.Avail = fromAvailSyCoin
		getState_Byte, _ = json.Marshal(&fromCustomerSYCoin)
		err = stub.PutState(JSBANK_SYCOIN_C+fromAccountId, getState_Byte)
		if err != nil {
			return nil, err
		}
	} else {
		fromMerchantSYCoin.Amount = fromTotalSyCoin
		getState_Byte, _ = json.Marshal(&fromMerchantSYCoin)
		err = stub.PutState(JSBANK_SYCOIN_M+fromAccountId, getState_Byte)
		if err != nil {
			return nil, err
		}
	}
	if toAccountType == ACCOUNT_TYPE_C {
		toCustomerSYCoin.Amount = toTotalSyCoin
		toCustomerSYCoin.Avail = toAvailSyCoin
		getState_Byte, _ = json.Marshal(&toCustomerSYCoin)
		err = stub.PutState(JSBANK_SYCOIN_C+toAccountId, getState_Byte)
		if err != nil {
			return nil, err
		}
	} else {
		toMerchantSYCoin.Amount = toTotalSyCoin
		getState_Byte, _ = json.Marshal(&toMerchantSYCoin)
		err = stub.PutState(JSBANK_SYCOIN_M+toAccountId, getState_Byte)
		if err != nil {
			return nil, err
		}
	}

	// 记入交易记录账本
	var paras []string
	// 1流水号
	paras = append(paras, serialNumber)
	// 2出金账户ID
	paras = append(paras, fromAccountId)
	// 3入金账户ID
	paras = append(paras, toAccountId)
	// 4交易类型
	if toAccountType == ACCOUNT_TYPE_C {
		paras = append(paras, TRANS_TYPE_GC)
	} else {
		paras = append(paras, TRANS_TYPE_GM)
	}
	// 5交易苏银币金额
	paras = append(paras, strconv.FormatInt(sycoin, 10)) //strconv.Itoa(sycoin))
	//6 出金账户余额
	paras = append(paras, strconv.FormatInt(fromTotalSyCoin, 10)) //strconv.Itoa(fromTotalSyCoin))
	// 7入金账户余额
	paras = append(paras, strconv.FormatInt(toTotalSyCoin, 10)) //strconv.Itoa(toTotalSyCoin))
	// 8交易日期
	paras = append(paras, date)
	// 9商品编码
	paras = append(paras, "")
	// 10摘要
	paras = append(paras, "")

	_, err = t.TransationInfoSet(stub, paras)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

//============================================================================================================
//	 Function CommodityTrade 客户购买商品
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易流水号
//		02 客户ECIF号
//		03 交易苏银币金额
//		04 商户号
//		05 商品编码
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) CommodityTrade(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 交易流水号
	var serialNumber string
	// 客户ECIF号
	var ecif string
	// 交易苏银币金额
	var sycoin int64
	// 商户号
	var merchantID string
	// 商品编码
	var productID string
	// 错误信息
	var errMsg string

	if len(args) != 5 {
		logger.Debugf("[CommodityTrade]参数错误。预期5个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("CommodityTrade function expects 5 argument")
	}

	serialNumber = args[0]
	ecif = args[1]
	sycoin, err := strconv.ParseInt(args[2], 10, 64) //strconv.Atoi(args[2])
	if err != nil {
		logger.Debugf("[ExchangePoint]参数错误。交易苏银币数应为数字,SyCoin=%s。参数列表=%v", args[2], args)
		errMsg = fmt.Sprintf("Invalid argument,Amount should be number. args[2]:%s", args[2])
		return nil, errors.New(errMsg)
	}
	merchantID = args[3]
	productID = args[4]

	// 客户账本
	jsBankCustomerSYCoin, err := t.GetCustomerAccount(stub, ecif)
	if err != nil {
		errMsg = fmt.Sprintf("Customer Account not exists. ECIF=%s", ecif)
		return nil, errors.New(errMsg)
	}

	// 商户账本
	jsBankMerchantSYCoin, err := t.GetMerchantAccount(stub, merchantID)
	if err != nil {
		errMsg = fmt.Sprintf("Merchant Account not exists. merchantID=%s", merchantID)
		return nil, errors.New(errMsg)
	}

	// 计入客户账本
	accountKey := JSBANK_SYCOIN_C + ecif
	jsBankCustomerSYCoin.Amount -= sycoin
	jsBankCustomerSYCoin.Avail -= sycoin
	custAccountRenewStr, _ := json.Marshal(&jsBankCustomerSYCoin)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", accountKey, string(custAccountRenewStr))
	err = stub.PutState(accountKey, custAccountRenewStr)
	if err != nil {
		return nil, err
	}

	// 计入商户账本
	accountKey = JSBANK_SYCOIN_M + merchantID
	jsBankMerchantSYCoin.Amount += sycoin
	mercAccountRenewStr, _ := json.Marshal(&jsBankMerchantSYCoin)

	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", accountKey, string(mercAccountRenewStr))
	err = stub.PutState(accountKey, mercAccountRenewStr)
	if err != nil {
		return nil, err
	}

	// 记入交易记录账本
	var paras []string
	// 1流水号
	paras = append(paras, serialNumber)
	// 2出金账户ID
	paras = append(paras, ecif)
	// 3入金账户ID
	paras = append(paras, merchantID)
	// 4交易类型
	paras = append(paras, TRANS_TYPE_BU)
	// 5交易苏银币金额
	paras = append(paras, strconv.FormatInt(sycoin, 10)) //strconv.Itoa(sycoin))
	// 6出金账户余额
	paras = append(paras, strconv.FormatInt(jsBankCustomerSYCoin.Amount, 10)) //strconv.Itoa(jsBankCustomerSYCoin.Amount))
	// 7入金账户余额
	paras = append(paras, strconv.FormatInt(jsBankMerchantSYCoin.Amount, 10)) //strconv.Itoa(jsBankMerchantSYCoin.Amount))
	// 8交易日期
	now := time.Now()
	date := now.Format("20060102")
	paras = append(paras, date)
	// 9商品编码
	paras = append(paras, productID)
	// 10摘要
	paras = append(paras, "")

	_, err = t.TransationInfoSet(stub, paras)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//============================================================================================================
//	 Function TransationInfoSet 记录交易
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易流水号
//		02 出金账户ID
//		03 入金账户ID
//		04 交易类型
//		05 交易苏银币金额
//		06 出金账户余额
//		07 入金账户余额
//		08 交易日期
//		09 商品编码
//		10 摘要
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) TransationInfoSet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// 交易KEY前30位
	var headID string
	// 1流水号
	var serialNumber string
	// 2出金账户ID
	var cashOutAccount string
	// 3入金账户ID
	var cashInAccount string
	// 4交易类型
	var transactionType string
	// 5交易苏银币金额
	var transactionAmount int64
	//6 出金账户余额
	var cashOutBalance int64
	// 7入金账户余额
	var cashInBalance int64
	// 8交易日期
	var transactionDate string
	// 9商品编码
	var procuctID string
	// 10摘要
	var memo string
	// 错误信息
	var err error
	var errMsg string
	// 交易账本
	var transactionInfo TransactionInfo

	if len(args) != 10 {
		logger.Debugf("[TransationInfoSet]参数错误。预期10个参数，实际%d个参数。参数列表=%v\r\n", len(args), args)
		return nil, errors.New("TransationInfoSet function expects 10 argument")
	}
	serialNumber = args[0]
	cashOutAccount = args[1]
	cashInAccount = args[2]
	transactionType = args[3]
	transactionAmount, err = strconv.ParseInt(args[4], 10, 64) //strconv.Atoi(args[4])
	if err != nil {
		errMsg = fmt.Sprintf("Invalid argument,transactionAmount should be number. args[4]:%s", args[4])
		return nil, errors.New(errMsg)
	}
	cashOutBalance, err = strconv.ParseInt(args[5], 10, 64) //strconv.Atoi(args[5])
	if err != nil {
		errMsg = fmt.Sprintf("Invalid argument,cashOutBalance should be number. args[5]:%s", args[5])
		return nil, errors.New(errMsg)
	}
	cashInBalance, err = strconv.ParseInt(args[6], 10, 64) //strconv.Atoi(args[6])
	if err != nil {
		errMsg = fmt.Sprintf("Invalid argument,cashInBalance should be number. args[6]:%s", args[6])
		return nil, errors.New(errMsg)
	}
	transactionDate = args[7]
	procuctID = args[8]
	memo = args[9]

	// 账本信息(入金记录）记录
	transactionInfo.SerialNumber = serialNumber
	transactionInfo.CashOutAccount = cashOutAccount
	transactionInfo.CashInAccount = cashInAccount
	transactionInfo.TransactionType = transactionType
	transactionInfo.TransactionAmount = transactionAmount
	transactionInfo.CashOutBalance = cashOutBalance
	transactionInfo.CashInBalance = cashInBalance
	transactionInfo.TransactionDate = transactionDate
	transactionInfo.ProcuctID = procuctID
	transactionInfo.Memo = memo

	jsonRespByte, err := json.Marshal(&transactionInfo)

	headID, err = t.BeforZeroEdit(stub, cashInAccount, 30)
	SerKey := JSBANK_TRADE + headID + serialNumber
	stub.PutState(SerKey, jsonRespByte)

	// 账本信息(出金记录）记录
	transactionInfo.SerialNumber = serialNumber
	transactionInfo.CashOutAccount = cashInAccount
	transactionInfo.CashInAccount = cashOutAccount
	transactionInfo.TransactionType = transactionType
	transactionInfo.TransactionAmount = transactionAmount * -1
	transactionInfo.CashOutBalance = cashInBalance
	transactionInfo.CashInBalance = cashOutBalance
	transactionInfo.TransactionDate = transactionDate
	transactionInfo.ProcuctID = procuctID
	transactionInfo.Memo = memo

	jsonRespByte, err = json.Marshal(&transactionInfo)

	headID, err = t.BeforZeroEdit(stub, cashOutAccount, 30)
	SerKey = JSBANK_TRADE + headID + serialNumber
	stub.PutState(SerKey, jsonRespByte)

	return nil, nil
}

//============================================================================================================
//	 Function Query Query路由函数
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//====================================
//Output
//01 []byte
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	// 查询兑换比率
	case FUNC_QUERY_EXCHANGE_RATE:
		return t.QueryExchangeRate(stub, function, args)
	// 银行账户查询
	case FUNC_ACCOUNT_B_QUERY:
		return t.AccountBQuery(stub, function, args)
	// 客户账户查询
	case FUNC_ACCOUNT_C_QUERY:
		return t.AccountCQuery(stub, function, args)
	// 商户账户查询
	case FUNC_ACCOUNT_M_QUERY:
		return t.AccountMQuery(stub, function, args)
	// 积分兑换查询
	case FUNC_POINT_EXCHANGE_QUERY:
		return t.PointExchangeQuery(stub, function, args)
	// 交易记录查询
	case FUNC_TRANSACTION_QUERY:
		return t.TransationInfoQuery(stub, function, args)
	// 指定条件查询
	case FUNC_TRANSACTION_RANGE_QUERY:
		return t.TransationInfoRangeQuery(stub, function, args)

	}
	return nil, errors.New("Invalid Function Call:" + function)
}

//============================================================================================================
//	 Function PointExchangeQuery 积分兑换苏银币查询
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易积分数
//====================================
//Output
//01 兑换比率账本:JSBankExchangeRate
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) PointExchangeQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var errMsg string
	if len(args) != 1 {
		logger.Debugf("[PointExchangeQuery Function]参数错误。预期1个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("PointExgSyCoin function expects 1 arguments")
	}
	pointStr := args[0]
	point, err := strconv.ParseInt(pointStr, 10, 64) //strconv.Atoi(pointStr)
	if err != nil {
		logger.Debugf("[PointExchangeQuery]参数错误。交易积分数额应为数字,POINT=%s。参数列表=%v", args[0], args)
		errMsg = fmt.Sprintf("Invalid argument,Amount should be number. args[0]:%s", args[0])
		return nil, errors.New(errMsg)
	}
	sycoin, rmb, _ := t.ExchangePointCalc(stub, point)

	ret := JSBankExchangeRate{
		Point:  point,
		SyCoin: sycoin,
		RMB:    rmb}
	res, _ := json.Marshal(&ret)
	return res, nil
}

//============================================================================================================
//	 Function AccountBQuery 查询银行全局账户
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//====================================
//Output
//01 银行账本:JSBankSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) AccountBQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	key := JSBANK_SYCOIN
	errMsg := ""

	logger.Debugf("[AccountBQuery]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}
	getState_String := string(getState_Byte)
	logger.Debugf("[AccountBQuery]Query Result for KEY:%s,VALUE=%s\n", key, getState_String)

	if getState_String == "" {
		errMsg = fmt.Sprintf("Query Result is Nil For Bank")
		return nil, errors.New(errMsg)
	}
	return getState_Byte, nil
}

//============================================================================================================
//	 Function AccountCQuery 查询客户账户
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 客户号（本行ECIF号）
//====================================
//Output
//01 客户账本:JSBankCustomerSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) AccountCQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		logger.Debugf("[AccountCQuery Function]参数错误。预期1个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("AccountCQuery function expects 1 arguments")
	}
	accountId := args[0]
	key := JSBANK_SYCOIN_C + accountId
	errMsg := ""

	logger.Debugf("[AccountCQuery]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}
	getState_String := string(getState_Byte)
	logger.Debugf("[AccountCQuery]Query Result for KEY:%s,VALUE=%s\n", key, getState_String)

	if getState_String == "" {
		errMsg = fmt.Sprintf("Query Result is Nil For Id:%s", accountId)
		return nil, errors.New(errMsg)
	}
	return []byte(getState_String), nil
}

//============================================================================================================
//	 Function AccountMQuery 查询商户账户
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 商户号
//====================================
//Output
//01 商户账本:JSBankMerchantSYCoin
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) AccountMQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		logger.Debugf("[AccountMQuery Function]参数错误。预期1个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("AccountMQuery function expects 1 arguments")
	}
	accountId := args[0]
	key := JSBANK_SYCOIN_M + accountId
	errMsg := ""

	logger.Debugf("[AccountMQuery]Query Stub for KEY:%s\n", key)
	getState_Byte, err := stub.GetState(key)
	if err != nil {
		errMsg = fmt.Sprintf("Fail to get value of KEY:%s", key)
		return nil, errors.New(errMsg)
	}
	getState_String := string(getState_Byte)
	logger.Debugf("[AccountMQuery]Query Result for KEY:%s,VALUE=%s\n", key, getState_String)

	if getState_String == "" {
		errMsg = fmt.Sprintf("Query Result is Nil For Id:%s", accountId)
		return nil, errors.New(errMsg)
	}
	return []byte(getState_String), nil
}

// 查询交易记录信息
func (t *Chaincode) TransationInfoQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 账本
	var transInfo TransactionInfo
	// 返回账本结构
	var listTrans ListTrans
	errMsg := ""
	/*
		if len(args) != 1 {
			logger.Debugf("[TransationInfoQuery]参数错误。预期1个参数，实际%d个参数。\n", len(args))
			return nil, errors.New("TransationInfoQuery function expects 1 argument")
		}
	*/
	sliceArgs := args
	sliceList := make([]TransactionInfo, 0, len(sliceArgs))
	for _, transNum := range sliceArgs {
		// 交易账本KEY
		key := JSBANK_TRADE + transNum
		// 交易记录取得
		getTradeInfo_Byte, err := stub.GetState(key)
		if err != nil {
			errMsg = fmt.Sprintf("Fail to get info of KEY:%s", key)
			return nil, errors.New(errMsg)
		}
		json.Unmarshal(getTradeInfo_Byte, &transInfo)
		sliceList = append(sliceList, transInfo)
	}
	listTrans.ListTrade = sliceList
	listTrans.Error = ""

	jsonRespByte, err := json.Marshal(&listTrans)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Customer Transaction format.%s", string(jsonRespByte))
		return nil, errors.New(errMsg)
	}
	return jsonRespByte, nil
}

//============================================================================================================
//	 Function TransationInfoRangeQuery 查询交易记录信息
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 客户/商户号
//		02 开始时间
//		03 结束时间
//====================================
//Output
//01 交易记录查询返回结构:listTrans
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) TransationInfoRangeQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// 交易KEY前30位
	var headID string
	// 账本
	var transInfo TransactionInfo
	// 返回账本结构
	var listTrans ListTrans
	// 账本数组
	var transArray []TransactionInfo

	var accountId string
	var startDate string // YYYYMMDD
	var endDate string   // YYYYMMDD

	errMsg := ""
	if len(args) != 3 {
		logger.Debugf("[TransationInfoRangeQuery]参数错误。预期3个参数，实际%d个参数。\n", len(args))
		return nil, errors.New("TransationInfoRangeQuery function expects 3 argument")
	}
	accountId = args[0]
	startDate = args[1]
	endDate = args[2]

	headID, _ = t.BeforZeroEdit(stub, accountId, 30)
	startKey := JSBANK_TRADE + headID + startDate + "000000" + "000000000000" + "AA"
	endKey := JSBANK_TRADE + headID + endDate + "999999" + "999999999999" + "ZZ"

	iter, err := stub.RangeQueryState(startKey, endKey)
	if err != nil {
		return nil, err
	}
	for iter.HasNext() {
		_, value, err := iter.Next()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(value, &transInfo)
		if err != nil {
			return nil, err
		}
		transArray = append(transArray, transInfo)
	}
	listTrans.ListTrade = transArray
	listTrans.Error = errMsg

	bytes, _ := json.Marshal(&listTrans)

	return bytes, nil
}

//============================================================================================================
//	 Function SYCoinIssue 银行发行苏银币
//============================================================================================================
//参数列表
//====================================
//Input
//01 方法名:function
//02 参数：args
//		01 交易流水号
// 		02 发行量
//		03 交易日期
//====================================
//Output
//01 nil
//02 error
//====================================
//============================================================================================================
func (t *Chaincode) SYCoinIssue(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 3 {
		logger.Debugf("[SYCoinIssue Function]参数错误。预期2个参数，实际%d个参数。参数列表=%v", len(args), args)
		return nil, errors.New("SYCoinIssue function expects 2 arguments")
	}
	sycoinStr := args[1]
	// 转换苏银币为整形
	sycoin, err := strconv.ParseInt(sycoinStr, 10, 64) //strconv.Atoi(sycoinStr)

	errMsg := ""
	if err != nil {
		logger.Debugf("[Gift]参数错误。交易苏银币数额应为数字,SYCOIN=%s。参数列表=%v", sycoinStr, args)
		errMsg = fmt.Sprintf("Invalid argument,SYCOIN should be number. args[5]:%s", sycoinStr)
		return nil, errors.New(errMsg)
	}

	// 银行账本更新
	bankKey := JSBANK_SYCOIN
	getState_Byte, err := stub.GetState(bankKey)
	if err != nil {
		errMsg = fmt.Sprintf("Cannot get Bank Account.KEY=%s", bankKey)
		return nil, errors.New(errMsg)
	}
	bankAccountStr := string(getState_Byte)
	jsBankSyCoin := JSBankSYCoin{}
	err = json.Unmarshal(getState_Byte, &jsBankSyCoin)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid Bank Account Format. DATA=%s", bankAccountStr)
		return nil, errors.New(errMsg)
	}
	jsBankSyCoin.TotalSyCoin += sycoin
	putState_Byte, _ := json.Marshal(&jsBankSyCoin)
	key := JSBANK_SYCOIN
	logger.Infof("[WRITE LEDGER].KEY=[%s],VALUE=[%s]\n", key, string(putState_Byte))
	err = stub.PutState(key, putState_Byte)
	if err != nil {
		return nil, err
	}
	//交易账本更新
	// 交易KEY前30位
	var headID string
	// 1流水号
	var serialNumber string

	// 交易账本
	var transactionInfo TransactionInfo

	serialNumber = args[0]
	// 账本信息(入金记录）记录
	transactionInfo.SerialNumber = serialNumber
	transactionInfo.CashOutAccount = ""
	transactionInfo.CashInAccount = JSBANK_SYCOIN
	transactionInfo.TransactionType = "SI"
	transactionInfo.TransactionAmount = sycoin
	transactionInfo.CashOutBalance = 0
	transactionInfo.CashInBalance = jsBankSyCoin.TotalSyCoin
	transactionInfo.TransactionDate = args[2]
	transactionInfo.ProcuctID = ""
	transactionInfo.Memo = ""

	jsonRespByte, err := json.Marshal(&transactionInfo)

	headID, err = t.BeforZeroEdit(stub, transactionInfo.CashInAccount, 30)
	SerKey := JSBANK_TRADE + headID + serialNumber
	stub.PutState(SerKey, jsonRespByte)

	return nil, nil
}

// Log定义
var logger *shim.ChaincodeLogger = shim.NewLogger("SYCOIN")

//============================================================================================================
//	 Function main函数
//============================================================================================================
func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		panic(err)
	}
}
