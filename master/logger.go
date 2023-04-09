package master

import (
	"context"
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

	return nil
}

// ListLog 获取任务执行日志列表
func (l *Logger) ListLog(name string, skip int, limit int) ([]*common.Log, error) {
	// 实例化任务执行日志过滤条件对象
	filter := common.NewLogFilter(name)

	// 实例化任务执行日志排序规则对象，按照开始时间倒序排序
	sorter := common.NewLogSorter(-1)

	// 查询任务执行日志
	opts := options.Find().SetSort(sorter).SetSkip(int64(skip)).SetLimit(int64(limit))
	cursor, err := l.Collection.Find(context.TODO(), filter, opts)
	defer func(cur *mongo.Cursor) {
		_ = cur.Close(context.TODO())
	}(cursor)
	if err != nil {
		return nil, err
	}

	// 遍历任务执行日志
	logList := make([]*common.Log, 0)
	for cursor.Next(context.TODO()) {
		// 实例化任务执行日志对象
		log := common.NewLog()

		// 反序列化 bson 数据
		if err := cursor.Decode(log); err != nil {
			continue // bson 数据格式不正确，跳过该条数据
		}
		logList = append(logList, log)
	}

	return logList, nil
}
