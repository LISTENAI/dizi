-- 简化的 gopher-lua-libs 示例脚本
-- 展示确认可用的核心功能

print("=== gopher-lua-libs 核心功能示例 ===")
print()

-- JSON 处理
print("1. JSON 处理:")
local json = require("json")
local data = {
    name = "Dizi MCP Server",
    version = "1.0.0",
    features = {"lua scripting", "extended libraries"},
    config = {port = 8082, enabled = true}
}

local json_str = json.encode(data)
print("JSON 编码:", json_str)

local decoded = json.decode(json_str)
print("解码结果 - 名称:", decoded.name)
print("解码结果 - 端口:", decoded.config.port)
print()

-- YAML 处理
print("2. YAML 处理:")
local yaml = require("yaml")
local yaml_data = {
    server = {host = "localhost", port = 8082},
    enabled = true
}

local yaml_str = yaml.encode(yaml_data)
print("YAML 输出:")
print(yaml_str)
print()

-- 时间处理
print("3. 时间处理:")
local time = require("time")
local now = time.unix()
print("当前时间戳:", math.floor(now))
print("格式化时间:", time.format(now, "2006-01-02 15:04:05"))
print()

-- 基本的字符串处理
print("4. 字符串处理:")
local strings = require("strings")
local text = "  Hello World  "
print("去除空格:", strings.trim_space(text))
print()

print("=== 核心功能测试完成 ===")
print("JSON、YAML、Time 和 Strings 模块已成功集成！")

print("5. 系统信息:")
local goos = require("goos")
local hostname, err = goos.hostname()
if err then error(err) end
print(hostname)

local crypto = require("crypto")

print("6. 加密:")
print("md5:")
print(crypto.md5("1\n"))

-- md5
if not(crypto.md5("1\n") == "b026324c6904b2a9cb4b88d6d61c81d1") then
    error("md5")
end

-- sha256
if not(crypto.sha256("1\n") == "4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865") then
    error("sha256")
end

local runtime = require("runtime")
print("7. 运行时:")
print("goos:")
print(runtime.goos())
print("goarch:")
print(runtime.goarch())


-- 设置结果变量
result = "Core gopher-lua-libs modules are working: JSON, YAML, Time, Strings"