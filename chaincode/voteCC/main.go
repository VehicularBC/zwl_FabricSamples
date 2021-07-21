package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const car = "CarType~ID"
const temp = "TempVehicle"
const formal = "FormalVehicle"
const voteThreshold = 2

//Smartcontract provides the function for managing
type SmartContract struct {
	contractapi.Contract
}
//发起人身份信息
type TempVehicle struct{
	Id 					string 		`json:"id"`						//临时车辆名称
	Threshold			int			`json:"threshold"`				//临时车辆成功注册所需要的背书者的数量
	EndorserList		[]string	`json:"endorser_list"`			//为白板车申请身份背书的车辆Id列表
}

type FormalVehicle struct {
	Id 					string		`json:"id"`						//正式车辆名称
	Uid					string		`json:"uid"`					//车辆名称所对应的SDK的wallet中的ID
	VoteWeight			int			`json:"vote_weight"`			//车辆投票权重
	EndorserList		[]string	`json:"endorser_list"`			//加入网络时的背书车辆的列表
}
type TempVoteCar struct {}
//计算临时车投票通过的门限
func calculateVoteThreshold() int{
	var threshold int
	//遍历正式车列表，如使用id切片分组作为车辆列表，可采用车辆列表的1/3的车辆数额作为投票门限

	//如果投票门限小于所设定的默认值，则返回默认值
	if threshold < voteThreshold{
		return voteThreshold
	}
	return threshold
}

func (s *SmartContract) VoteForTempCar(ctx contractapi.TransactionContextInterface, objectKey  string) (bool,error){
	myKey := s.GetSDKuserId(ctx)
	compositeKey, err := CreateCarVoteCompositeKey(ctx, temp , objectKey, myKey)
	if err != nil {
		return  false, err
	}
	if tempCar, _ := ctx.GetStub().GetState(compositeKey); tempCar != nil {
		return false, fmt.Errorf("该正式车已投票")
	}
	//创建临时车辆表项
	err = ctx.GetStub().PutState(compositeKey, []byte{0x00})
	if err != nil{
		return false, err
	}
	return true, nil
}

func (s *SmartContract) GetVoteResult(ctx contractapi.TransactionContextInterface, objectKey string) (bool, error) {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(car, []string{temp, objectKey})
	if err != nil{
		return false, err
	}
	defer resultsIterator.Close()
	if !resultsIterator.HasNext() {
		return false, fmt.Errorf("No variable by the name %s exists", objectKey)
	}
	var flag int
	for resultsIterator.HasNext(){
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return false, err
		}
		if queryResponse != nil{
			flag++
		}
	}
	if flag > calculateVoteThreshold() {
		return false, fmt.Errorf("该车辆投票过程已结束")
	}
	if flag == calculateVoteThreshold() {
		//此处应加入删除所有表项的代码
		return true, nil
	}
	//临时表项存在
	//临时表项不存在
	return false , nil
	/*var voteList []string
	for resultsIterator.HasNext(){
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return 0, err
		}
		_, keyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if err != nil{
			return 0, err
		}
		
		_ = keyParts[3]
		/*if voteList != nil {
			for _, id := range voteList {
				if id == voteId {
					continue
				} else {
					voteList = append(voteList, voteId)
				}
			}
		}else{
			voteList = append(voteList, voteId)
		}
		voteList = append(voteList, "voteId")*/
}
//*****
func CreateCarVoteCompositeKey(ctx contractapi.TransactionContextInterface, attribute ...string) (string, error){
	attr := attribute
	//attr := []string{attribute,id}
	compositeKey, err := ctx.GetStub().CreateCompositeKey(car, attr)
	if err != nil{
		return "",fmt.Errorf("the composite key failed to be created")
	}
	return compositeKey, nil
}

//获得SDK的user身份ID，每个user的msp有一个唯一身份ID
func (s *SmartContract) GetSDKuserId(ctx contractapi.TransactionContextInterface) string {
	uid, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return ""
	}
	return uid
}

func (s *SmartContract) GetSDKcreatorId(ctx contractapi.TransactionContextInterface) string{
	uid, err := ctx.GetStub().GetCreator()
	if err != nil {
		return ""
	}
	return string(uid)
}

//向账本中添加临时车辆表项
func (s *SmartContract) CreateTempVehicle(ctx contractapi.TransactionContextInterface, id , userId string ) (bool,error){
	//创建复合键名
	compositeKey, err := CreateCarVoteCompositeKey(ctx,temp,id)
	if err != nil {
		return false, err
	}
	//创建临时车辆表项
	newTempCar := &TempVehicle{
		Id:           compositeKey,
		Threshold:    calculateVoteThreshold()-1,
		EndorserList: []string{userId},
	}
	newTempCarJson,err := json.Marshal(newTempCar)
	if err != nil{
		return false, fmt.Errorf("JSON format conversion failed")
	}
	err = ctx.GetStub().PutState(compositeKey, newTempCarJson)
	if err != nil{
		return false, err
	}
	return true, nil
}
//添加临时投票车辆表项 ****
func (s *SmartContract) CreateTempVoteCar(ctx contractapi.TransactionContextInterface, objectKey , myKey string ) (bool,error){
	//attr := []string{objectKey, myKey}
	compositeKey, err := CreateCarVoteCompositeKey(ctx, temp , objectKey, myKey)
	if err != nil {
		return  false, err
	}
	//创建临时车辆表项
	newTempCar := &TempVoteCar{}
	newTempCarJson,err := json.Marshal(newTempCar)
	if err != nil{
		return false, fmt.Errorf("JSON format conversion failed")
	}
	err = ctx.GetStub().PutState(compositeKey, newTempCarJson)
	if err != nil{
		return false, err
	}
	return true, nil
}

//创建正式车辆表项
func (s *SmartContract) TransToFormalVehicle(ctx contractapi.TransactionContextInterface, id string) error{

	//检测该车辆是否已经存在于正式车辆表项中
	if ok, _ := s.IsCompositeExisted(ctx, car, formal, id ); ok == true {
		return fmt.Errorf("The car is not a legal formal vehicle")
	}
	//创建复合键名,搜索临时表项
	tempCompositeKey, err := CreateCarVoteCompositeKey(ctx, temp , id)
	if err != nil {
		return err
	}
	tempCar, _ := s.GetTempCar(ctx, tempCompositeKey)
	//创建正式车辆表项
	var a = tempCar.Id
	newFormalCar := &FormalVehicle{
		Id:           a,
		Uid:          "",
		VoteWeight:   2,
		EndorserList: tempCar.EndorserList,
	}
	//转化为Json格式添加账本
	newFormalCarJson, err := json.Marshal(newFormalCar)
	if err != nil{
		return fmt.Errorf("JSON format conversion failed")
	}
	formalCompositeKey, err := CreateCarVoteCompositeKey(ctx, formal ,id)
	err = ctx.GetStub().PutState(formalCompositeKey, newFormalCarJson)
	if err != nil{
		return err
	}
	return  nil
}
//******

func (s *SmartContract) Vote(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	userId := s.GetSDKuserId(ctx)
	if tempCar, _ := s.GetTempVoteCar(ctx, id, userId); tempCar == true {
		return false, fmt.Errorf("该正式车已投票")
	}
	var flag int
	attrs := []string{temp, id}

	queryResult, _ := ctx.GetStub().GetStateByPartialCompositeKey(car, attrs)
	defer queryResult.Close()
	for queryResult.HasNext(){
		queryResponse, err := queryResult.Next()
		if err != nil {
			return false, err
		}
		if queryResponse != nil{
			flag++
		}
	}
	if flag >= calculateVoteThreshold() {
		return false, fmt.Errorf("该车辆投票过程已结束")
	}
	_, err := s.CreateTempVoteCar(ctx, id, userId)
	flag += 1
	if err != nil {
		return false, fmt.Errorf("创建临时表项失败")
	}
	if flag == calculateVoteThreshold() {
		//此处应加入删除所有表项的代码
		return true, nil
	}
	//临时表项存在
	//临时表项不存在
	return false , nil
}
//检查账本中是否存在某一复合键
func (s *SmartContract) IsCompositeExisted(ctx contractapi.TransactionContextInterface, objectType string, attributes string, id string) (bool, error) {
	attr := []string{attributes,id}
	compositeKey, err := ctx.GetStub().CreateCompositeKey(objectType, attr)
	if err != nil{
		return false,fmt.Errorf("the composite key failed to be created")
	}
	tempJson, err := ctx.GetStub().GetState(compositeKey)
	return tempJson != nil, err
}
//为临时车辆投票
func (s *SmartContract) InitFormalVehicle(ctx contractapi.TransactionContextInterface, id string) (*FormalVehicle, error){
	userId := s.GetSDKuserId(ctx)
	if ok, _ := s.IsCompositeExisted(ctx, car,formal, id); ok != false {
		return &FormalVehicle{},fmt.Errorf("The car has been created")
	}
	//创建复合键名,搜索临时表项
	tempCompositeKey, err := CreateCarVoteCompositeKey(ctx, formal ,id)
	if err != nil {
		return &FormalVehicle{}, err
	}
	newFormalVehicle := &FormalVehicle{
		Id:           tempCompositeKey,
		Uid:          userId,
		VoteWeight:   1,
		EndorserList: []string{},
	}
	formalVehicleJson,err  := json.Marshal(newFormalVehicle)
	if err != nil{
		return &FormalVehicle{},nil
	}
	_ = ctx.GetStub().PutState(tempCompositeKey,formalVehicleJson)
	return newFormalVehicle,nil
}

// GetAsset returns the basic asset with id given from the world state
func (s *SmartContract) GetTempCar(ctx contractapi.TransactionContextInterface, key string) (*TempVehicle, error) {
	compositeKey, err := CreateCarVoteCompositeKey(ctx, temp , key)
	if err != nil {
		return  nil, err
	}
	existing, err:= ctx.GetStub().GetState(compositeKey)
	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
	}
	ba := new(TempVehicle)

	err = json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Voter", key)
	}
	return ba, nil
}
//******
func (s *SmartContract) GetTempVoteCar(ctx contractapi.TransactionContextInterface, objectKey , myKey string) (bool, error) {
	//attr := []string{objectKey, myKey}
	compositeKey, err := CreateCarVoteCompositeKey(ctx, temp , objectKey , myKey)
	if err != nil {
		return  false, err
	}
	existing, err:= ctx.GetStub().GetState(compositeKey)
	if existing == nil {
		return false, nil
	}
	return true, nil
}

func (s *SmartContract) GetFormalCar(ctx contractapi.TransactionContextInterface, key string) (*FormalVehicle, error) {
	compositeKey, err := CreateCarVoteCompositeKey(ctx, formal , key)
	if err != nil {
		return nil,err
	}

	existing, err:= ctx.GetStub().GetState(compositeKey)

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
	}

	ba := new(FormalVehicle)

	err = json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Voter", key)
	}
	return ba, nil
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


