-- ytac caption helper: exposes the current subtitle as a single line in
-- user-data/ytac/caption so the terminal status line can show it without
-- mpv's own multi-line terminal subtitles (which duplicate and wrap).
local width = tonumber(mp.get_opt("ytac-width") or "") or 80
local max_chars = math.max(width - 4, 10) -- room for the "💬 " prefix

local function truncate(text)
    local out, n = {}, 0
    for ch in text:gmatch("[%z\1-\127\194-\244][\128-\191]*") do
        n = n + 1
        if n > max_chars then
            out[#out] = "…"
            break
        end
        out[#out + 1] = ch
    end
    return table.concat(out)
end

local function publish(text)
    local line = ""
    if text ~= "" then
        line = "💬 " .. truncate(text)
    end
    mp.set_property_native("user-data/ytac/caption", line)
end

-- Lines of the previous cue. YouTube's auto-generated ("rolling") captions
-- repeat the previous line on top of each new cue, e.g.
--   "A"  ->  "A\nB"  ->  "B\nC"
-- so lines already seen in the previous cue are dropped to show only new text.
local prev_lines = {}

local function split_lines(text)
    local lines = {}
    for line in text:gmatch("[^\n]+") do
        line = line:gsub("^%s+", ""):gsub("%s+$", "")
        if line ~= "" then lines[#lines + 1] = line end
    end
    return lines
end

mp.observe_property("sub-text", "string", function(_, text)
    if not text then return end
    local lines = split_lines(text)
    -- Keep the previous caption through short gaps between cues.
    if #lines == 0 then return end

    local fresh = {}
    for _, line in ipairs(lines) do
        if not prev_lines[line] then fresh[#fresh + 1] = line end
    end
    if #fresh == 0 then fresh = { lines[#lines] } end

    prev_lines = {}
    for _, line in ipairs(lines) do prev_lines[line] = true end

    publish(table.concat(fresh, " "))
end)

local function reset()
    prev_lines = {}
    publish("")
end

-- Clear when the track changes, captions are toggled off, or a new song starts.
mp.observe_property("sid", "native", reset)
mp.register_event("start-file", reset)

publish("")
