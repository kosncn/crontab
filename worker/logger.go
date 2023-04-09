package worker

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"crontab/common"
)

// GlobalLogger 日志管理器对象
var GlobalLogger = NewLogger()

// Logger 日志管理器
type Logger struct {
	Client     *mongo.Client
	Collection *mongo.Collection
	LogChan    chan *common.Log
	BatchChan  chan *common.Batch
}

// NewLogger 实例化日志管理器对象
func NewLogger() *Logger {
	return &Logger{}
}

// Init 初始化日志管理器对象
func (l *Logger) Init() error {
	// 建立 mongodb 连接
	opts := options.Client().
		ApplyURI(GlobalConfig.MongoDBURI).
		SetConnectTimeout(time.Duration(GlobalConfig.MongoDBConnectTimeout) * time.Millisecond)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return err
	}

	// 选择 db 和 collection
	l.Client = client
	l.Collection = client.Database("cron").Collection("log")
	l.LogChan = make(chan *common.Log, GlobalConfig.ChanSize)
	l.BatchChan = make(chan *common.Batch, GlobalConfig.ChanSize)

	// 启动日志储存协程
	go l.WriteLoop()

	return nil
}

// Save 保存日志
func (l *Logger) Save(log *common.Log) {
	select {
	case l.LogChan <- log:
	default:
		// 日志批次已经存满，丢弃当前日志
	}
}

// WriteLoop 日志储存协程
func (l *Logger) WriteLoop() {
	var batch *common.Batch
	var commitTimer *time.Timer

	for {
		select {
		case log := <-l.LogChan:
			if batch == nil {

				// 初始化日志批次
				batch = common.NewBatch()

				// 日志批次自动提交超时
				commitTimer = time.AfterFunc(
					time.Duration(GlobalConfig.LogCommitTimeout)*time.Millisecond,
					func(b *common.Batch) func() {
						return func() {
							l.BatchChan <- b
						}
					}(batch),
				)
			}

			// 追加新日志到日志批次中
			batch.Logs = append(batch.Logs, log)

			// 判断日志批次是否已满，满了就储存到 mongodb 中
			if len(batch.Logs) >= GlobalConfig.BatchSize {
				// 保存日志
				_, _ = l.Collection.InsertMany(context.TODO(), batch.Logs)

				// 清空日志批次
				batch = nil

				// 取消定时器
				commitTimer.Stop()
			}
		case b := <-l.BatchChan: // 过期的日志批次
			// 判断过期日志批次是否为当前日志批次
			if b != batch {
				continue // 跳过已经被提交的日志批次
			}

			// 将日志批次写入 mongodb 中
			if _, err := l.Collection.InsertMany(context.TODO(), b.Logs); err != nil {
				fmt.Println(err)
			}

			// 清空日志批次
			batch = nil
		}
	}
}
