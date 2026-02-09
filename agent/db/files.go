package db

import (
	"context"
	"github.com/xiehqing/common/pkg/ormx"
)

// CreateFile 创建文件
func (q *Queries) CreateFile(ctx context.Context, args CreateFileArgs) (File, error) {
	var f = &File{
		SessionID: args.SessionID,
		Path:      args.Path,
		Content:   args.Content,
		Version:   args.Version,
	}
	f.ID = args.ID
	err := ormx.Insert(q.db, f)
	if err != nil {
		return *f, err
	}
	return *f, nil
}

func (q *Queries) DeleteFile(ctx context.Context, id string) error {
	return q.db.Where("id = ?", id).Delete(&File{}).Error
}

func (q *Queries) DeleteSessionFiles(ctx context.Context, sessionID string) error {
	return q.db.Where("session_id = ?", sessionID).Delete(&File{}).Error
}

func (q *Queries) GetFile(ctx context.Context, id string) (File, error) {
	var f File
	err := q.db.Where("id = ?", id).First(&f).Error
	return f, err
}

func (q *Queries) GetFileByPathAndSession(ctx context.Context, arg GetFileByPathAndSessionArgs) (File, error) {
	var f File
	err := q.db.Where("path = ? and session_id = ?", arg.Path, arg.SessionID).Order("version DESC, created_at DESC").First(&f).Error
	return f, err
}

func (q *Queries) ListFilesByPath(ctx context.Context, path string) ([]File, error) {
	var files []File
	err := q.db.Where("path = ?", path).Order("version ASC, created_at ASC").Find(&files).Error
	return files, err
}

func (q *Queries) ListFilesBySession(ctx context.Context, sessionID string) ([]File, error) {
	var files []File
	err := q.db.Where("session_id = ?", sessionID).Order("version ASC, created_at ASC").Find(&files).Error
	return files, err
}

const listLatestSessionFilesSQL = `SELECT f.id, f.session_id, f.path, f.content, f.version, f.created_at, f.updated_at
FROM files f
INNER JOIN (
    SELECT path, MAX(version) as max_version, MAX(created_at) as max_created_at
    FROM files
    GROUP BY path
) latest ON f.path = latest.path AND f.version = latest.max_version AND f.created_at = latest.max_created_at
WHERE f.session_id = ?
ORDER BY f.path;`

func (q *Queries) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]File, error) {
	var files []File
	err := q.db.Raw(listLatestSessionFilesSQL, sessionID).Scan(&files).Error
	return files, err
}

func (q *Queries) ListNewFiles(ctx context.Context) ([]File, error) {
	var files []File
	err := q.db.Where("is_new = 1").Order("version DESC, created_at DESC").Scan(&files).Error
	return files, err
}
