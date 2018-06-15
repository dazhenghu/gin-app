package session

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    token2 "github.com/dazhenghu/ginApp/safe/token"
    "errors"
    "github.com/dazhenghu/ginApp/types"
    "fmt"
)


func GenerateSessionToken(c *gin.Context, key string) (token string, err error) {
    tokenObj := token2.NewToken("")
    token = tokenObj.GenerateToken()
    session := sessions.Default(c)
    tokens := session.Get(key)
    var tokenList types.SliceString
    fmt.Printf("tokens:%+v\n", tokens)
    if tokens != nil {
        tokenList = types.NewSliceStringFromSlice(tokens.([]string))
    } else {
        tokenList = types.NewSliceString()
    }

    (&tokenList).Append(token)
    session.Set(key, tokenList.ToSlice())
    err = session.Save()
    return
}

func CheckSessionToken(c *gin.Context, key string, token string) (err error) {
    session := sessions.Default(c)
    sessionTokens := session.Get(key).([]string)
    for _, val := range sessionTokens {
        if val == token {

            return
        }
    }

    err = errors.New("invalid token")
    return
}
