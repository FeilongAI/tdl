# 备份功能设计文档

## 需求
将本地文件夹树结构备份到 Telegram Forum（话题群）中，每个文件夹对应一个话题。

## 用户场景
```
本地路径: D:\Program\GoWork\src\backup
目标: 上传到 Telegram forum 群组
忽略前缀: D:\Program\GoWork\src

文件夹结构:
backup/
├── 工作文档/
│   ├── 项目A/
│   │   ├── file1.pdf
│   │   └── file2.docx
│   └── 项目B/
│       └── file3.xlsx
└── 个人资料/
    └── photo.jpg

期望的话题结构:
Forum Group
├── Topic: backup
├── Topic: backup/工作文档
│   ├── Topic: backup/工作文档/项目A  (包含 file1.pdf, file2.docx)
│   └── Topic: backup/工作文档/项目B  (包含 file3.xlsx)
└── Topic: backup/个人资料  (包含 photo.jpg)
```

## 技术方案

### 1. 命令接口
```bash
tdl backup --chat <forum_group_id> --path <local_path> --strip-prefix <prefix>
```

参数说明：
- `--chat`: Forum 群组ID（必须是启用了 Forum 的超级群组）
- `--path`: 本地备份文件夹路径
- `--strip-prefix`: 要忽略的路径前缀（可选，默认为 path 的父目录）

### 2. 实现步骤

#### Step 1: 遍历文件夹树
- 使用 `filepath.WalkDir` 遍历所有文件和文件夹
- 为每个文件夹记录相对路径

#### Step 2: 创建话题映射
- 为每个文件夹创建对应的 Forum Topic
- 使用相对路径作为话题标题
- 将话题ID存储到数据库中（防止重复创建）

#### Step 3: 上传文件
- 遍历文件，根据所在文件夹找到对应的话题ID
- 使用 `--chat` 和 `--topic` 参数上传文件

### 3. 数据结构

```go
type BackupOptions struct {
    Chat        string   // Forum 群组ID
    Path        string   // 本地备份路径
    StripPrefix string   // 要忽略的路径前缀
    Includes    []string // 包含的文件扩展名
    Excludes    []string // 排除的文件扩展名
}

type TopicMapping struct {
    RelativePath string // 相对路径
    TopicID      int    // Telegram 话题ID
    TopicName    string // 话题名称
}
```

### 4. 存储方案
使用 KV 存储保存路径到话题ID的映射：
- Key: `backup:<chat_id>:<relative_path>`
- Value: `topic_id`

### 5. API 调用
```go
// 创建话题
api := tg.NewClient(...)
req := &tg.ChannelsCreateForumTopicRequest{
    Channel:  inputChannel,
    Title:    folderName,
    RandomID: rand.Int63(),
}
updates, err := api.ChannelsCreateForumTopic(ctx, req)

// 获取创建的话题ID
topicID := extractTopicIDFromUpdates(updates)
```

## 实现计划
1. 创建 `cmd/backup.go` 定义命令
2. 创建 `app/backup/backup.go` 实现核心逻辑
3. 创建 `app/backup/topic.go` 实现话题管理
4. 创建 `app/backup/storage.go` 实现映射存储
5. 测试和完善