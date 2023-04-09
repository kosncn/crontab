package common

// Batch 日志批次
type Batch struct {
	Logs []interface{} // 多条日志
}

// NewBatch 实例化日志批次对象
func NewBatch() *Batch {
	return &Batch{}
}
