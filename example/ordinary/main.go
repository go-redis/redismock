package ordinary

import (
	"context"
	"github.com/go-redis/redis/v9"
)

const itemKey = "item_cache"

var ctx = context.TODO()

func ItemCache(db *redis.Client, itemID string) (item string, err error) {
	item, err = db.HGet(ctx, itemKey, itemID).Result()
	if err == redis.Nil {
		// call api
		item = "information"

		err = db.HSet(ctx, itemKey, itemID, item).Err()
	}
	return
}
