package cmd

import (
	"fmt"

	"github.com/armylong/armylong-go/internal/common/webcache"
	"github.com/spf13/cobra"
)

func GetRedisData(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	// cacheKey, _ := cmd.Flags().GetString("cache_key")
	cacheKey := args[0]

	cacheValue, _ := webcache.RedisClient.Get(ctx, cacheKey).Result()

	fmt.Printf("cache_key: %s, cache_value: %s\n", cacheKey, cacheValue)
}
