package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/fx"
	"time"
	"usdt/scraping/data"
	"usdt/scraping/req"
	"usdt/scraping/sqlc-config/config"
	"usdt/scraping/sqlc-transfer-alternative/transfer_alternative"
	"usdt/scraping/sqlc-transfer/transfer"
	"usdt/scraping/tg"
)

var configData = &config.Config{}
var transferAlternativeQueries = data.TokenTransfers{}
var success = make(chan data.TokenTransfers, 10)
var fail = make(chan data.TokenTransfers, 10)

func Entrance() {
	var v Config
	conf, err := v.getConf()
	if err != nil {
		panic(err)
	}
	//初始化数据库
	db, err := sql.Open(conf.Mysql.DriverName, conf.Mysql.DataSourceName)
	if err != nil {
		panic(err)
	}

	configData, err = data.NewConfig(db)
	if err != nil {
		panic(err)
	}

	//fmt.Println(configData)
	ch := make(chan interface{}, 10)

	go Grab(ch)
	go DataFiltering2(ch)
	go DeleteData(db)
	for i := 0; i < 10; i++ {
		go Success(db, success)
		go Fail(db, fail)
	}

	// 监控数据库数据变更
	UpdateConfig(db)
}

func UpdateConfig(db *sql.DB) {
	copyConf := *configData
	for {
		time.Sleep(time.Second)
		conf, err := data.NewConfig(db)
		if err != nil {
			fmt.Printf("UpdateConfig -> data.NewConfig(db) 错误：%v\n", err)
			continue
		}
		// 比较copyConf与conf是否相等
		//fmt.Println(copyConf, *conf, !compareConfig(copyConf, *conf))

		if !compareConfig(copyConf, *conf) {
			fmt.Printf("配置文件更新：%v,旧数据：%v\n", *conf, copyConf)
			// 更新copyConf
			configData = conf
			copyConf = *conf
		}

	}
}

// compareConfig 比较两个config.Config是否相等
func compareConfig(c1 config.Config, c2 config.Config) bool {
	if c1.Status == c2.Status && c1.Url == c2.Url && c1.StartBlock == c2.StartBlock && c1.FirstAmount == c2.FirstAmount && c1.SecondAmount == c2.SecondAmount && c1.TimeLimit == c2.TimeLimit && c1.AmountCondition == c2.AmountCondition {
		return true
	}
	return false
}

func Grab(ch chan interface{}) {
	for {
		//if configData.Status {
		limit := 20
		offset := 0

		for {
			// 获取接口数据
			record, err := GetPageData(limit, offset)
			if err != nil {
				fmt.Printf("Grab -> GetPageData(limit, offset) 错误：%v\n", err)
				time.Sleep(time.Second)
				continue
			}

			var d = record.(*data.Record)
			// 确保抓取的区块数据不为空
			if len(d.TokenTransfers) > 0 {
				transferAlternativeQueries = d.TokenTransfers[0]
				// 判断当前区块是不是已确认
				if d.TokenTransfers[0].Confirmed {
					//// 数据筛选处理
					//DataFiltering(db, d)

					ch <- record

					// 接口数据翻页处理
					offset = offset + limit
					if offset > (d.Total - limit) {
						break
					}
				}
			} else {
				break
			}
		}
		configData.StartBlock = configData.StartBlock + 1
		//time.Sleep(time.Second)
		//}
	}
}

func DataFiltering2(ch chan interface{}) {

	fx.From(func(source chan<- interface{}) {
		for c := range ch {
			source <- c
		}
	}).Walk(func(item interface{}, pipe chan<- interface{}) {
		itemRecord := item.(*data.Record)
		for _, tokenTransfer := range itemRecord.TokenTransfers {
			if tokenTransfer.TokenAbbr == "USDT" {
				pipe <- tokenTransfer
			}
		}
	}).ForEach(func(item interface{}) {
		itemRecord := item.(data.TokenTransfers)
		fromString, err := decimal.NewFromString(itemRecord.Quant)
		if err != nil {
			fmt.Printf("DataFiltering2 -> decimal.NewFromString(itemRecord.Quant) 错误：%v\n", err)
		} else {
			// 计算转账金额
			num := fromString.Div(decimal.NewFromInt(1000000))
			newFromString, err := decimal.NewFromString(configData.FirstAmount)

			if err != nil {
				fmt.Printf("DataFiltering2 -> decimal.NewFromString(configData.FirstAmount) 错误：%v\n", err)
			} else {
				if num.IsPositive() {
					// 判断转账金额是否满足设置的条件
					if num.IsPositive() && num.Cmp(newFromString) <= 0 {

						success <- itemRecord
					}

					fail <- itemRecord
				}
			}
		}

	})
}

// 金额条件满足入库
func Success(db *sql.DB, itemRecord chan data.TokenTransfers) {
	ctx := context.Background()
	query := transfer_alternative.New(db)
	r := transfer.New(db)
	for {
		item := <-itemRecord
		fromString, err := decimal.NewFromString(item.Quant)
		if err != nil {
			fmt.Printf("Success -> decimal.NewFromString(item.Quant) 错误：%v\n", err)
			continue
		}
		num := fromString.Div(decimal.NewFromInt(1000000))

		// 如果再次监听到相同的1U转账 则通知
		err = r.IsExistTransfer(ctx, item.ToAddress)
		if err != nil && err != sql.ErrNoRows {
			// 调用小飞机通知接口
			telegram := tg.TG{
				Url: configData.TgBotUrl + "/sendMessage",
			}
			tgTxt := map[string]interface{}{
				"chat_id": configData.TgID,
				"text":    fmt.Sprintf("监测到 \n%v\n向\n%v\n转账金额\n%v USDT", item.FromAddress, item.ToAddress, num),
			}
			err := telegram.SendMsg(tgTxt, nil)
			if err != nil {
				fmt.Printf("Success -> telegram.SendMsg(tgTxt, nil) 错误：%v\n", err)
			}
		}

		// 计算转账金额
		err = query.CreateTransferAlternative(ctx, transfer_alternative.CreateTransferAlternativeParams{
			Amount:        num.String(),
			FromAddress:   item.FromAddress,
			ToAddress:     item.ToAddress,
			Block:         int32(item.Block),
			TransactionID: item.TransactionId,
			Time:          time.Unix(item.Time/1000, 0),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if err != nil {
			fmt.Printf("Success -> query.CreateTransferAlternative 错误：%v\n", err)
			continue
		}
		fmt.Println("1:", <-itemRecord, num)

	}
}

// 金额条件不满足入库
func Fail(db *sql.DB, itemRecord chan data.TokenTransfers) {
	ctx := context.Background()
	query := transfer_alternative.New(db)
	r := transfer.New(db)

	for {
		item := <-itemRecord
		fromString, err := decimal.NewFromString(item.Quant)
		if err != nil {
			fmt.Printf("Fail -> decimal.NewFromString(item.Quant) 错误：%v\n", err)
			continue
		}

		// 计算转账金额
		num := fromString.Div(decimal.NewFromInt(1000000))

		de, err := decimal.NewFromString(configData.SecondAmount)
		if err != nil {
			fmt.Printf("Fail -> decimal.NewFromString(configData.SecondAmount) 错误：%v\n", err)
			continue
		}
		if num.Cmp(de) >= 0 {
			transf, err := query.TransferAlternative(ctx, transfer_alternative.TransferAlternativeParams{
				FromAddress: item.ToAddress,
				ToAddress:   item.FromAddress,
			})
			if err != nil {
				if err != sql.ErrNoRows {
					fmt.Printf("Fail -> query.TransferAlternative 错误：%v\n", err)
				}
				continue
			}

			fmt.Println(transf, item.Time, transf.Time)
			s := time.Unix(item.Time/1000, 0).Sub(transf.Time).Seconds()
			if int32(s) < configData.TimeLimit {

				err := r.CreateTransfer(ctx, transfer.CreateTransferParams{
					Amount:        num.String(),
					FromAddress:   item.FromAddress,
					ToAddress:     item.ToAddress,
					Block:         int32(item.Block),
					TransactionID: item.TransactionId,
					Time:          time.Unix(item.Time/1000, 0),
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				})

				if err != nil {
					fmt.Printf("Fail -> r.CreateTransfer 错误：%v\n", err)
					continue
				}

				fmt.Println("2:", <-itemRecord, num)
			}
		}

	}

}

func GetPageData(limit int, offset int) (interface{}, error) {
	r := req.Http{}
	var d data.Record
	url := fmt.Sprintf("%v/api/token_trc20/transfers?limit=%v&start=%v&sort=-timestamp&count=true&block=%v", configData.Url, limit, offset, configData.StartBlock)
	err := r.Get(url, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func DataFiltering(db *sql.DB, record *data.Record) {
	index := 0
	end := len(record.TokenTransfers)
	for index < end {
		d := record.TokenTransfers[index]
		// 只处理usdt转账的数据
		if d.TokenAbbr == "USDT" {
			fromString, err := decimal.NewFromString(d.Quant)
			if err == nil {
				// 计算转账金额
				num := fromString.Div(decimal.NewFromInt(1000000))
				newFromString, err := decimal.NewFromString(configData.FirstAmount)

				if err == nil {
					ctx := context.Background()
					query := transfer_alternative.New(db)

					// 判断转账金额是否满足设置的条件
					if num.Equal(newFromString) {
						err := query.CreateTransferAlternative(ctx, transfer_alternative.CreateTransferAlternativeParams{
							Amount:        num.String(),
							FromAddress:   d.FromAddress,
							ToAddress:     d.ToAddress,
							Block:         int32(d.Block),
							TransactionID: d.TransactionId,
							Time:          time.Unix(d.Time/1000, 0),
						})
						if err == nil {
							fmt.Println("1:", record.TokenTransfers[index], num)
						}
					} else {
						de, err := decimal.NewFromString(configData.SecondAmount)
						if err == nil {
							if num.Cmp(de) >= 0 {
								transf, err := query.TransferAlternative(ctx, transfer_alternative.TransferAlternativeParams{
									FromAddress: d.ToAddress,
									ToAddress:   d.FromAddress,
								})
								if err == nil && err != sql.ErrNoRows {
									fmt.Println(transf, err, err == sql.ErrNoRows, d.Time, transf.Time)
									s := time.Unix(d.Time/1000, 0).Sub(transf.Time).Seconds()
									if int32(s) < configData.TimeLimit {
										//balance, err := GetUSDTBalance(configData, d.ToAddress)
										//if err == nil {
										//	l, err := decimal.NewFromString(configData.AmountCondition)
										//	if err == nil {
										//		if decimal.NewFromFloat(balance).Cmp(l) >= 0 {
										r := transfer.New(db)
										r.CreateTransfer(ctx, transfer.CreateTransferParams{
											Amount:        num.String(),
											FromAddress:   d.FromAddress,
											ToAddress:     d.ToAddress,
											Block:         int32(d.Block),
											TransactionID: d.TransactionId,
											Time:          time.Unix(d.Time/1000, 0),
										})

										fmt.Println("2:", record.TokenTransfers[index], num)
										//}
										//}
										//}
									}
								}
							}
							//fmt.Println("2:", record.TokenTransfers[index], num)
						}

					}
					//fmt.Println(record.TokenTransfers[index], num)
				}
			}
		}

		index = index + 1
	}
}

func GetUSDTBalance(address string) (float64, error) {
	r := req.Http{}
	var d data.RecordBalance
	url := fmt.Sprintf("%v/api/account/tokens?address=%v&start=0&limit=20&token=&hidden=0&show=0&sortType=0", configData.Url, address)
	err := r.Get(url, &d)
	if err != nil {
		return 0, err
	}
	for _, i2 := range d.Data {
		if i2.TokenAbbr == "USDT" {
			return i2.AmountInUsd, nil
		}
	}
	return 0, nil
}

// 定时任务 10分钟执行一次 删除数据库中超时的数据
func DeleteData(db *sql.DB) {
	query := transfer_alternative.New(db)

	for {
		if transferAlternativeQueries.Time > 0 {
			ctx := context.Background()
			t := transferAlternativeQueries.Time
			// 13位时间戳转时间
			tm := time.Unix(t/1000, 0)
			//计算tm时间减去configData.TimeLimit的时间
			tm = tm.Add(time.Duration(-configData.TimeLimit) * time.Second)

			err := query.DeleteTransferAlternativeByTime(ctx, tm)
			if err != nil {
				fmt.Println("数据删除脚本执行失败: ", err)
			}
			fmt.Println("数据删除脚本执行成功")
			time.Sleep(time.Minute * 10)
		}
	}
}

//
//package main
//
//import (
//"context"
//"database/sql"
//"fmt"
//_ "github.com/go-sql-driver/mysql"
//"github.com/shopspring/decimal"
//"time"
//"usdt/scraping/data"
//"usdt/scraping/req"
//"usdt/scraping/sqlc-config/config"
//"usdt/scraping/sqlc-transfer-alternative/transfer_alternative"
//"usdt/scraping/sqlc-transfer/transfer"
//)
//
//var configData = &config.Config{}
//var transferAlternativeQueries = data.TokenTransfers{}
//
//func Entrance() {
//	var v Config
//	conf, err := v.getConf()
//	if err != nil {
//		panic(err)
//	}
//	//初始化数据库
//	db, err := sql.Open(conf.Mysql.DriverName, conf.Mysql.DataSourceName)
//	if err != nil {
//		panic(err)
//	}
//
//	configData, err = data.NewConfig(db)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(configData)
//
//	go Grab(db)
//	go DeleteData(db)
//	// 监控数据库数据变更
//	UpdateConfig(db)
//}
//
//func UpdateConfig(db *sql.DB) {
//	copyConf := *configData
//	for {
//		time.Sleep(time.Second)
//		conf, err := data.NewConfig(db)
//		if err == nil && err != sql.ErrNoRows {
//			// 比较copyConf与conf是否相等
//			if !compareConfig(copyConf, *conf) {
//				fmt.Printf("配置文件更新：%v,旧数据：%v\n", *conf, copyConf)
//				// 更新copyConf
//				configData = conf
//				copyConf = *conf
//			}
//		}
//	}
//}
//
//// compareConfig 比较两个config.Config是否相等
//func compareConfig(c1 config.Config, c2 config.Config) bool {
//	if c1.Status == c2.Status && c1.Url == c2.Url && c1.StartBlock == c2.StartBlock {
//		return true
//	}
//	return false
//}
//
//func Grab(db *sql.DB) {
//	for {
//		fmt.Println(configData)
//		if configData.Status > 0 {
//			limit := 20
//			offset := 0
//
//			for {
//				// 获取接口数据
//				record, err := GetPageData(limit, offset)
//				if err != nil {
//					time.Sleep(time.Second)
//					continue
//				}
//
//				var d = record.(*data.Record)
//				// 确保抓取的区块数据不为空
//				if len(d.TokenTransfers) > 0 {
//					transferAlternativeQueries = d.TokenTransfers[0]
//					// 判断当前区块是不是已确认
//					if d.TokenTransfers[0].Confirmed {
//						// 数据筛选处理
//						DataFiltering(db, d)
//						// 接口数据翻页处理
//						offset = offset + limit
//						if offset > (d.Total - limit) {
//							break
//						}
//					}
//				}
//			}
//			configData.StartBlock = configData.StartBlock + 1
//			//fmt.Println(d)
//			//break
//			time.Sleep(time.Second)
//		}
//	}
//}
//
//func GetPageData(limit int, offset int) (interface{}, error) {
//	r := req.Http{}
//	var d data.Record
//	url := fmt.Sprintf("%v/api/token_trc20/transfers?limit=%v&start=%v&sort=-timestamp&count=true&block=%v", configData.Url, limit, offset, configData.StartBlock)
//	err := r.Get(url, &d)
//	if err != nil {
//		return nil, err
//	}
//	return &d, nil
//}
//
//func DataFiltering(db *sql.DB, record *data.Record) {
//	index := 0
//	end := len(record.TokenTransfers)
//	for index < end {
//		d := record.TokenTransfers[index]
//		// 只处理usdt转账的数据
//		if d.TokenAbbr == "USDT" {
//			fromString, err := decimal.NewFromString(d.Quant)
//			if err == nil {
//				// 计算转账金额
//				num := fromString.Div(decimal.NewFromInt(1000000))
//				newFromString, err := decimal.NewFromString(configData.FirstAmount)
//
//				if err == nil {
//					ctx := context.Background()
//					query := transfer_alternative.New(db)
//
//					// 判断转账金额是否满足设置的条件
//					if num.Equal(newFromString) {
//						err := query.CreateTransferAlternative(ctx, transfer_alternative.CreateTransferAlternativeParams{
//							Amount:        num.String(),
//							FromAddress:   d.FromAddress,
//							ToAddress:     d.ToAddress,
//							Block:         int32(d.Block),
//							TransactionID: d.TransactionId,
//							Time:          time.Unix(d.Time/1000, 0),
//						})
//						if err == nil {
//							fmt.Println("1:", record.TokenTransfers[index], num)
//						}
//					} else {
//						de, err := decimal.NewFromString(configData.SecondAmount)
//						if err == nil {
//							if num.Cmp(de) >= 0 {
//								transf, err := query.TransferAlternative(ctx, transfer_alternative.TransferAlternativeParams{
//									FromAddress: d.ToAddress,
//									ToAddress:   d.FromAddress,
//								})
//								if err == nil && err != sql.ErrNoRows {
//									fmt.Println(transf, err, err == sql.ErrNoRows, d.Time, transf.Time)
//									s := time.Unix(d.Time/1000, 0).Sub(transf.Time).Seconds()
//									if int32(s) < configData.TimeLimit {
//										//balance, err := GetUSDTBalance(configData, d.ToAddress)
//										//if err == nil {
//										//	l, err := decimal.NewFromString(configData.AmountCondition)
//										//	if err == nil {
//										//		if decimal.NewFromFloat(balance).Cmp(l) >= 0 {
//										r := transfer.New(db)
//										r.CreateTransfer(ctx, transfer.CreateTransferParams{
//											Amount:        num.String(),
//											FromAddress:   d.FromAddress,
//											ToAddress:     d.ToAddress,
//											Block:         int32(d.Block),
//											TransactionID: d.TransactionId,
//											Time:          time.Unix(d.Time/1000, 0),
//										})
//
//										fmt.Println("2:", record.TokenTransfers[index], num)
//										//}
//										//}
//										//}
//									}
//								}
//							}
//							//fmt.Println("2:", record.TokenTransfers[index], num)
//						}
//
//					}
//					//fmt.Println(record.TokenTransfers[index], num)
//				}
//			}
//		}
//
//		index = index + 1
//	}
//}
//
//func GetUSDTBalance(address string) (float64, error) {
//	r := req.Http{}
//	var d data.RecordBalance
//	url := fmt.Sprintf("%v/api/account/tokens?address=%v&start=0&limit=20&token=&hidden=0&show=0&sortType=0", configData.Url, address)
//	err := r.Get(url, &d)
//	if err != nil {
//		return 0, err
//	}
//	for _, i2 := range d.Data {
//		if i2.TokenAbbr == "USDT" {
//			return i2.AmountInUsd, nil
//		}
//	}
//	return 0, nil
//}
//
//// 定时任务 10分钟执行一次 删除数据库中超时的数据
//func DeleteData(db *sql.DB) {
//	for {
//		if transferAlternativeQueries.Time > 0 {
//			ctx := context.Background()
//			query := transfer_alternative.New(db)
//			t := transferAlternativeQueries.Time
//			// 13位时间戳转时间
//			tm := time.Unix(t/1000, 0)
//			//计算tm时间减去configData.TimeLimit的时间
//			tm = tm.Add(time.Duration(-configData.TimeLimit) * time.Second)
//
//			err := query.DeleteTransferAlternativeByTime(ctx, tm)
//			if err != nil {
//				fmt.Println("数据删除脚本执行失败: ", err)
//			}
//			fmt.Println("数据删除脚本执行成功")
//			time.Sleep(time.Minute * 10)
//		}
//	}
//}
