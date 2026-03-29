package cmd

import (
	"fmt"

	yangfenBusiness "github.com/armylong/armylong-go/internal/business/yangfen"
	"github.com/spf13/cobra"
)

func YangfenHandler(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	action := ""
	if len(args) > 0 {
		action = args[0]
	}

	uid, _ := cmd.Flags().GetString("uid")
	amount, _ := cmd.Flags().GetInt("amount")
	toUid, _ := cmd.Flags().GetString("to-uid")
	expireSec, _ := cmd.Flags().GetInt64("expire-sec")
	transactionId, _ := cmd.Flags().GetString("transaction-id")

	if action == "" {
		fmt.Println("错误: action 不能为空")
		fmt.Println("可用命令: balance, recharge, consume, transfer, refund, transactions, clear")
		return
	}

	switch action {
	case "balance": // 查询余额
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		balance, err := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		if err != nil {
			fmt.Printf("查询余额失败: %v\n", err)
			return
		}
		fmt.Printf("用户 %s 当前余额: %d\n", uid, balance)
		return

	case "recharge": // 充值
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return
		}
		err := yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, amount, expireSec)
		if err != nil {
			fmt.Printf("充值失败: %v\n", err)
			return
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 充值成功，当前余额: %d\n", balance)
		return

	case "consume": // 消费
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return
		}
		err := yangfenBusiness.YangfenBusiness.Consume(ctx, uid, amount)
		if err != nil {
			fmt.Printf("消费失败: %v\n", err)
			return
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 消费成功，当前余额: %d\n", balance)
		return

	case "transfer": // 转账
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		if toUid == "" {
			fmt.Println("错误: to-uid 不能为空")
			return
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return
		}
		err := yangfenBusiness.YangfenBusiness.Transfer(ctx, uid, toUid, amount)
		if err != nil {
			fmt.Printf("转账失败: %v\n", err)
			return
		}
		fromBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		toBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, toUid)
		fmt.Printf("✓ 转账成功\n")
		fmt.Printf("  转出账户 %s 余额: %d\n", uid, fromBalance)
		fmt.Printf("  转入账户 %s 余额: %d\n", toUid, toBalance)
		return

	case "refund": // 退款
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		if transactionId == "" {
			fmt.Println("错误: transaction-id 不能为空")
			return
		}
		err := yangfenBusiness.YangfenBusiness.Refund(ctx, uid, transactionId)
		if err != nil {
			fmt.Printf("退款失败: %v\n", err)
			return
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 退款成功，当前余额: %d\n", balance)
		return

	case "transactions": // 查看交易记录
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		transactions, err := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		if err != nil {
			fmt.Printf("获取交易记录失败: %v\n", err)
			return
		}
		fmt.Printf("用户 %s 交易记录 (共 %d 条):\n", uid, len(transactions))
		fmt.Println("========================================")
		for i, t := range transactions {
			fmt.Printf("[%d] ID: %v\n", i+1, t["id"])
			fmt.Printf("    类型: %v\n", t["type"])
			fmt.Printf("    金额: %v\n", t["amount"])
			fmt.Printf("    余额: %v\n", t["balance"])
			fmt.Printf("    描述: %v\n", t["description"])
			fmt.Println("----------------------------------------")
		}
		return

	case "clear": // 清除数据
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return
		}
		err := yangfenBusiness.YangfenBusiness.ClearData(ctx, uid)
		if err != nil {
			fmt.Printf("清除数据失败: %v\n", err)
			return
		}
		fmt.Printf("✓ 用户 %s 数据已清除\n", uid)
		return

	default:
		fmt.Printf("未知命令: %s\n", action)
		fmt.Println("可用命令: balance, recharge, consume, transfer, refund, transactions, clear")
	}
}
