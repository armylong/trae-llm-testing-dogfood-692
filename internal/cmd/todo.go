package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/armylong-go/internal/common/webcache"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func TodoHandler(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	taskType := ""
	if len(args) > 0 {
		taskType = args[0]
	}
	taskId, _ := cmd.Flags().GetInt64("task_id")
	title, _ := cmd.Flags().GetString("title")
	desc, _ := cmd.Flags().GetString("desc")
	sort, _ := cmd.Flags().GetInt64("sort")
	expireAt, _ := cmd.Flags().GetString("expire_at")

	if taskType == "" {
		fmt.Println("错误: task_type 不能为空")
		fmt.Println("可用命令: get, create, sort, complete, expire")
		return
	}

	switch taskType {
	case "get":
		taskData, err := getTodoTask(ctx, taskId)
		if err != nil {
			fmt.Printf("获取任务失败: %v\n", err)
			return
		}
		printTask(taskId, taskData)
		return
	case "create":
		res, err := createTodoTask(ctx, title, desc, sort, expireAt)
		if err != nil {
			fmt.Printf("创建任务失败: %v\n", err)
			return
		}
		if res {
			fmt.Println("✓ 任务创建成功")
		} else {
			fmt.Println("任务已存在，更新完成")
		}
		return
	case "sort":
		if taskId == 0 {
			fmt.Println("错误: sort 命令需要指定 task_id")
			return
		}
		err := updateTaskSort(ctx, taskId, sort)
		if err != nil {
			fmt.Printf("更新排序失败: %v\n", err)
			return
		}
		fmt.Printf("✓ 任务 %d 排序值已更新为 %d\n", taskId, sort)
		return
	case "complete":
		if taskId == 0 {
			fmt.Println("错误: complete 命令需要指定 task_id")
			return
		}
		err := completeTodoTask(ctx, taskId)
		if err != nil {
			fmt.Printf("完成任务失败: %v\n", err)
			return
		}
		fmt.Printf("✓ 任务 %d 已标记为完成\n", taskId)
		return
	case "expire":
		count, err := expireTodoTasks(ctx)
		if err != nil {
			fmt.Printf("检测过期任务失败: %v\n", err)
			return
		}
		if count > 0 {
			fmt.Printf("✓ 检测到 %d 个过期任务并已标记\n", count)
		} else {
			fmt.Println("✓ 未发现过期任务")
		}
		return
	default:
		fmt.Printf("未知命令: %s\n", taskType)
		fmt.Println("可用命令: get, create, sort, complete, expire")
	}

}

type TodoTaskRequest struct {
	TaskType     string    `json:"task_type"`
	TaskId       int64     `json:"task_id"`
	TaskData     *TaskData `json:"task_data"`
	TaskDataJson string    `json:"task_data_json"`
}

type TaskData struct {
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Sort        int64  `json:"sort"`      // 数字越大越靠前
	Status      int    `json:"status"`    // 0已删除 1正常 2已完成 3已过期
	ExpireAt    string `json:"expire_at"` // 过期时间，格式：2006-01-02 15:04:05
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at"`
	CompletedAt string `json:"completed_at"`
}

const (
	statusDeleted   = 0
	statusNormal    = 1
	statusCompleted = 2
	statusExpired   = 3
)

func getTodoTaskKey(ctx context.Context) string {
	return "todo:task:map"
}

// 生成任务id
func generateTaskId(ctx context.Context) int64 {
	return cast.ToInt64(time.Now().Format("20060102150405"))
}

// 创建任务
func createTodoTask(ctx context.Context, title, desc string, sort int64, expireAt string) (bool, error) {
	if title == "" {
		return false, fmt.Errorf("title is empty")
	}
	if desc == "" {
		return false, fmt.Errorf("desc is empty")
	}

	// 验证过期时间格式
	if expireAt != "" {
		_, err := time.Parse("2006-01-02 15:04:05", expireAt)
		if err != nil {
			return false, fmt.Errorf("expire_at format error, expected: 2006-01-02 15:04:05")
		}
	}

	// 生成任务id
	taskId := generateTaskId(ctx)
	fmt.Printf("生成任务id: %d\n", taskId)

	taskData := &TaskData{
		Title:     title,
		Desc:      desc,
		Sort:      sort,
		Status:    statusNormal,
		ExpireAt:  expireAt,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return false, fmt.Errorf("json marshal error: %v", err)
	}

	// 更新redis(hset)
	res, err := webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return res == 1, err
}

// 获取所有任务
func getTodoTaskMap(ctx context.Context) (map[int64]TaskData, error) {
	taskKey := getTodoTaskKey(ctx)
	taskDataJson, err := webcache.RedisClient.HGetAll(ctx, taskKey).Result()
	if err != nil {
		return nil, err
	}
	taskMap := make(map[int64]TaskData)
	for taskIdStr, taskDataJson := range taskDataJson {
		taskId := cast.ToInt64(taskIdStr)
		taskData := TaskData{}
		if err := json.Unmarshal([]byte(taskDataJson), &taskData); err != nil {
			continue
		}
		taskMap[taskId] = taskData
	}
	return taskMap, nil
}

// 获取任务详情
func getTodoTask(ctx context.Context, taskId int64) (*TaskData, error) {
	if taskId == 0 {
		return nil, fmt.Errorf("task_id is empty")
	}
	taskMap, err := getTodoTaskMap(ctx)
	if err != nil {
		return nil, err
	}
	taskData, exists := taskMap[taskId]
	if !exists {
		return nil, fmt.Errorf("task_id %d not found", taskId)
	}
	return &taskData, nil
}

// 更新任务排序值
func updateTaskSort(ctx context.Context, taskId int64, sort int64) error {
	if taskId == 0 {
		return fmt.Errorf("task_id is empty")
	}

	taskData, err := getTodoTask(ctx, taskId)
	if err != nil {
		return err
	}

	// 更新排序值
	taskData.Sort = sort
	taskData.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return err
}

// 标记任务完成
func completeTodoTask(ctx context.Context, taskId int64) error {
	if taskId == 0 {
		return fmt.Errorf("task_id is empty")
	}

	taskData, err := getTodoTask(ctx, taskId)
	if err != nil {
		return err
	}

	// 检查任务状态
	if taskData.Status == statusDeleted {
		return fmt.Errorf("task has been deleted")
	}
	if taskData.Status == statusCompleted {
		return fmt.Errorf("task already completed")
	}

	// 更新状态为完成
	taskData.Status = statusCompleted
	taskData.CompletedAt = time.Now().Format("2006-01-02 15:04:05")
	taskData.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	taskDataJson, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
	return err
}

// 检测并标记过期任务
func expireTodoTasks(ctx context.Context) (int, error) {
	taskMap, err := getTodoTaskMap(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	expireCount := 0

	for taskId, taskData := range taskMap {
		// 只检查正常状态的任务
		if taskData.Status != statusNormal {
			continue
		}

		// 检查是否有过期时间
		if taskData.ExpireAt == "" {
			continue
		}

		// 解析过期时间
		expireTime, err := time.Parse("2006-01-02 15:04:05", taskData.ExpireAt)
		if err != nil {
			continue
		}

		// 如果已过期，更新状态
		if now.After(expireTime) {
			taskData.Status = statusExpired
			taskData.UpdatedAt = now.Format("2006-01-02 15:04:05")

			taskDataJson, err := json.Marshal(taskData)
			if err != nil {
				continue
			}

			_, err = webcache.RedisClient.HSet(ctx, getTodoTaskKey(ctx), taskId, string(taskDataJson)).Result()
			if err == nil {
				expireCount++
			}
		}
	}

	return expireCount, nil
}

// 获取状态文本
func getStatusText(status int) string {
	switch status {
	case statusDeleted:
		return "已删除"
	case statusNormal:
		return "正常"
	case statusCompleted:
		return "已完成"
	case statusExpired:
		return "已过期"
	default:
		return "未知"
	}
}

// 打印任务信息
func printTask(taskId int64, taskData *TaskData) {
	fmt.Println("========================================")
	fmt.Printf("任务ID:     %d\n", taskId)
	fmt.Printf("标题:       %s\n", taskData.Title)
	fmt.Printf("描述:       %s\n", taskData.Desc)
	fmt.Printf("状态:       %s\n", getStatusText(taskData.Status))
	fmt.Printf("排序值:     %d\n", taskData.Sort)
	if taskData.ExpireAt != "" {
		fmt.Printf("过期时间:   %s\n", taskData.ExpireAt)
	}
	fmt.Printf("创建时间:   %s\n", taskData.CreatedAt)
	fmt.Printf("更新时间:   %s\n", taskData.UpdatedAt)
	if taskData.CompletedAt != "" {
		fmt.Printf("完成时间:   %s\n", taskData.CompletedAt)
	}
	fmt.Println("========================================")
}
