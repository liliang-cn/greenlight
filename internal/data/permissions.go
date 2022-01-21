package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Permissions slice用来存放某个用户的权限代码 如"movies:read" 和 "movies:write"
type Permissions []string

// Include 检查是否在 slice 中
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

// PermissionModel 权限模型类型
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser 方法返回指定用户的所有权限
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code 
		FROM permissions 
		INNER JOIN users_permissions ON permissions.id = users_permissions.permission_id
		INNER JOIN users ON users.id = users_permissions.user_id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser 给指定用户添加权限
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
	INSERT INTO users_permissions
	SELECT $1, permissions.id 
	FROM permissions WHERE permissions.code=ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
