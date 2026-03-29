package cmd

/*
	氧分CLI命令测试套件
	测试范围：balance、recharge、consume、transfer、refund、transactions、clear
	测试覆盖：正常流程、异常流程、边界场景、并发操作

	测试用例列表：
	1. 基础命令测试
	   - TestYangfen_NoAction: 无action参数
	   - TestYangfen_UnknownAction: 未知action

	2. balance子命令
	   - TestBalance_WithoutUID: 缺少uid参数
	   - TestBalance_NewUser: 新用户余额查询(0余额)

	3. recharge子命令
	   - TestRecharge_Success: 充值成功
	   - TestRecharge_WithoutUID: 缺少uid参数
	   - TestRecharge_InvalidAmount: 金额不合法(≤0)
	   - TestRecharge_MultipleTimes: 多次充值累加

	4. consume子命令
	   - TestConsume_Success: 消费成功
	   - TestConsume_WithoutUID: 缺少uid参数
	   - TestConsume_InvalidAmount: 金额不合法(≤0)
	   - TestConsume_InsufficientBalance: 余额不足
	   - TestConsume_ExactBalance: 余额刚好用完

	5. transfer子命令
	   - TestTransfer_Success: 转账成功
	   - TestTransfer_WithoutFromUID: 缺少转出用户
	   - TestTransfer_WithoutToUID: 缺少转入用户
	   - TestTransfer_SelfTransfer: 不能转给自己
	   - TestTransfer_InvalidAmount: 金额不合法(≤0)
	   - TestTransfer_InsufficientBalance: 余额不足

	6. refund子命令
	   - TestRefund_WithoutUID: 缺少uid参数
	   - TestRefund_WithoutTransactionID: 缺少交易ID
	   - TestRefund_TransactionNotFound: 交易不存在
	   - TestRefund_Valid: 有效退款
	   - TestRefund_NotConsume: 只能退款消费记录
	   - TestRefund_Duplicate: 重复退款漏洞检测

	7. transactions子命令
	   - TestTransactions_WithoutUID: 缺少uid参数
	   - TestTransactions_NewUser: 新用户无记录
	   - TestTransactions_WithRecords: 有交易记录

	8. clear子命令
	   - TestClear_WithoutUID: 缺少uid参数
	   - TestClear_Success: 清除数据成功

	9. 集成测试
	   - TestFullWorkflow: 完整业务流程
	   - TestConcurrentOperations: 并发操作数据一致性
	   - TestEdgeCase_LargeAmount: 大额充值消费
	   - TestEdgeCase_AmountOne: 最小金额(1)测试
*/

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/armylong/armylong-go/internal/business/yangfen"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// 测试用户ID
const (
	testUID1 = "test_user_001"
	testUID2 = "test_user_002"
	testUID3 = "test_user_003"
)

// setupTest 初始化测试环境，清除测试数据
func setupTest(t *testing.T) {
	ctx := context.Background()
	_ = yangfen.YangfenBusiness.ClearData(ctx, testUID1)
	_ = yangfen.YangfenBusiness.ClearData(ctx, testUID2)
	_ = yangfen.YangfenBusiness.ClearData(ctx, testUID3)
}

// executeCommand 执行Cobra命令并捕获输出（需重定向stdout）
func executeCommand(args ...string) (string, error) {
	// 创建根命令
	rootCmd := &cobra.Command{Use: "test"}
	yangfenCmd := &cobra.Command{
		Use:   "yangfen [action]",
		Short: "氧分管理",
		Run:   YangfenHandler,
	}
	yangfenCmd.Flags().StringP("uid", "", "", "用户ID")
	yangfenCmd.Flags().IntP("amount", "", 0, "金额")
	yangfenCmd.Flags().StringP("to-uid", "", "", "转账目标用户ID")
	yangfenCmd.Flags().Int64P("expire-sec", "", 0, "过期时间（秒）")
	yangfenCmd.Flags().StringP("transaction-id", "", "", "交易ID")
	rootCmd.AddCommand(yangfenCmd)

	// 由于YangfenHandler使用fmt.Println直接输出到stdout，需要重定向stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	// 恢复stdout
	w.Close()
	os.Stdout = oldStdout

	// 读取输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

// TestYangfen_NoAction 测试无action参数场景
func TestYangfen_NoAction(t *testing.T) {
	output, _ := executeCommand("yangfen")
	assert.Contains(t, output, "错误: action 不能为空")
	assert.Contains(t, output, "可用命令: balance, recharge, consume, transfer, refund, transactions, clear")
}

// TestYangfen_UnknownAction 测试未知action
func TestYangfen_UnknownAction(t *testing.T) {
	output, _ := executeCommand("yangfen", "unknown")
	assert.Contains(t, output, "未知命令: unknown")
}

// TestBalance_WithoutUID 测试查询余额-缺少uid参数
func TestBalance_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "balance")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestBalance_NewUser 测试查询余额-新用户（余额为0）
func TestBalance_NewUser(t *testing.T) {
	setupTest(t)
	output, _ := executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 0", testUID1))
}

// TestRecharge_Success 测试充值-成功场景
func TestRecharge_Success(t *testing.T) {
	setupTest(t)
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 100")
}

// TestRecharge_WithoutUID 测试充值-缺少uid参数
func TestRecharge_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "recharge", "--amount", "100")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestRecharge_InvalidAmount 测试充值-金额不合法
func TestRecharge_InvalidAmount(t *testing.T) {
	// 金额为0
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "0")
	assert.Contains(t, output, "错误: amount 必须大于0")

	// 金额为负数
	output, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "-10")
	assert.Contains(t, output, "错误: amount 必须大于0")
}

// TestRecharge_MultipleTimes 测试充值-多次充值累加
func TestRecharge_MultipleTimes(t *testing.T) {
	setupTest(t)
	// 第一次充值
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 100")

	// 第二次充值
	output, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "50", "--expire-sec", "3600")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 150")
}

// TestConsume_Success 测试消费-成功场景
func TestConsume_Success(t *testing.T) {
	setupTest(t)
	// 先充值
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")

	// 再消费
	output, _ := executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "30")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 70")
}

// TestConsume_WithoutUID 测试消费-缺少uid参数
func TestConsume_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "consume", "--amount", "10")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestConsume_InvalidAmount 测试消费-金额不合法
func TestConsume_InvalidAmount(t *testing.T) {
	output, _ := executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "0")
	assert.Contains(t, output, "错误: amount 必须大于0")

	output, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "-5")
	assert.Contains(t, output, "错误: amount 必须大于0")
}

// TestConsume_InsufficientBalance 测试消费-余额不足
func TestConsume_InsufficientBalance(t *testing.T) {
	setupTest(t)
	// 充值50
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "50", "--expire-sec", "3600")

	// 尝试消费100
	output, _ := executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "100")
	assert.Contains(t, output, "消费失败: 余额不足")
}

// TestConsume_ExactBalance 测试消费-余额刚好用完
func TestConsume_ExactBalance(t *testing.T) {
	setupTest(t)
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "50", "--expire-sec", "3600")

	output, _ := executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "50")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 0")
}

// TestTransfer_Success 测试转账-成功场景
func TestTransfer_Success(t *testing.T) {
	setupTest(t)
	// 给用户1充值
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	// 用户2初始余额为0
	_, _ = executeCommand("yangfen", "balance", "--uid", testUID2)

	// 执行转账
	output, _ := executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID2, "--amount", "30")
	assert.Contains(t, output, "✓ 转账成功")
	assert.Contains(t, output, fmt.Sprintf("转出账户 %s 余额: 70", testUID1))
	assert.Contains(t, output, fmt.Sprintf("转入账户 %s 余额: 30", testUID2))
}

// TestTransfer_WithoutFromUID 测试转账-缺少转出用户
func TestTransfer_WithoutFromUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "transfer", "--to-uid", testUID2, "--amount", "10")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestTransfer_WithoutToUID 测试转账-缺少转入用户
func TestTransfer_WithoutToUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "transfer", "--uid", testUID1, "--amount", "10")
	assert.Contains(t, output, "错误: to-uid 不能为空")
}

// TestTransfer_SelfTransfer 测试转账-不能转给自己
func TestTransfer_SelfTransfer(t *testing.T) {
	setupTest(t)
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")

	output, _ := executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID1, "--amount", "10")
	assert.Contains(t, output, "转账失败: 不能转给自己")
}

// TestTransfer_InvalidAmount 测试转账-金额不合法
func TestTransfer_InvalidAmount(t *testing.T) {
	output, _ := executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID2, "--amount", "0")
	assert.Contains(t, output, "错误: amount 必须大于0")

	output, _ = executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID2, "--amount", "-5")
	assert.Contains(t, output, "错误: amount 必须大于0")
}

// TestTransfer_InsufficientBalance 测试转账-余额不足
func TestTransfer_InsufficientBalance(t *testing.T) {
	setupTest(t)
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "50", "--expire-sec", "3600")

	output, _ := executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID2, "--amount", "100")
	assert.Contains(t, output, "转账失败: 余额不足")
}

// TestRefund_WithoutUID 测试退款-缺少uid参数
func TestRefund_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "refund", "--transaction-id", "TX12345")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestRefund_WithoutTransactionID 测试退款-缺少交易ID
func TestRefund_WithoutTransactionID(t *testing.T) {
	output, _ := executeCommand("yangfen", "refund", "--uid", testUID1)
	assert.Contains(t, output, "错误: transaction-id 不能为空")
}

// TestRefund_TransactionNotFound 测试退款-交易不存在
func TestRefund_TransactionNotFound(t *testing.T) {
	setupTest(t)
	output, _ := executeCommand("yangfen", "refund", "--uid", testUID1, "--transaction-id", "TX_NON_EXISTENT")
	assert.Contains(t, output, "退款失败: 交易记录不存在")
}

// TestTransactions_WithoutUID 测试查询交易记录-缺少uid
func TestTransactions_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "transactions")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestTransactions_NewUser 测试查询交易记录-新用户无记录
func TestTransactions_NewUser(t *testing.T) {
	setupTest(t)
	output, _ := executeCommand("yangfen", "transactions", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 交易记录 (共 0 条):", testUID1))
}

// TestTransactions_WithRecords 测试查询交易记录-有交易记录
func TestTransactions_WithRecords(t *testing.T) {
	setupTest(t)
	// 执行充值和消费
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	_, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "30")

	output, _ := executeCommand("yangfen", "transactions", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 交易记录 (共 2 条):", testUID1))
	assert.Contains(t, output, "类型: recharge")
	assert.Contains(t, output, "类型: consume")
}

// TestClear_WithoutUID 测试清除数据-缺少uid
func TestClear_WithoutUID(t *testing.T) {
	output, _ := executeCommand("yangfen", "clear")
	assert.Contains(t, output, "错误: uid 不能为空")
}

// TestClear_Success 测试清除数据-成功
func TestClear_Success(t *testing.T) {
	setupTest(t)
	// 先充值
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")

	// 清除数据
	output, _ := executeCommand("yangfen", "clear", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("✓ 用户 %s 数据已清除", testUID1))

	// 验证余额为0
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 0", testUID1))
}

// TestFullWorkflow 完整业务流程测试
func TestFullWorkflow(t *testing.T) {
	setupTest(t)

	// 1. 用户1充值
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "200", "--expire-sec", "3600")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 200")

	// 2. 用户1消费
	output, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "50")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 150")

	// 3. 用户1转账给用户2
	output, _ = executeCommand("yangfen", "transfer", "--uid", testUID1, "--to-uid", testUID2, "--amount", "70")
	assert.Contains(t, output, "✓ 转账成功")
	assert.Contains(t, output, fmt.Sprintf("转出账户 %s 余额: 80", testUID1))
	assert.Contains(t, output, fmt.Sprintf("转入账户 %s 余额: 70", testUID2))

	// 4. 查询用户1余额
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 80", testUID1))

	// 5. 查询用户2余额
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID2)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 70", testUID2))

	// 6. 查询用户1交易记录（充值、消费、转出）
	output, _ = executeCommand("yangfen", "transactions", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 交易记录 (共 3 条):", testUID1))

	// 7. 查询用户2交易记录（转入）
	output, _ = executeCommand("yangfen", "transactions", "--uid", testUID2)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 交易记录 (共 1 条):", testUID2))
	assert.Contains(t, output, "类型: transfer_in")

	// 8. 用户2消费
	output, _ = executeCommand("yangfen", "consume", "--uid", testUID2, "--amount", "20")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 50")

	// 9. 清除用户1数据
	output, _ = executeCommand("yangfen", "clear", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("✓ 用户 %s 数据已清除", testUID1))

	// 10. 验证用户1数据已清除
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 0", testUID1))

	// 11. 用户2余额不受影响
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID2)
	assert.Contains(t, output, fmt.Sprintf("用户 %s 当前余额: 50", testUID2))
}

// getTransactionID 从交易记录中获取一个消费类型的交易ID
func getConsumeTransactionID(uid string) string {
	ctx := context.Background()
	transactions, _ := yangfen.YangfenBusiness.GetTransactions(ctx, uid)
	for _, t := range transactions {
		if t["type"] == "consume" {
			return fmt.Sprint(t["id"])
		}
	}
	return ""
}

// TestRefund_Valid 测试退款-有效退款场景
func TestRefund_Valid(t *testing.T) {
	setupTest(t)
	// 先充值再消费
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	_, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "30")

	// 获取消费交易ID
	txID := getConsumeTransactionID(testUID1)
	if txID == "" {
		t.Fatal("未能获取消费交易ID")
	}

	// 执行退款
	output, _ := executeCommand("yangfen", "refund", "--uid", testUID1, "--transaction-id", txID)
	assert.Contains(t, output, "✓ 退款成功，当前余额: 100")
}

// TestRefund_NotConsume 测试退款-只能退款消费记录
func TestRefund_NotConsume(t *testing.T) {
	setupTest(t)
	// 充值
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")

	// 获取充值交易ID（非消费类型）
	ctx := context.Background()
	transactions, _ := yangfen.YangfenBusiness.GetTransactions(ctx, testUID1)
	if len(transactions) == 0 {
		t.Fatal("未找到交易记录")
	}
	txID := fmt.Sprint(transactions[0]["id"])

	// 尝试退款充值记录
	output, _ := executeCommand("yangfen", "refund", "--uid", testUID1, "--transaction-id", txID)
	assert.Contains(t, output, "退款失败: 只能退款消费记录")
}

// TestRefund_Duplicate 测试退款-重复退款漏洞检测
// 业务风险：同一笔消费交易如果可以多次退款，会导致用户氧分被非法套取
// Bug检测：此测试用例专门用于验证"同一笔交易不能重复退款"的业务规则
func TestRefund_Duplicate(t *testing.T) {
	setupTest(t)
	// 先充值再消费
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "100", "--expire-sec", "3600")
	_, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "30")

	// 验证消费后余额：100 - 30 = 70
	output, _ := executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, "当前余额: 70")

	// 获取消费交易ID
	txID := getConsumeTransactionID(testUID1)
	if txID == "" {
		t.Fatal("未能获取消费交易ID")
	}

	// 第一次退款 - 应该成功，余额回到 70 + 30 = 100
	output, _ = executeCommand("yangfen", "refund", "--uid", testUID1, "--transaction-id", txID)
	assert.Contains(t, output, "✓ 退款成功，当前余额: 100")

	// 第二次退款 - 同一笔交易，应该失败（防止重复退款漏洞）
	// 如果业务逻辑没有防重机制，退款会再次成功，余额变成 100 + 30 = 130
	output, _ = executeCommand("yangfen", "refund", "--uid", testUID1, "--transaction-id", txID)

	// 核心断言：第二次退款应该失败
	// Bug：如果输出包含"退款成功"，说明存在【重复退款漏洞】
	assert.NotContains(t, output, "✓ 退款成功", "【BUG发现】同一笔交易不能重复退款！余额将异常增加")

	// 辅助验证：余额应该保持100不变
	output, _ = executeCommand("yangfen", "balance", "--uid", testUID1)
	assert.Contains(t, output, "当前余额: 100", "【漏洞确认】重复退款导致余额异常！")
}

// TestConcurrentOperations 并发操作测试 - 验证数据一致性
func TestConcurrentOperations(t *testing.T) {
	setupTest(t)
	_, _ = executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "1000", "--expire-sec", "3600")

	// 并发多次消费
	errChan := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "10")
			errChan <- err
		}()
	}

	// 等待所有操作完成
	for i := 0; i < 10; i++ {
		<-errChan
	}

	// 验证最终余额
	output, _ := executeCommand("yangfen", "balance", "--uid", testUID1)
	// 应该是 1000 - 10*10 = 900
	assert.Contains(t, output, "900")
}

// TestEdgeCase_LargeAmount 边界测试 - 大额充值和消费
func TestEdgeCase_LargeAmount(t *testing.T) {
	setupTest(t)
	// 大额充值
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "1000000", "--expire-sec", "86400")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 1000000")

	// 大额消费
	output, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "999999")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 1")
}

// TestEdgeCase_AmountOne 边界测试 - 最小金额1
func TestEdgeCase_AmountOne(t *testing.T) {
	setupTest(t)
	output, _ := executeCommand("yangfen", "recharge", "--uid", testUID1, "--amount", "1", "--expire-sec", "3600")
	assert.Contains(t, output, "✓ 充值成功，当前余额: 1")

	output, _ = executeCommand("yangfen", "consume", "--uid", testUID1, "--amount", "1")
	assert.Contains(t, output, "✓ 消费成功，当前余额: 0")
}
