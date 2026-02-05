package login

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
)

const (
	CodeOK          = 0
	CodeNameExists  = 1001
	CodeInvalidArg  = 1002
	CodeDBError     = 1003
)

type RegisterResult struct {
	Code      int32
	AccountID int64
	Message   string
}

func Register(ctx context.Context, db *sql.DB, username, password string) (RegisterResult, error) {
	if username == "" || password == "" {
		return RegisterResult{Code: CodeInvalidArg, Message: "参数不合法"}, nil
	}

	// Existence check (unique index will also protect us).
	var existingID int64
	if err := db.QueryRowContext(ctx, "SELECT id FROM admin_account WHERE username=? LIMIT 1", username).Scan(&existingID); err == nil {
		return RegisterResult{Code: CodeNameExists, Message: "名称已存在"}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return RegisterResult{Code: CodeDBError, Message: "系统错误"}, err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return RegisterResult{Code: CodeDBError, Message: "系统错误"}, err
	}
	sum := md5.Sum(append(salt, []byte(password)...)) // md5(salt || password)
	hashHex := hex.EncodeToString(sum[:])
	saltHex := hex.EncodeToString(salt)

	res, err := db.ExecContext(ctx,
		"INSERT INTO admin_account (username, password_hash, password_salt) VALUES (?,?,?)",
		username, hashHex, saltHex,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return RegisterResult{Code: CodeNameExists, Message: "名称已存在"}, nil
		}
		return RegisterResult{Code: CodeDBError, Message: "系统错误"}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return RegisterResult{Code: CodeDBError, Message: "系统错误"}, err
	}
	return RegisterResult{Code: CodeOK, AccountID: id, Message: "注册成功"}, nil
}

func Login(ctx context.Context, db *sql.DB, username, password string) (bool, string, error) {
	if username == "" || password == "" {
		return false, "名称或密码不能为空", nil
	}

	var (
		id      int64
		hashHex string
		saltHex string
	)
	err := db.QueryRowContext(ctx,
		"SELECT id, password_hash, password_salt FROM admin_account WHERE username=? LIMIT 1",
		username,
	).Scan(&id, &hashHex, &saltHex)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "账号不存在", nil
		}
		return false, "系统错误", err
	}

	saltBytes, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, "系统错误", err
	}
	sum := md5.Sum(append(saltBytes, []byte(password)...))
	gotHashHex := hex.EncodeToString(sum[:])
	if subtle.ConstantTimeCompare([]byte(gotHashHex), []byte(hashHex)) != 1 {
		return false, "密码错误", nil
	}
	_ = id
	return true, "登录成功", nil
}

