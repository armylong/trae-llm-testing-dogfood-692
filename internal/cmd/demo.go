package cmd // 包名=目录名

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// DemoHandler demo命令执行逻辑
// go run main.go demo 张三 -m "测试消息" -e -a 25 -H 篮球 -H 编程
func DemoHandler(cmd *cobra.Command, args []string) {
	username := "匿名用户"
	if len(args) > 0 {
		username = args[0]
	}

	// 读取标志参数
	toggle, _ := cmd.Flags().GetBool("toggle")
	enable, _ := cmd.Flags().GetBool("enable")
	message, _ := cmd.Flags().GetString("message")
	age, _ := cmd.Flags().GetInt("age")
	hobbies, _ := cmd.Flags().GetStringSlice("hobby")

	// 输出结果
	if toggle {
		fmt.Println("===== Toggle模式 =====")
	}
	fmt.Printf("用户：%s\n", username)
	fmt.Printf("功能启用：%t\n", enable)
	fmt.Printf("消息：%s\n", message)
	fmt.Printf("年龄：%d\n", age)
	fmt.Printf("爱好：%s\n", strings.Join(hobbies, ", "))
}

// UvLampStatisticsRecalculateHandler 紫外线灯统计重算逻辑
// go run main.go UvLampStatisticsRecalculate --date 2026-03-22 --tenant-id 1001
func UvLampStatisticsRecalculateHandler(cmd *cobra.Command, args []string) {
	date, _ := cmd.Flags().GetString("date")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")
	tenantID, _ := cmd.Flags().GetInt64("tenant-id")

	fmt.Println("===== 紫外线灯统计重算 =====")
	fmt.Printf("日期：%s\n开始日期：%s\n结束日期：%s\n门店ID：%d\n", date, startDate, endDate, tenantID)
}
