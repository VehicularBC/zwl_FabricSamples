package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Smartcontract provides the function for managing
type SmartContract struct {
	contractapi.Contract
}
//发起人身份信息
type Voter struct{
	ID         string  `json:"id"`   //投票人ID
	Weight     int  `json:"weight"`   //投票人权重
	Voted      bool `json:"voted"`    //是否已经投票标记
	// vote       int  `json:"vote"`     //当前投票索引
}
//冒泡排序
func BubbleSort(slice []int){
	for  range slice{
		for i := range slice{
			if i < len(slice)-1{
				if slice[i] > slice[i+1]{
					slice[i],slice[i+1] = slice[i+1], slice[i]
				}
			}
		}
	}
}
//快速排序
func QuickSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	splitdata := arr[0]          //第一个数据
	low := make([]int, 0, 0)     //比我小的数据
	hight := make([]int, 0, 0)   //比我大的数据
	mid := make([]int, 0, 0)     //与我一样大的数据
	mid = append(mid, splitdata) //加入一个
	for i := 1; i < len(arr); i++ {
		if arr[i] < splitdata {
			low = append(low, arr[i])
		} else if arr[i] > splitdata {
			hight = append(hight, arr[i])
		} else {
			mid = append(mid, arr[i])
		}
	}
	low, hight = QuickSort(low), QuickSort(hight)
	myarr := append(append(low, mid...), hight...)
	return myarr
}
func SliceBubbleSort(num int)[]int{
	rand.Seed(time.Now().Unix())
	numarr := make([]int, 0)
	for i:=0;i<num;i++{
		numarr = append(numarr, rand.Intn(1000))
	}
	BubbleSort(numarr)
	return numarr
}
func SliceQuickSort(num int)[]int{
	rand.Seed(time.Now().Unix())
	numarr := make([]int, 0)
	for i:=0;i<num;i++{
		numarr = append(numarr, rand.Intn(100))
	}
	a := QuickSort(numarr)
	return a
}
//投票人初始化
func (s *SmartContract) InitVoterLedger(ctx contractapi.TransactionContextInterface) error {
	voters := []Voter{
		{ID: "car1", Weight: 1, Voted: false},
		{ID: "car2", Weight: 1, Voted: false},
		{ID: "car3", Weight: 1, Voted: false},
		{ID: "car4", Weight: 1, Voted: false},
		{ID: "car5", Weight: 1, Voted: false},
		{ID: "car6", Weight: 1, Voted: false},
	}

	for _, voter := range voters {
		voterJSON, err := json.Marshal(voter)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(voter.ID, voterJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}
//检测账本中是否存在输入的键
func (s *SmartContract) ExistInquire(ctx contractapi.TransactionContextInterface, everything string) (bool,error){
	existsJSON, err :=ctx.GetStub().GetState(everything)
	if err != nil{
		return false, fmt.Errorf("can't read from world state: %v" , err)
	}
	return existsJSON != nil, nil
}
//正常创建键
func (s *SmartContract) CreateVoter(ctx contractapi.TransactionContextInterface, id string) error {

	voter := Voter{
		ID:     "id",
		Weight: 1,
		Voted:  false,
	}
	voterJson, err := json.Marshal(voter)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	err = ctx.GetStub().PutState(id, voterJson)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	return nil
}
//冒泡排序创造键
func (s *SmartContract) CreateVoterBubbleSort(ctx contractapi.TransactionContextInterface, id string, num int) error {
	a := SliceBubbleSort(num)
	voter := Voter{
		ID:     "id",
		Weight: a[6],
		Voted:  false,
	}
	voterJson, err := json.Marshal(voter)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	err = ctx.GetStub().PutState(id, voterJson)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	return nil
}
//快速排序创造键
func (s *SmartContract) CreateVoterQuickSort(ctx contractapi.TransactionContextInterface, id string, num int) error {
	a := SliceQuickSort(num)
	voter := Voter{
		ID:     "id",
		Weight: a[6],
		Voted:  false,
	}
	voterJson, err := json.Marshal(voter)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	err = ctx.GetStub().PutState(id, voterJson)
	if err != nil{
		return fmt.Errorf("can't write the world state %v", err)
	}
	return nil
}
//正常删除键
func (s *SmartContract) DeleteVoter(ctx contractapi.TransactionContextInterface, id string) error{
	temp, err := s.ExistInquire(ctx, id)
	if temp != true {
		 fmt.Print("there is no key named this")
	}
	if err != nil{
		return fmt.Errorf("there is something wrong in reading world state %v", err)
	}
	err = ctx.GetStub().DelState(id)
	if err != nil{
		return fmt.Errorf(" there is something wrong in ddleting key %v", err)
	}
	return nil
}
//百次正常创造键
func (s *SmartContract) HundredCreateVoter(ctx contractapi.TransactionContextInterface, id string, num int) error {
	a := SliceBubbleSort(num)
	for _, value := range a{
		strep := "car"+string(value)
		err := s.CreateVoter(ctx, strep)
		if err != nil{
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}
//投票委托
func (s *SmartContract) DelegateVote(ctx contractapi.TransactionContextInterface, ownerName string, delegatedName string) error {
	owner, err := s.GetVoter(ctx, ownerName) //获得owner的结构
	if err != nil {
		return nil
	}
	delegatedvoter, err := s.GetVoter(ctx, delegatedName) //获得delegatedvoter结构
	if err != nil {
		return  nil
	}

	//检测是否已投票
	if owner.Voted == true || delegatedvoter.Voted == true {
		return fmt.Errorf("the delegated voter %s has voted", delegatedName)
	}
	//检测是否投给本人
	if owner.ID == delegatedvoter.ID {
		return fmt.Errorf("cannot delegate the vote to yourself")
	}
	delegatedvoter.Weight = owner.Weight + delegatedvoter.Weight
	owner.Weight = 0
	owner.Voted = true

	ownerJSON, errOwner := json.Marshal(owner)
	if errOwner != nil {
		return  nil
	}
	delegatedvoterJSON, errdelegatedvoter := json.Marshal(delegatedvoter)
	if errdelegatedvoter != nil {
		return  nil
	}
	err = ctx.GetStub().PutState(ownerName, ownerJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	err = ctx.GetStub().PutState(delegatedName, delegatedvoterJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}
// GetAsset returns the basic asset with id given from the world state
func (s *SmartContract) GetVoter(ctx contractapi.TransactionContextInterface, id string) (*Voter, error) {
	existing, err:= ctx.GetStub().GetState(id)

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(Voter)

	err = json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Voter", id)
	}
	return ba, nil
}//GetVoter和 GetProposal应该可以写成一个函数
func (s *SmartContract) GetAllVoter(ctx contractapi.TransactionContextInterface) ([]*Voter, error){
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var voters []*Voter
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var voter Voter
		err = json.Unmarshal(queryResponse.Value, &voter)
		if err != nil {
			return nil, err
		}
		voters = append(voters, &voter)
	}

	return voters, nil
}
//从世界状态中返回所输入name的proposal
/*func (s *SmartContract) CalculateVoteName(ctx contractapi.TransactionContextInterface, votedname string) (string, error){
	resultsIterator, err := ctx.GetStub().GetStateByRange("proposal1", "proposal6")
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()
	//var proposals []*Proposal
	maxvote := 0
	var votedproposal string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}

		var proposal Proposal
		err = json.Unmarshal(queryResponse.Value, &proposal)
		if err != nil {
			return "", err
		}
		if maxvote < proposal.VoteCount{
			maxvote = proposal.VoteCount
			votedproposal = proposal.Name
		}
	}
	 votedJSON, err := json.Marshal(votedproposal)
	err = ctx.GetStub().PutState(votedname,votedJSON)
	if err != nil{
		return "",nil
	}
	return votedproposal, nil
}*/
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