package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端结构体
type RedisClient struct {
	client      *redis.Client
	config      RedisConfig
	ctx         context.Context
	cancel      context.CancelFunc
	onError     func(error)
	onReconnect func()
}

// RedisConfig Redis配置结构体
type RedisConfig struct {
	Address            string        // Redis服务器地址，格式：host:port
	Password           string        // 密码
	Database           int           // 数据库编号，默认0
	PoolSize           int           // 连接池大小，默认10
	MinIdleConns       int           // 最小空闲连接数，默认0
	MaxConnAge         time.Duration // 连接最大存活时间，默认0（永不过期）
	PoolTimeout        time.Duration // 获取连接超时时间，默认4秒
	IdleTimeout        time.Duration // 空闲连接超时时间，默认5分钟
	IdleCheckFrequency time.Duration // 空闲连接检查频率，默认1分钟
	DialTimeout        time.Duration // 连接超时时间，默认5秒
	ReadTimeout        time.Duration // 读取超时时间，默认3秒
	WriteTimeout       time.Duration // 写入超时时间，默认3秒
	MaxRetries         int           // 最大重试次数，默认3
	MinRetryBackoff    time.Duration // 最小重试间隔，默认8毫秒
	MaxRetryBackoff    time.Duration // 最大重试间隔，默认512毫秒
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	// 设置默认值
	if config.Address == "" {
		config.Address = "localhost:6379"
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.PoolTimeout == 0 {
		config.PoolTimeout = 4 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.IdleCheckFrequency == 0 {
		config.IdleCheckFrequency = 1 * time.Minute
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.MinRetryBackoff == 0 {
		config.MinRetryBackoff = 8 * time.Millisecond
	}
	if config.MaxRetryBackoff == 0 {
		config.MaxRetryBackoff = 512 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 创建Redis客户端选项
	opts := &redis.Options{
		Addr:            config.Address,
		Password:        config.Password,
		DB:              config.Database,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		ConnMaxLifetime: 1 * time.Hour, // 替代 MaxConnAge
		PoolTimeout:     config.PoolTimeout,
		ConnMaxIdleTime: config.IdleTimeout, // 替代 IdleTimeout
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		MaxRetries:      config.MaxRetries,
	}

	client := redis.NewClient(opts)

	redisClient := &RedisClient{
		client: client,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 测试连接
	if err := redisClient.Ping(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return redisClient, nil
}

// SetCallbacks 设置回调函数
func (r *RedisClient) SetCallbacks(onError func(error), onReconnect func()) {
	r.onError = onError
	r.onReconnect = onReconnect
}

// Close 关闭Redis客户端
func (r *RedisClient) Close() error {
	r.cancel()
	return r.client.Close()
}

// Ping 测试连接
func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetContext 获取上下文
func (r *RedisClient) GetContext() context.Context {
	return r.ctx
}

// handleError 处理错误
func (r *RedisClient) handleError(operation string, err error) error {
	if err != nil && r.onError != nil {
		go r.onError(fmt.Errorf("Redis %s error: %w", operation, err))
	}
	return err
}

// =============================================================================
// 字符串操作
// =============================================================================

// Set 设置字符串值
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	return r.handleError("Set", err)
}

// Get 获取字符串值
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 键不存在返回空字符串
	}
	return val, r.handleError("Get", err)
}

// GetSet 设置新值并返回旧值
func (r *RedisClient) GetSet(key string, value interface{}) (string, error) {
	val, err := r.client.GetSet(r.ctx, key, value).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, r.handleError("GetSet", err)
}

// Incr 自增1
func (r *RedisClient) Incr(key string) (int64, error) {
	val, err := r.client.Incr(r.ctx, key).Result()
	return val, r.handleError("Incr", err)
}

// IncrBy 自增指定值
func (r *RedisClient) IncrBy(key string, value int64) (int64, error) {
	val, err := r.client.IncrBy(r.ctx, key, value).Result()
	return val, r.handleError("IncrBy", err)
}

// Decr 自减1
func (r *RedisClient) Decr(key string) (int64, error) {
	val, err := r.client.Decr(r.ctx, key).Result()
	return val, r.handleError("Decr", err)
}

// DecrBy 自减指定值
func (r *RedisClient) DecrBy(key string, value int64) (int64, error) {
	val, err := r.client.DecrBy(r.ctx, key, value).Result()
	return val, r.handleError("DecrBy", err)
}

// MSet 批量设置多个键值对
func (r *RedisClient) MSet(pairs ...interface{}) error {
	err := r.client.MSet(r.ctx, pairs...).Err()
	return r.handleError("MSet", err)
}

// MGet 批量获取多个键的值
func (r *RedisClient) MGet(keys ...string) ([]interface{}, error) {
	vals, err := r.client.MGet(r.ctx, keys...).Result()
	return vals, r.handleError("MGet", err)
}

// =============================================================================
// 哈希操作
// =============================================================================

// HSet 设置哈希字段值
func (r *RedisClient) HSet(key, field string, value interface{}) error {
	err := r.client.HSet(r.ctx, key, field, value).Err()
	return r.handleError("HSet", err)
}

// HGet 获取哈希字段值
func (r *RedisClient) HGet(key, field string) (string, error) {
	val, err := r.client.HGet(r.ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, r.handleError("HGet", err)
}

// HMSet 批量设置哈希字段值
func (r *RedisClient) HMSet(key string, fields map[string]interface{}) error {
	err := r.client.HMSet(r.ctx, key, fields).Err()
	return r.handleError("HMSet", err)
}

// HMGet 批量获取哈希字段值
func (r *RedisClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	vals, err := r.client.HMGet(r.ctx, key, fields...).Result()
	return vals, r.handleError("HMGet", err)
}

// HGetAll 获取哈希所有字段和值
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	vals, err := r.client.HGetAll(r.ctx, key).Result()
	return vals, r.handleError("HGetAll", err)
}

// HKeys 获取哈希所有字段名
func (r *RedisClient) HKeys(key string) ([]string, error) {
	vals, err := r.client.HKeys(r.ctx, key).Result()
	return vals, r.handleError("HKeys", err)
}

// HVals 获取哈希所有值
func (r *RedisClient) HVals(key string) ([]string, error) {
	vals, err := r.client.HVals(r.ctx, key).Result()
	return vals, r.handleError("HVals", err)
}

// HExists 检查哈希字段是否存在
func (r *RedisClient) HExists(key, field string) (bool, error) {
	exists, err := r.client.HExists(r.ctx, key, field).Result()
	return exists, r.handleError("HExists", err)
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(key string, fields ...string) (int64, error) {
	count, err := r.client.HDel(r.ctx, key, fields...).Result()
	return count, r.handleError("HDel", err)
}

// HLen 获取哈希字段数量
func (r *RedisClient) HLen(key string) (int64, error) {
	count, err := r.client.HLen(r.ctx, key).Result()
	return count, r.handleError("HLen", err)
}

// HIncrBy 哈希字段自增
func (r *RedisClient) HIncrBy(key, field string, incr int64) (int64, error) {
	val, err := r.client.HIncrBy(r.ctx, key, field, incr).Result()
	return val, r.handleError("HIncrBy", err)
}

// =============================================================================
// 列表操作
// =============================================================================

// LPush 从左侧推入列表
func (r *RedisClient) LPush(key string, values ...interface{}) (int64, error) {
	count, err := r.client.LPush(r.ctx, key, values...).Result()
	return count, r.handleError("LPush", err)
}

// RPush 从右侧推入列表
func (r *RedisClient) RPush(key string, values ...interface{}) (int64, error) {
	count, err := r.client.RPush(r.ctx, key, values...).Result()
	return count, r.handleError("RPush", err)
}

// LPop 从左侧弹出列表元素
func (r *RedisClient) LPop(key string) (string, error) {
	val, err := r.client.LPop(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, r.handleError("LPop", err)
}

// RPop 从右侧弹出列表元素
func (r *RedisClient) RPop(key string) (string, error) {
	val, err := r.client.RPop(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, r.handleError("RPop", err)
}

// LRange 获取列表指定范围的元素
func (r *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	vals, err := r.client.LRange(r.ctx, key, start, stop).Result()
	return vals, r.handleError("LRange", err)
}

// LLen 获取列表长度
func (r *RedisClient) LLen(key string) (int64, error) {
	count, err := r.client.LLen(r.ctx, key).Result()
	return count, r.handleError("LLen", err)
}

// LIndex 获取列表指定位置的元素
func (r *RedisClient) LIndex(key string, index int64) (string, error) {
	val, err := r.client.LIndex(r.ctx, key, index).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, r.handleError("LIndex", err)
}

// LSet 设置列表指定位置的元素
func (r *RedisClient) LSet(key string, index int64, value interface{}) error {
	err := r.client.LSet(r.ctx, key, index, value).Err()
	return r.handleError("LSet", err)
}

// LTrim 保留列表指定范围的元素
func (r *RedisClient) LTrim(key string, start, stop int64) error {
	err := r.client.LTrim(r.ctx, key, start, stop).Err()
	return r.handleError("LTrim", err)
}

// =============================================================================
// 集合操作
// =============================================================================

// SAdd 向集合添加成员
func (r *RedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	count, err := r.client.SAdd(r.ctx, key, members...).Result()
	return count, r.handleError("SAdd", err)
}

// SMembers 获取集合所有成员
func (r *RedisClient) SMembers(key string) ([]string, error) {
	members, err := r.client.SMembers(r.ctx, key).Result()
	return members, r.handleError("SMembers", err)
}

// SIsMember 检查成员是否在集合中
func (r *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	exists, err := r.client.SIsMember(r.ctx, key, member).Result()
	return exists, r.handleError("SIsMember", err)
}

// SCard 获取集合成员数量
func (r *RedisClient) SCard(key string) (int64, error) {
	count, err := r.client.SCard(r.ctx, key).Result()
	return count, r.handleError("SCard", err)
}

// SRem 从集合删除成员
func (r *RedisClient) SRem(key string, members ...interface{}) (int64, error) {
	count, err := r.client.SRem(r.ctx, key, members...).Result()
	return count, r.handleError("SRem", err)
}

// SPop 随机移除并返回集合中的成员
func (r *RedisClient) SPop(key string) (string, error) {
	member, err := r.client.SPop(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return member, r.handleError("SPop", err)
}

// SRandMember 随机返回集合中的成员
func (r *RedisClient) SRandMember(key string) (string, error) {
	member, err := r.client.SRandMember(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return member, r.handleError("SRandMember", err)
}

// =============================================================================
// 有序集合操作
// =============================================================================

// ZAdd 向有序集合添加成员
func (r *RedisClient) ZAdd(key string, members ...*redis.Z) (int64, error) {
	// 转换指针切片为值切片
	zMembers := make([]redis.Z, len(members))
	for i, member := range members {
		zMembers[i] = *member
	}
	count, err := r.client.ZAdd(r.ctx, key, zMembers...).Result()
	return count, r.handleError("ZAdd", err)
}

// ZScore 获取有序集合成员的分数
func (r *RedisClient) ZScore(key, member string) (float64, error) {
	score, err := r.client.ZScore(r.ctx, key, member).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return score, r.handleError("ZScore", err)
}

// ZRange 获取有序集合指定范围的成员
func (r *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	members, err := r.client.ZRange(r.ctx, key, start, stop).Result()
	return members, r.handleError("ZRange", err)
}

// ZRangeWithScores 获取有序集合指定范围的成员及分数
func (r *RedisClient) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	members, err := r.client.ZRangeWithScores(r.ctx, key, start, stop).Result()
	return members, r.handleError("ZRangeWithScores", err)
}

// ZRank 获取有序集合成员的排名
func (r *RedisClient) ZRank(key, member string) (int64, error) {
	rank, err := r.client.ZRank(r.ctx, key, member).Result()
	if err == redis.Nil {
		return -1, nil
	}
	return rank, r.handleError("ZRank", err)
}

// ZCard 获取有序集合成员数量
func (r *RedisClient) ZCard(key string) (int64, error) {
	count, err := r.client.ZCard(r.ctx, key).Result()
	return count, r.handleError("ZCard", err)
}

// ZRem 从有序集合删除成员
func (r *RedisClient) ZRem(key string, members ...interface{}) (int64, error) {
	count, err := r.client.ZRem(r.ctx, key, members...).Result()
	return count, r.handleError("ZRem", err)
}

// =============================================================================
// 键操作
// =============================================================================

// Exists 检查键是否存在
func (r *RedisClient) Exists(keys ...string) (int64, error) {
	count, err := r.client.Exists(r.ctx, keys...).Result()
	return count, r.handleError("Exists", err)
}

// Del 删除键
func (r *RedisClient) Del(keys ...string) (int64, error) {
	count, err := r.client.Del(r.ctx, keys...).Result()
	return count, r.handleError("Del", err)
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(key string, expiration time.Duration) (bool, error) {
	success, err := r.client.Expire(r.ctx, key, expiration).Result()
	return success, r.handleError("Expire", err)
}

// ExpireAt 设置键在指定时间过期
func (r *RedisClient) ExpireAt(key string, tm time.Time) (bool, error) {
	success, err := r.client.ExpireAt(r.ctx, key, tm).Result()
	return success, r.handleError("ExpireAt", err)
}

// TTL 获取键的剩余过期时间
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	duration, err := r.client.TTL(r.ctx, key).Result()
	return duration, r.handleError("TTL", err)
}

// Keys 查找匹配模式的键
func (r *RedisClient) Keys(pattern string) ([]string, error) {
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	return keys, r.handleError("Keys", err)
}

// Type 获取键的类型
func (r *RedisClient) Type(key string) (string, error) {
	keyType, err := r.client.Type(r.ctx, key).Result()
	return keyType, r.handleError("Type", err)
}

// Rename 重命名键
func (r *RedisClient) Rename(key, newkey string) error {
	err := r.client.Rename(r.ctx, key, newkey).Err()
	return r.handleError("Rename", err)
}

// =============================================================================
// JSON操作助手
// =============================================================================

// SetJSON 设置JSON对象
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.Set(key, string(jsonData), expiration)
}

// GetJSON 获取JSON对象
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	jsonStr, err := r.Get(key)
	if err != nil {
		return err
	}
	if jsonStr == "" {
		return redis.Nil
	}
	return json.Unmarshal([]byte(jsonStr), dest)
}

// HSetJSON 设置哈希字段JSON值
func (r *RedisClient) HSetJSON(key, field string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.HSet(key, field, string(jsonData))
}

// HGetJSON 获取哈希字段JSON值
func (r *RedisClient) HGetJSON(key, field string, dest interface{}) error {
	jsonStr, err := r.HGet(key, field)
	if err != nil {
		return err
	}
	if jsonStr == "" {
		return redis.Nil
	}
	return json.Unmarshal([]byte(jsonStr), dest)
}

// =============================================================================
// 发布订阅操作
// =============================================================================

// Publish 发布消息到频道
func (r *RedisClient) Publish(channel string, message interface{}) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *RedisClient) Subscribe(channels ...string) *redis.PubSub {
	return r.client.Subscribe(r.ctx, channels...)
}

// PSubscribe 按模式订阅频道
func (r *RedisClient) PSubscribe(patterns ...string) *redis.PubSub {
	return r.client.PSubscribe(r.ctx, patterns...)
}

// =============================================================================
// 事务操作
// =============================================================================

// TxPipeline 创建事务管道
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Watch 监视键用于事务
func (r *RedisClient) Watch(fn func(*redis.Tx) error, keys ...string) error {
	return r.client.Watch(r.ctx, fn, keys...)
}

// =============================================================================
// 连接池信息
// =============================================================================

// PoolStats 获取连接池统计信息
func (r *RedisClient) PoolStats() *redis.PoolStats {
	return r.client.PoolStats()
}
