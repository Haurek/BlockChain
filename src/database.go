package BlockChain

import "badger"

// WriteToDB write key-value to database
func WriteToDB(db *badger.DB, tablePrefix, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append(tablePrefix, key...), value)
		return err
	})
}

// ReadFromDB read value from database
func ReadFromDB(db *badger.DB, tablePrefix, key []byte) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append(tablePrefix, key...))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})
		return err
	})
	return value, err
}

// UpdateInDB update value
func UpdateInDB(db *badger.DB, tablePrefix, key, updatedValue []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append(tablePrefix, key...), updatedValue)
		return err
	})
}

// DeleteFromDB delete value
func DeleteFromDB(db *badger.DB, tablePrefix, key []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(append(tablePrefix, key...))
		return err
	})
}

//// 导入包
//import (
//	"database/sql"
//	"fmt"
//	_ "github.com/go-sql-driver/mysql"
//)
//
//// 声明一个全局变量db
//var db *sql.DB
//
//// 创建一个初始化数据库的函数
//func initDB(dsn string) (err error) {
//	db, err = sql.Open("mysql", dsn)
//	if err != nil {
//		return
//	}
//	err = db.Ping()
//	if err != nil {
//		return
//	}
//	db.SetMaxIdleConns(10)
//	db.SetMaxOpenConns(10)
//	fmt.Println("数据库连接成功")
//	return
//}
//
//func createTableBlocks() error {
//	// 准备创建表格的 SQL 语句
//	sqlStatement := `
//		CREATE TABLE blocks (
//			id INT  PRIMARY KEY UNIQUE,
//			Hash VARCHAR(64) NOT NULL,
//			TimeStamp VARCHAR(20) NOT NULL,
//			PrevHash VARCHAR(64)
//		)
//	`
//
//	// 执行 SQL 语句创建表格
//	_, err := db.Exec(sqlStatement)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func dropTable(tableName string) error {
//	// 准备删除表格的 SQL 语句
//	sqlStatement := fmt.Sprintf(`
//		DROP TABLE %s
//	`, tableName)
//
//	// 执行 SQL 语句删除表格
//	_, err := db.Exec(sqlStatement)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func insertBlock(index int, hash string, timestamp string, prevHash string) error {
//	// 准备插入数据的 SQL 语句
//	sqlStatement := `
//		INSERT INTO blocks (id, Hash, TimeStamp, PrevHash)
//		VALUES (?, ?, ?, ?)
//	`
//
//	// 执行 SQL 语句插入数据
//	_, err := db.Exec(sqlStatement, index, hash, timestamp, prevHash)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// 查询所有的数据
//func queryBlocks() error {
//	//查询数据的语句
//	sqlStatement := `
//		SELECT id, Hash, TimeStamp, PrevHash
//		FROM blocks
//	`
//	//执行查询操作
//	rows, err := db.Query(sqlStatement)
//	if err != nil {
//		return err
//
//	}
//	defer rows.Close()
//	//遍历结果集
//	for rows.Next() {
//		var id int
//		var hash string
//		var timestamp string
//		var prevHash string
//		err := rows.Scan(&id, &hash, &timestamp, &prevHash)
//		if err != nil {
//			return err
//		}
//		fmt.Println(id, hash, timestamp, prevHash)
//	}
//	return nil
//}
//
//// 根据id查询数据库的数据
//func queryBlocksById(id int) error {
//	//查询数据的语句
//	sqlStatement := `
//		SELECT id, Hash, TimeStamp, PrevHash
//		FROM blocks
//		WHERE id = ?
//	`
//	//执行查询操作
//	rows, err := db.Query(sqlStatement, id)
//	if err != nil {
//		return err
//
//	}
//	defer rows.Close()
//	//如果没找到
//	if !rows.Next() {
//		fmt.Println("queryBlockById Error: 没有id为", id, "的数据")
//		return err
//	}
//
//	var index int
//	var hash string
//	var timestamp string
//	var prevHash string
//	err = rows.Scan(&index, &hash, &timestamp, &prevHash)
//	if err != nil {
//		return err
//	}
//	fmt.Println(index, hash, timestamp, prevHash)
//
//	return nil
//}
//
//func updateBlock(id int, hash string, timestamp string, prevHash string) error {
//	// 查询原有数据，获取需要保留的字段值
//	var originalHash, originalTimestamp, originalPrevHash string
//	err := db.QueryRow("SELECT Hash, TimeStamp, PrevHash FROM blocks WHERE id = ?", id).Scan(&originalHash, &originalTimestamp, &originalPrevHash)
//	if err != nil {
//		return err
//	}
//
//	// 如果传入的参数为空字符串，则保留原有字段值
//	if hash == "" {
//		hash = originalHash
//	}
//	if timestamp == "" {
//		timestamp = originalTimestamp
//	}
//	if prevHash == "" {
//		prevHash = originalPrevHash
//	}
//
//	// 准备更新数据的 SQL 语句
//	sqlStatement := `
//		UPDATE blocks
//		SET Hash = ?, TimeStamp = ?, PrevHash = ?
//		WHERE id = ?
//	`
//
//	// 执行 SQL 语句更新数据
//	_, err = db.Exec(sqlStatement, hash, timestamp, prevHash, id)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func deleteBlockByID(id int) error {
//	// 准备删除数据的 SQL 语句
//	sqlStatement := `
//		DELETE FROM blocks
//		WHERE id = ?
//	`
//
//	// 执行 SQL 语句删除数据
//	_, err := db.Exec(sqlStatement, id)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func HandleErr(err error, msg string) {
//	if err != nil {
//		fmt.Println(msg, "Error:", err)
//	}
//
//}
