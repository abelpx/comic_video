package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"encoding/json"
	"comic_video/internal/domain/entity"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr, password string, db int) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	result, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Incr 递增
func (c *Client) Incr(ctx context.Context, key string) error {
	return c.client.Incr(ctx, key).Err()
}

// IncrBy 按指定值递增
func (c *Client) IncrBy(ctx context.Context, key string, value int64) error {
	return c.client.IncrBy(ctx, key, value).Err()
}

// HSet 设置哈希字段
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// LPush 左推入列表
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// RPop 右弹出列表
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// SAdd 添加到集合
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SRem 从集合移除
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SetTaskStatus 存储任务进度（JSON序列化）
func (c *Client) SetTaskStatus(ctx context.Context, task *entity.Task, expiration time.Duration) error {
	key := "task:" + task.ID.String() + ":status"
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, expiration)
}

// GetTaskStatus 查询任务进度
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*entity.Task, error) {
	key := "task:" + taskID + ":status"
	val, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var task entity.Task
	if err := json.Unmarshal([]byte(val), &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
} 