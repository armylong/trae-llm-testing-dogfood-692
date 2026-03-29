package yangfen

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/armylong/armylong-go/internal/common/webcache"
)

// TestClearAllData 清除所有测试数据，初始化环境
// 在运行其他测试前先运行此方法
func TestClearAllData(t *testing.T) {
	ctx := context.Background()

	for i := 1; i <= 20; i++ {
		uid := strconv.Itoa(i)
		webcache.RedisClient.Del(ctx, YangfenBalanceKey+uid)
		webcache.RedisClient.Del(ctx, YangfenTransactionKey+uid)
		webcache.RedisClient.Del(ctx, YangfenExpireKey+uid)
	}

	t.Log("所有测试数据已清除")
}

// TestBug1_TransferConcurrency 转账并发问题测试
// Bug描述: 转账时没有使用事务/锁，并发转账可能导致数据不一致
// 复现步骤:
// 1. 用户1充值100
// 2. 用户1同时向10个用户各转账20（并发执行，总共需要200）
// 3. 预期：只有5笔转账成功，其余因余额不足失败
// 4. 实际：可能多笔转账都成功，导致数据不一致（转出金额超过原始余额）
func TestBug1_TransferConcurrency(t *testing.T) {
	ctx := context.Background()
	uid1 := "1"

	TestClearAllData(t)

	// 用户1充值100
	err := YangfenBusiness.Recharge(ctx, uid1, 100, 100)
	if err != nil {
		t.Fatalf("充值失败: %v", err)
	}

	balance, _ := YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("用户1充值后余额: %d", balance)

	// 10个并发转账：用户1同时向用户2-11各转账20
	transferCount := 10
	transferAmount := 20

	var wg sync.WaitGroup
	errs := make([]error, transferCount)

	wg.Add(transferCount)

	for i := 0; i < transferCount; i++ {
		go func(index int) {
			defer wg.Done()
			toUid := strconv.Itoa(index + 2)
			errs[index] = YangfenBusiness.Transfer(ctx, uid1, toUid, transferAmount)
		}(i)
	}

	wg.Wait()

	// 统计结果
	successCount := 0
	failCount := 0
	for _, e := range errs {
		if e == nil {
			successCount++
		} else {
			failCount++
		}
	}

	// 检查各用户余额
	balance1, _ := YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("用户1最终余额: %d", balance1)

	totalReceived := 0
	for i := 2; i <= 11; i++ {
		uid := strconv.Itoa(i)
		b, _ := YangfenBusiness.GetBalance(ctx, uid)
		if b > 0 {
			t.Logf("用户%s余额: %d", uid, b)
			totalReceived += b
		}
	}

	t.Logf("成功转账: %d笔, 失败: %d笔", successCount, failCount)
	t.Logf("总转出金额: %d (原始余额100, 每笔%d)", totalReceived, transferAmount)

	// Bug复现：如果总转出金额超过原始余额，说明并发问题存在
	if totalReceived > 100 {
		t.Errorf("Bug复现! 总转出金额%d超过原始余额100，数据不一致!", totalReceived)
	}

	// 正确情况：成功5笔，用户1余额=0
	// Bug情况：成功超过5笔，用户1余额可能为负或数据不一致
}

// TestBug2_RefundAfterExpire 退款金额错误测试
// Bug描述: 退款时直接返还原始消费金额，没有考虑充值时的"过期时间"
// 复现步骤:
// 1. 用户1充值100，设置1秒后过期
// 2. 消费50（余额剩50）
// 3. 等待过期，触发过期检查（余额被清零）
// 4. 退款消费记录
// 5. 实际：退款成功，余额变成50
// 6. 问题：这笔积分已经过期了，退款应该失败或特殊处理
func TestBug2_RefundAfterExpire(t *testing.T) {
	ctx := context.Background()
	uid1 := "1"

	TestClearAllData(t)

	// 用户1充值100，设置1秒后过期
	err := YangfenBusiness.Recharge(ctx, uid1, 100, 1)
	if err != nil {
		t.Fatalf("充值失败: %v", err)
	}

	balance, _ := YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("用户1充值后余额: %d (1秒后过期)", balance)

	// 消费50
	err = YangfenBusiness.Consume(ctx, uid1, 50)
	if err != nil {
		t.Fatalf("消费失败: %v", err)
	}

	balance, _ = YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("消费50后余额: %d", balance)

	// 获取交易记录，找到消费记录
	transactions, _ := YangfenBusiness.GetTransactions(ctx, uid1)
	var consumeTxId string
	for _, tx := range transactions {
		if tx["type"] == "consume" {
			consumeTxId = tx["id"].(string)
			break
		}
	}

	if consumeTxId == "" {
		t.Fatal("找不到消费记录")
	}
	t.Logf("消费交易ID: %s", consumeTxId)

	// 等待过期
	time.Sleep(2 * time.Second)
	t.Log("等待2秒，积分已过期")

	// 触发过期检查（充值0会触发checkAndClearExpired）
	YangfenBusiness.Recharge(ctx, uid1, 0, 100)

	// 验证余额已被清零
	balance, _ = YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("过期后余额: %d (应该为0)", balance)

	// 退款 - 这里是bug所在：积分已过期，但退款仍然成功
	err = YangfenBusiness.Refund(ctx, uid1, consumeTxId)
	if err != nil {
		t.Fatalf("退款失败: %v", err)
	}

	balance, _ = YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("退款后余额: %d", balance)

	// Bug复现：退款成功，但原始积分已过期，用户凭空获得50积分
	if balance > 0 {
		t.Errorf("Bug复现! 积分已过期，但退款成功，用户凭空获得%d积分", balance)
	}
}

// TestBug3_ConsumeBonusNotApplied 消费奖励未发放测试
// Bug描述: 代码中计算了 bonusRate 但没有实际使用
// 复现步骤:
// 1. 用户1充值200
// 2. 消费100（满100应该有双倍积分奖励）
// 3. 预期：消费100后，获得100奖励，余额=200
// 4. 实际：消费100后，余额=100，奖励未发放
func TestBug3_ConsumeBonusNotApplied(t *testing.T) {
	ctx := context.Background()
	uid1 := "1"

	TestClearAllData(t)

	// 用户1充值200
	err := YangfenBusiness.Recharge(ctx, uid1, 200, 100)
	if err != nil {
		t.Fatalf("充值失败: %v", err)
	}

	balance, _ := YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("用户1充值后余额: %d", balance)

	// 消费100（满100应该有双倍积分奖励）
	err = YangfenBusiness.Consume(ctx, uid1, 100)
	if err != nil {
		t.Fatalf("消费失败: %v", err)
	}

	balance, _ = YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("消费100后余额: %d", balance)

	// 预期：如果奖励生效，余额应该是 200-100+100=200
	// 实际：余额是100，说明奖励没有发放
	expectedWithBonus := 200
	if balance != expectedWithBonus {
		t.Errorf("Bug复现! 消费满100应获得双倍积分奖励，预期余额%d，实际余额%d", expectedWithBonus, balance)
	}
}

// TestBug4_TransactionNotClearedAfterExpire 余额过期后交易记录未清理测试
// Bug描述: 积分过期时只清零了余额，没有清理交易记录
// 复现步骤:
// 1. 用户1充值100
// 2. 手动设置过期时间为过去（模拟积分过期）
// 3. 触发过期检查（调用getBalance会触发checkAndClearExpired）
// 4. 查看余额：已清零
// 5. 查看交易记录：仍然存在
// 6. 问题：交易记录与余额不一致
func TestBug4_TransactionNotClearedAfterExpire(t *testing.T) {
	ctx := context.Background()
	uid1 := "1"

	TestClearAllData(t)

	// 用户1充值100
	err := YangfenBusiness.Recharge(ctx, uid1, 100, 100)
	if err != nil {
		t.Fatalf("充值失败: %v", err)
	}

	balance, _ := YangfenBusiness.GetBalance(ctx, uid1)
	t.Logf("用户1充值后余额: %d", balance)

	// 查看交易记录
	transactions, _ := YangfenBusiness.GetTransactions(ctx, uid1)
	t.Logf("过期前交易记录数: %d", len(transactions))

	// 手动设置过期时间为过去（模拟积分已过期）
	pastTime := time.Now().Add(-48 * time.Hour).Unix()
	webcache.RedisClient.Set(ctx, YangfenExpireKey+uid1, strconv.FormatInt(pastTime, 10), 0)
	t.Log("已设置过期时间为48小时前")

	// 触发过期检查（充值会触发checkAndClearExpired）
	err = YangfenBusiness.Recharge(ctx, uid1, 0, 100)
	if err == nil {
		t.Log("充值0触发过期检查")
	}

	// 再次查看余额和交易记录
	balance, _ = YangfenBusiness.GetBalance(ctx, uid1)
	transactions, _ = YangfenBusiness.GetTransactions(ctx, uid1)

	t.Logf("过期后余额: %d (预期为0)", balance)
	t.Logf("过期后交易记录数: %d (预期为0)", len(transactions))

	// Bug: 余额已清零，但交易记录仍然存在
	if balance == 0 && len(transactions) > 0 {
		t.Errorf("Bug复现! 余额已清零，但交易记录未清理，存在%d条记录", len(transactions))
	}
}

// TestAllBugs 运行所有Bug测试
func TestAllBugs(t *testing.T) {
	fmt.Println("\n========== 清除测试数据 ==========")
	TestClearAllData(t)

	fmt.Println("\n========== Bug1: 转账并发问题 ==========")
	TestBug1_TransferConcurrency(t)

	fmt.Println("\n========== Bug2: 退款金额错误 ==========")
	TestBug2_RefundAfterExpire(t)

	fmt.Println("\n========== Bug3: 消费奖励未发放 ==========")
	TestBug3_ConsumeBonusNotApplied(t)

	fmt.Println("\n========== Bug4: 交易记录未清理 ==========")
	TestBug4_TransactionNotClearedAfterExpire(t)
}
