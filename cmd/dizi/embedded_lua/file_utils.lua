-- File Utilities Lua script
-- Provides file operation utilities

function read_file_safe(filename)
    local file = io.open(filename, "r")
    if not file then
        return nil, "Cannot open file: " .. filename
    end
    local content = file:read("*all")
    file:close()
    return content, nil
end

function write_file_safe(filename, content)
    local file = io.open(filename, "w")
    if not file then
        return false, "Cannot create file: " .. filename
    end
    file:write(content)
    file:close()
    return true, nil
end

function get_file_info(filename)
    local file = io.open(filename, "r")
    if not file then
        return "File not found: " .. filename
    end
    file:close()
    
    -- Basic file info (limited in standard Lua)
    return "File exists: " .. filename
end

-- Example usage based on parameters
if action and filename then
    if action == "read" then
        local content, err = read_file_safe(filename)
        if err then
            result = "Error: " .. err
        else
            result = "File content:\n" .. content
        end
    elseif action == "info" then
        result = get_file_info(filename)
    elseif action == "write" and content then
        local success, err = write_file_safe(filename, content)
        if err then
            result = "Error: " .. err
        else
            result = "Successfully wrote to: " .. filename
        end
    else
        result = "Usage: action=read/write/info, filename=path [, content=text]"
    end
else
    result = "File utilities loaded. Use with action and filename parameters."
end