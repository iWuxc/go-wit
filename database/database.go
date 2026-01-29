package database

import (
	"github.com/iWuxc/go-wit/log"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

var ErrInvalidDBDriver = errors.New("unsupported sql driver .")

type DB = *gorm.DB

type DataConf struct {
	// Debug print to sql if true.
	Debug bool `json:"debug"`
	// Driver implement of mysql/sqlite .
	Driver string `json:"driver"`
	// Source DSN for Driver.
	// MySQL:   username:password@protocol(address)/dbname?param=value
	//		    root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local
	// SQLite:  test.db
	Source string `json:"source"`

	// DBConf db conn pool conf.
	DBConf struct {
		MaxOpenConn     int           `json:"max_open_conn" mapstructure:"max_open_conn"`
		MaxIdleConn     int           `json:"max_idle_conn" mapstructure:"max_idle_conn"`
		ConnMaxLifeTime time.Duration `json:"conn_max_life_time" mapstructure:"conn_max_life_time"`
	} `json:"db_conf" mapstructure:"db_conf"`
}

// NewDB init gorm .
func NewDB(conf *DataConf) (DB, func(), error) {
	driver, err := newDBDriver(conf.Driver, conf.Source)
	if err != nil {
		return nil, nil, err
	}

	logLevel := gormLog.Warn
	if conf.Debug {
		logLevel = gormLog.Info
	}

	db, err := gorm.Open(driver, &gorm.Config{
		Logger: gormLog.New(log.GetInstance(), gormLog.Config{
			LogLevel:                  logLevel,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
		}),
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	sqlDB.SetMaxOpenConns(conf.DBConf.MaxOpenConn)
	sqlDB.SetMaxIdleConns(conf.DBConf.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Minute * conf.DBConf.ConnMaxLifeTime)

	return db, func() {
		_ = sqlDB.Close()
	}, nil
}

// newDBDriver get gorm dialect .
func newDBDriver(dbDriver, dbDSN string) (gorm.Dialector, error) {
	switch dbDriver {
	case "mysql":
		return mysql.Open(dbDSN), nil
	case "sqlite":
		return sqlite.Open(dbDSN), nil
	default:
		return nil, ErrInvalidDBDriver
	}
}
