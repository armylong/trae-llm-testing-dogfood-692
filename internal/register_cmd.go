package internal // 包名=目录名，适配internal/根目录

import (
	"github.com/armylong/armylong-go/internal/cmd" // 导入handler所在包
	"github.com/armylong/go-library/service/command"
	"github.com/spf13/cobra"
)

// RegisterCmd 集中注册所有子命令（修正所有Cobra语法错误）
func RegisterCmd(command command.BaseCommand) {
	command.AddCliCommand(&cobra.Command{
		Use:   "demo username message age [hobby]", // Cobra核心：Use定义命令名+使用方式
		Short: "演示参数接收",
		Args:  cobra.MaximumNArgs(5),
		Run:   cmd.DemoHandler, // 调用internal/cmd下的执行逻辑
	})

	todoCmd := &cobra.Command{
		Use:   "todo [task_type]", // 替换Name→Use
		Short: "任务管理",
		Run:   cmd.TodoHandler,
	}
	todoCmd.Flags().Int64P("task_id", "", 0, "任务ID（可选）")
	todoCmd.Flags().StringP("title", "", "", "任务标题（create时必填）")
	todoCmd.Flags().StringP("desc", "", "", "任务描述（create时必填）")
	todoCmd.Flags().Int64P("sort", "", 0, "任务排序值，数字越大越靠前（可选）")
	todoCmd.Flags().StringP("expire_at", "", "", "过期时间，格式：2006-01-02 15:04:05（可选）")
	command.AddCliCommand(todoCmd)

	// -----------------------------------------------------------------------------------------------------------------
	yangfenCmd := &cobra.Command{
		Use:   "yangfen [action]",
		Short: "氧分管理",
		Run:   cmd.YangfenHandler,
	}
	yangfenCmd.Flags().StringP("uid", "", "", "用户ID")
	yangfenCmd.Flags().IntP("amount", "", 0, "金额")
	yangfenCmd.Flags().StringP("to-uid", "", "", "转账目标用户ID")
	yangfenCmd.Flags().Int64P("expire-sec", "", 0, "过期时间（秒）")
	yangfenCmd.Flags().StringP("transaction-id", "", "", "交易ID")
	command.AddCliCommand(yangfenCmd)
}
