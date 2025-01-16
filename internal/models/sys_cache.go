package models

import (
	"time"
	g "wp_template_display/internal/global"
)

type SysCache struct {
	Model
	Key    string    `json:"key" gorm:"size:20;not null;unique;comment:键"`
	Value  string    `json:"value" gorm:"type:text;comment:值"` // 统一转为JSON字符串
	Expire time.Time `json:"expire" gorm:"comment:过期时间"`
}

// 设置缓存
func CacheSet(key string, value string) {
	cache := SysCache{
		Key:    key,
		Value:  value,
		Expire: time.Now().Add(time.Hour * 24),
	}
	g.DB.Create(&cache)
}

// 获取缓存
func CacheGet(key string) string {
	// 查询
	cache := SysCache{}
	g.DB.Where("key = ?", key).First(&cache)

	// 如果缓存存在，则返回缓存值
	if cache.Id > 0 {
		if cache.Expire.After(time.Now()) {
			return cache.Value
		} else {
			// 如果缓存过期，则删除缓存
			g.DB.Delete(&cache)
		}
	}

	return ""
}

// 删除缓存
func CacheDel(key string) error {
	tx := g.DB.Where("key = ?", key).Delete(&SysCache{})
	return tx.Error
}
