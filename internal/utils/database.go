package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	db    *gorm.DB
	query *gorm.DB
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		query: db,
	}
}

// WithContext 设置上下文
func (qb *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	qb.query = qb.query.WithContext(ctx)
	return qb
}

// Select 选择字段
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	if len(fields) > 0 {
		qb.query = qb.query.Select(fields)
	}
	return qb
}

// Where 添加WHERE条件
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Where(condition, args...)
	return qb
}

// WhereIn 添加IN条件
func (qb *QueryBuilder) WhereIn(field string, values interface{}) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s IN ?", field), values)
	return qb
}

// WhereNotIn 添加NOT IN条件
func (qb *QueryBuilder) WhereNotIn(field string, values interface{}) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s NOT IN ?", field), values)
	return qb
}

// WhereBetween 添加BETWEEN条件
func (qb *QueryBuilder) WhereBetween(field string, start, end interface{}) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), start, end)
	return qb
}

// WhereNull 添加IS NULL条件
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s IS NULL", field))
	return qb
}

// WhereNotNull 添加IS NOT NULL条件
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s IS NOT NULL", field))
	return qb
}

// WhereLike 添加LIKE条件
func (qb *QueryBuilder) WhereLike(field, pattern string) *QueryBuilder {
	qb.query = qb.query.Where(fmt.Sprintf("%s LIKE ?", field), pattern)
	return qb
}

// OrWhere 添加OR WHERE条件
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Or(condition, args...)
	return qb
}

// Join 添加JOIN
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	qb.query = qb.query.Joins(fmt.Sprintf("JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin 添加LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	qb.query = qb.query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// RightJoin 添加RIGHT JOIN
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	qb.query = qb.query.Joins(fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition))
	return qb
}

// OrderBy 添加排序
func (qb *QueryBuilder) OrderBy(field string, direction ...string) *QueryBuilder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	qb.query = qb.query.Order(fmt.Sprintf("%s %s", field, dir))
	return qb
}

// GroupBy 添加分组
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.query = qb.query.Group(strings.Join(fields, ", "))
	return qb
}

// Having 添加HAVING条件
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Having(condition, args...)
	return qb
}

// Limit 设置限制
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query = qb.query.Limit(limit)
	return qb
}

// Offset 设置偏移
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query = qb.query.Offset(offset)
	return qb
}

// Paginate 分页
func (qb *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
	offset := (page - 1) * pageSize
	qb.query = qb.query.Offset(offset).Limit(pageSize)
	return qb
}

// Preload 预加载关联
func (qb *QueryBuilder) Preload(association string, conditions ...interface{}) *QueryBuilder {
	qb.query = qb.query.Preload(association, conditions...)
	return qb
}

// Find 查找记录
func (qb *QueryBuilder) Find(dest interface{}) error {
	return qb.query.Find(dest).Error
}

// First 查找第一条记录
func (qb *QueryBuilder) First(dest interface{}) error {
	return qb.query.First(dest).Error
}

// Last 查找最后一条记录
func (qb *QueryBuilder) Last(dest interface{}) error {
	return qb.query.Last(dest).Error
}

// Count 统计记录数
func (qb *QueryBuilder) Count() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (qb *QueryBuilder) Exists() (bool, error) {
	count, err := qb.Count()
	return count > 0, err
}

// Update 更新记录
func (qb *QueryBuilder) Update(column string, value interface{}) error {
	return qb.query.Update(column, value).Error
}

// Updates 批量更新记录
func (qb *QueryBuilder) Updates(values interface{}) error {
	return qb.query.Updates(values).Error
}

// Delete 删除记录
func (qb *QueryBuilder) Delete(value interface{}) error {
	return qb.query.Delete(value).Error
}

// GetQuery 获取GORM查询对象
func (qb *QueryBuilder) GetQuery() *gorm.DB {
	return qb.query
}

// PaginationResult 分页结果
type PaginationResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// PaginateWithCount 分页查询并统计总数
func (qb *QueryBuilder) PaginateWithCount(dest interface{}, page, pageSize int) (*PaginationResult, error) {
	// 先统计总数
	total, err := qb.Count()
	if err != nil {
		return nil, err
	}
	
	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	hasNext := page < totalPages
	hasPrev := page > 1
	
	// 查询分页数据
	err = qb.Paginate(page, pageSize).Find(dest)
	if err != nil {
		return nil, err
	}
	
	return &PaginationResult{
		Data:       dest,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// Transaction 事务处理
func Transaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	return db.Transaction(fn)
}

// BatchInsert 批量插入
func BatchInsert(db *gorm.DB, data interface{}, batchSize int) error {
	return db.CreateInBatches(data, batchSize).Error
}

// BulkUpdate 批量更新
func BulkUpdate(db *gorm.DB, model interface{}, updates map[string]interface{}, whereConditions map[string]interface{}) error {
	query := db.Model(model)
	
	// 添加WHERE条件
	for field, value := range whereConditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}
	
	return query.Updates(updates).Error
}

// DatabaseStats 数据库统计信息
type DatabaseStats struct {
	OpenConnections int           `json:"open_connections"`
	InUse          int           `json:"in_use"`
	Idle           int           `json:"idle"`
	WaitCount      int64         `json:"wait_count"`
	WaitDuration   time.Duration `json:"wait_duration"`
	MaxIdleClosed  int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64      `json:"max_lifetime_closed"`
}

// GetDatabaseStats 获取数据库连接统计
func GetDatabaseStats(db *gorm.DB) (*DatabaseStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	
	stats := sqlDB.Stats()
	
	return &DatabaseStats{
		OpenConnections:   stats.OpenConnections,
		InUse:            stats.InUse,
		Idle:             stats.Idle,
		WaitCount:        stats.WaitCount,
		WaitDuration:     stats.WaitDuration,
		MaxIdleClosed:    stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}, nil
}

// SlowQueryLogger 慢查询日志记录器
type SlowQueryLogger struct {
	logger.Interface
	SlowThreshold time.Duration
}

// NewSlowQueryLogger 创建慢查询日志记录器
func NewSlowQueryLogger(slowThreshold time.Duration) *SlowQueryLogger {
	return &SlowQueryLogger{
		Interface:     logger.Default,
		SlowThreshold: slowThreshold,
	}
}

// Trace 记录SQL执行轨迹
func (l *SlowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	
	if elapsed > l.SlowThreshold {
		sql, rows := fc()
		LogWarn(ctx, "Slow query detected", map[string]interface{}{
			"sql":      sql,
			"rows":     rows,
			"elapsed":  elapsed,
			"error":    err,
		})
	}
	
	// 调用原始的Trace方法
	l.Interface.Trace(ctx, begin, fc, err)
}

// OptimizeQuery 查询优化建议
func OptimizeQuery(db *gorm.DB, sql string) ([]string, error) {
	var suggestions []string
	
	// 检查是否使用了索引
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)
	rows, err := db.Raw(explainSQL).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// 简化的优化建议逻辑
	sqlUpper := strings.ToUpper(sql)
	
	if strings.Contains(sqlUpper, "SELECT *") {
		suggestions = append(suggestions, "避免使用SELECT *，明确指定需要的字段")
	}
	
	if strings.Contains(sqlUpper, "ORDER BY") && !strings.Contains(sqlUpper, "LIMIT") {
		suggestions = append(suggestions, "ORDER BY查询建议添加LIMIT限制结果集大小")
	}
	
	if strings.Contains(sqlUpper, "LIKE '%") {
		suggestions = append(suggestions, "避免使用前缀通配符LIKE '%xxx'，考虑使用全文索引")
	}
	
	if !strings.Contains(sqlUpper, "WHERE") && strings.Contains(sqlUpper, "SELECT") {
		suggestions = append(suggestions, "考虑添加WHERE条件限制查询范围")
	}
	
	return suggestions, nil
}
