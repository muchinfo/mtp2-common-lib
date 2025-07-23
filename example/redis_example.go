package main

import (
	"fmt"
	"log"
	"time"

	"github.com/muchinfo/mtp2-common-lib/redis"
	redisLib "github.com/redis/go-redis/v9"
)

// 示例数据结构
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	CreateAt int64  `json:"create_at"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Stock       int     `json:"stock"`
}

func RunRedisExample() {
	// 创建Redis客户端配置
	config := redis.RedisConfig{
		Address:      "localhost:6379", // Redis服务器地址
		Password:     "",               // 如果有密码则填写
		Database:     0,                // 使用数据库0
		PoolSize:     10,               // 连接池大小
		MinIdleConns: 2,                // 最小空闲连接
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读取超时
		WriteTimeout: 3 * time.Second,  // 写入超时
		MaxRetries:   3,                // 最大重试次数
		PoolTimeout:  4 * time.Second,  // 获取连接超时
		IdleTimeout:  5 * time.Minute,  // 空闲连接超时
	}

	// 创建Redis客户端
	client, err := redis.NewRedisClient(config)
	if err != nil {
		log.Fatalf("❌ Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// 设置回调函数
	client.SetCallbacks(
		func(err error) {
			log.Printf("⚠️ Redis Error: %v", err)
		},
		func() {
			log.Println("🔄 Redis Reconnected")
		},
	)

	log.Println("🎉 Redis client created successfully!")

	// 演示基础操作
	demonstrateBasicOperations(client)

	// 演示字符串操作
	demonstrateStringOperations(client)

	// 演示哈希操作
	demonstrateHashOperations(client)

	// 演示列表操作
	demonstrateListOperations(client)

	// 演示集合操作
	demonstrateSetOperations(client)

	// 演示有序集合操作
	demonstrateSortedSetOperations(client)

	// 演示JSON操作
	demonstrateJSONOperations(client)

	// 演示键操作
	demonstrateKeyOperations(client)

	// 演示发布订阅
	demonstratePubSub(client)

	// 演示事务操作
	demonstrateTransactions(client)

	// 演示缓存模式
	demonstrateCachingPatterns(client)

	// 演示性能监控
	demonstratePerformanceMonitoring(client)

	log.Println("✅ Redis example completed successfully!")
}

func demonstrateBasicOperations(client *redis.RedisClient) {
	log.Println("\n🔧 === 基础操作演示 ===")

	// 测试连接
	if err := client.Ping(); err != nil {
		log.Printf("❌ Ping failed: %v", err)
		return
	}
	log.Println("✅ Redis connection is healthy")

	// 获取连接池统计
	stats := client.PoolStats()
	log.Printf("📊 Pool Stats - Total: %d, Idle: %d, Stale: %d",
		stats.TotalConns, stats.IdleConns, stats.StaleConns)
}

func demonstrateStringOperations(client *redis.RedisClient) {
	log.Println("\n📝 === 字符串操作演示 ===")

	// 设置和获取字符串
	err := client.Set("user:1:name", "Alice", 10*time.Minute)
	if err != nil {
		log.Printf("❌ Set failed: %v", err)
		return
	}

	name, err := client.Get("user:1:name")
	if err != nil {
		log.Printf("❌ Get failed: %v", err)
		return
	}
	log.Printf("✅ Retrieved name: %s", name)

	// 计数器操作
	client.Set("page:views", "100", 0)
	views, err := client.IncrBy("page:views", 5)
	if err != nil {
		log.Printf("❌ IncrBy failed: %v", err)
	} else {
		log.Printf("✅ Page views after increment: %d", views)
	}

	// 批量操作
	err = client.MSet("product:1", "Laptop", "product:2", "Phone", "product:3", "Tablet")
	if err != nil {
		log.Printf("❌ MSet failed: %v", err)
	} else {
		log.Println("✅ Multiple products set")
	}

	products, err := client.MGet("product:1", "product:2", "product:3")
	if err != nil {
		log.Printf("❌ MGet failed: %v", err)
	} else {
		log.Printf("✅ Retrieved products: %v", products)
	}

	// 原子操作
	oldValue, err := client.GetSet("user:1:name", "Bob")
	if err != nil {
		log.Printf("❌ GetSet failed: %v", err)
	} else {
		log.Printf("✅ GetSet - old value: %s", oldValue)
	}
}

func demonstrateHashOperations(client *redis.RedisClient) {
	log.Println("\n🗂️ === 哈希操作演示 ===")

	userKey := "user:123"

	// 设置用户信息
	userInfo := map[string]interface{}{
		"name":      "Charlie",
		"email":     "charlie@example.com",
		"age":       28,
		"city":      "New York",
		"active":    true,
		"balance":   1500.50,
		"join_date": time.Now().Format("2006-01-02"),
	}

	err := client.HMSet(userKey, userInfo)
	if err != nil {
		log.Printf("❌ HMSet failed: %v", err)
		return
	}
	log.Println("✅ User information set")

	// 获取单个字段
	email, err := client.HGet(userKey, "email")
	if err != nil {
		log.Printf("❌ HGet failed: %v", err)
	} else {
		log.Printf("✅ User email: %s", email)
	}

	// 获取多个字段
	fields, err := client.HMGet(userKey, "name", "age", "city")
	if err != nil {
		log.Printf("❌ HMGet failed: %v", err)
	} else {
		log.Printf("✅ User info: name=%v, age=%v, city=%v", fields[0], fields[1], fields[2])
	}

	// 获取所有字段
	allFields, err := client.HGetAll(userKey)
	if err != nil {
		log.Printf("❌ HGetAll failed: %v", err)
	} else {
		log.Printf("✅ All user fields count: %d", len(allFields))
		for k, v := range allFields {
			log.Printf("   %s: %s", k, v)
		}
	}

	// 字段操作
	exists, err := client.HExists(userKey, "phone")
	if err != nil {
		log.Printf("❌ HExists failed: %v", err)
	} else {
		log.Printf("✅ Phone field exists: %v", exists)
	}

	// 添加新字段
	err = client.HSet(userKey, "phone", "+1-555-1234")
	if err != nil {
		log.Printf("❌ HSet phone failed: %v", err)
	} else {
		log.Println("✅ Phone number added")
	}

	// 数值操作
	newBalance, err := client.HIncrBy(userKey, "login_count", 1)
	if err != nil {
		log.Printf("❌ HIncrBy failed: %v", err)
	} else {
		log.Printf("✅ Login count incremented to: %d", newBalance)
	}
}

func demonstrateListOperations(client *redis.RedisClient) {
	log.Println("\n📜 === 列表操作演示 ===")

	queueKey := "message:queue"
	historyKey := "user:browsing:history"

	// 消息队列示例 (FIFO)
	log.Println("📤 Message Queue Example:")

	// 添加消息到队列尾部
	messages := []interface{}{"Welcome message", "Notification 1", "Notification 2", "Alert message"}
	count, err := client.RPush(queueKey, messages...)
	if err != nil {
		log.Printf("❌ RPush failed: %v", err)
	} else {
		log.Printf("✅ Added %d messages to queue, total: %d", len(messages), count)
	}

	// 从队列头部处理消息
	for i := 0; i < 2; i++ {
		message, err := client.LPop(queueKey)
		if err != nil {
			log.Printf("❌ LPop failed: %v", err)
		} else if message != "" {
			log.Printf("✅ Processed message: %s", message)
		}
	}

	// 查看剩余消息
	remaining, err := client.LRange(queueKey, 0, -1)
	if err != nil {
		log.Printf("❌ LRange failed: %v", err)
	} else {
		log.Printf("✅ Remaining messages: %v", remaining)
	}

	// 浏览历史示例 (最新的在前面)
	log.Println("🌐 Browsing History Example:")

	pages := []interface{}{
		"https://example.com/home",
		"https://example.com/products",
		"https://example.com/about",
		"https://example.com/contact",
	}

	// 添加到历史记录头部
	for _, page := range pages {
		client.LPush(historyKey, page)
		log.Printf("✅ Visited: %s", page)
	}

	// 获取最近5条历史记录
	history, err := client.LRange(historyKey, 0, 4)
	if err != nil {
		log.Printf("❌ Failed to get history: %v", err)
	} else {
		log.Printf("✅ Recent browsing history: %v", history)
	}

	// 限制历史记录数量 (保留最新10条)
	err = client.LTrim(historyKey, 0, 9)
	if err != nil {
		log.Printf("❌ LTrim failed: %v", err)
	} else {
		log.Println("✅ History trimmed to last 10 entries")
	}
}

func demonstrateSetOperations(client *redis.RedisClient) {
	log.Println("\n🎯 === 集合操作演示 ===")

	tagsKey := "article:123:tags"
	followersKey := "user:456:followers"

	// 文章标签示例
	log.Println("🏷️ Article Tags Example:")

	tags := []interface{}{"golang", "redis", "database", "caching", "performance"}
	count, err := client.SAdd(tagsKey, tags...)
	if err != nil {
		log.Printf("❌ SAdd failed: %v", err)
	} else {
		log.Printf("✅ Added %d tags, total unique: %d", len(tags), count)
	}

	// 检查标签是否存在
	exists, err := client.SIsMember(tagsKey, "redis")
	if err != nil {
		log.Printf("❌ SIsMember failed: %v", err)
	} else {
		log.Printf("✅ Tag 'redis' exists: %v", exists)
	}

	// 获取所有标签
	allTags, err := client.SMembers(tagsKey)
	if err != nil {
		log.Printf("❌ SMembers failed: %v", err)
	} else {
		log.Printf("✅ All tags: %v", allTags)
	}

	// 关注者示例
	log.Println("👥 Followers Example:")

	followers := []interface{}{"user:100", "user:200", "user:300", "user:400"}
	client.SAdd(followersKey, followers...)

	followerCount, err := client.SCard(followersKey)
	if err != nil {
		log.Printf("❌ SCard failed: %v", err)
	} else {
		log.Printf("✅ Follower count: %d", followerCount)
	}

	// 随机获取一个关注者
	randomFollower, err := client.SRandMember(followersKey)
	if err != nil {
		log.Printf("❌ SRandMember failed: %v", err)
	} else if randomFollower != "" {
		log.Printf("✅ Random follower: %s", randomFollower)
	}

	// 移除一个关注者
	removed, err := client.SRem(followersKey, "user:200")
	if err != nil {
		log.Printf("❌ SRem failed: %v", err)
	} else {
		log.Printf("✅ Removed %d followers", removed)
	}
}

func demonstrateSortedSetOperations(client *redis.RedisClient) {
	log.Println("\n🏆 === 有序集合操作演示 ===")

	leaderboardKey := "game:leaderboard"

	// 游戏排行榜示例
	log.Println("🎮 Game Leaderboard Example:")

	// 添加玩家分数
	players := []*redisLib.Z{
		{Score: 1500, Member: "player:alice"},
		{Score: 2300, Member: "player:bob"},
		{Score: 1800, Member: "player:charlie"},
		{Score: 2100, Member: "player:diana"},
		{Score: 1950, Member: "player:eve"},
	}

	count, err := client.ZAdd(leaderboardKey, players...)
	if err != nil {
		log.Printf("❌ ZAdd failed: %v", err)
	} else {
		log.Printf("✅ Added %d players to leaderboard", count)
	}

	// 获取排行榜前3名 (分数从高到低)
	topPlayers, err := client.ZRangeWithScores(leaderboardKey, -3, -1)
	if err != nil {
		log.Printf("❌ ZRangeWithScores failed: %v", err)
	} else {
		log.Println("🥇 Top 3 Players:")
		for i, player := range topPlayers {
			log.Printf("   %d. %s - %.0f points", len(topPlayers)-i, player.Member, player.Score)
		}
	}

	// 获取特定玩家的排名和分数
	playerName := "player:charlie"
	rank, err := client.ZRank(leaderboardKey, playerName)
	if err != nil {
		log.Printf("❌ ZRank failed: %v", err)
	} else {
		log.Printf("✅ %s rank: %d (0-based)", playerName, rank)
	}

	score, err := client.ZScore(leaderboardKey, playerName)
	if err != nil {
		log.Printf("❌ ZScore failed: %v", err)
	} else {
		log.Printf("✅ %s score: %.0f", playerName, score)
	}

	// 获取排行榜总人数
	totalPlayers, err := client.ZCard(leaderboardKey)
	if err != nil {
		log.Printf("❌ ZCard failed: %v", err)
	} else {
		log.Printf("✅ Total players: %d", totalPlayers)
	}

	// 更新玩家分数
	newPlayers := []*redisLib.Z{
		{Score: 2500, Member: "player:alice"}, // Alice 提升分数
	}
	client.ZAdd(leaderboardKey, newPlayers...)
	log.Printf("✅ Updated player:alice score to 2500")
}

func demonstrateJSONOperations(client *redis.RedisClient) {
	log.Println("\n📋 === JSON操作演示 ===")

	// 用户对象示例
	user := User{
		ID:       12345,
		Name:     "David Wilson",
		Email:    "david@example.com",
		Age:      32,
		CreateAt: time.Now().Unix(),
	}

	userKey := "user:profile:12345"

	// 存储JSON对象
	err := client.SetJSON(userKey, user, time.Hour)
	if err != nil {
		log.Printf("❌ SetJSON failed: %v", err)
		return
	}
	log.Printf("✅ User profile saved: %s", userKey)

	// 读取JSON对象
	var retrievedUser User
	err = client.GetJSON(userKey, &retrievedUser)
	if err != nil {
		log.Printf("❌ GetJSON failed: %v", err)
	} else {
		log.Printf("✅ Retrieved user: %s (ID: %d, Age: %d)",
			retrievedUser.Name, retrievedUser.ID, retrievedUser.Age)
	}

	// 产品目录示例
	products := []Product{
		{ID: 1, Name: "Gaming Laptop", Price: 1299.99, Category: "Electronics", Stock: 15},
		{ID: 2, Name: "Wireless Mouse", Price: 29.99, Category: "Accessories", Stock: 50},
		{ID: 3, Name: "Mechanical Keyboard", Price: 89.99, Category: "Accessories", Stock: 25},
	}

	catalogKey := "products:catalog"

	// 使用哈希存储多个产品
	for _, product := range products {
		fieldKey := fmt.Sprintf("product:%d", product.ID)
		err := client.HSetJSON(catalogKey, fieldKey, product)
		if err != nil {
			log.Printf("❌ HSetJSON failed for product %d: %v", product.ID, err)
		}
	}
	log.Printf("✅ Product catalog saved with %d products", len(products))

	// 读取特定产品
	var laptop Product
	err = client.HGetJSON(catalogKey, "product:1", &laptop)
	if err != nil {
		log.Printf("❌ HGetJSON failed: %v", err)
	} else {
		log.Printf("✅ Retrieved product: %s - $%.2f (Stock: %d)",
			laptop.Name, laptop.Price, laptop.Stock)
	}

	// 获取所有产品信息
	allProductFields, err := client.HKeys(catalogKey)
	if err != nil {
		log.Printf("❌ HKeys failed: %v", err)
	} else {
		log.Printf("✅ Total products in catalog: %d", len(allProductFields))
		for _, field := range allProductFields {
			log.Printf("   - %s", field)
		}
	}
}

func demonstrateKeyOperations(client *redis.RedisClient) {
	log.Println("\n🔑 === 键操作演示 ===")

	// 创建一些测试键
	testKeys := map[string]interface{}{
		"temp:session:abc123": "user_data",
		"temp:cache:page1":    "cached_content",
		"temp:counter:daily":  100,
	}

	for key, value := range testKeys {
		client.Set(key, value, 0)
	}
	log.Printf("✅ Created %d test keys", len(testKeys))

	// 检查键是否存在
	exists, err := client.Exists("temp:session:abc123", "temp:cache:page1", "nonexistent:key")
	if err != nil {
		log.Printf("❌ Exists failed: %v", err)
	} else {
		log.Printf("✅ Keys exist count: %d", exists)
	}

	// 获取键的类型
	keyType, err := client.Type("temp:counter:daily")
	if err != nil {
		log.Printf("❌ Type failed: %v", err)
	} else {
		log.Printf("✅ Key type: %s", keyType)
	}

	// 设置过期时间
	success, err := client.Expire("temp:session:abc123", 5*time.Minute)
	if err != nil {
		log.Printf("❌ Expire failed: %v", err)
	} else if success {
		log.Println("✅ Session expiry set to 5 minutes")
	}

	// 查看TTL
	ttl, err := client.TTL("temp:session:abc123")
	if err != nil {
		log.Printf("❌ TTL failed: %v", err)
	} else {
		log.Printf("✅ Session TTL: %v", ttl)
	}

	// 模式匹配查找键
	keys, err := client.Keys("temp:*")
	if err != nil {
		log.Printf("❌ Keys failed: %v", err)
	} else {
		log.Printf("✅ Found %d temp keys:", len(keys))
		for _, key := range keys {
			log.Printf("   - %s", key)
		}
	}

	// 重命名键
	err = client.Rename("temp:counter:daily", "temp:counter:renamed")
	if err != nil {
		log.Printf("❌ Rename failed: %v", err)
	} else {
		log.Println("✅ Key renamed successfully")
	}

	// 清理测试键
	allTempKeys, _ := client.Keys("temp:*")
	if len(allTempKeys) > 0 {
		deleted, err := client.Del(allTempKeys...)
		if err != nil {
			log.Printf("❌ Cleanup failed: %v", err)
		} else {
			log.Printf("✅ Cleaned up %d temp keys", deleted)
		}
	}
}

func demonstratePubSub(client *redis.RedisClient) {
	log.Println("\n📢 === 发布订阅演示 ===")

	channel := "notifications"

	// 创建订阅者
	pubsub := client.Subscribe(channel)
	defer pubsub.Close()

	log.Printf("✅ Subscribed to channel: %s", channel)

	// 在goroutine中发布消息
	go func() {
		time.Sleep(100 * time.Millisecond) // 确保订阅已建立

		messages := []string{
			"System maintenance scheduled",
			"New feature released",
			"Performance optimization completed",
		}

		for i, msg := range messages {
			time.Sleep(200 * time.Millisecond)
			err := client.Publish(channel, fmt.Sprintf("Message %d: %s", i+1, msg))
			if err != nil {
				log.Printf("❌ Publish failed: %v", err)
			} else {
				log.Printf("📤 Published: Message %d", i+1)
			}
		}

		// 发送结束信号
		client.Publish(channel, "END")
	}()

	// 接收消息
	messageCount := 0
	for {
		msg, err := pubsub.ReceiveMessage(client.GetContext())
		if err != nil {
			log.Printf("❌ ReceiveMessage failed: %v", err)
			break
		}

		if msg.Payload == "END" {
			log.Println("📬 Received end signal")
			break
		}

		messageCount++
		log.Printf("📥 Received: %s", msg.Payload)

		// 防止无限等待
		if messageCount >= 10 {
			break
		}
	}

	log.Printf("✅ PubSub demo completed, received %d messages", messageCount)
}

func demonstrateTransactions(client *redis.RedisClient) {
	log.Println("\n💳 === 事务操作演示 ===")

	// 银行转账示例
	accountA := "account:alice"
	accountB := "account:bob"

	// 初始化账户余额
	client.Set(accountA, "1000", 0)
	client.Set(accountB, "500", 0)
	log.Println("💰 Initial balances - Alice: $1000, Bob: $500")

	// 执行转账事务 (Alice向Bob转账$200)
	transferAmount := 200

	pipe := client.TxPipeline()
	pipe.DecrBy(client.GetContext(), accountA, int64(transferAmount))
	pipe.IncrBy(client.GetContext(), accountB, int64(transferAmount))
	pipe.Set(client.GetContext(), "transfer:log",
		fmt.Sprintf("Transfer $%d from Alice to Bob at %s", transferAmount, time.Now().Format(time.RFC3339)),
		time.Hour)

	results, err := pipe.Exec(client.GetContext())
	if err != nil {
		log.Printf("❌ Transaction failed: %v", err)
		return
	}

	log.Printf("✅ Transaction completed with %d operations", len(results))

	// 验证转账结果
	balanceA, _ := client.Get(accountA)
	balanceB, _ := client.Get(accountB)
	transferLog, _ := client.Get("transfer:log")

	log.Printf("💰 Final balances - Alice: $%s, Bob: $%s", balanceA, balanceB)
	log.Printf("📝 Transfer log: %s", transferLog)

	// 批量操作示例
	log.Println("📦 Batch Operations Example:")

	pipe2 := client.TxPipeline()

	// 批量创建用户会话
	sessions := map[string]string{
		"session:user1": "active",
		"session:user2": "active",
		"session:user3": "active",
	}

	for key, value := range sessions {
		pipe2.Set(client.GetContext(), key, value, 30*time.Minute)
	}

	// 更新计数器
	pipe2.Incr(client.GetContext(), "stats:active_sessions")
	pipe2.Set(client.GetContext(), "stats:last_update", time.Now().Unix(), 0)

	batchResults, err := pipe2.Exec(client.GetContext())
	if err != nil {
		log.Printf("❌ Batch operation failed: %v", err)
	} else {
		log.Printf("✅ Batch operation completed: %d operations", len(batchResults))
	}

	// 清理测试数据
	client.Del(accountA, accountB, "transfer:log", "stats:active_sessions", "stats:last_update")
	for key := range sessions {
		client.Del(key)
	}
}

func demonstrateCachingPatterns(client *redis.RedisClient) {
	log.Println("\n🎯 === 缓存模式演示 ===")

	// 缓存穿透保护示例
	log.Println("🛡️ Cache-aside Pattern:")

	getUserFromCache := func(userID int) (*User, error) {
		key := fmt.Sprintf("cache:user:%d", userID)

		var user User
		err := client.GetJSON(key, &user)
		if err == redisLib.Nil {
			// 缓存未命中，模拟从数据库获取
			log.Printf("🔍 Cache miss for user %d, fetching from database...", userID)

			// 模拟数据库查询
			time.Sleep(50 * time.Millisecond)
			user = User{
				ID:       userID,
				Name:     fmt.Sprintf("User%d", userID),
				Email:    fmt.Sprintf("user%d@example.com", userID),
				Age:      25 + userID%10,
				CreateAt: time.Now().Unix(),
			}

			// 将结果存入缓存
			client.SetJSON(key, user, 10*time.Minute)
			log.Printf("💾 User %d cached for 10 minutes", userID)

			return &user, nil
		} else if err != nil {
			return nil, err
		}

		log.Printf("⚡ Cache hit for user %d", userID)
		return &user, nil
	}

	// 测试缓存模式
	for i := 1; i <= 3; i++ {
		user, err := getUserFromCache(100 + i)
		if err != nil {
			log.Printf("❌ Failed to get user %d: %v", 100+i, err)
		} else {
			log.Printf("✅ Got user: %s (%s)", user.Name, user.Email)
		}
	}

	// 再次获取相同用户 (应该命中缓存)
	log.Println("🔄 Second request (should hit cache):")
	user, _ := getUserFromCache(101)
	log.Printf("✅ Got cached user: %s", user.Name)

	// 热门内容排行示例
	log.Println("🔥 Popular Content Ranking:")

	// 模拟内容访问
	contents := []string{"article:123", "video:456", "post:789", "tutorial:101"}

	for _, content := range contents {
		// 随机访问次数
		views := 1 + (time.Now().UnixNano() % 50)
		client.ZAdd("popular:content", &redisLib.Z{
			Score:  float64(views),
			Member: content,
		})
		log.Printf("📈 %s viewed %d times", content, views)
	}

	// 获取热门排行
	popular, err := client.ZRangeWithScores("popular:content", -5, -1)
	if err != nil {
		log.Printf("❌ Failed to get popular content: %v", err)
	} else {
		log.Println("🏆 Top Popular Content:")
		for i, item := range popular {
			log.Printf("   %d. %s - %.0f views", len(popular)-i, item.Member, item.Score)
		}
	}

	// 分布式锁示例
	log.Println("🔒 Distributed Lock Example:")

	lockKey := "lock:critical_section"
	lockValue := fmt.Sprintf("process_%d_%d", 12345, time.Now().UnixNano())

	// 尝试获取锁
	err = client.Set(lockKey, lockValue, 30*time.Second)
	// 注意：实际的分布式锁需要使用 SET NX EX 命令，这里简化演示

	if err != nil {
		log.Printf("❌ Failed to acquire lock: %v", err)
	} else {
		log.Printf("✅ Lock acquired: %s", lockKey)

		// 模拟临界区操作
		time.Sleep(100 * time.Millisecond)
		log.Println("⚙️ Performing critical operation...")

		// 释放锁
		client.Del(lockKey)
		log.Println("🔓 Lock released")
	}
}

func demonstratePerformanceMonitoring(client *redis.RedisClient) {
	log.Println("\n📊 === 性能监控演示 ===")

	// 连接池统计
	stats := client.PoolStats()
	log.Printf("🔧 Connection Pool Stats:")
	log.Printf("   Total Connections: %d", stats.TotalConns)
	log.Printf("   Idle Connections: %d", stats.IdleConns)
	log.Printf("   Stale Connections: %d", stats.StaleConns)
	log.Printf("   Hits: %d", stats.Hits)
	log.Printf("   Misses: %d", stats.Misses)
	log.Printf("   Timeouts: %d", stats.Timeouts)

	// 性能测试
	log.Println("⚡ Performance Test:")

	testKey := "perf:test"
	iterations := 1000

	// 写入性能测试
	start := time.Now()
	for i := 0; i < iterations; i++ {
		client.Set(fmt.Sprintf("%s:%d", testKey, i), fmt.Sprintf("value%d", i), time.Minute)
	}
	writeTime := time.Since(start)

	log.Printf("✅ Write Performance: %d operations in %v (%.2f ops/sec)",
		iterations, writeTime, float64(iterations)/writeTime.Seconds())

	// 读取性能测试
	start = time.Now()
	for i := 0; i < iterations; i++ {
		client.Get(fmt.Sprintf("%s:%d", testKey, i))
	}
	readTime := time.Since(start)

	log.Printf("✅ Read Performance: %d operations in %v (%.2f ops/sec)",
		iterations, readTime, float64(iterations)/readTime.Seconds())

	// 批量操作性能
	start = time.Now()
	keys := make([]string, iterations)
	for i := 0; i < iterations; i++ {
		keys[i] = fmt.Sprintf("%s:%d", testKey, i)
	}
	client.MGet(keys...)
	batchTime := time.Since(start)

	log.Printf("✅ Batch Read Performance: %d operations in %v (%.2f ops/sec)",
		iterations, batchTime, float64(iterations)/batchTime.Seconds())

	// 清理测试数据
	log.Println("🧹 Cleaning up performance test data...")
	client.Del(keys...)

	// 最终连接池统计
	finalStats := client.PoolStats()
	log.Printf("🔧 Final Pool Stats:")
	log.Printf("   Total Connections: %d", finalStats.TotalConns)
	log.Printf("   Idle Connections: %d", finalStats.IdleConns)
}
