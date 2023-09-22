package limiter

import (
	"github.com/shareed2k/go_limiter"
	"github.com/vndee/lensquery-backend/pkg/database"
)

var Limiter *go_limiter.Limiter

func InitLimter() error {
	Limiter = go_limiter.NewLimiter(database.RedisClient)
	return nil
}
