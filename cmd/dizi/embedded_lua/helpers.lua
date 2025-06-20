-- Lua REPL helper functions

function help()
    print("Dizi Lua REPL Commands:")
    print("  :help or help()   - Show this help")
    print("  :quit or :exit    - Exit REPL")
    print("  :clear            - Clear screen")
    print("  :reset            - Reset Lua state")
    print("  :version          - Show Lua version")
    print("  :vars             - Show all variables")
    print("  :history          - Show command history")
    print("")
    print("Navigation:")
    print("  ↑↓ arrows         - Browse history")
    print("  ←→ arrows         - Move cursor")
    print("  Tab               - Auto-completion")
    print("  Ctrl+C            - Interrupt")
    print("")
    print("Examples:")
    print("  2 + 3             - Simple expression")
    print("  x = 10; x * 2     - Variables")
    print("  print('hello')    - Output")
    print("  function f(x) return x*x end; f(5) - Functions")
end

function vars()
    print("User-defined variables:")
    local count = 0
    local builtins = {
        "_G", "_VERSION", "assert", "collectgarbage", "dofile", "error", 
        "getfenv", "getmetatable", "ipairs", "load", "loadfile", "loadstring",
        "module", "next", "pairs", "pcall", "print", "rawequal", "rawget", 
        "rawset", "require", "select", "setfenv", "setmetatable", "tonumber",
        "tostring", "type", "unpack", "xpcall", "newproxy", "coroutine",
        "debug", "io", "math", "os", "package", "string", "table", "channel",
        "help", "vars"
    }
    
    local function isBuiltin(name)
        for _, builtin in ipairs(builtins) do
            if name == builtin then return true end
        end
        return false
    end
    
    for k, v in pairs(_G) do
        if type(k) == "string" and not k:match("^_") and not isBuiltin(k) then
            print("  " .. k .. " = " .. tostring(v) .. " (" .. type(v) .. ")")
            count = count + 1
        end
    end
    if count == 0 then
        print("  No user-defined variables")
    end
end