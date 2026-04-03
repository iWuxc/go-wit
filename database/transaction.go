package database

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/utils"
	"gorm.io/gorm"
)

type Func func(ctx context.Context, db DB) error

// Trans 数据库事务支持。
// 调用方如需超时控制，请在传入的 ctx 中设置 deadline，例如：
//
//	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
//	defer cancel()
//	Trans(ctx, db, fn)
func Trans(ctx context.Context, db DB, fns ...Func) (e error) {
	tx := db.WithContext(ctx).Begin()
	defer func(tx *gorm.DB) {
		if err := recover(); err != nil {
			e = errors.New(fmt.Sprintf("%v", err))
			tx.Rollback()
			log.Errorf("database trans panic recovered: %v \n%s", err, utils.Stack(3))
		}
	}(tx)

	if e = tx.Error; e != nil {
		tx.Rollback()
		return
	}

	for _, fn := range fns {
		if e = fn(ctx, tx); e != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return nil
}
