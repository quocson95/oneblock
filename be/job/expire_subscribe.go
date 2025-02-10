package job

import (
	"be/common"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func JobExpireSubscribe() error {
	offset := 0
	limit := 1000
	for {
		ml, err := common.GetAllSubscribeByUser(0, false, offset, limit)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				zap.L().With(zap.Error(err)).Error("get all subscribe failed")
				return err
			}
			break
		}
		if len(ml) == 0 {
			break
		}
		nowUnix := time.Now().Unix()
		for _, v := range ml {
			if v.ExpireUnix < nowUnix {
				v.Delete()
			}
		}
		offset += limit
	}
	return nil
}
