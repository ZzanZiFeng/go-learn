# 文件上传 (File Upload)

## 概述

Gin 提供了简单的文件上传处理方式，支持单文件和多文件上传。

## 与 Express.js 对比

### Express + Multer

```javascript
const multer = require('multer');

const storage = multer.diskStorage({
  destination: './uploads/',
  filename: (req, file, cb) => {
    cb(null, Date.now() + '-' + file.originalname);
  }
});

const upload = multer({
  storage,
  limits: { fileSize: 5 * 1024 * 1024 },
  fileFilter: (req, file, cb) => {
    if (file.mimetype.startsWith('image/')) {
      cb(null, true);
    } else {
      cb(new Error('Only images allowed'));
    }
  }
});

app.post('/upload', upload.single('file'), (req, res) => {
  res.json({ filename: req.file.filename });
});
```

### Gin

```go
r.POST("/upload", func(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename)
    if err := c.SaveUploadedFile(file, "./uploads/"+filename); err != nil {
        c.JSON(500, gin.H{"error": "Failed to save file"})
        return
    }

    c.JSON(200, gin.H{"filename": filename})
})
```

## 单文件上传

### 基本上传

```go
r.POST("/upload", func(c *gin.Context) {
    // 获取上传的文件
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file uploaded"})
        return
    }

    // 文件信息
    fmt.Println("Filename:", file.Filename)
    fmt.Println("Size:", file.Size)
    fmt.Println("Header:", file.Header)

    // 保存文件
    dst := "./uploads/" + file.Filename
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(500, gin.H{"error": "Failed to save"})
        return
    }

    c.JSON(200, gin.H{
        "filename": file.Filename,
        "size":     file.Size,
    })
})
```

### 带表单数据

```go
type UploadForm struct {
    Title       string                `form:"title" binding:"required"`
    Description string                `form:"description"`
    File        *multipart.FileHeader `form:"file" binding:"required"`
}

r.POST("/upload", func(c *gin.Context) {
    var form UploadForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    dst := "./uploads/" + form.File.Filename
    if err := c.SaveUploadedFile(form.File, dst); err != nil {
        c.JSON(500, gin.H{"error": "Failed to save"})
        return
    }

    c.JSON(200, gin.H{
        "title":    form.Title,
        "filename": form.File.Filename,
    })
})
```

## 多文件上传

### 同字段多文件

```go
r.POST("/uploads", func(c *gin.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    files := form.File["files"]  // 获取所有文件
    uploaded := []string{}

    for _, file := range files {
        filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename)
        dst := "./uploads/" + filename

        if err := c.SaveUploadedFile(file, dst); err != nil {
            continue
        }
        uploaded = append(uploaded, filename)
    }

    c.JSON(200, gin.H{
        "count": len(uploaded),
        "files": uploaded,
    })
})
```

### 不同字段多文件

```go
type MultiFileForm struct {
    Avatar    *multipart.FileHeader   `form:"avatar"`
    Documents []*multipart.FileHeader `form:"documents"`
}

r.POST("/profile", func(c *gin.Context) {
    var form MultiFileForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 处理头像
    if form.Avatar != nil {
        c.SaveUploadedFile(form.Avatar, "./uploads/avatars/"+form.Avatar.Filename)
    }

    // 处理文档
    for _, doc := range form.Documents {
        c.SaveUploadedFile(doc, "./uploads/documents/"+doc.Filename)
    }

    c.JSON(200, gin.H{"message": "uploaded"})
})
```

## 文件验证

### 验证文件大小

```go
const maxFileSize = 5 * 1024 * 1024  // 5MB

func validateFileSize(file *multipart.FileHeader, maxSize int64) error {
    if file.Size > maxSize {
        return fmt.Errorf("file size exceeds %d bytes", maxSize)
    }
    return nil
}

r.POST("/upload", func(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := validateFileSize(file, maxFileSize); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 保存文件...
})

// 全局限制
r.MaxMultipartMemory = 8 << 20  // 8MB
```

### 验证文件类型

```go
var allowedImageTypes = map[string]bool{
    "image/jpeg": true,
    "image/png":  true,
    "image/gif":  true,
    "image/webp": true,
}

var allowedDocTypes = map[string]bool{
    "application/pdf":    true,
    "application/msword": true,
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

func validateFileType(file *multipart.FileHeader, allowed map[string]bool) error {
    // 打开文件
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()

    // 读取前 512 字节检测类型
    buffer := make([]byte, 512)
    _, err = src.Read(buffer)
    if err != nil {
        return err
    }

    // 检测 MIME 类型
    contentType := http.DetectContentType(buffer)

    if !allowed[contentType] {
        return fmt.Errorf("file type %s not allowed", contentType)
    }

    return nil
}

r.POST("/upload/image", func(c *gin.Context) {
    file, err := c.FormFile("image")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := validateFileType(file, allowedImageTypes); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 保存文件...
})
```

### 验证文件扩展名

```go
var allowedExtensions = map[string]bool{
    ".jpg":  true,
    ".jpeg": true,
    ".png":  true,
    ".gif":  true,
    ".pdf":  true,
}

func validateExtension(filename string, allowed map[string]bool) error {
    ext := strings.ToLower(filepath.Ext(filename))
    if !allowed[ext] {
        return fmt.Errorf("extension %s not allowed", ext)
    }
    return nil
}
```

## 安全文件名

```go
import (
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/google/uuid"
)

// 清理文件名
func sanitizeFilename(filename string) string {
    // 获取扩展名
    ext := filepath.Ext(filename)

    // 获取基础名称（不含扩展名）
    base := strings.TrimSuffix(filename, ext)

    // 移除危险字符
    reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
    base = reg.ReplaceAllString(base, "_")

    // 限制长度
    if len(base) > 50 {
        base = base[:50]
    }

    return base + ext
}

// 生成唯一文件名
func generateFilename(originalName string) string {
    ext := filepath.Ext(originalName)
    return uuid.New().String() + ext
}

// 生成带时间戳的文件名
func generateTimestampFilename(originalName string) string {
    ext := filepath.Ext(originalName)
    timestamp := time.Now().Format("20060102-150405")
    randomStr := uuid.New().String()[:8]
    return fmt.Sprintf("%s-%s%s", timestamp, randomStr, ext)
}

r.POST("/upload", func(c *gin.Context) {
    file, _ := c.FormFile("file")

    // 使用安全文件名
    filename := generateFilename(file.Filename)
    dst := filepath.Join("./uploads", filename)

    c.SaveUploadedFile(file, dst)
    c.JSON(200, gin.H{"filename": filename})
})
```

## 文件存储

### 按日期分目录

```go
func getUploadPath() string {
    now := time.Now()
    path := fmt.Sprintf("./uploads/%d/%02d/%02d",
        now.Year(), now.Month(), now.Day())

    // 确保目录存在
    os.MkdirAll(path, 0755)

    return path
}

r.POST("/upload", func(c *gin.Context) {
    file, _ := c.FormFile("file")

    uploadPath := getUploadPath()
    filename := generateFilename(file.Filename)
    dst := filepath.Join(uploadPath, filename)

    c.SaveUploadedFile(file, dst)

    // 返回相对路径
    relativePath := strings.TrimPrefix(dst, "./uploads/")
    c.JSON(200, gin.H{"path": relativePath})
})
```

### 按类型分目录

```go
func getUploadPathByType(contentType string) string {
    var subdir string
    switch {
    case strings.HasPrefix(contentType, "image/"):
        subdir = "images"
    case strings.HasPrefix(contentType, "video/"):
        subdir = "videos"
    case contentType == "application/pdf":
        subdir = "documents"
    default:
        subdir = "others"
    }

    path := filepath.Join("./uploads", subdir)
    os.MkdirAll(path, 0755)
    return path
}
```

## 图片处理

### 生成缩略图

```go
import (
    "image"
    "image/jpeg"
    _ "image/png"

    "github.com/nfnt/resize"
)

func createThumbnail(src string, dst string, width, height uint) error {
    file, err := os.Open(src)
    if err != nil {
        return err
    }
    defer file.Close()

    img, _, err := image.Decode(file)
    if err != nil {
        return err
    }

    thumbnail := resize.Thumbnail(width, height, img, resize.Lanczos3)

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    return jpeg.Encode(out, thumbnail, &jpeg.Options{Quality: 80})
}

r.POST("/upload/image", func(c *gin.Context) {
    file, _ := c.FormFile("image")

    filename := generateFilename(file.Filename)
    originalPath := "./uploads/original/" + filename
    thumbnailPath := "./uploads/thumbnails/" + filename

    // 保存原图
    c.SaveUploadedFile(file, originalPath)

    // 生成缩略图
    createThumbnail(originalPath, thumbnailPath, 200, 200)

    c.JSON(200, gin.H{
        "original":  filename,
        "thumbnail": filename,
    })
})
```

## 流式上传

### 大文件处理

```go
r.POST("/upload/large", func(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    defer file.Close()

    filename := generateFilename(header.Filename)
    dst, err := os.Create("./uploads/" + filename)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create file"})
        return
    }
    defer dst.Close()

    // 流式复制，适合大文件
    written, err := io.Copy(dst, file)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to save file"})
        return
    }

    c.JSON(200, gin.H{
        "filename": filename,
        "size":     written,
    })
})
```

### 分片上传

```go
type ChunkInfo struct {
    FileID     string `form:"file_id" binding:"required"`
    ChunkIndex int    `form:"chunk_index" binding:"required"`
    TotalChunks int   `form:"total_chunks" binding:"required"`
}

r.POST("/upload/chunk", func(c *gin.Context) {
    var info ChunkInfo
    if err := c.ShouldBind(&info); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    file, err := c.FormFile("chunk")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 保存分片
    chunkPath := fmt.Sprintf("./temp/%s/chunk_%d", info.FileID, info.ChunkIndex)
    os.MkdirAll(filepath.Dir(chunkPath), 0755)
    c.SaveUploadedFile(file, chunkPath)

    c.JSON(200, gin.H{
        "file_id":     info.FileID,
        "chunk_index": info.ChunkIndex,
        "received":    true,
    })
})

r.POST("/upload/complete", func(c *gin.Context) {
    var req struct {
        FileID      string `json:"file_id" binding:"required"`
        Filename    string `json:"filename" binding:"required"`
        TotalChunks int    `json:"total_chunks" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 合并分片
    finalPath := "./uploads/" + generateFilename(req.Filename)
    finalFile, _ := os.Create(finalPath)
    defer finalFile.Close()

    for i := 0; i < req.TotalChunks; i++ {
        chunkPath := fmt.Sprintf("./temp/%s/chunk_%d", req.FileID, i)
        chunk, _ := os.ReadFile(chunkPath)
        finalFile.Write(chunk)
    }

    // 清理临时文件
    os.RemoveAll("./temp/" + req.FileID)

    c.JSON(200, gin.H{"filename": filepath.Base(finalPath)})
})
```

## 文件服务

### 静态文件服务

```go
// 整个目录
r.Static("/files", "./uploads")

// 单个文件
r.StaticFile("/favicon.ico", "./resources/favicon.ico")

// 使用 http.FileSystem
r.StaticFS("/assets", http.Dir("assets"))
```

### 文件下载

```go
r.GET("/download/:filename", func(c *gin.Context) {
    filename := c.Param("filename")

    // 验证文件名，防止路径遍历
    if strings.Contains(filename, "..") {
        c.JSON(400, gin.H{"error": "Invalid filename"})
        return
    }

    filepath := "./uploads/" + filename

    // 检查文件是否存在
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(404, gin.H{"error": "File not found"})
        return
    }

    c.FileAttachment(filepath, filename)
})
```

### 带权限的文件访问

```go
r.GET("/files/:id", AuthMiddleware(), func(c *gin.Context) {
    fileID := c.Param("id")
    userID := c.GetInt("userID")

    // 从数据库获取文件信息
    file, err := fileService.GetByID(fileID)
    if err != nil {
        c.JSON(404, gin.H{"error": "File not found"})
        return
    }

    // 检查权限
    if file.OwnerID != userID && !file.IsPublic {
        c.JSON(403, gin.H{"error": "Access denied"})
        return
    }

    c.File(file.Path)
})
```

## 完整示例

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const (
    maxFileSize    = 10 << 20  // 10MB
    uploadDir      = "./uploads"
)

var allowedTypes = map[string]bool{
    "image/jpeg":      true,
    "image/png":       true,
    "image/gif":       true,
    "application/pdf": true,
}

type FileInfo struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    Size     int64  `json:"size"`
    Type     string `json:"type"`
    URL      string `json:"url"`
}

func init() {
    os.MkdirAll(uploadDir, 0755)
}

func generateFilename(original string) string {
    ext := filepath.Ext(original)
    return uuid.New().String() + ext
}

func validateFile(file *multipart.FileHeader) error {
    // 检查大小
    if file.Size > maxFileSize {
        return fmt.Errorf("file too large (max %d MB)", maxFileSize>>20)
    }

    // 检查类型
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()

    buffer := make([]byte, 512)
    src.Read(buffer)
    contentType := http.DetectContentType(buffer)

    if !allowedTypes[contentType] {
        return fmt.Errorf("file type %s not allowed", contentType)
    }

    return nil
}

func main() {
    r := gin.Default()
    r.MaxMultipartMemory = maxFileSize

    // 静态文件服务
    r.Static("/files", uploadDir)

    // 单文件上传
    r.POST("/upload", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "No file uploaded"})
            return
        }

        if err := validateFile(file); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        filename := generateFilename(file.Filename)
        dst := filepath.Join(uploadDir, filename)

        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(500, gin.H{"error": "Failed to save file"})
            return
        }

        c.JSON(200, FileInfo{
            ID:       strings.TrimSuffix(filename, filepath.Ext(filename)),
            Filename: filename,
            Size:     file.Size,
            URL:      "/files/" + filename,
        })
    })

    // 多文件上传
    r.POST("/uploads", func(c *gin.Context) {
        form, err := c.MultipartForm()
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        files := form.File["files"]
        results := []FileInfo{}
        errors := []string{}

        for _, file := range files {
            if err := validateFile(file); err != nil {
                errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
                continue
            }

            filename := generateFilename(file.Filename)
            dst := filepath.Join(uploadDir, filename)

            if err := c.SaveUploadedFile(file, dst); err != nil {
                errors = append(errors, fmt.Sprintf("%s: failed to save", file.Filename))
                continue
            }

            results = append(results, FileInfo{
                Filename: filename,
                Size:     file.Size,
                URL:      "/files/" + filename,
            })
        }

        c.JSON(200, gin.H{
            "uploaded": results,
            "errors":   errors,
        })
    })

    // 文件下载
    r.GET("/download/:filename", func(c *gin.Context) {
        filename := c.Param("filename")

        if strings.Contains(filename, "..") {
            c.JSON(400, gin.H{"error": "Invalid filename"})
            return
        }

        path := filepath.Join(uploadDir, filename)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            c.JSON(404, gin.H{"error": "File not found"})
            return
        }

        c.FileAttachment(path, filename)
    })

    // 删除文件
    r.DELETE("/files/:filename", func(c *gin.Context) {
        filename := c.Param("filename")

        if strings.Contains(filename, "..") {
            c.JSON(400, gin.H{"error": "Invalid filename"})
            return
        }

        path := filepath.Join(uploadDir, filename)
        if err := os.Remove(path); err != nil {
            if os.IsNotExist(err) {
                c.JSON(404, gin.H{"error": "File not found"})
                return
            }
            c.JSON(500, gin.H{"error": "Failed to delete"})
            return
        }

        c.JSON(200, gin.H{"message": "File deleted"})
    })

    r.Run(":8080")
}
```

## 总结

| 功能 | 方法 |
|------|------|
| 单文件 | `c.FormFile("file")` |
| 多文件 | `c.MultipartForm()` |
| 保存 | `c.SaveUploadedFile(file, dst)` |
| 下载 | `c.File()` / `c.FileAttachment()` |
| 静态服务 | `r.Static()` |
| 大小限制 | `r.MaxMultipartMemory` |

**下一节**：[Swagger](./10-swagger.md) - 学习 API 文档
