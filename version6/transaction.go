package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

const reward = 12.5		//目前的挖矿奖励

//1. 定义交易结构
type Transaction struct {
	TxHash []byte		//交易hash
	//一个交易中有多个输入和多个输出
	Vin []TxInput		//交易输入
	Vout []TxOutput		//交易输出
	TimeStamp uint64	//产生时间
}

//定义交易输入结构
type TxInput struct {
	TxHash []byte		//上一个输出所在交易的hash
	OutputIndex int64	//上一个输出的编号 就是其在VOut中的序号
	ScriptSig []byte	//输入脚本，这里先简单的使用地址 	//TODO
}
//定义交易输出结构
type TxOutput struct {
	Value float64		//输出的金额
	ScriptPubKey []byte	//输出脚本，这里先用公钥
}

//设置交易hash
func (tx * Transaction)SetTxHash() {
	//采用序列化
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil{
		log.Panic("设置交易hash失败！", err)
	}
	data := buffer.Bytes()
	//sha256
	hash := sha256.Sum256(data)
	//赋值
	tx.TxHash = hash[:]
}

//交易id计数器
var txId uint64 = 0

//2. 提供创建交易方法
//创建普通交易
func NewTransaction(from, to string, amount float64, bc BlockChain) *Transaction {
	//1. 找到合理的UTXO集合，用于转账. utxo是一个map，key为交易哈希，val为交易编号数组
	utxosMap, resAmount := bc.FindSuitableUTXO(from, amount)
	//判断最合适的UTXO是否满足需求，不满足直接返回
	if resAmount < amount{
		fmt.Printf("转账账户当前的最高余额是:%.8f\n", resAmount)
		return nil
	}
	//2. 把这些交易转换为input输入
	var inputArray []TxInput 
	var outputArray []TxOutput
	// 遍历map
	for txId, indexArray := range utxosMap{
		// 遍历编号数组
		for _, index := range indexArray{
			//转换
			input := TxInput{
				TxHash:      []byte(txId),
				OutputIndex: index,
				ScriptSig:   []byte(from),		//TODO 这里要解锁验证
			}
			//加入到输入组中:
			inputArray = append(inputArray, input)
		}
	}
	//3. 再创建对应的输出
	output := TxOutput{
		Value:        amount,	//这的钱是转账需要的钱
		ScriptPubKey: []byte(to),
	}
	outputArray = append(outputArray, output)
	//判断是否要找零
	if resAmount > amount {
		//找零, 添加一个输出
		outputArray = append(outputArray, TxOutput{
			Value:        resAmount-amount,
			ScriptPubKey: []byte(from),
		})
	}
	//4. 生成交易
	newTx := Transaction{
		TxHash:    nil,
		Vin:       inputArray,
		Vout:      outputArray,
		TimeStamp: uint64(time.Now().Unix()),
	}
	newTx.SetTxHash()
	return &newTx
}


//3. 创建挖矿交易
//挖矿交易
func createCoinBaseTx(address string, commitment string)  *Transaction{
	//创建coinbase
	input := TxInput{
		TxHash:      nil,					//铸币交易没有输入
		OutputIndex: -1,					//没有上一个输出
		ScriptSig:   []byte(commitment),	//coinBase域，可以写想保存的内容
	}
	Output := TxOutput{
		Value:        reward,				//铸币交易
		ScriptPubKey: []byte(address),		//用公钥签名
	}
	tx :=  &Transaction{
		TxHash:    nil,
		Vin:       []TxInput{input},		//把铸币交易放进区块中
		Vout:      []TxOutput{Output},
		TimeStamp: uint64(time.Now().Unix()),
	}
	//设置交易的Hash值
	tx.SetTxHash()
	return tx
}

//判断一个交易是否是挖矿交易
func (tx *Transaction)isCoinbaseTx() bool {
	//条件
	//1. 交易只有一个输入
	//2. 输入中上一个交易hash为空
	//3. 输入中的上一个输出编号为-1
	if len(tx.Vin) == 1 && bytes.Equal(tx.Vin[0].TxHash , []byte{}) && tx.Vin[0].OutputIndex == -1{
		return true
	}
	return false
}



