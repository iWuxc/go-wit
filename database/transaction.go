package database

import (
	"context"
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/utils"
	"gorm.io/gorm"
	"time"
)

type Func func(ctx context.Context, db DB) error

// Trans 数据库事物支持 .
func Trans(ctx context.Context, db DB, fns ...Func) (e error) {
	tx := db.Begin()
	defer func(tx *gorm.DB) {
		if err := recover(); err != nil {
			e = errors.New(fmt.Sprintf("%v", err))
			tx.Rollback()
			log.Errorf("database trans panic recovered: %v \n%s", err, utils.Stack(3))
		}
	}(tx)

	time.AfterFunc(time.Minute, func() {
		if tx != nil {
			tx.Rollback()
		}
	})

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
