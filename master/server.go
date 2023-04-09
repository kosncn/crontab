package master

import (
	"net/http"
	"strconv"
	"time"

	"crontab/common"
)

// GlobalServer 服务对象
var GlobalServer = NewServer()

// Server 服务
type Server struct {
	HTTPServer *http.Server
}

// NewServer 实例化服务对象
func NewServer() *Server {
	return &Server{}
}

// Init 初始化服务对象
func (m *Server) Init() error {
	// 配置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/task/save", handleSaveTask)
	mux.HandleFunc("/task/delete", handleDeleteTask)
	mux.HandleFunc("/task/list", handleListTask)
	mux.HandleFunc("/task/kill", handleKillTask)
	mux.HandleFunc("/task/log", handleTaskLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	// 配置静态文件服务
	fileHandler := http.FileServer(http.Dir(GlobalConfig.WebPath))
	mux.Handle("/", http.StripPrefix("/", fileHandler))

	// 服务对象赋值
	m.HTTPServer = &http.Server{
		Addr:         GlobalConfig.Addr,
		Handler:      mux,
		ReadTimeout:  time.Duration(GlobalConfig.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(GlobalConfig.WriteTimeout) * time.Millisecond,
	}

	return nil
}

// handleSaveTask 保存任务接口
// POST {"task": `{"name": "xxx", "shell": "echo hello", "cronExpr": "* * * * *"}`}
func handleSaveTask(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 解析 POST 表单
	if err := r.ParseForm(); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 序列化任务数据
	task := common.NewTask()
	if err := task.Unmarshal([]byte(r.PostForm.Get("task"))); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 保存任务至 etcd 中
	oldTask, err := GlobalManager.SaveTask(task)
	if err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回旧任务响应
	data, _ := response.Build(common.StateSuccess, "", oldTask)
	_, _ = w.Write(data)
}

// handleDeleteTask 删除任务接口
// POST {"name": "task1"}
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 解析 POST 表单
	if err := r.ParseForm(); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 从 etcd 中删除任务
	oldTask, err := GlobalManager.DeleteTask(r.PostForm.Get("name"))
	if err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回旧任务响应
	data, _ := response.Build(common.StateSuccess, "", oldTask)
	_, _ = w.Write(data)
}

// handleListTask 获取任务列表接口
// GET /task/list
func handleListTask(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 从 etcd 中获取任务列表
	listTask, err := GlobalManager.ListTask()
	if err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回旧任务列表响应
	data, _ := response.Build(common.StateSuccess, "", listTask)
	_, _ = w.Write(data)
}

// handleKillTask 杀死任务接口
// POST {"name": "task1"}
func handleKillTask(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 解析 POST 表单
	if err := r.ParseForm(); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 通知 worker 服务杀死任务
	if err := GlobalManager.KillTask(r.PostForm.Get("name")); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回成功响应
	data, _ := response.Build(common.StateSuccess, "", nil)
	_, _ = w.Write(data)
}

// handleTaskLog 获取任务日志接口
// GET /task/log?name=task1&skip=0&limit=10
func handleTaskLog(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 解析 GET 参数
	if err := r.ParseForm(); err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 获取 GET 参数
	name := r.Form.Get("name")
	skip, err := strconv.Atoi(r.Form.Get("skip"))
	if err != nil {
		skip = 0
	}
	limit, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		limit = 10
	}

	// 从 mongodb 中获取任务执行日志列表
	logList, err := GlobalLogger.ListLog(name, skip, limit)
	if err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回任务执行日志列表响应
	data, _ := response.Build(common.StateSuccess, "", logList)
	_, _ = w.Write(data)
}

// handleWorkerList 获取服务注册接口
// GET /worker/list
func handleWorkerList(w http.ResponseWriter, r *http.Request) {
	// 实例化通讯响应对象
	response := common.NewResponse()

	// 从 etcd 中获取服务注册列表
	workerList, err := GlobalManager.ListWorker()
	if err != nil {
		data, _ := response.Build(common.StateFailure, err.Error(), nil)
		_, _ = w.Write(data)
	}

	// 返回服务注册列表响应
	data, _ := response.Build(common.StateSuccess, "", workerList)
	_, _ = w.Write(data)
}
