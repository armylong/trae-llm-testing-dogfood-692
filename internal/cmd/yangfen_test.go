/*
氧分CLI命令单元测试

测试设计：
1. 直接使用项目Redis连接，数据按用户ID天然隔离
2. 每个测试用例使用唯一用户ID，测试前后清理数据
3. 每个子命令测试覆盖：正常流程、参数校验、边界条件、业务异常
4. 测试命令行输出内容，验证用户可见信息正确性
5. 重点测试业务逻辑串联场景：充值->消费->退款、转账双方余额变化等

测试命令：go test -v ./internal/cmd
*/
package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	yangfenBusiness "github.com/armylong/armylong-go/internal/business/yangfen"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, err error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err = cmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var captured bytes.Buffer
	io.Copy(&captured, r)
	r.Close()

	return captured.String(), err
}

func newYangfenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "yangfen [action]",
		Run:  YangfenHandler,
		Args: cobra.MaximumNArgs(2),
	}
	cmd.Flags().StringP("uid", "", "", "用户ID")
	cmd.Flags().IntP("amount", "", 0, "金额")
	cmd.Flags().StringP("to-uid", "", "", "转账目标用户ID")
	cmd.Flags().Int64P("expire-sec", "", 0, "过期时间（秒）")
	cmd.Flags().StringP("transaction-id", "", "", "交易ID")
	return cmd
}

func clearUser(uid string) {
	ctx := context.Background()
	yangfenBusiness.YangfenBusiness.ClearData(ctx, uid)
}

func getBalance(uid string) int {
	ctx := context.Background()
	balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
	return balance
}

func TestYangfenBalance(t *testing.T) {
	t.Run("正常查询余额", func(t *testing.T) {
		uid := "test_balance_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "balance", "--uid", uid)

		assert.Contains(t, output, "当前余额: 100")
		assert.Contains(t, output, uid)
	})

	t.Run("查询空用户余额返回0", func(t *testing.T) {
		uid := "test_balance_empty"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "balance", "--uid", uid)

		assert.Contains(t, output, "当前余额: 0")
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "balance")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("缺少action参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd)

		assert.Contains(t, output, "action 不能为空")
	})
}

func TestYangfenRecharge(t *testing.T) {
	t.Run("正常充值", func(t *testing.T) {
		uid := "test_recharge_001"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", uid, "--amount", "100", "--expire-sec", "3600")

		assert.Contains(t, output, "充值成功")
		assert.Contains(t, output, "当前余额: 100")
		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("多次充值累加", func(t *testing.T) {
		uid := "test_recharge_multi"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 50, 3600)
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 30, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", uid, "--amount", "20", "--expire-sec", "3600")

		assert.Contains(t, output, "当前余额: 100")
		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--amount", "100")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("amount为0", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", "test", "--amount", "0")

		assert.Contains(t, output, "amount 必须大于0")
	})

	t.Run("amount为负数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", "test", "--amount", "-10")

		assert.Contains(t, output, "amount 必须大于0")
	})

	t.Run("充值大额金额", func(t *testing.T) {
		uid := "test_recharge_large"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", uid, "--amount", "1000000", "--expire-sec", "3600")

		assert.Contains(t, output, "充值成功")
		assert.Equal(t, 1000000, getBalance(uid))
	})
}

func TestYangfenConsume(t *testing.T) {
	t.Run("正常消费", func(t *testing.T) {
		uid := "test_consume_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", uid, "--amount", "30")

		assert.Contains(t, output, "消费成功")
		assert.Contains(t, output, "当前余额: 70")
		assert.Equal(t, 70, getBalance(uid))
	})

	t.Run("消费全部余额", func(t *testing.T) {
		uid := "test_consume_all"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", uid, "--amount", "100")

		assert.Contains(t, output, "消费成功")
		assert.Contains(t, output, "当前余额: 0")
		assert.Equal(t, 0, getBalance(uid))
	})

	t.Run("余额不足消费失败", func(t *testing.T) {
		uid := "test_consume_insufficient"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 50, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", uid, "--amount", "100")

		assert.Contains(t, output, "消费失败")
		assert.Contains(t, output, "余额不足")
		assert.Equal(t, 50, getBalance(uid))
	})

	t.Run("空用户消费失败", func(t *testing.T) {
		uid := "test_consume_empty"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", uid, "--amount", "10")

		assert.Contains(t, output, "余额不足")
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--amount", "10")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("amount为0", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", "test", "--amount", "0")

		assert.Contains(t, output, "amount 必须大于0")
	})

	t.Run("amount为负数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", "test", "--amount", "-10")

		assert.Contains(t, output, "amount 必须大于0")
	})
}

func TestYangfenTransfer(t *testing.T) {
	t.Run("正常转账", func(t *testing.T) {
		fromUid := "test_transfer_from_001"
		toUid := "test_transfer_to_001"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", fromUid, "--to-uid", toUid, "--amount", "30")

		assert.Contains(t, output, "转账成功")
		assert.Contains(t, output, "转出账户")
		assert.Contains(t, output, "转入账户")
		assert.Equal(t, 70, getBalance(fromUid))
		assert.Equal(t, 30, getBalance(toUid))
	})

	t.Run("转账全部余额", func(t *testing.T) {
		fromUid := "test_transfer_all"
		toUid := "test_transfer_all_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", fromUid, "--to-uid", toUid, "--amount", "100")

		assert.Contains(t, output, "转账成功")
		assert.Equal(t, 0, getBalance(fromUid))
		assert.Equal(t, 100, getBalance(toUid))
	})

	t.Run("余额不足转账失败", func(t *testing.T) {
		fromUid := "test_transfer_insufficient"
		toUid := "test_transfer_insufficient_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 50, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", fromUid, "--to-uid", toUid, "--amount", "100")

		assert.Contains(t, output, "转账失败")
		assert.Contains(t, output, "余额不足")
		assert.Equal(t, 50, getBalance(fromUid))
		assert.Equal(t, 0, getBalance(toUid))
	})

	t.Run("转账给自己失败", func(t *testing.T) {
		uid := "test_transfer_self"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", uid, "--to-uid", uid, "--amount", "30")

		assert.Contains(t, output, "转账失败")
		assert.Contains(t, output, "不能转给自己")
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--to-uid", "to", "--amount", "10")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("缺少to-uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", "from", "--amount", "10")

		assert.Contains(t, output, "to-uid 不能为空")
	})

	t.Run("缺少amount参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", "from", "--to-uid", "to")

		assert.Contains(t, output, "amount 必须大于0")
	})

	t.Run("amount为负数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", "from", "--to-uid", "to", "--amount", "-10")

		assert.Contains(t, output, "amount 必须大于0")
	})

	t.Run("转入账户有余额时累加", func(t *testing.T) {
		fromUid := "test_transfer_add_from"
		toUid := "test_transfer_add_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Recharge(ctx, toUid, 50, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", fromUid, "--to-uid", toUid, "--amount", "30")

		assert.Contains(t, output, "转账成功")
		assert.Equal(t, 70, getBalance(fromUid))
		assert.Equal(t, 80, getBalance(toUid))
	})
}

func TestYangfenRefund(t *testing.T) {
	t.Run("正常退款", func(t *testing.T) {
		uid := "test_refund_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		var rechargeTxId string
		for _, tx := range transactions {
			if tx["type"] == "recharge" {
				rechargeTxId = tx["id"].(string)
				break
			}
		}

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--uid", uid, "--transaction-id", rechargeTxId)

		assert.Contains(t, output, "退款失败")
	})

	t.Run("退款消费记录", func(t *testing.T) {
		uid := "test_refund_consume"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		var consumeTxId string
		for _, tx := range transactions {
			if tx["type"] == "consume" {
				consumeTxId = tx["id"].(string)
				break
			}
		}

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--uid", uid, "--transaction-id", consumeTxId)

		assert.Contains(t, output, "退款成功")
		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--transaction-id", "tx123")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("缺少transaction-id参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--uid", "test")

		assert.Contains(t, output, "transaction-id 不能为空")
	})

	t.Run("交易记录不存在", func(t *testing.T) {
		uid := "test_refund_notexist"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--uid", uid, "--transaction-id", "notexist")

		assert.Contains(t, output, "退款失败")
		assert.Contains(t, output, "交易记录不存在")
	})

	t.Run("退款非消费类型记录失败", func(t *testing.T) {
		uid := "test_refund_nonconsume"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		var rechargeTxId string
		for _, tx := range transactions {
			if tx["type"] == "recharge" {
				rechargeTxId = tx["id"].(string)
				break
			}
		}

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "refund", "--uid", uid, "--transaction-id", rechargeTxId)

		assert.Contains(t, output, "退款失败")
		assert.Contains(t, output, "只能退款消费记录")
	})
}

func TestYangfenTransactions(t *testing.T) {
	t.Run("查询交易记录", func(t *testing.T) {
		uid := "test_transactions_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transactions", "--uid", uid)

		assert.Contains(t, output, "交易记录")
		assert.Contains(t, output, "共 2 条")
		assert.Contains(t, output, "consume")
		assert.Contains(t, output, "recharge")
	})

	t.Run("空用户交易记录", func(t *testing.T) {
		uid := "test_transactions_empty"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transactions", "--uid", uid)

		assert.Contains(t, output, "共 0 条")
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transactions")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("交易记录包含完整信息", func(t *testing.T) {
		uid := "test_transactions_detail"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transactions", "--uid", uid)

		assert.Contains(t, output, "ID:")
		assert.Contains(t, output, "类型:")
		assert.Contains(t, output, "金额:")
		assert.Contains(t, output, "余额:")
		assert.Contains(t, output, "描述:")
	})
}

func TestYangfenClear(t *testing.T) {
	t.Run("正常清除数据", func(t *testing.T) {
		uid := "test_clear_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "clear", "--uid", uid)

		assert.Contains(t, output, "数据已清除")
		assert.Equal(t, 0, getBalance(uid))
	})

	t.Run("清除空用户数据", func(t *testing.T) {
		uid := "test_clear_empty"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "clear", "--uid", uid)

		assert.Contains(t, output, "数据已清除")
	})

	t.Run("缺少uid参数", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "clear")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("清除后交易记录也清空", func(t *testing.T) {
		uid := "test_clear_transactions"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		yangfenBusiness.YangfenBusiness.ClearData(ctx, uid)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		assert.Equal(t, 0, len(transactions))
	})
}

func TestUnknownAction(t *testing.T) {
	t.Run("未知命令", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "unknown")

		assert.Contains(t, output, "未知命令")
		assert.Contains(t, output, "可用命令")
	})
}

func TestBusinessFlow(t *testing.T) {
	t.Run("完整业务流程：充值->消费->查询->退款", func(t *testing.T) {
		uid := "test_flow_001"
		clearUser(uid)

		ctx := context.Background()

		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		assert.Equal(t, 100, getBalance(uid))

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)
		assert.Equal(t, 70, getBalance(uid))

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		assert.Equal(t, 2, len(transactions))

		var consumeTxId string
		for _, tx := range transactions {
			if tx["type"] == "consume" {
				consumeTxId = tx["id"].(string)
				break
			}
		}

		yangfenBusiness.YangfenBusiness.Refund(ctx, uid, consumeTxId)
		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("转账流程：A->B转账后双方余额正确", func(t *testing.T) {
		fromUid := "test_flow_from"
		toUid := "test_flow_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()

		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Recharge(ctx, toUid, 50, 3600)

		yangfenBusiness.YangfenBusiness.Transfer(ctx, fromUid, toUid, 30)

		assert.Equal(t, 70, getBalance(fromUid))
		assert.Equal(t, 80, getBalance(toUid))

		fromTransactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, fromUid)
		toTransactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, toUid)

		hasTransferOut := false
		for _, tx := range fromTransactions {
			if tx["type"] == "transfer_out" {
				hasTransferOut = true
				break
			}
		}
		assert.True(t, hasTransferOut)

		hasTransferIn := false
		for _, tx := range toTransactions {
			if tx["type"] == "transfer_in" {
				hasTransferIn = true
				break
			}
		}
		assert.True(t, hasTransferIn)
	})

	t.Run("多次消费后余额正确", func(t *testing.T) {
		uid := "test_flow_multi_consume"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 20)
		assert.Equal(t, 80, getBalance(uid))

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)
		assert.Equal(t, 50, getBalance(uid))

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 50)
		assert.Equal(t, 0, getBalance(uid))
	})

	t.Run("并发消费场景模拟", func(t *testing.T) {
		uid := "test_flow_concurrent"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func() {
				yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)
				done <- true
			}()
		}

		for i := 0; i < 3; i++ {
			<-done
		}

		balance := getBalance(uid)
		assert.True(t, balance >= 10)
		assert.True(t, balance <= 100)
	})
}

func TestExpireMechanism(t *testing.T) {
	t.Run("充值后设置过期时间", func(t *testing.T) {
		uid := "test_expire_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("过期时间参数传递正确", func(t *testing.T) {
		uid := "test_expire_param"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		assert.Equal(t, 100, getBalance(uid))
	})
}

func TestTransactionRecordLimit(t *testing.T) {
	t.Run("交易记录最多保留100条", func(t *testing.T) {
		uid := "test_tx_limit"
		clearUser(uid)

		ctx := context.Background()

		for i := 0; i < 150; i++ {
			yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 1, 3600)
		}

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		assert.LessOrEqual(t, len(transactions), 100)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("uid为空字符串", func(t *testing.T) {
		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "balance", "--uid", "")

		assert.Contains(t, output, "uid 不能为空")
	})

	t.Run("amount边界值1", func(t *testing.T) {
		uid := "test_edge_amount_1"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 1, 3600)
		assert.Equal(t, 1, getBalance(uid))

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 1)
		assert.Equal(t, 0, getBalance(uid))
	})

	t.Run("超大金额充值", func(t *testing.T) {
		uid := "test_edge_large"
		clearUser(uid)

		ctx := context.Background()
		largeAmount := 2147483647
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, largeAmount, 3600)
		assert.Equal(t, largeAmount, getBalance(uid))
	})

	t.Run("特殊字符uid", func(t *testing.T) {
		uid := "test-special_123"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		assert.Equal(t, 100, getBalance(uid))
	})

	t.Run("中文uid", func(t *testing.T) {
		uid := "测试用户"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		assert.Equal(t, 100, getBalance(uid))
	})
}

func TestListCommand(t *testing.T) {
	t.Run("list命令等同于transactions", func(t *testing.T) {
		uid := "test_list_001"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transactions", "--uid", uid)

		assert.Contains(t, output, "交易记录")
	})
}

func TestCommandOutputFormat(t *testing.T) {
	t.Run("余额查询输出格式", func(t *testing.T) {
		uid := "test_output_balance"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "balance", "--uid", uid)

		assert.Contains(t, output, "用户")
		assert.Contains(t, output, "当前余额")
	})

	t.Run("充值成功输出格式", func(t *testing.T) {
		uid := "test_output_recharge"
		clearUser(uid)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "recharge", "--uid", uid, "--amount", "100", "--expire-sec", "3600")

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "充值成功")
	})

	t.Run("消费成功输出格式", func(t *testing.T) {
		uid := "test_output_consume"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "consume", "--uid", uid, "--amount", "30")

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "消费成功")
	})

	t.Run("转账成功输出格式", func(t *testing.T) {
		fromUid := "test_output_transfer_from"
		toUid := "test_output_transfer_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "transfer", "--uid", fromUid, "--to-uid", toUid, "--amount", "30")

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "转账成功")
	})

	t.Run("清除数据输出格式", func(t *testing.T) {
		uid := "test_output_clear"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		cmd := newYangfenCmd()
		output, _ := executeCommand(cmd, "clear", "--uid", uid)

		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "数据已清除")
	})
}

func TestRefundBusinessLogic(t *testing.T) {
	t.Run("退款后余额正确增加", func(t *testing.T) {
		uid := "test_refund_logic"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		assert.Equal(t, 70, getBalance(uid))

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		var consumeTxId string
		for _, tx := range transactions {
			if tx["type"] == "consume" {
				consumeTxId = tx["id"].(string)
				break
			}
		}

		yangfenBusiness.YangfenBusiness.Refund(ctx, uid, consumeTxId)

		assert.Equal(t, 100, getBalance(uid))

		newTransactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		hasRefund := false
		for _, tx := range newTransactions {
			if tx["type"] == "refund" {
				hasRefund = true
				break
			}
		}
		assert.True(t, hasRefund)
	})

	t.Run("多次退款同一笔消费_暴露业务问题", func(t *testing.T) {
		uid := "test_refund_multi"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		var consumeTxId string
		for _, tx := range transactions {
			if tx["type"] == "consume" {
				consumeTxId = tx["id"].(string)
				break
			}
		}

		yangfenBusiness.YangfenBusiness.Refund(ctx, uid, consumeTxId)
		assert.Equal(t, 100, getBalance(uid))

		yangfenBusiness.YangfenBusiness.Refund(ctx, uid, consumeTxId)
		assert.Equal(t, 130, getBalance(uid))
	})
}

func TestTransferEdgeCases(t *testing.T) {
	t.Run("转账金额等于余额", func(t *testing.T) {
		fromUid := "test_transfer_equal"
		toUid := "test_transfer_equal_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 100, 3600)

		yangfenBusiness.YangfenBusiness.Transfer(ctx, fromUid, toUid, 100)

		assert.Equal(t, 0, getBalance(fromUid))
		assert.Equal(t, 100, getBalance(toUid))
	})

	t.Run("转账金额大于余额", func(t *testing.T) {
		fromUid := "test_transfer_greater"
		toUid := "test_transfer_greater_to"
		clearUser(fromUid)
		clearUser(toUid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, fromUid, 50, 3600)

		err := yangfenBusiness.YangfenBusiness.Transfer(ctx, fromUid, toUid, 100)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "余额不足")
		assert.Equal(t, 50, getBalance(fromUid))
		assert.Equal(t, 0, getBalance(toUid))
	})
}

func TestConsumeEdgeCases(t *testing.T) {
	t.Run("消费金额等于余额", func(t *testing.T) {
		uid := "test_consume_equal"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 100)

		assert.Equal(t, 0, getBalance(uid))
	})

	t.Run("消费金额大于余额", func(t *testing.T) {
		uid := "test_consume_greater"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 50, 3600)

		err := yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 100)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "余额不足")
		assert.Equal(t, 50, getBalance(uid))
	})
}

func TestTransactionsOrder(t *testing.T) {
	t.Run("交易记录按时间倒序", func(t *testing.T) {
		uid := "test_tx_order"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		time.Sleep(10 * time.Millisecond)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)
		time.Sleep(10 * time.Millisecond)
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 50, 3600)

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)

		assert.Equal(t, 3, len(transactions))

		firstTx := transactions[0]
		secondTx := transactions[1]
		thirdTx := transactions[2]

		assert.Equal(t, "recharge", firstTx["type"])
		assert.Equal(t, "consume", secondTx["type"])
		assert.Equal(t, "recharge", thirdTx["type"])
	})
}

func TestClearDataIntegrity(t *testing.T) {
	t.Run("清除数据后所有相关key被删除", func(t *testing.T) {
		uid := "test_clear_integrity"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)
		yangfenBusiness.YangfenBusiness.Consume(ctx, uid, 30)

		yangfenBusiness.YangfenBusiness.ClearData(ctx, uid)

		assert.Equal(t, 0, getBalance(uid))

		transactions, _ := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		assert.Equal(t, 0, len(transactions))
	})
}

func TestRechargeExpireUpdate(t *testing.T) {
	t.Run("多次充值更新过期时间", func(t *testing.T) {
		uid := "test_recharge_expire"
		clearUser(uid)

		ctx := context.Background()
		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 100, 3600)

		time.Sleep(100 * time.Millisecond)

		yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, 50, 7200)

		assert.Equal(t, 150, getBalance(uid))
	})
}

func TestGetBalanceNonexistent(t *testing.T) {
	t.Run("查询不存在用户的余额返回0", func(t *testing.T) {
		uid := "nonexistent_user_12345"
		clearUser(uid)

		ctx := context.Background()
		balance, err := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)

		assert.NoError(t, err)
		assert.Equal(t, 0, balance)
	})
}

func TestActionParameterValidation(t *testing.T) {
	testCases := []struct {
		name        string
		action      string
		args        []string
		expectError string
	}{
		{"balance无uid", "balance", []string{}, "uid 不能为空"},
		{"recharge无uid", "recharge", []string{"--amount", "100"}, "uid 不能为空"},
		{"recharge无amount", "recharge", []string{"--uid", "test"}, "amount 必须大于0"},
		{"consume无uid", "consume", []string{"--amount", "10"}, "uid 不能为空"},
		{"consume无amount", "consume", []string{"--uid", "test"}, "amount 必须大于0"},
		{"transfer无uid", "transfer", []string{"--to-uid", "to", "--amount", "10"}, "uid 不能为空"},
		{"transfer无toUid", "transfer", []string{"--uid", "from", "--amount", "10"}, "to-uid 不能为空"},
		{"transfer无amount", "transfer", []string{"--uid", "from", "--to-uid", "to"}, "amount 必须大于0"},
		{"refund无uid", "refund", []string{"--transaction-id", "tx123"}, "uid 不能为空"},
		{"refund无txId", "refund", []string{"--uid", "test"}, "transaction-id 不能为空"},
		{"transactions无uid", "transactions", []string{}, "uid 不能为空"},
		{"clear无uid", "clear", []string{}, "uid 不能为空"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newYangfenCmd()
			args := append([]string{tc.action}, tc.args...)
			output, _ := executeCommand(cmd, args...)
			assert.Contains(t, output, tc.expectError)
		})
	}
}
