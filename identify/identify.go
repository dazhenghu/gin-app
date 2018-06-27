package identify

import (
    "github.com/gin-gonic/gin"
    "path/filepath"
    "bytes"
    "github.com/dchest/captcha"
    "net/http"
    "time"
    "github.com/dazhenghu/ginApp/errorDefine"
    "path"
    "strings"
)

/**
验证码功能
*/

type captchaHandler struct {
    imgWidth  int
    imgHeight int
}

/**
生成实例
 */
func CaptchaNew(imgWidth, imgHeight int) *captchaHandler {
    return &captchaHandler{imgWidth, imgHeight}
}

/**
执行操作，生成随机码并返回
 */
func (ch *captchaHandler) Handle(context *gin.Context) error {
    dir, _ := path.Split(context.Request.URL.Path)
    name := context.Param("name") // 请求文件名
    ext := filepath.Ext(name)
    id := name[:len(name)-len(ext)]
    if ext == "" || id == "" {
        return errorDefine.ERROR_NOT_FOUND
    }

    if context.Param("reload") != "" {
        captcha.Reload(id)
    }

    lang := strings.ToLower(context.Param("lang"))
    download := path.Base(dir) == "download"

    return ch.serve(context, lang, download)
}

func (ch *captchaHandler) serve(context *gin.Context, lang string, download bool) error {
    name := context.Param("name") // 请求文件名
    ext := filepath.Ext(name)
    id := name[:len(name)-len(ext)]

    context.Header("Cache-Control", "no-cache, no-store, must-revalidate")
    context.Header("Pragma", "no-cache")
    context.Header("Expires", "0")

    var content bytes.Buffer
    switch ext {
    case ".png":
        context.Header("Content-Type", "image/png")
        captcha.WriteImage(context.Writer, id, ch.imgWidth, ch.imgHeight)
    case ".wav":
        context.Header("Content-Type", "audio/x-wav")
        captcha.WriteAudio(&content, id, lang)
    default:
        return errorDefine.ERROR_NOT_FOUND
    }

    if download {
        context.Header("Content-Type", "application/octet-stream")
    }

    http.ServeContent(context.Writer, context.Request, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
    return nil
}

//func (h *captchaHandler) serve(w http.ResponseWriter, r *http.Request, id, ext, lang string, download bool) error {
//    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
//    w.Header().Set("Pragma", "no-cache")
//    w.Header().Set("Expires", "0")
//
//    var content bytes.Buffer
//    switch ext {
//    case ".png":
//        w.Header().Set("Content-Type", "image/png")
//        WriteImage(&content, id, h.imgWidth, h.imgHeight)
//    case ".wav":
//        w.Header().Set("Content-Type", "audio/x-wav")
//        WriteAudio(&content, id, lang)
//    default:
//        return ErrNotFound
//    }
//
//    if download {
//        w.Header().Set("Content-Type", "application/octet-stream")
//    }
//    http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
//    return nil
//}
//
//func (h *captchaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//    dir, file := path.Split(r.URL.Path)
//    ext := path.Ext(file)
//    id := file[:len(file)-len(ext)]
//    if ext == "" || id == "" {
//        http.NotFound(w, r)
//        return
//    }
//    if r.FormValue("reload") != "" {
//        Reload(id)
//    }
//    lang := strings.ToLower(r.FormValue("lang"))
//    download := path.Base(dir) == "download"
//    if h.serve(w, r, id, ext, lang, download) == ErrNotFound {
//        http.NotFound(w, r)
//    }
//    // Ignore other errors.
//}
