# Lua 配置文件目录

此目录用于存储游戏活动的 Lua 配置文件。

## 文件命名格式

配置文件应按照以下格式命名：
- `{iid}{年月日}.lua`
- 例如：`bydr20251104.lua`、`dsc20251104.lua`、`cqsj20251104.lua`、`yxds20251104.lua`

其中：
- `iid`：游戏ID（bydr、dsc、cqsj、yxds）
- `年月日`：8位数字，格式为 YYYYMMDD（例如：20251104 表示 2025年11月4日）

## 配置文件来源

配置文件可以从以下位置获取：
- 原始路径：`D:\workSpace\Data\Excel\Activity\Lua`
- 项目路径：`config/lua`（当前目录）

系统会优先从项目内的 `config/lua` 目录读取配置文件，如果不存在，则回退到原始路径。

## 配置文件内容

每个配置文件应包含以下内容：
- 邮件标题配置（`mailtitle_daily_zd`、`mailtitle_total_zd`、`mailtitle_daily_gh`、`mailtitle_total_gh`）
- 邮件内容配置（`mailcontent_daily_zd`、`mailcontent_total_zd`、`mailcontent_daily_gh`、`mailcontent_total_gh`）
- 奖励配置（`daily_reward_zd`、`total_reward_zd`、`daily_reward_gh`、`total_reward_gh`）

## 使用说明

1. 将 Lua 配置文件复制到此目录
2. 在网页界面中选择对应的游戏类型（bydr/dsc/cqsj/yxds）
3. 点击"加载默认配置"按钮，系统会自动选择最新日期的配置文件并加载
