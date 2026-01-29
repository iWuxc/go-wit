package redis

// enqueue.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// KEYS[2] -> go-kit:{<queue>}:pending
// --
// ARGV[1] -> task message data
// ARGV[2] -> task ID
// ARGV[3] -> current unix time in nsec
//
// Output:
// Returns 1 if successfully enqueued
// Returns 0 if task ID already exists
const enqueue = `
if redis.call("EXISTS", KEYS[1]) == 1 then
	return 0
end
redis.call("HSET", KEYS[1],
           "msg", ARGV[1],
           "state", "pending",
           "pending_since", ARGV[3])
redis.call("LPUSH", KEYS[2], ARGV[2])
return 1
`

// enqueueUnique .
//
// KEYS[1] -> unique key
// KEYS[2] -> go-kit:{<queue>}:t:<taskid>
// KEYS[3] -> go-kit:{<queue>}:pending
// --
// ARGV[1] -> task ID
// ARGV[2] -> uniqueness lock TTL
// ARGV[3] -> task message data
// ARGV[4] -> current unix time in nsec
//
// Output:
// Returns 1 if successfully enqueued
// Returns 0 if task ID conflicts with another task
// Returns -1 if task unique key already exists
const enqueueUnique = `
local ok = redis.call("SET", KEYS[1], ARGV[1], "NX", "EX", ARGV[2])
if not ok then
  return -1 
end
if redis.call("EXISTS", KEYS[2]) == 1 then
  return 0
end
redis.call("HSET", KEYS[2],
           "msg", ARGV[3],
           "state", "pending",
           "pending_since", ARGV[4],
           "unique_key", KEYS[1])
redis.call("LPUSH", KEYS[3], ARGV[1])
return 1
`

// Input:
// KEYS[1] -> go-kit:{<queue>}:pending
// KEYS[2] -> go-kit:{<queue>}:paused
// KEYS[3] -> go-kit:{<queue>}:active
// KEYS[4] -> go-kit:{<queue>}:lease
// --
// ARGV[1] -> initial lease expiration Unix time
// ARGV[2] -> task key prefix
//
// Output:
// Returns nil if no processable task is found in the given queue.
// Returns an encoded TaskMessage.
//
// Note: dequeue checks whether a queue is paused first, before
// calling RPOPLPUSH to pop a task from the queue.
const dequeue = `
if redis.call("EXISTS", KEYS[2]) == 0 then
	local id = redis.call("RPOPLPUSH", KEYS[1], KEYS[3])
	if id then
		local key = ARGV[2] .. id
		redis.call("HSET", key, "state", "active")
		redis.call("HDEL", key, "pending_since")
		redis.call("ZADD", KEYS[4], ARGV[1], id)
		return redis.call("HGET", key, "msg")
	end
end
return nil`

// done .
// KEYS[1] -> go-kit:{<queue>}:active
// KEYS[2] -> go-kit:{<queue>}:lease
// KEYS[3] -> go-kit:{<queue>}:t:<task_id>
// KEYS[4] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[5] -> go-kit:{<queue>}:processed
// -------
// ARGV[1] -> task ID
// ARGV[2] -> stats expiration timestamp
// ARGV[3] -> max int64 value
const done = `
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[2], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("DEL", KEYS[3]) == 0 then
  return redis.error_reply("NOT FOUND")
end
local n = redis.call("INCR", KEYS[4])
if tonumber(n) == 1 then
	redis.call("EXPIREAT", KEYS[4], ARGV[2])
end
local total = redis.call("GET", KEYS[5])
if tonumber(total) == tonumber(ARGV[3]) then
	redis.call("SET", KEYS[5], 1)
else
	redis.call("INCR", KEYS[5])
end
return redis.status_reply("OK")
`

// doneUnique .
// KEYS[1] -> go-kit:{<queue>}:active
// KEYS[2] -> go-kit:{<queue>}:lease
// KEYS[3] -> go-kit:{<queue>}:t:<task_id>
// KEYS[4] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[5] -> go-kit:{<queue>}:processed
// KEYS[6] -> unique key
// -------
// ARGV[1] -> task ID
// ARGV[2] -> stats expiration timestamp
// ARGV[3] -> max int64 value
const doneUnique = `
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[2], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("DEL", KEYS[3]) == 0 then
  return redis.error_reply("NOT FOUND")
end
local n = redis.call("INCR", KEYS[4])
if tonumber(n) == 1 then
	redis.call("EXPIREAT", KEYS[4], ARGV[2])
end
local total = redis.call("GET", KEYS[5])
if tonumber(total) == tonumber(ARGV[3]) then
	redis.call("SET", KEYS[5], 1)
else
	redis.call("INCR", KEYS[5])
end
if redis.call("GET", KEYS[6]) == ARGV[1] then
  redis.call("DEL", KEYS[6])
end
return redis.status_reply("OK")
`

// markAsComplete
// KEYS[1] -> go-kit:{<queue>}:active
// KEYS[2] -> go-kit:{<queue>}:lease
// KEYS[3] -> go-kit:{<queue>}:completed
// KEYS[4] -> go-kit:{<queue>}:t:<task_id>
// KEYS[5] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[6] -> go-kit:{<queue>}:processed
//
// ARGV[1] -> task ID
// ARGV[2] -> stats expiration timestamp
// ARGV[3] -> task expiration time in unix time
// ARGV[4] -> task message data
// ARGV[5] -> max int64 value
const markAsComplete = `
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[2], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZADD", KEYS[3], ARGV[3], ARGV[1]) ~= 1 then
  redis.redis.error_reply("INTERNAL")
end
redis.call("HSET", KEYS[4], "msg", ARGV[4], "state", "completed")
local n = redis.call("INCR", KEYS[5])
if tonumber(n) == 1 then
	redis.call("EXPIREAT", KEYS[5], ARGV[2])
end
local total = redis.call("GET", KEYS[6])
if tonumber(total) == tonumber(ARGV[5]) then
	redis.call("SET", KEYS[6], 1)
else
	redis.call("INCR", KEYS[6])
end
return redis.status_reply("OK")
`

// markAsCompleteUnique .
// KEYS[1] -> go-kit:{<queue>}:active
// KEYS[2] -> go-kit:{<queue>}:lease
// KEYS[3] -> go-kit:{<queue>}:completed
// KEYS[4] -> go-kit:{<queue>}:t:<task_id>
// KEYS[5] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[6] -> go-kit:{<queue>}:processed
// KEYS[7] -> go-kit:{<queue>}:unique:{<checksum>}
//
// ARGV[1] -> task ID
// ARGV[2] -> stats expiration timestamp
// ARGV[3] -> task expiration time in unix time
// ARGV[4] -> task message data
// ARGV[5] -> max int64 value
const markAsCompleteUnique = `
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[2], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZADD", KEYS[3], ARGV[3], ARGV[1]) ~= 1 then
  redis.redis.error_reply("INTERNAL")
end
redis.call("HSET", KEYS[4], "msg", ARGV[4], "state", "completed")
local n = redis.call("INCR", KEYS[5])
if tonumber(n) == 1 then
	redis.call("EXPIREAT", KEYS[5], ARGV[2])
end
local total = redis.call("GET", KEYS[6])
if tonumber(total) == tonumber(ARGV[5]) then
	redis.call("SET", KEYS[6], 1)
else
	redis.call("INCR", KEYS[6])
end
if redis.call("GET", KEYS[7]) == ARGV[1] then
  redis.call("DEL", KEYS[7])
end
return redis.status_reply("OK")
`

// requeue .
// KEYS[1] -> go-kit:{<queue>}:active
// KEYS[2] -> go-kit:{<queue>}:lease
// KEYS[3] -> go-kit:{<queue>}:pending
// KEYS[4] -> go-kit:{<queue>}:t:<task_id>
// ARGV[1] -> task ID
// Note: Use RPUSH to push to the head of the queue.
const requeue = `
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[2], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
redis.call("RPUSH", KEYS[3], ARGV[1])
redis.call("HSET", KEYS[4], "state", "pending")
return redis.status_reply("OK")
`

// schedule .
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// KEYS[2] -> go-kit:{<queue>}:scheduled
// -------
// ARGV[1] -> task message data
// ARGV[2] -> process_at time in Unix time
// ARGV[3] -> task ID
//
// Output:
// Returns 1 if successfully enqueued
// Returns 0 if task ID already exists
const schedule = `
if redis.call("EXISTS", KEYS[1]) == 1 then
	return 0
end
redis.call("HSET", KEYS[1],
           "msg", ARGV[1],
           "state", "scheduled")
redis.call("ZADD", KEYS[2], ARGV[2], ARGV[3])
return 1
`

// scheduleUnique
// KEYS[1] -> unique key
// KEYS[2] -> go-kit:{<queue>}:t:<task_id>
// KEYS[3] -> go-kit:{<queue>}:scheduled
// -------
// ARGV[1] -> task ID
// ARGV[2] -> uniqueness lock TTL
// ARGV[3] -> score (process_at timestamp)
// ARGV[4] -> task message
//
// Output:
// Returns 1 if successfully scheduled
// Returns 0 if task ID already exists
// Returns -1 if task unique key already exists
const scheduleUnique = `
local ok = redis.call("SET", KEYS[1], ARGV[1], "NX", "EX", ARGV[2])
if not ok then
  return -1
end
if redis.call("EXISTS", KEYS[2]) == 1 then
  return 0
end
redis.call("HSET", KEYS[2],
           "msg", ARGV[4],
           "state", "scheduled",
           "unique_key", KEYS[1])
redis.call("ZADD", KEYS[3], ARGV[3], ARGV[1])
return 1
`

// retry .
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// KEYS[2] -> go-kit:{<queue>}:active
// KEYS[3] -> go-kit:{<queue>}:lease
// KEYS[4] -> go-kit:{<queue>}:retry
// KEYS[5] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[6] -> go-kit:{<queue>}:failed:<yyyy-mm-dd>
// KEYS[7] -> go-kit:{<queue>}:processed
// KEYS[8] -> go-kit:{<queue>}:failed
// -------
// ARGV[1] -> task ID
// ARGV[2] -> updated base.TaskMessage value
// ARGV[3] -> retry_at UNIX timestamp
// ARGV[4] -> stats expiration timestamp
// ARGV[5] -> is_failure (bool)
// ARGV[6] -> max int64 value
const retry = `
if redis.call("LREM", KEYS[2], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[3], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
redis.call("ZADD", KEYS[4], ARGV[3], ARGV[1])
redis.call("HSET", KEYS[1], "msg", ARGV[2], "state", "retry")
if tonumber(ARGV[5]) == 1 then
	local n = redis.call("INCR", KEYS[5])
	if tonumber(n) == 1 then
		redis.call("EXPIREAT", KEYS[5], ARGV[4])
	end
	local m = redis.call("INCR", KEYS[6])
	if tonumber(m) == 1 then
		redis.call("EXPIREAT", KEYS[6], ARGV[4])
	end
    local total = redis.call("GET", KEYS[7])
    if tonumber(total) == tonumber(ARGV[6]) then
    	redis.call("SET", KEYS[7], 1)
    	redis.call("SET", KEYS[8], 1)
    else
    	redis.call("INCR", KEYS[7])
    	redis.call("INCR", KEYS[8])
    end
end
return redis.status_reply("OK")
`

// archive .
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// KEYS[2] -> go-kit:{<queue>}:active
// KEYS[3] -> go-kit:{<queue>}:lease
// KEYS[4] -> go-kit:{<queue>}:archived
// KEYS[5] -> go-kit:{<queue>}:processed:<yyyy-mm-dd>
// KEYS[6] -> go-kit:{<queue>}:failed:<yyyy-mm-dd>
// KEYS[7] -> go-kit:{<queue>}:processed
// KEYS[8] -> go-kit:{<queue>}:failed
// -------
// ARGV[1] -> task ID
// ARGV[2] -> updated base.TaskMessage value
// ARGV[3] -> died_at UNIX timestamp
// ARGV[4] -> cutoff timestamp (e.g., 90 days ago)
// ARGV[5] -> max number of tasks in archive (e.g., 100)
// ARGV[6] -> stats expiration timestamp
// ARGV[7] -> max int64 value
const archive = `
if redis.call("LREM", KEYS[2], 0, ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
if redis.call("ZREM", KEYS[3], ARGV[1]) == 0 then
  return redis.error_reply("NOT FOUND")
end
redis.call("ZADD", KEYS[4], ARGV[3], ARGV[1])
redis.call("ZREMRANGEBYSCORE", KEYS[4], "-inf", ARGV[4])
redis.call("ZREMRANGEBYRANK", KEYS[4], 0, -ARGV[5])
redis.call("HSET", KEYS[1], "msg", ARGV[2], "state", "archived")
local n = redis.call("INCR", KEYS[5])
if tonumber(n) == 1 then
	redis.call("EXPIREAT", KEYS[5], ARGV[6])
end
local m = redis.call("INCR", KEYS[6])
if tonumber(m) == 1 then
	redis.call("EXPIREAT", KEYS[6], ARGV[6])
end
local total = redis.call("GET", KEYS[7])
if tonumber(total) == tonumber(ARGV[7]) then
   	redis.call("SET", KEYS[7], 1)
   	redis.call("SET", KEYS[8], 1)
else
  	redis.call("INCR", KEYS[7])
   	redis.call("INCR", KEYS[8])
end
return redis.status_reply("OK")
`

// forward .
// KEYS[1] -> source queue (e.g. go-kit:{<queue>:scheduled or go-kit:{<queue>}:retry})
// KEYS[2] -> go-kit:{<queue>}:pending
// ARGV[1] -> current unix time in seconds
// ARGV[2] -> task key prefix
// ARGV[3] -> current unix time in nsec
// Note: Script moves tasks up to 100 at a time to keep the runtime of script short.
const forward = `
local ids = redis.call("ZRANGEBYSCORE", KEYS[1], "-inf", ARGV[1], "LIMIT", 0, 100)
for _, id in ipairs(ids) do
	redis.call("LPUSH", KEYS[2], id)
	redis.call("ZREM", KEYS[1], id)
	redis.call("HSET", ARGV[2] .. id,
               "state", "pending",
               "pending_since", ARGV[3])
end
return table.getn(ids)
`

// deleteExpiredCompletedTasks .
// KEYS[1] -> go-kit:{<queue>}:completed
// ARGV[1] -> current time in unix time
// ARGV[2] -> task key prefix
// ARGV[3] -> batch size (i.e. maximum number of tasks to delete)
//
// Returns the number of tasks deleted.
const deleteExpiredCompletedTasks = `
local ids = redis.call("ZRANGEBYSCORE", KEYS[1], "-inf", ARGV[1], "LIMIT", 0, tonumber(ARGV[3]))
for _, id in ipairs(ids) do
	redis.call("DEL", ARGV[2] .. id)
	redis.call("ZREM", KEYS[1], id)
end
return table.getn(ids)
`

// listLeaseExpired.
// KEYS[1] -> go-kit:{<queue>}:lease
// ARGV[1] -> cutoff in unix time
// ARGV[2] -> task key prefix
const listLeaseExpired = `
local res = {}
local ids = redis.call("ZRANGEBYSCORE", KEYS[1], "-inf", ARGV[1])
for _, id in ipairs(ids) do
	local key = ARGV[2] .. id
	table.insert(res, redis.call("HGET", key, "msg"))
end
return res
`

// writeServerState .
// KEYS[1]  -> go-kit:servers:{<host:pid:sid>}
// KEYS[2]  -> go-kit:workers:{<host:pid:sid>}
// ARGV[1]  -> TTL in seconds
// ARGV[2]  -> server info
// ARGV[3:] -> alternate key-value pair of (worker id, worker data)
// Note: Add key to ZSET with expiration time as score.
// ref: https://github.com/antirez/redis/issues/135#issuecomment-2361996
const writeServerState = `
redis.call("SETEX", KEYS[1], ARGV[1], ARGV[2])
redis.call("DEL", KEYS[2])
for i = 3, table.getn(ARGV)-1, 2 do
	redis.call("HSET", KEYS[2], ARGV[i], ARGV[i+1])
end
redis.call("EXPIRE", KEYS[2], ARGV[1])
return redis.status_reply("OK")
`

// clearServerState .
// KEYS[1] -> go-kit:servers:{<host:pid:sid>}
// KEYS[2] -> go-kit:workers:{<host:pid:sid>}
const clearServerState = `
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[2])
return redis.status_reply("OK")
`

// writeSchedulerEntries .
// KEYS[1]  -> go-kit:schedulers:{<schedulerID>}
// ARGV[1]  -> TTL in seconds
// ARGV[2:] -> scheduler entries
const writeSchedulerEntries = `
redis.call("DEL", KEYS[1])
for i = 2, #ARGV do
	redis.call("LPUSH", KEYS[1], ARGV[i])
end
redis.call("EXPIRE", KEYS[1], ARGV[1])
return redis.status_reply("OK")
`

// recordSchedulerEnqueueEvent .
// KEYS[1] -> go-kit:scheduler_history:<entryID>
// ARGV[1] -> enqueued_at timestamp
// ARGV[2] -> serialized SchedulerEnqueueEvent data
// ARGV[3] -> max number of events to be persisted
const recordSchedulerEnqueueEvent = `
redis.call("ZREMRANGEBYRANK", KEYS[1], 0, -ARGV[3])
redis.call("ZADD", KEYS[1], ARGV[1], ARGV[2])
return redis.status_reply("OK")
`


// CurrentStats .
// returns a current state of the queues.
//
// KEYS[1] ->  go-kit:<queue>:pending
// KEYS[2] ->  go-kit:<queue>:active
// KEYS[3] ->  go-kit:<queue>:scheduled
// KEYS[4] ->  go-kit:<queue>:retry
// KEYS[5] ->  go-kit:<queue>:archived
// KEYS[6] ->  go-kit:<queue>:completed
// KEYS[7] ->  go-kit:<queue>:processed:<yyyy-mm-dd>
// KEYS[8] ->  go-kit:<queue>:failed:<yyyy-mm-dd>
// KEYS[9] ->  go-kit:<queue>:processed
// KEYS[10] -> go-kit:<queue>:failed
// KEYS[11] -> go-kit:<queue>:paused
//
// ARGV[1] -> task key prefix
const CurrentStats = `
local res = {}
local pendingTaskCount = redis.call("LLEN", KEYS[1])
table.insert(res, KEYS[1])
table.insert(res, pendingTaskCount)
table.insert(res, KEYS[2])
table.insert(res, redis.call("LLEN", KEYS[2]))
table.insert(res, KEYS[3])
table.insert(res, redis.call("ZCARD", KEYS[3]))
table.insert(res, KEYS[4])
table.insert(res, redis.call("ZCARD", KEYS[4]))
table.insert(res, KEYS[5])
table.insert(res, redis.call("ZCARD", KEYS[5]))
table.insert(res, KEYS[6])
table.insert(res, redis.call("ZCARD", KEYS[6]))
for i=7,10 do
    local count = 0
	local n = redis.call("GET", KEYS[i])	
	if n then
	    count = tonumber(n)
	end
	table.insert(res, KEYS[i])
	table.insert(res, count)
end
table.insert(res, KEYS[11])
table.insert(res, redis.call("EXISTS", KEYS[11]))
table.insert(res, "oldest_pending_since")
if pendingTaskCount > 0 then
	local id = redis.call("LRANGE", KEYS[1], -1, -1)[1]
	table.insert(res, redis.call("HGET", ARGV[1] .. id, "pending_since"))
else
	table.insert(res, 0)
end
return res`

// MemoryUsed .
// Computes memory usage for the given queue by sampling tasks
// from each redis list/zset. Returns approximate memory usage value
// in bytes.
//
// KEYS[1] -> go-kit:{queue}:active
// KEYS[2] -> go-kit:{queue}:pending
// KEYS[3] -> go-kit:{queue}:scheduled
// KEYS[4] -> go-kit:{queue}:retry
// KEYS[5] -> go-kit:{queue}:archived
// KEYS[6] -> go-kit:{queue}:completed
//
// ARGV[1] -> go-kit:{queue}:t:
// ARGV[2] -> sample_size (e.g 20)
const MemoryUsed = `
local sample_size = tonumber(ARGV[2])
if sample_size <= 0 then
    return redis.error_reply("sample size must be a positive number")
end
local mem_used = 0
for i=1,2 do
    local ids = redis.call("LRANGE", KEYS[i], 0, sample_size - 1)
    local sample_total = 0
    if (table.getn(ids) > 0) then
        for _, id in ipairs(ids) do
            local bytes = redis.call("MEMORY", "USAGE", ARGV[1] .. id)
            sample_total = sample_total + bytes
        end
        local n = redis.call("LLEN", KEYS[i])
        local avg = sample_total / table.getn(ids)
        mem_used = mem_used + (avg * n)
    end
    local m = redis.call("MEMORY", "USAGE", KEYS[i])
    if (m) then
        mem_used = mem_used + m
    end
end
for i=3,6 do
    local ids = redis.call("ZRANGE", KEYS[i], 0, sample_size - 1)
    local sample_total = 0
    if (table.getn(ids) > 0) then
        for _, id in ipairs(ids) do
            local bytes = redis.call("MEMORY", "USAGE", ARGV[1] .. id)
            sample_total = sample_total + bytes
        end
        local n = redis.call("ZCARD", KEYS[i])
        local avg = sample_total / table.getn(ids)
        mem_used = mem_used + (avg * n)
    end
    local m = redis.call("MEMORY", "USAGE", KEYS[i])
    if (m) then
        mem_used = mem_used + m
    end
end
return mem_used
`

// HistoricalStats .
const HistoricalStats = `
local res = {}
for _, key in ipairs(KEYS) do
	local n = redis.call("GET", key)
	if not n then
		n = 0
	end
	table.insert(res, tonumber(n))
end
return res
`

// TaskInfo .
//
// KEYS[1] -> task key (go-kit:{<queue>}:t:<taskid>)
// ARGV[1] -> task id
// ARGV[2] -> current time in Unix time (seconds)
// ARGV[3] -> queue key prefix (go-kit:{<queue>}:)
//
// Output:
// Tuple of {msg, state, nextProcessAt, result}
// msg: encoded task message
// state: string describing the state of the task
// nextProcessAt: unix time in seconds, zero if not applicable.
// result: result data associated with the task
//
// If the task key doesn't exist, it returns error with a message "NOT FOUND"
const TaskInfo = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return redis.error_reply("NOT FOUND")
end
local msg, state, result = unpack(redis.call("HMGET", KEYS[1], "msg", "state", "result"))
if state == "scheduled" or state == "retry" then
	return {msg, state, redis.call("ZSCORE", ARGV[3] .. state, ARGV[1]), result}
end
if state == "pending" then
	return {msg, state, ARGV[2], result}
end
return {msg, state, 0, result}
`

// ListMessage .
// returns a list of TaskInfo in Redis list with the given key.
//
// KEYS[1] -> key for id list (e.g. go-kit:{<queue>}:pending)
// ARGV[1] -> start offset
// ARGV[2] -> stop offset
// ARGV[3] -> task key prefix
const ListMessage = `
local ids = redis.call("LRange", KEYS[1], ARGV[1], ARGV[2])
local data = {}
for _, id in ipairs(ids) do
	local key = ARGV[3] .. id
	local msg, result = unpack(redis.call("HMGET", key, "msg","result"))
	table.insert(data, msg)
	table.insert(data, result)
end
return data
`

// ListZSetEntries .
// returns a list of message and score pairs in Redis sorted-set with the given key.
//
// KEYS[1] -> key for ids set (e.g. go-kit:{<queue>}:scheduled)
// ARGV[1] -> min
// ARGV[2] -> max
// ARGV[3] -> task key prefix
//
// Returns an array populated with
// [msg1, score1, result1, msg2, score2, result2, ..., msgN, scoreN, resultN]
const ListZSetEntries = `
local data = {}
local id_score_pairs = redis.call("ZRANGE", KEYS[1], ARGV[1], ARGV[2], "WITHSCORES")
for i = 1, table.getn(id_score_pairs), 2 do
	local id = id_score_pairs[i]
	local score = id_score_pairs[i+1]
	local key = ARGV[3] .. id
	local msg, res = unpack(redis.call("HMGET", key, "msg", "result"))
	table.insert(data, msg)
	table.insert(data, score)
	table.insert(data, res)
end
return data
`

// RunTask .
// finds a task that matches the id from the given queue and updates it to pending state.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// KEYS[2] -> go-kit:{<queue>}:pending
// --
// ARGV[1] -> task ID
// ARGV[2] -> queue key prefix; go-kit:{<queue>}:
//
// Output:
// Numeric code indicating the status:
// Returns 1 if task is successfully updated.
// Returns 0 if task is not found.
// Returns -1 if task is in active state.
// Returns -2 if task is in pending state.
// Returns error reply if unexpected error occurs.
const RunTask = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
local state = redis.call("HGET", KEYS[1], "state")
if state == "active" then
	return -1
elseif state == "pending" then
	return -2
end
local n = redis.call("ZREM", ARGV[2] .. state, ARGV[1])
if n == 0 then
	return redis.error_reply("internal error: task id not found in zset " .. tostring(state))
end
redis.call("LPUSH", KEYS[2], ARGV[1])
redis.call("HSET", KEYS[1], "state", "pending")
return 1
`

// RunALL
// (one of: scheduled, retry, archived) to pending state.
//
// Input:
// KEYS[1] -> zset which holds task ids (e.g. go-kit:{<queue>}:scheduled)
// KEYS[2] -> go-kit:{<queue>}:pending
// --
// ARGV[1] -> task key prefix
//
// Output:
// integer: number of tasks updated to pending state.
const RunALL = `
local ids = redis.call("ZRANGE", KEYS[1], 0, -1)
for _, id in ipairs(ids) do
	redis.call("LPUSH", KEYS[2], id)
	redis.call("HSET", ARGV[1] .. id, "state", "pending")
end
redis.call("DEL", KEYS[1])
return table.getn(ids)
`

// ArchiveALLPending .
// moves all pending tasks from the given queue to archived state.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:pending
// KEYS[2] -> go-kit:{<queue>}:archived
// --
// ARGV[1] -> current timestamp
// ARGV[2] -> cutoff timestamp (e.g., 90 days ago)
// ARGV[3] -> max number of tasks in archive (e.g., 100)
// ARGV[4] -> task key prefix (go-kit:{<queue>}:t:)
//
// Output:
// integer: Number of tasks archived
const ArchiveALLPending = `
local ids = redis.call("LRANGE", KEYS[1], 0, -1)
for _, id in ipairs(ids) do
	redis.call("ZADD", KEYS[2], ARGV[1], id)
	redis.call("HSET", ARGV[4] .. id, "state", "archived")
end
redis.call("ZREMRANGEBYSCORE", KEYS[2], "-inf", ARGV[2])
redis.call("ZREMRANGEBYRANK", KEYS[2], 0, -ARGV[3])
redis.call("DEL", KEYS[1])
return table.getn(ids)
`

// ArchiveALL
// archives all tasks in either scheduled or retry state from the given queue.
//
// Input:
// KEYS[1] -> ZSET to move task from (e.g., go-kit:{<queue>}:retry)
// KEYS[2] -> go-kit:{<queue>}:archived
// --
// ARGV[1] -> current timestamp
// ARGV[2] -> cutoff timestamp (e.g., 90 days ago)
// ARGV[3] -> max number of tasks in archive (e.g., 100)
// ARGV[4] -> task key prefix (go-kit:{<queue>}:t:)
//
// Output:
// integer: number of tasks archived
const ArchiveALL = `
local ids = redis.call("ZRANGE", KEYS[1], 0, -1)
for _, id in ipairs(ids) do
	redis.call("ZADD", KEYS[2], ARGV[1], id)
	redis.call("HSET", ARGV[4] .. id, "state", "archived")
end
redis.call("ZREMRANGEBYSCORE", KEYS[2], "-inf", ARGV[2])
redis.call("ZREMRANGEBYRANK", KEYS[2], 0, -ARGV[3])
redis.call("DEL", KEYS[1])
return table.getn(ids)
`

// ArchiveTask
// archives a task given a task id.
//
// Input:
// KEYS[1] -> task key (go-kit:{<queue>}:t:<task_id>)
// KEYS[2] -> archived key (go-kit:{<queue>}:archived)
// --
// ARGV[1] -> id of the task to archive
// ARGV[2] -> current timestamp
// ARGV[3] -> cutoff timestamp (e.g., 90 days ago)
// ARGV[4] -> max number of tasks in archived state (e.g., 100)
// ARGV[5] -> queue key prefix (go-kit:{<queue>}:)
//
// Output:
// Numeric code indicating the status:
// Returns 1 if task is successfully archived.
// Returns 0 if task is not found.
// Returns -1 if task is already archived.
// Returns -2 if task is in active state.
// Returns error reply if unexpected error occurs.
const ArchiveTask = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
local state = redis.call("HGET", KEYS[1], "state")
if state == "active" then
	return -2
end
if state == "archived" then
	return -1
end
if state == "pending" then
	if redis.call("LREM", ARGV[5] .. state, 1, ARGV[1]) == 0 then
		return redis.error_reply("task id not found in list " .. tostring(state))
	end
else 
	if redis.call("ZREM", ARGV[5] .. state, ARGV[1]) == 0 then
		return redis.error_reply("task id not found in zset " .. tostring(state))
	end
end
redis.call("ZADD", KEYS[2], ARGV[2], ARGV[1])
redis.call("HSET", KEYS[1], "state", "archived")
redis.call("ZREMRANGEBYSCORE", KEYS[2], "-inf", ARGV[3])
redis.call("ZREMRANGEBYRANK", KEYS[2], 0, -ARGV[4])
return 1
`

// DeleteTask .
// finds a task that matches the id from the given queue and deletes it.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:t:<task_id>
// --
// ARGV[1] -> task ID
// ARGV[2] -> queue key prefix
//
// Output:
// Numeric code indicating the status:
// Returns 1 if task is successfully deleted.
// Returns 0 if task is not found.
// Returns -1 if task is in active state.
const DeleteTask = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
local state = redis.call("HGET", KEYS[1], "state")
if state == "active" then
	return -1
end
if state == "pending" then
	if redis.call("LREM", ARGV[2] .. state, 0, ARGV[1]) == 0 then
		return redis.error_reply("task is not found in list: " .. tostring(state))
	end
else
	if redis.call("ZREM", ARGV[2] .. state, ARGV[1]) == 0 then
		return redis.error_reply("task is not found in zset: " .. tostring(state))
	end
end
local unique_key = redis.call("HGET", KEYS[1], "unique_key")
if unique_key and unique_key ~= "" and redis.call("GET", unique_key) == ARGV[1] then
	redis.call("DEL", unique_key)
end
return redis.call("DEL", KEYS[1])
`

// DeleteALL
// deletes tasks from the given zset.
//
// Input:
// KEYS[1] -> zset holding the task ids.
// --
// ARGV[1] -> task key prefix
//
// Output:
// integer: number of tasks deleted
const DeleteALL = `
local ids = redis.call("ZRANGE", KEYS[1], 0, -1)
for _, id in ipairs(ids) do
	local task_key = ARGV[1] .. id
	local unique_key = redis.call("HGET", task_key, "unique_key")
	if unique_key and unique_key ~= "" and redis.call("GET", unique_key) == id then
		redis.call("DEL", unique_key)
	end
	redis.call("DEL", task_key)
end
redis.call("DEL", KEYS[1])
return table.getn(ids)
`

// DeleteALLPending
// deletes all pending tasks from the given queue.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:pending
// --
// ARGV[1] -> task key prefix
//
// Output:
// integer: number of tasks deleted
const DeleteALLPending = `
local ids = redis.call("LRANGE", KEYS[1], 0, -1)
for _, id in ipairs(ids) do
	redis.call("DEL", ARGV[1] .. id)
end
redis.call("DEL", KEYS[1])
return table.getn(ids)
`

// RemoveQueueForce
// removes the given queue regardless of whether the queue is empty.
// It is only check whether active queue is empty before removing.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}
// KEYS[2] -> go-kit:{<queue>}:active
// KEYS[3] -> go-kit:{<queue>}:scheduled
// KEYS[4] -> go-kit:{<queue>}:retry
// KEYS[5] -> go-kit:{<queue>}:archived
// KEYS[6] -> go-kit:{<queue>}:lease
// --
// ARGV[1] -> task key prefix
//
// Output:
// Numeric code to indicate the status.
// Returns 1 if successfully removed.
// Returns -2 if the queue has active tasks.
const RemoveQueueForce = `
local active = redis.call("LLEN", KEYS[2])
if active > 0 then
    return -2
end
for _, id in ipairs(redis.call("LRANGE", KEYS[1], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("LRANGE", KEYS[2], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[3], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[4], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[5], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("LRANGE", KEYS[1], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("LRANGE", KEYS[2], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[3], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[4], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[5], 0, -1)) do
	redis.call("DEL", ARGV[1] .. id)
end
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[2])
redis.call("DEL", KEYS[3])
redis.call("DEL", KEYS[4])
redis.call("DEL", KEYS[5])
redis.call("DEL", KEYS[6])
return 1
`

// RemoveQueue
// removes the given queue. It checks whether queue is empty before removing.
//
// Input:
// KEYS[1] -> go-kit:{<queue>}:pending
// KEYS[2] -> go-kit:{<queue>}:active
// KEYS[3] -> go-kit:{<queue>}:scheduled
// KEYS[4] -> go-kit:{<queue>}:retry
// KEYS[5] -> go-kit:{<queue>}:archived
// KEYS[6] -> go-kit:{<queue>}:lease
// --
// ARGV[1] -> task key prefix
//
// Output:
// Numeric code to indicate the status
// Returns 1 if successfully removed.
// Returns -1 if queue is not empty
const RemoveQueue = `
local ids = {}
for _, id in ipairs(redis.call("LRANGE", KEYS[1], 0, -1)) do
	table.insert(ids, id)
end
for _, id in ipairs(redis.call("LRANGE", KEYS[2], 0, -1)) do
	table.insert(ids, id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[3], 0, -1)) do
	table.insert(ids, id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[4], 0, -1)) do
	table.insert(ids, id)
end
for _, id in ipairs(redis.call("ZRANGE", KEYS[5], 0, -1)) do
	table.insert(ids, id)
end
if table.getn(ids) > 0 then
	return -1
end
for _, id in ipairs(ids) do
	redis.call("DEL", ARGV[1] .. id)
end
for _, id in ipairs(ids) do
	redis.call("DEL", ARGV[1] .. id)
end
redis.call("DEL", KEYS[1])
redis.call("DEL", KEYS[2])
redis.call("DEL", KEYS[3])
redis.call("DEL", KEYS[4])
redis.call("DEL", KEYS[5])
redis.call("DEL", KEYS[6])
return 1`

// ListServerKeys .
// returns the list of server info.
// Note: Script also removes stale keys.
const ListServerKeys = `
local now = tonumber(ARGV[1])
local keys = redis.call("ZRANGEBYSCORE", KEYS[1], now, "+inf")
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now-1)
return keys`

// ListWorkers .
// returns the list of worker stats.
// Note: Script also removes stale keys.
const ListWorkers = `
local now = tonumber(ARGV[1])
local keys = redis.call("ZRANGEBYSCORE", KEYS[1], now, "+inf")
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now-1)
return keys`

// listSchedulerKeys .
// returns the list of scheduler entries.
// Note: Script also removes stale keys.
const listSchedulerKeys = `
local now = tonumber(ARGV[1])
local keys = redis.call("ZRANGEBYSCORE", KEYS[1], now, "+inf")
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now-1)
return keys`


