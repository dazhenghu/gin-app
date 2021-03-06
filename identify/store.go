package identify

import (
    "github.com/dchest/captcha"
    "sync"
    "github.com/gin-gonic/gin"
    "errors"
    "time"
    "github.com/gin-contrib/sessions"
    "github.com/dazhenghu/ginApp/consts"
    "github.com/dazhenghu/ginApp/logs"
    "fmt"
)

type sessionStore struct {
    captcha.Store
    contextIdMap map[string]*contextWithTime // 保存id与gin.context的映射关系，用于将id保存至session中
    expirePeriod time.Duration               // 有效时长，超过这个时长后过期删除
    gcPeriod     time.Duration               // 定时清理contextIdMap中过期元素的周期时长

    sessionMutex sync.RWMutex
    memoryMutex sync.RWMutex
}

type contextWithTime struct {
    context *gin.Context
    createTime time.Time
}
// 无效的id，请刷新
var STORE_ERR_REMOVE_EMPTY = errors.New("invalid id, please reload")
// 重复的id
var STORE_ERR_ID_EXISTS    = errors.New("id exists, please reload")

var sessionStoreInstance *sessionStore
var once sync.Once

func GetSessionStore() *sessionStore {
    once.Do(func() {
        sessionStoreInstance = &sessionStore{
            contextIdMap: make(map[string]*contextWithTime),
        }
    })

    return sessionStoreInstance
}

func (ss *sessionStore) Init(expirePeriod time.Duration, gcPeriod time.Duration)  {
    sessStore := GetSessionStore()
    sessStore.expirePeriod = expirePeriod
    sessStore.gcPeriod     = gcPeriod

    go ss.clearCacheData()
}

/**
定期清除contextIdMap中的过期数据
 */
func (ss *sessionStore) clearCacheData() {
    for {
        time.Sleep(ss.gcPeriod)

        for id, contextTime := range ss.contextIdMap {
            if contextTime != nil && contextTime.createTime.Add(ss.expirePeriod).Before(time.Now()) {
                // 已经过期
                ss.RemoveContextId(id)
            }
        }
    }
}

/**
设置校验码，存储至session中
 */
func (ss *sessionStore) Set(id string, digits []byte) {
    ss.sessionMutex.Lock()
    defer ss.sessionMutex.Unlock()

    contextWithTime, ok := ss.contextIdMap[id]

    if ok {
        sess := sessions.Default(contextWithTime.context)
        sess.Set(ss.keyByid(id), digits)
        sess.Save()
    }
}

/**
从session中读取校验码
 */
func (ss *sessionStore) Get(id string, clear bool) (digits []byte)  {
    contextWithTime, ok := ss.contextIdMap[id]
    if !ok {
        logs.Error(fmt.Sprintf("captcha get err, index val empty:%s", id))
        return
    }

    sess := sessions.Default(contextWithTime.context) // 获取用户session

    overdue := ss.expirePeriod > 0 && contextWithTime.createTime.Add(ss.expirePeriod).Before(time.Now())
    if overdue || clear {
        // 到期了或者是需要删除的
        ss.sessionMutex.Lock()
        defer ss.sessionMutex.Unlock()

        if overdue {
            // 是过期
            digits = nil
        } else if clear {
            // 未过期但是要读取后删除，先读取
            digits = sess.Get(ss.keyByid(id)).([]byte)
        }
        // 删除存储在session中的对应id信息
        sess.Delete(ss.keyByid(id))
        sess.Save()
        // 删除内存中保存的context与id对应关系
        ss.RemoveContextId(id)
        return
    }

    digitsVal := sess.Get(ss.keyByid(id))
    if digitsVal != nil {
        digits = digitsVal.([]byte)
    }
    return
}

/**
根据id生成对应key
 */
func (ss *sessionStore) keyByid(id string) string {
    return consts.SESSION_KEY_INDENTIFY + "_" + id
}

/**
在内存中添加context与id的映射数据
 */
func (ss *sessionStore) PushContextId(context *gin.Context, id string) error {
    ss.memoryMutex.Lock()
    defer ss.memoryMutex.Unlock()
    _, ok := ss.contextIdMap[id]
    if ok {
        return STORE_ERR_ID_EXISTS
    }
    ss.contextIdMap[id] = &contextWithTime{
        context: context,
        createTime: time.Now(),
    }
    return nil
}

/**
删除停留在内存中的context与id的映射数据
 */
func (ss *sessionStore) RemoveContextId(id string) error {
    ss.memoryMutex.Lock()
    defer ss.memoryMutex.Unlock()
    _, ok := ss.contextIdMap[id]

    if !ok {
        // 本来就没有这个数据，则返回删除成功
        return nil
    }

    delete(ss.contextIdMap, id)
    return nil
}

