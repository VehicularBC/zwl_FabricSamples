# 链码Fabric-Sample环境

测试链码时可以直接下载该分支，保证链码写入者和测试者的环境一致性。

## 21.7.21 更新

新增函数`VoteForTempCar()`和`GetVoteResult()`。使用参考如下：

```go
result, err := contract.SubmitTransaction("GetVoteResult","Tempcar1")//获取当前投票状态
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}
fmt.Println(string(result))

result, err = contract.SubmitTransaction("VoteForTempCar","Tempcar1")//投票
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}
fmt.Println(string(result))

result, err := contract.SubmitTransaction("GetVoteResult","Tempcar1")//获取当前投票状态
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}
fmt.Println(string(result))
//返回示例：
//临时车从未被投票：错误"No variable by the name 某某id exists", true(代表正常投票)，false(未达到阈值)
//临时车已被其他车投票，自己投票后到阈值：false, ture，true
//临时车已到阈值，自己投票后超过阈值:true, true(仍会继续投票),error(该车辆投票过程已结束)
//某正式车重复投票：...,"正式车已投票",...投票状态见以上情况
```

