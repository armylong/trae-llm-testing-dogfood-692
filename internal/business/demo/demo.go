package demo

import (
	"context"

	"github.com/armylong/armylong-go/internal/common/config"
	"github.com/armylong/armylong-go/internal/common/webcache"
)

type demoBusiness struct{}

var DemoBusiness = &demoBusiness{}

func (b *demoBusiness) SetMessage(ctx context.Context, message string) (res string, err error) {
	res, err = webcache.RedisClient.Set(ctx, config.DemoMessageCacheKey, message, 0).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}

func (b *demoBusiness) GetMessage(ctx context.Context) (res string, err error) {
	res, err = webcache.RedisClient.Get(ctx, config.DemoMessageCacheKey).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}
