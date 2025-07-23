package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// 测试数据结构
type TestUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	CreateAt int64  `json:"create_at"`
}

func getTestRedisClient() (*RedisClient, error) {
	config := RedisConfig{
		Address:  "localhost:6379",
		Database: 1, // 使用数据库1进行测试
		Password: "",
	}

	return NewRedisClient(config)
}

func TestRedisClient_Connection(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	// 测试连接
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// 测试连接池统计
	stats := client.PoolStats()
	if stats == nil {
		t.Error("PoolStats returned nil")
	} else {
		t.Logf("Pool stats - TotalConns: %d, IdleConns: %d, StaleConns: %d",
			stats.TotalConns, stats.IdleConns, stats.StaleConns)
	}
}

func TestRedisClient_StringOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	// 清理测试数据
	client.Del("test:string", "test:string2", "test:string3")

	// 测试Set和Get
	err = client.Set("test:string", "hello world", time.Minute)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	value, err := client.Get("test:string")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if value != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", value)
	}

	// 测试GetSet
	oldValue, err := client.GetSet("test:string", "new value")
	if err != nil {
		t.Errorf("GetSet failed: %v", err)
	}
	if oldValue != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", oldValue)
	}

	// 测试Incr和Decr
	client.Set("test:counter", "10", time.Minute)

	newVal, err := client.Incr("test:counter")
	if err != nil {
		t.Errorf("Incr failed: %v", err)
	}
	if newVal != 11 {
		t.Errorf("Expected 11, got %d", newVal)
	}

	newVal, err = client.DecrBy("test:counter", 5)
	if err != nil {
		t.Errorf("DecrBy failed: %v", err)
	}
	if newVal != 6 {
		t.Errorf("Expected 6, got %d", newVal)
	}

	// 测试MSet和MGet
	err = client.MSet("test:string2", "value2", "test:string3", "value3")
	if err != nil {
		t.Errorf("MSet failed: %v", err)
	}

	values, err := client.MGet("test:string2", "test:string3")
	if err != nil {
		t.Errorf("MGet failed: %v", err)
	}
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// 清理测试数据
	client.Del("test:string", "test:string2", "test:string3", "test:counter")
}

func TestRedisClient_HashOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	hashKey := "test:hash"
	client.Del(hashKey)

	// 测试HSet和HGet
	err = client.HSet(hashKey, "name", "Alice")
	if err != nil {
		t.Errorf("HSet failed: %v", err)
	}

	value, err := client.HGet(hashKey, "name")
	if err != nil {
		t.Errorf("HGet failed: %v", err)
	}
	if value != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", value)
	}

	// 测试HMSet和HMGet
	fields := map[string]interface{}{
		"age":    30,
		"email":  "alice@example.com",
		"active": true,
	}
	err = client.HMSet(hashKey, fields)
	if err != nil {
		t.Errorf("HMSet failed: %v", err)
	}

	values, err := client.HMGet(hashKey, "name", "age", "email")
	if err != nil {
		t.Errorf("HMGet failed: %v", err)
	}
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// 测试HGetAll
	allFields, err := client.HGetAll(hashKey)
	if err != nil {
		t.Errorf("HGetAll failed: %v", err)
	}
	if len(allFields) != 4 { // name, age, email, active
		t.Errorf("Expected 4 fields, got %d", len(allFields))
	}

	// 测试HExists
	exists, err := client.HExists(hashKey, "name")
	if err != nil {
		t.Errorf("HExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected field 'name' to exist")
	}

	// 测试HLen
	length, err := client.HLen(hashKey)
	if err != nil {
		t.Errorf("HLen failed: %v", err)
	}
	if length != 4 {
		t.Errorf("Expected length 4, got %d", length)
	}

	// 测试HDel
	deleted, err := client.HDel(hashKey, "active")
	if err != nil {
		t.Errorf("HDel failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	// 清理测试数据
	client.Del(hashKey)
}

func TestRedisClient_ListOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	listKey := "test:list"
	client.Del(listKey)

	// 测试LPush和RPush
	count, err := client.LPush(listKey, "item1", "item2")
	if err != nil {
		t.Errorf("LPush failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	count, err = client.RPush(listKey, "item3", "item4")
	if err != nil {
		t.Errorf("RPush failed: %v", err)
	}
	if count != 4 {
		t.Errorf("Expected count 4, got %d", count)
	}

	// 测试LLen
	length, err := client.LLen(listKey)
	if err != nil {
		t.Errorf("LLen failed: %v", err)
	}
	if length != 4 {
		t.Errorf("Expected length 4, got %d", length)
	}

	// 测试LRange
	items, err := client.LRange(listKey, 0, -1)
	if err != nil {
		t.Errorf("LRange failed: %v", err)
	}
	if len(items) != 4 {
		t.Errorf("Expected 4 items, got %d", len(items))
	}

	// 测试LIndex
	item, err := client.LIndex(listKey, 0)
	if err != nil {
		t.Errorf("LIndex failed: %v", err)
	}
	if item != "item2" { // 因为LPush是从左边插入
		t.Errorf("Expected 'item2', got '%s'", item)
	}

	// 测试LPop和RPop
	leftItem, err := client.LPop(listKey)
	if err != nil {
		t.Errorf("LPop failed: %v", err)
	}
	if leftItem != "item2" {
		t.Errorf("Expected 'item2', got '%s'", leftItem)
	}

	rightItem, err := client.RPop(listKey)
	if err != nil {
		t.Errorf("RPop failed: %v", err)
	}
	if rightItem != "item4" {
		t.Errorf("Expected 'item4', got '%s'", rightItem)
	}

	// 清理测试数据
	client.Del(listKey)
}

func TestRedisClient_SetOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	setKey := "test:set"
	client.Del(setKey)

	// 测试SAdd
	count, err := client.SAdd(setKey, "member1", "member2", "member3")
	if err != nil {
		t.Errorf("SAdd failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// 测试SCard
	size, err := client.SCard(setKey)
	if err != nil {
		t.Errorf("SCard failed: %v", err)
	}
	if size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}

	// 测试SIsMember
	exists, err := client.SIsMember(setKey, "member1")
	if err != nil {
		t.Errorf("SIsMember failed: %v", err)
	}
	if !exists {
		t.Error("Expected member1 to exist in set")
	}

	// 测试SMembers
	members, err := client.SMembers(setKey)
	if err != nil {
		t.Errorf("SMembers failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// 测试SRem
	removed, err := client.SRem(setKey, "member2")
	if err != nil {
		t.Errorf("SRem failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}

	// 清理测试数据
	client.Del(setKey)
}

func TestRedisClient_SortedSetOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	zsetKey := "test:zset"
	client.Del(zsetKey)

	// 测试ZAdd
	members := []*redis.Z{
		{Score: 1, Member: "member1"},
		{Score: 2, Member: "member2"},
		{Score: 3, Member: "member3"},
	}
	count, err := client.ZAdd(zsetKey, members...)
	if err != nil {
		t.Errorf("ZAdd failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// 测试ZCard
	size, err := client.ZCard(zsetKey)
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}

	// 测试ZScore
	score, err := client.ZScore(zsetKey, "member2")
	if err != nil {
		t.Errorf("ZScore failed: %v", err)
	}
	if score != 2 {
		t.Errorf("Expected score 2, got %f", score)
	}

	// 测试ZRange
	rangeMembers, err := client.ZRange(zsetKey, 0, -1)
	if err != nil {
		t.Errorf("ZRange failed: %v", err)
	}
	if len(rangeMembers) != 3 {
		t.Errorf("Expected 3 members, got %d", len(rangeMembers))
	}

	// 测试ZRank
	rank, err := client.ZRank(zsetKey, "member1")
	if err != nil {
		t.Errorf("ZRank failed: %v", err)
	}
	if rank != 0 { // 最小分数，排名第0
		t.Errorf("Expected rank 0, got %d", rank)
	}

	// 清理测试数据
	client.Del(zsetKey)
}

func TestRedisClient_KeyOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	testKey := "test:key"
	client.Del(testKey)

	// 测试Set和Exists
	err = client.Set(testKey, "test value", 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	exists, err := client.Exists(testKey)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if exists != 1 {
		t.Errorf("Expected exists 1, got %d", exists)
	}

	// 测试Type
	keyType, err := client.Type(testKey)
	if err != nil {
		t.Errorf("Type failed: %v", err)
	}
	if keyType != "string" {
		t.Errorf("Expected type 'string', got '%s'", keyType)
	}

	// 测试Expire和TTL
	success, err := client.Expire(testKey, time.Minute)
	if err != nil {
		t.Errorf("Expire failed: %v", err)
	}
	if !success {
		t.Error("Expected Expire to return true")
	}

	ttl, err := client.TTL(testKey)
	if err != nil {
		t.Errorf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("Expected TTL between 0 and 1 minute, got %v", ttl)
	}

	// 测试Rename
	newKey := "test:renamed_key"
	client.Del(newKey)

	err = client.Rename(testKey, newKey)
	if err != nil {
		t.Errorf("Rename failed: %v", err)
	}

	// 验证新键存在，旧键不存在
	exists, _ = client.Exists(newKey)
	if exists != 1 {
		t.Error("Expected renamed key to exist")
	}

	exists, _ = client.Exists(testKey)
	if exists != 0 {
		t.Error("Expected original key to not exist")
	}

	// 清理测试数据
	client.Del(newKey)
}

func TestRedisClient_JSONOperations(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	jsonKey := "test:json"
	hashKey := "test:json_hash"
	client.Del(jsonKey, hashKey)

	// 测试SetJSON和GetJSON
	user := TestUser{
		ID:       1,
		Name:     "Alice",
		Email:    "alice@example.com",
		Age:      30,
		CreateAt: time.Now().Unix(),
	}

	err = client.SetJSON(jsonKey, user, time.Hour)
	if err != nil {
		t.Errorf("SetJSON failed: %v", err)
	}

	var retrievedUser TestUser
	err = client.GetJSON(jsonKey, &retrievedUser)
	if err != nil {
		t.Errorf("GetJSON failed: %v", err)
	}

	if retrievedUser.Name != user.Name {
		t.Errorf("Expected name '%s', got '%s'", user.Name, retrievedUser.Name)
	}
	if retrievedUser.Age != user.Age {
		t.Errorf("Expected age %d, got %d", user.Age, retrievedUser.Age)
	}

	// 测试HSetJSON和HGetJSON
	err = client.HSetJSON(hashKey, "user:1", user)
	if err != nil {
		t.Errorf("HSetJSON failed: %v", err)
	}

	var hashUser TestUser
	err = client.HGetJSON(hashKey, "user:1", &hashUser)
	if err != nil {
		t.Errorf("HGetJSON failed: %v", err)
	}

	if hashUser.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, hashUser.Email)
	}

	// 清理测试数据
	client.Del(jsonKey, hashKey)
}

func TestRedisClient_PubSub(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	channel := "test:channel"
	message := "hello pub/sub"

	// 创建订阅者
	pubsub := client.Subscribe(channel)
	defer pubsub.Close()

	// 等待订阅确认
	_, err = pubsub.Receive(client.GetContext())
	if err != nil {
		t.Errorf("Failed to receive subscription confirmation: %v", err)
		return
	}

	// 发布消息
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := client.Publish(channel, message)
		if err != nil {
			t.Errorf("Publish failed: %v", err)
		}
	}()

	// 接收消息
	msg, err := pubsub.ReceiveMessage(client.GetContext())
	if err != nil {
		t.Errorf("ReceiveMessage failed: %v", err)
		return
	}

	if msg.Payload != message {
		t.Errorf("Expected message '%s', got '%s'", message, msg.Payload)
	}
}

func TestRedisClient_Transaction(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	key1 := "test:tx1"
	key2 := "test:tx2"
	client.Del(key1, key2)

	// 测试事务管道
	pipe := client.TxPipeline()
	pipe.Set(client.GetContext(), key1, "value1", 0)
	pipe.Set(client.GetContext(), key2, "value2", 0)
	pipe.Incr(client.GetContext(), key1+"_counter")

	cmds, err := pipe.Exec(client.GetContext())
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	if len(cmds) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(cmds))
	}

	// 验证结果
	val1, _ := client.Get(key1)
	val2, _ := client.Get(key2)
	counter, _ := client.Get(key1 + "_counter")

	if val1 != "value1" {
		t.Errorf("Expected value1, got %s", val1)
	}
	if val2 != "value2" {
		t.Errorf("Expected value2, got %s", val2)
	}
	if counter != "1" {
		t.Errorf("Expected counter 1, got %s", counter)
	}

	// 清理测试数据
	client.Del(key1, key2, key1+"_counter")
}

func TestRedisClient_ErrorHandling(t *testing.T) {
	client, err := getTestRedisClient()
	if err != nil {
		t.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	var errorReceived bool
	var lastError error

	// 设置错误回调
	client.SetCallbacks(func(err error) {
		errorReceived = true
		lastError = err
		t.Logf("Error callback triggered: %v", err)
	}, func() {
		t.Log("Reconnect callback triggered")
	})

	// 尝试对不存在的键执行不当操作
	// 这应该不会触发错误回调，因为Redis会返回nil或0
	_, err = client.Get("nonexistent:key")
	if err != nil {
		t.Errorf("Get on nonexistent key should not error: %v", err)
	}

	// 尝试获取不存在键的JSON
	var dummy struct{}
	err = client.GetJSON("nonexistent:json", &dummy)
	if err != redis.Nil {
		t.Errorf("Expected redis.Nil error, got: %v", err)
	}

	// 测试格式错误的JSON
	client.Set("test:bad_json", "invalid json", time.Minute)
	err = client.GetJSON("test:bad_json", &dummy)
	if err == nil {
		t.Error("Expected JSON unmarshal error, got nil")
	}

	// 使用变量避免编译警告
	_ = errorReceived
	_ = lastError

	client.Del("test:bad_json")
}

// 基准测试
func BenchmarkRedisClient_Set(b *testing.B) {
	client, err := getTestRedisClient()
	if err != nil {
		b.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:set:%d", i)
		err := client.Set(key, fmt.Sprintf("value%d", i), time.Hour)
		if err != nil {
			b.Errorf("Set failed: %v", err)
		}
	}
}

func BenchmarkRedisClient_Get(b *testing.B) {
	client, err := getTestRedisClient()
	if err != nil {
		b.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	// 预设一些数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench:get:%d", i)
		client.Set(key, fmt.Sprintf("value%d", i), time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:get:%d", i%1000)
		_, err := client.Get(key)
		if err != nil {
			b.Errorf("Get failed: %v", err)
		}
	}
}

func BenchmarkRedisClient_HSet(b *testing.B) {
	client, err := getTestRedisClient()
	if err != nil {
		b.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	hashKey := "bench:hash"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field := fmt.Sprintf("field%d", i)
		value := fmt.Sprintf("value%d", i)
		err := client.HSet(hashKey, field, value)
		if err != nil {
			b.Errorf("HSet failed: %v", err)
		}
	}
}

func BenchmarkRedisClient_SetJSON(b *testing.B) {
	client, err := getTestRedisClient()
	if err != nil {
		b.Skipf("Redis server not available: %v", err)
		return
	}
	defer client.Close()

	user := TestUser{
		ID:       1,
		Name:     "BenchUser",
		Email:    "bench@example.com",
		Age:      25,
		CreateAt: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:json:%d", i)
		err := client.SetJSON(key, user, time.Hour)
		if err != nil {
			b.Errorf("SetJSON failed: %v", err)
		}
	}
}
