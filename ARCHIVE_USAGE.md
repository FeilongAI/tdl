# Archive 功能使用指南

## 功能简介

`archive` 命令可以将本地文件夹树形结构备份到 Telegram 论坛群组（Forum Group）中，每个文件夹会对应创建一个话题（Topic），文件会上传到对应的话题中。

## 前提条件

1. 创建或使用一个已存在的 Telegram 超级群组
2. 在群组设置中启用"话题"功能（Topics/Forum）
3. 确保你的账号在该群组中有创建话题和发送文件的权限

## 基本使用

### 1. 登录 tdl

首先使用你的自定义 AppID 登录：

```bash
# 使用验证码登录
tdl login -T code

# 或使用二维码登录
tdl login -T qr
```

### 2. 获取论坛群组 ID

使用 chat list 命令查看你的群组列表：

```bash
tdl chat ls
```

找到你的论坛群组，记下它的 ID 或用户名。

### 3. 归档文件

基本命令格式：

```bash
tdl archive --chat <forum_group_id> --path <local_folder_path>
```

#### 示例 1: 备份工作文件夹

```bash
# 假设你的备份文件夹在 D:\Program\GoWork\src\backup
# 论坛群组 ID 为 -1001234567890

tdl archive --chat -1001234567890 --path "D:\Program\GoWork\src\backup"
```

这会：
- 自动以 `D:\Program\GoWork\src` 为根目录
- 为 `backup` 文件夹创建一个话题
- 为 `backup` 下的每个子文件夹创建对应的话题
- 将文件上传到对应的话题中

#### 示例 2: 自定义路径前缀

```bash
# 如果你想以不同的前缀为根目录
tdl archive --chat -1001234567890 --path "D:\Program\GoWork\src\backup" --strip-prefix "D:\Program\GoWork"
```

这会让话题名称从 `src/backup` 开始，而不是 `backup`。

#### 示例 3: 过滤文件类型

```bash
# 只上传图片文件
tdl archive --chat -1001234567890 --path "D:\Photos" -i .jpg -i .png -i .gif

# 排除视频文件
tdl archive --chat -1001234567890 --path "D:\Documents" -e .mp4 -e .avi
```

#### 示例 4: 其他选项

```bash
# 图片以照片格式上传（而不是文件）
tdl archive --chat -1001234567890 --path "D:\Photos" --photo

# 上传后删除本地文件
tdl archive --chat -1001234567890 --path "D:\Temp" --rm
```

## 完整参数说明

| 参数 | 简写 | 必填 | 说明 |
|------|------|------|------|
| `--chat` | `-c` | 是 | 论坛群组 ID 或用户名 |
| `--path` | `-p` | 是 | 本地文件夹路径 |
| `--strip-prefix` | | 否 | 要忽略的路径前缀（默认为 path 的父目录） |
| `--include` | `-i` | 否 | 包含的文件扩展名（可多次使用） |
| `--exclude` | `-e` | 否 | 排除的文件扩展名（可多次使用） |
| `--photo` | | 否 | 将图片以照片格式上传 |
| `--rm` | | 否 | 上传后删除本地文件 |

## 话题映射存储

- 话题映射关系会自动保存在 tdl 的数据库中
- 下次上传相同路径时，会复用已存在的话题
- 映射关系保存在键 `archive:topics:<chat_id>` 中

## 注意事项

1. **权限要求**：确保你的账号在论坛群组中有"管理话题"权限
2. **文件夹层级**：支持多层文件夹嵌套，每一层都会创建对应的话题
3. **文件大小限制**：受 Telegram 限制，单个文件最大 2GB（普通用户）或 4GB（Premium 用户）
4. **断点续传**：如果中途中断，下次运行会跳过已创建的话题，继续上传文件
5. **话题名称**：使用文件夹的相对路径作为话题标题

## 实际案例

### 场景：备份百度网盘下载的文件

```bash
# 1. 从百度网盘下载文件到本地
# 假设下载到 D:\BaiduNetdisk\我的文档

# 2. 创建或使用论坛群组（假设 ID 为 -1001234567890）

# 3. 归档到 Telegram
tdl archive --chat -1001234567890 --path "D:\BaiduNetdisk\我的文档" --strip-prefix "D:\BaiduNetdisk"

# 结果：
# - 创建话题"我的文档"
# - 如果有子文件夹"工作/项目A"，会创建话题"我的文档/工作/项目A"
# - 文件会上传到对应的话题中
```

### 场景：定期备份工作文件

```bash
# 每天备份工作文件到 Telegram
tdl archive --chat @my_backup_forum --path "D:\Work\Projects" -e .tmp -e .cache
```

## 故障排除

### 问题 1：权限错误
```
Error: create forum topic: CHAT_ADMIN_REQUIRED
```
**解决方案**：确保你在群组中有管理员权限并启用了"管理话题"权限。

### 问题 2：找不到群组
```
Error: get forum chat: USERNAME_NOT_FOUND
```
**解决方案**：使用群组的数字 ID 而不是用户名，可以通过 `tdl chat ls` 查看。

### 问题 3：不是论坛群组
```
Error: forum peer is not a channel
```
**解决方案**：确保在群组设置中启用了"话题"功能。

## 高级技巧

### 1. 结合脚本批量备份

创建批处理文件 `backup.bat`：

```batch
@echo off
tdl archive --chat -1001234567890 --path "D:\Documents" --strip-prefix "D:\"
tdl archive --chat -1001234567890 --path "D:\Photos" --strip-prefix "D:\" --photo
tdl archive --chat -1001234567890 --path "D:\Videos" --strip-prefix "D:\"
```

### 2. 配合任务计划程序定时备份

在 Windows 任务计划程序中设置定时运行 `backup.bat`。

## 常见问题

**Q: 如何更新已上传的文件？**
A: 重新运行 archive 命令，新文件会上传到已存在的话题中。

**Q: 如何删除已创建的话题？**
A: 目前需要手动在 Telegram 客户端中删除话题。

**Q: 话题数量有限制吗？**
A: 取决于 Telegram 的限制，建议合理组织文件夹结构。

**Q: 支持增量备份吗？**
A: 当前版本会上传所有文件，未来版本可能支持增量备份。