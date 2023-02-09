local hashmemcached = require "resty.hashmemcached"

local memc, err = hashmemcached.new({
    {'127.0.0.1', 11210, 1},
    {'127.0.0.1', 11211, 1}
}, 'hashmemcached')

memc:set_timeout(1000) -- 1 sec


local tcpsock, err = ngx.req.socket(true)
if err then
	ngx.log(ngx.ERR, "ngx.req.socket:", err)
	ngx.exit(0)
end

local function cleanup()
    ngx.log(ngx.WARN, "do cleanup")
    ngx.exit(0)
end

local ok, err = ngx.on_abort(cleanup)
if not ok then
    ngx.log(ngx.ERR, "failed to register the on_abort callback: ", err)
    ngx.exit(0)
end


local function split(str, pat)
   local t = {}  -- NOTE: use {n = 0} in Lua-5.0
   local fpat = "(.-)" .. pat
   local last_end = 1
   local s, e, cap = str:find(fpat, 1)
   while s do
      if s ~= 1 or cap ~= "" then
	 table.insert(t,cap)
      end
      last_end = e+1
      s, e, cap = str:find(fpat, last_end)
   end
   if last_end <= #str then
      cap = str:sub(last_end)
      table.insert(t, cap)
   end
   return t
end


while not ngx.worker.exiting() do
	local line, err = tcpsock:receive()
	if err and "timeout" ~= err then
		ngx.log(ngx.WARN, "receive failed:", err)
		break
	end
	if line == nil then
		break
	end

	local command = split(line, " ")
	if #command < 1 then
		break
	end
	local req_data = nil

	local name = string.lower(command[1])
	if "set" == name or "add" == name or "replace" == name or "hset" == name then
		req_data, err = tcpsock:receive(tonumber(command[#command])+2)
		if err then
			ngx.log(ngx.WARN, "receive value failed:", err)
			break
		end

		if "\r\n" ~= req_data:sub(-2, -1) then
			ngx.log(ngx.WARN, "receive last is not \\r\\n")
			break
		end
	end

	if #command < 2 then
		tcpsock:send("END\r\n")
		break
	end

	local key = string.lower(command[2])
	if "set" == name then
		local expire = tonumber(command[4])
		local ok, err = memc:set(key, req_data, expire)
		if not ok then
			tcpsock:send(err .. "\r\n")
		else
			tcpsock:send("STORED" .. "\r\n")
		end

	elseif "add" == name then
		local expire = tonumber(command[4])
   		local ok, err = memc:add(key, req_data, repire)
   		if not ok then
			tcpsock:send(err .. "\r\n")
		else
   			tcpsock:send("STORED" .. "\r\n")
   		end

	elseif "get" == name then
		local res, flags, err = memc:get(key)
                if err then
			tcpsock:send(err .. "\r\n")
                else
			if res and "table" ~= res then
      				tcpsock:send("VALUE "..command[2].." 0 ".. #res -2 .. "\r\n" .. res  .. "END" .. "\r\n")
   			else
      				tcpsock:send("END" .. "\r\n") 
   			end
		end
	elseif "getrange" == name or "gets" == name then
		local offset = tonumber(command[3])
		local limit = tonumber(command[4])
		local res, flags, err = memc:get(key)
                if err then
			ngx.log(ngx.ERR, "getrange err: " .. err)
			tcpsock:send("END" .. "\r\n")
                else
			if res and "table" ~= res then
				local count = limit == -1 and limit or offset+limit
				local val = string.sub(res, offset+1, count)
      				tcpsock:send("VALUE "..command[2].." 0 ".. #val .. "\r\n" .. val  .. "\r\nEND" .. "\r\n")
   			else
      				tcpsock:send("END" .. "\r\n") 
   			end
		end
   	elseif "delete" == name then
   		local ok, err = memc:delete(key)
   		if not ok then
			tcpsock:send(err .. "\r\n")
   		else
   			tcpsock:send("DELETED" .. "\r\n")
		end
	end
end

-- put it into the connection pool of size 100,
-- with 10 seconds max idle timeout
local ok, err = memc:set_keepalive(10000, 1000)
if not ok then
    ngx.say("cannot set keepalive: ", err)
    return
end

