-- System Information Lua script
-- Collects and formats system information

function get_system_info()
    local info = {}
    info.lua_version = _VERSION
    info.timestamp = os.date("%Y-%m-%d %H:%M:%S")
    info.platform = "Dizi MCP Server"
    return info
end

function format_table(t, indent)
    indent = indent or 0
    local result = "{\n"
    for k, v in pairs(t) do
        result = result .. string.rep("  ", indent + 1)
        if type(k) == "string" then
            result = result .. k .. ": "
        end
        
        if type(v) == "table" then
            result = result .. format_table(v, indent + 1)
        elseif type(v) == "string" then
            result = result .. '"' .. v .. '"'
        else
            result = result .. tostring(v)
        end
        result = result .. "\n"
    end
    result = result .. string.rep("  ", indent) .. "}"
    return result
end

-- Set result
local sys_info = get_system_info()
result = "System Information:\n" .. format_table(sys_info)