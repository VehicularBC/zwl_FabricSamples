package main

import (
	"fmt"
	"time"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
	"math"
)
type SmartContract struct {
	contractapi.Contract
}


const (
	compositeIndexName = "carID~trustEvaluation~txID"
 	declineFactor = 0.5
 	timeSlot = 60 //一个时隙60s
 	a = 0.2
 	b = 0.3
 	c = 0.4
 	lamdaWeight = 0.5
)
type estimator struct {
	Id		string
	Slot	int64
}

//创建对某个id的信任值，信任值本应该由认证车在每个时隙中认证其他车辆的比例（认证贡献）来决定，此处使用设定一个贡献值来模拟，
// 复合键中包含了目标id，信任值，信任值创建时间戳，自身id
func (s *SmartContract) CreateTrustValue(ctx contractapi.TransactionContextInterface, id, value, myself string) (string, error){
	if v, _:= strconv.ParseFloat(value,64); v > 1.0 || v == 0.0 {
		return "", fmt.Errorf("认证贡献不能超过1或小于等于0")
	}
	ts,_ := ctx.GetStub().GetTxTimestamp()
	timeStamp := time.Unix(ts.Seconds,0)
	time_stamp := strconv.FormatInt(timeStamp.Unix(),10)
	attr := []string{id, value, time_stamp, myself}
	//attr := []string{attribute,id}
	compositeKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, attr)
	if err != nil{
		return "",fmt.Errorf("the composite key failed to be created")
	}
	err = ctx.GetStub().PutState(compositeKey,[]byte{0x00})
	if err != nil {
		return "", fmt.Errorf("the composite key failed to be write in world state")
	}
	return compositeKey, nil
}

func (s *SmartContract) GetLatestTrustValue(ctx contractapi.TransactionContextInterface, id string) (float64, error) {
	//var totalTrust float64
	//将所有的评价者的每个周期的值存入hash表，一个周期只统计第一次的评价值，后续评价值无效。
	// 通过遍历hash表来计算不同评价者不同周期内的评价值的总和。
	trustMp := make(map[estimator]float64)
	estimatorTrustMp := make(map[string]float64)
	deltaKeyIterator, deltaErr := ctx.GetStub().GetStateByPartialCompositeKey(compositeIndexName,[]string{id})
	if deltaErr != nil {
		return 0.0, fmt.Errorf("failed to create the ReadIterator of %s for %s",id, deltaErr.Error())
	}
	defer deltaKeyIterator.Close()
	if !deltaKeyIterator.HasNext() {
		return 0.0, fmt.Errorf("the key of %s does not exist", id)
	}
	for deltaKeyIterator.HasNext() {
		nextKey, nextErr := deltaKeyIterator.Next()
		if nextErr != nil {
			return 0.0, fmt.Errorf(nextErr.Error())
		}
		_, keyParts, splitKeyErr := ctx.GetStub().SplitCompositeKey(nextKey.Key)
		if splitKeyErr != nil {
			return 0.0, fmt.Errorf(splitKeyErr.Error())
		}
		if keyParts[0] == id {
			valueStr := keyParts[1]
			ts := keyParts[2]
			estimatorId := keyParts[3]
			value, convErr := strconv.ParseFloat(valueStr,64)
			if convErr != nil {
				return 0.0, fmt.Errorf("String %s failed to convert to float", valueStr)
			}
			timeNow := time.Now().Unix()
			timeStamp,_ := strconv.ParseInt(ts,10,64)
			curSlot :=  (timeNow -  timeStamp) / timeSlot
			if _, v := trustMp[estimator{estimatorId, curSlot}]; v == false {
				trustMp[estimator{estimatorId, curSlot}] = value
			}
		}
	}
	for key, value := range trustMp {
		if _, ok := estimatorTrustMp[key.Id]; ok != true {
			estimatorTrustMp[key.Id] = math.Pow(lamdaWeight,float64(key.Slot)) * value
		}
	}
	var avaTrust, tempTrust float64
	for _, v := range estimatorTrustMp {
		tempTrust += v
	}
	avaTrust = tempTrust / float64(len(estimatorTrustMp))
	return avaTrust, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
