-- Hello World Lua script for Dizi
-- This script demonstrates basic Lua functionality

function greet(name)
    if name == nil or name == "" then
        name = "World"
    end
    return "Hello, " .. name .. "!"
end

function calculate(operation, a, b)
    local num_a = tonumber(a) or 0
    local num_b = tonumber(b) or 0
    
    if operation == "add" then
        return num_a + num_b
    elseif operation == "multiply" then
        return num_a * num_b
    elseif operation == "subtract" then
        return num_a - num_b
    elseif operation == "divide" then
        if num_b == 0 then
            return "Error: Division by zero"
        end
        return num_a / num_b
    else
        return "Unknown operation: " .. tostring(operation)
    end
end

-- Main execution
if name then
    result = greet(name)
elseif operation and a and b then
    result = operation .. "(" .. a .. ", " .. b .. ") = " .. tostring(calculate(operation, a, b))
else
    result = greet() .. " Use parameters: name, or operation+a+b"
end